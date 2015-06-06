// Tideland Go Library - Web - Unit Tests
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package web_test

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/logger"
	"github.com/tideland/golib/web"
)

//--------------------
// INIT
//--------------------

func init() {
	logger.SetLevel(logger.LevelDebug)
}

//--------------------
// TESTS
//--------------------

// TestGetJSON tests the GET command with a JSON result.
func TestGetJSON(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Setup the test server.
	mux, ts, err := web.StartTestServer()
	assert.Nil(err)
	defer ts.Close()
	err = mux.Register("test", "json", NewTestHandler("json", assert))
	assert.Nil(err)
	// Perform test requests.
	resp, err := web.DoTestRequest(ts, &web.TestRequest{
		Method: "GET",
		Path:   "/test/json/4711",
		Header: web.TestSettings{"Accept": "application/json"},
	})
	assert.Nil(err, "Local JSON GET.")
	var data TestRequestData
	err = json.Unmarshal(resp.Body, &data)
	assert.Nil(err)
	assert.Equal(data.ResourceID, "4711")
}

// TestPutJSON tests the PUT command with a JSON payload and result.
func TestPutJSON(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Setup the test server.
	mux, ts, err := web.StartTestServer()
	assert.Nil(err)
	defer ts.Close()
	err = mux.Register("test", "json", NewTestHandler("json", assert))
	assert.Nil(err)
	// Perform test requests.
	reqData := TestRequestData{"foo", "bar", "4711"}
	reqBuf, _ := json.Marshal(reqData)
	resp, err := web.DoTestRequest(ts, &web.TestRequest{
		Method: "PUT",
		Path:   "/test/json/4711",
		Header: web.TestSettings{"Content-Type": "application/json", "Accept": "application/json"},
		Body:   reqBuf,
	})
	assert.Nil(err)
	var recvData TestRequestData
	err = json.Unmarshal(resp.Body, &recvData)
	assert.Nil(err)
	assert.Equal(recvData, reqData)
}

// TestGetXML tests the GET command with an XML result.
func TestGetXML(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Setup the test server.
	mux, ts, err := web.StartTestServer()
	assert.Nil(err)
	defer ts.Close()
	err = mux.Register("test", "xml", NewTestHandler("xml", assert))
	assert.Nil(err)
	// Perform test requests.
	resp, err := web.DoTestRequest(ts, &web.TestRequest{
		Method: "GET",
		Path:   "/test/xml/4711",
		Header: web.TestSettings{"Accept": "application/xml"},
	})
	assert.Nil(err)
	assert.Substring("<ResourceID>4711</ResourceID>", string(resp.Body))
}

// TestPutXML tests the PUT command with a XML payload and result.
func TestPutXML(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Setup the test server.
	mux, ts, err := web.StartTestServer()
	assert.Nil(err)
	defer ts.Close()
	err = mux.Register("test", "xml", NewTestHandler("xml", assert))
	assert.Nil(err)
	// Perform test requests.
	reqData := TestRequestData{"foo", "bar", "4711"}
	reqBuf, _ := xml.Marshal(reqData)
	resp, err := web.DoTestRequest(ts, &web.TestRequest{
		Method: "PUT",
		Path:   "/test/xml/4711",
		Header: web.TestSettings{"Content-Type": "application/xml", "Accept": "application/xml"},
		Body:   reqBuf,
	})
	assert.Nil(err)
	var recvData TestRequestData
	err = xml.Unmarshal(resp.Body, &recvData)
	assert.Nil(err)
	assert.Equal(recvData, reqData)
}

// TestPutGOB tests the PUT command with a GOB payload and result.
func TestPutGOB(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Setup the test server.
	mux, ts, err := web.StartTestServer()
	assert.Nil(err)
	defer ts.Close()
	err = mux.Register("test", "gob", NewTestHandler("putgob", assert))
	assert.Nil(err)
	// Perform test requests.
	reqData := TestCounterData{"test", 4711}
	reqBuf := new(bytes.Buffer)
	err = gob.NewEncoder(reqBuf).Encode(reqData)
	assert.Nil(err, "GOB encode.")
	t.Logf("%q", reqBuf.String())
	resp, err := web.DoTestRequest(ts, &web.TestRequest{
		Method: "POST",
		Path:   "/test/gob",
		Header: web.TestSettings{"Content-Type": "application/vnd.tideland.gob"},
		Body:   reqBuf.Bytes(),
	})
	var respData TestCounterData
	err = gob.NewDecoder(bytes.NewBuffer(resp.Body)).Decode(&respData)
	assert.Nil(err, "GOB decode.")
	assert.Equal(respData.ID, "test", "GOB decoded 'id'.")
	assert.Equal(respData.Count, int64(4711), "GOB decoded 'count'.")
}

// TestLongPath tests the setting of long path tail as resource ID.
func TestLongPath(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Setup the test server.
	mux, ts, err := web.StartTestServer()
	assert.Nil(err)
	defer ts.Close()
	err = mux.Register("content", "blog", NewTestHandler("default", assert))
	assert.Nil(err)
	// Perform test requests.
	resp, err := web.DoTestRequest(ts, &web.TestRequest{
		Method: "GET",
		Path:   "/content/blog/2014/09/30/just-a-test",
	})
	assert.Nil(err)
	assert.Substring("<li>Resource ID: 2014/09/30/just-a-test</li>", string(resp.Body))
}

// TestFallbackDefault tests the fallback to default.
func TestFallbackDefault(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Setup the test server.
	mux, ts, err := web.StartTestServer()
	assert.Nil(err)
	defer ts.Close()
	err = mux.Register("default", "default", NewTestHandler("default", assert))
	assert.Nil(err)
	// Perform test requests.
	resp, err := web.DoTestRequest(ts, &web.TestRequest{
		Method: "GET",
		Path:   "/x/y",
	})
	assert.Nil(err)
	assert.Substring("<li>Resource: y</li>", string(resp.Body))
}

// TestHandlerStack tests a complete handler stack.
func TestHandlerStack(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Setup the test server.
	sm := web.NewCookieSceneManager(5 * time.Minute)
	mux, ts, err := web.StartTestServer(web.Scenes(sm))
	assert.Nil(err)
	defer ts.Close()
	err = mux.RegisterAll(web.Registrations{
		{"authentication", "login", NewTestHandler("login", assert)},
		{"test", "stack", NewAuthHandler("foo", assert)},
		{"test", "stack", NewTestHandler("stack", assert)},
	})
	assert.Nil(err)
	// Perform test requests.
	resp, err := web.DoTestRequest(ts, &web.TestRequest{
		Method: "GET",
		Path:   "/test/stack",
	})
	sceneID := resp.Cookies["sceneID"]
	assert.Substring("<li>Resource: login</li>", string(resp.Body))
	resp, err = web.DoTestRequest(ts, &web.TestRequest{
		Method:  "GET",
		Path:    "/test/stack",
		Cookies: web.TestSettings{"sceneID": sceneID},
		Header:  web.TestSettings{"password": "foo"},
	})
	assert.Nil(err)
	assert.Substring("<li>Resource: stack</li>", string(resp.Body))
	resp, err = web.DoTestRequest(ts, &web.TestRequest{
		Method:  "GET",
		Path:    "/test/stack",
		Cookies: web.TestSettings{"sceneID": sceneID},
		Header:  web.TestSettings{"password": "foo"},
	})
	assert.Nil(err)
	assert.Substring("<li>Resource: stack</li>", string(resp.Body))
}

// TestMethodNotSupported tests the handling of a not support HTTP method.
func TestMethodNotSupported(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Setup the test server.
	mux, ts, err := web.StartTestServer()
	assert.Nil(err)
	defer ts.Close()
	err = mux.Register("test", "method", NewTestHandler("method", assert))
	assert.Nil(err)
	// Perform test requests.
	resp, err := web.DoTestRequest(ts, &web.TestRequest{
		Method: "OPTION",
		Path:   "/test/method",
	})
	assert.Nil(err)
	assert.Substring("OPTION", string(resp.Body))
}

//--------------------
// AUTHENTICATION HANDLER
//--------------------

type AuthHandler struct {
	password string
	assert   audit.Assertion
}

func NewAuthHandler(password string, assert audit.Assertion) *AuthHandler {
	return &AuthHandler{password, assert}
}

func (ah *AuthHandler) ID() string {
	return ah.password
}

func (ah *AuthHandler) Init(mux web.Multiplexer, domain, resource string) error {
	return nil
}

func (ah *AuthHandler) Get(ctx web.Context) (bool, error) {
	logger.Infof("scene ID: %s", ctx.Scene().ID())
	password, err := ctx.Scene().Fetch("password")
	if err == nil {
		logger.Infof("scene is logged in")
		return true, nil
	}
	password = ctx.Request().Header.Get("password")
	if password != ah.password {
		ctx.Redirect("authentication", "login", "")
		return false, nil
	}
	logger.Infof("logging scene in")
	ctx.Scene().Store("password", password)
	return true, nil
}

//--------------------
// TEST HANDLER
//--------------------

type TestRequestData struct {
	Domain     string
	Resource   string
	ResourceID string
}

type TestCounterData struct {
	ID    string
	Count int64
}

type TestErrorData struct {
	Error string
}

const testTemplateHTML = `
<?DOCTYPE html?>
<html>
<head><title>Test</title></head>
<body>
<ul>
<li>Domain: {{.Domain}}</li>
<li>Resource: {{.Resource}}</li>
<li>Resource ID: {{.ResourceID}}</li>
</ul>
</body>
</html>
`

type TestHandler struct {
	id     string
	assert audit.Assertion
}

func NewTestHandler(id string, assert audit.Assertion) *TestHandler {
	return &TestHandler{id, assert}
}

func (th *TestHandler) ID() string {
	return th.id
}

func (th *TestHandler) Init(mux web.Multiplexer, domain, resource string) error {
	mux.ParseTemplate("test:context:html", testTemplateHTML, "text/html")
	return nil
}

func (th *TestHandler) Get(ctx web.Context) (bool, error) {
	data := TestRequestData{ctx.Domain(), ctx.Resource(), ctx.ResourceID()}
	switch {
	case ctx.AcceptsContentType(web.ContentTypeXML):
		logger.Infof("get XML")
		ctx.WriteXML(data)
	case ctx.AcceptsContentType(web.ContentTypeJSON):
		logger.Infof("get JSON")
		ctx.WriteJSON(data, true)
	default:
		logger.Infof("get HTML")
		ctx.RenderTemplate("test:context:html", data)
	}
	return true, nil
}

func (th *TestHandler) Head(ctx web.Context) (bool, error) {
	return false, nil
}

func (th *TestHandler) Put(ctx web.Context) (bool, error) {
	var data TestRequestData
	switch {
	case ctx.HasContentType(web.ContentTypeJSON):
		err := ctx.ReadJSON(&data)
		if err != nil {
			ctx.WriteJSON(TestErrorData{err.Error()}, true)
		} else {
			ctx.WriteJSON(data, true)
		}
	case ctx.HasContentType(web.ContentTypeXML):
		err := ctx.ReadXML(&data)
		if err != nil {
			ctx.WriteXML(TestErrorData{err.Error()})
		} else {
			ctx.WriteXML(data)
		}
	}

	return true, nil
}

func (th *TestHandler) Post(ctx web.Context) (bool, error) {
	var data TestCounterData
	err := ctx.ReadGOB(&data)
	if err != nil {
		ctx.WriteGOB(err)
	} else {
		ctx.WriteGOB(data)
	}
	return true, nil
}

func (th *TestHandler) Delete(ctx web.Context) (bool, error) {
	return false, nil
}

// EOF
