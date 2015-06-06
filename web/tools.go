// Tideland Go Library - Web - Tools
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package web

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/loop"
	"github.com/tideland/golib/scene"
)

//--------------------
// COOKIE SCENE MANAGER
//--------------------

const (
	sceneID = "sceneID"
)

// sceneResponse wraps the manager responses.
type sceneResponse struct {
	id    string
	scene scene.Scene
}

// sceneRequest wraps the manager requests.
type sceneRequest struct {
	id           string
	responseChan chan *sceneResponse
}

// cookieSceneManager implements the SceneManager interface using
// Cookies. The inactivity timeout of scenes has to be specified.
type cookieSceneManager struct {
	timeout     time.Duration
	scenes      map[string]scene.Scene
	requestChan chan *sceneRequest
	loop        loop.Loop
}

// NewCookieSceneManager creates a scene manager using the cookie
// "sceneID" to identify and manage the scene of a client session.
func NewCookieSceneManager(timeout time.Duration) SceneManager {
	m := &cookieSceneManager{
		timeout:     timeout,
		scenes:      make(map[string]scene.Scene),
		requestChan: make(chan *sceneRequest),
	}
	m.loop = loop.Go(m.backendLoop)
	return m
}

// Scene is specified on the SceneManager interface.
func (m *cookieSceneManager) Scene(ctx Context) (scene.Scene, error) {
	cookie, err := ctx.Request().Cookie(sceneID)
	if err != nil && err != http.ErrNoCookie {
		return nil, err
	}
	request := &sceneRequest{
		responseChan: make(chan *sceneResponse, 1),
	}
	if err == http.ErrNoCookie {
		request.id = ""
	} else {
		request.id = cookie.Value
	}
	select {
	case m.requestChan <- request:
	case <-m.loop.IsStopping():
		return nil, errors.New(ErrSceneManagement, errorMessages, "stopping")
	}
	m.requestChan <- request
	select {
	case response := <-request.responseChan:
		cookie = &http.Cookie{
			Name:  sceneID,
			Value: response.id,
		}
		http.SetCookie(ctx.ResponseWriter(), cookie)
		return response.scene, nil
	case <-m.loop.IsStopping():
		return nil, errors.New(ErrSceneManagement, errorMessages, "stopping")
	}
}

// Stop is specified on the SceneManager interface.
func (m *cookieSceneManager) Stop() error {
	return m.loop.Stop()
}

// backendLoop manages the scenes and cleans them periodically.
func (m *cookieSceneManager) backendLoop(l loop.Loop) error {
	ticker := time.Tick(5 * time.Minute)
	for {
		select {
		case <-m.loop.ShallStop():
			return nil
		case request := <-m.requestChan:
			m.requestScene(request)
		case <-ticker:
			m.expire()
		}
	}
}

// requestScene trieves a scene or creates a new one.
func (m *cookieSceneManager) requestScene(request *sceneRequest) {
	// Check availability.
	scn, ok := m.scenes[request.id]
	if ok {
		status, _ := scn.Status()
		if status == scene.Active {
			// Active scene found.
			response := &sceneResponse{
				id:    scn.ID().String(),
				scene: scn,
			}
			request.responseChan <- response
			return
		}
	}
	// New scene (first request or expired).
	scn = scene.StartLimited(m.timeout, 0)
	response := &sceneResponse{
		id:    scn.ID().String(),
		scene: scn,
	}
	m.scenes[response.id] = response.scene
	request.responseChan <- response
}

// expire checks for stopped scenes and removes them.
func (m *cookieSceneManager) expire() {
	ids := []string{}
	for id, scn := range m.scenes {
		status, _ := scn.Status()
		if status == scene.Over {
			ids = append(ids, id)
		}
	}
	for _, id := range ids {
		delete(m.scenes, id)
	}
}

// SceneID takes a response of a HTTP request and returns
// a scene ID it received as cookie. It can be used to establish
// session when doing client calls.
func SceneID(resp *http.Response) (string, bool) {
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == sceneID {
			return cookie.Value, true
		}
	}
	return "", false
}

//--------------------
// KEY VALUE
//--------------------

// KeyValue assigns a value to a key.
type KeyValue struct {
	Key   string
	Value interface{}
}

// String prints the encoded form key=value for URLs.
func (kv KeyValue) String() string {
	return fmt.Sprintf("%v=%v", url.QueryEscape(kv.Key), url.QueryEscape(fmt.Sprintf("%v", kv.Value)))
}

// KeyValues is a number of key/value pairs.
type KeyValues []KeyValue

// String prints the encoded form key=value joind by & for URLs.
func (kvs KeyValues) String() string {
	kvss := make([]string, len(kvs))
	for i, kv := range kvs {
		kvss[i] = kv.String()
	}
	return strings.Join(kvss, "&")
}

// EOF
