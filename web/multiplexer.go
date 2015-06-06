// Tideland Go Library - Web - Multiplexer
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
	"strings"
	"sync"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/logger"
	"github.com/tideland/golib/monitoring"
)

//--------------------
// HANDLER LIST
//--------------------

// handlerListEntry is one entry in a list of resource handlers.
type handlerListEntry struct {
	resourceHandler ResourceHandler
	next            *handlerListEntry
}

// handle lets the resource handler process the request.
func (hle *handlerListEntry) handle(ctx Context) (bool, error) {
	switch ctx.Request().Method {
	case "GET":
		rh, ok := hle.resourceHandler.(GetResourceHandler)
		if !ok {
			return false, errors.New(ErrNoGetHandler, errorMessages, hle.id(ctx))
		}
		return rh.Get(ctx)
	case "HEAD":
		rh, ok := hle.resourceHandler.(HeadResourceHandler)
		if !ok {
			return false, errors.New(ErrNoGetHandler, errorMessages, hle.id(ctx))
		}
		return rh.Head(ctx)
	case "PUT":
		rh, ok := hle.resourceHandler.(PutResourceHandler)
		if !ok {
			return false, errors.New(ErrNoPutHandler, errorMessages, hle.id(ctx))
		}
		return rh.Put(ctx)
	case "POST":
		rh, ok := hle.resourceHandler.(PostResourceHandler)
		if !ok {
			return false, errors.New(ErrNoPostHandler, errorMessages, hle.id(ctx))
		}
		return rh.Post(ctx)
	case "DELETE":
		rh, ok := hle.resourceHandler.(DeleteResourceHandler)
		if !ok {
			return false, errors.New(ErrNoDeleteHandler, errorMessages, hle.id(ctx))
		}
		return rh.Delete(ctx)
	case "OPTIONS":
		rh, ok := hle.resourceHandler.(OptionsResourceHandler)
		if !ok {
			return false, errors.New(ErrNoDeleteHandler, errorMessages, hle.id(ctx))
		}
		return rh.Options(ctx)
	}
	return false, errors.New(ErrMethodNotSupported, errorMessages, ctx.Request().Method)
}

// id returns the ID of this handler in a given context.
func (hle *handlerListEntry) id(ctx Context) string {
	return fmt.Sprintf("%s@%s/%s", hle.resourceHandler.ID(), ctx.Domain(), ctx.Resource())
}

// handlerList maintains a list of handlers responsible
// for one domain and resource.
type handlerList struct {
	head *handlerListEntry
}

// register adds a new resource handler.
func (hl *handlerList) register(handler ResourceHandler) error {
	if hl.head == nil {
		hl.head = &handlerListEntry{handler, nil}
		return nil
	}
	current := hl.head
	for {
		if current.resourceHandler == handler {
			return errors.New(ErrDuplicateHandler, errorMessages, handler.ID())
		}
		if current.next == nil {
			break
		}
		current = current.next
	}
	current.next = &handlerListEntry{handler, nil}
	return nil
}

// deregister removes a resource handler.
func (hl *handlerList) deregister(id string) {
	var head, tail *handlerListEntry
	current := hl.head
	for current != nil {
		if current.resourceHandler.ID() != id {
			if head == nil {
				head = current
				tail = current
			} else {
				tail.next = current
				tail = tail.next
			}
		}
		current = current.next
	}
	hl.head = head
}

// handle lets all resource handlers process the request.
func (hl *handlerList) handle(ctx *context) error {
	current := hl.head
	for current != nil {
		goOn, err := current.handle(ctx)
		if err != nil {
			return err
		}
		if !goOn {
			return nil
		}
		current = current.next
	}
	return nil
}

//--------------------
// MAPPING
//--------------------

// mapping maps domains and resources to lists of
// resource handlers.
type mapping struct {
	handlers map[string]*handlerList
}

// newMapping returns a new handler mapping.
func newMapping() *mapping {
	return &mapping{
		handlers: make(map[string]*handlerList),
	}
}

// register adds a resource handler.
func (m *mapping) register(domain, resource string, handler ResourceHandler) error {
	location := m.location(domain, resource)
	hl, ok := m.handlers[location]
	if !ok {
		hl = &handlerList{}
		m.handlers[location] = hl
	}
	return hl.register(handler)
}

// deregister removes a resource handler.
func (m *mapping) deregister(domain, resource string, id string) {
	location := m.location(domain, resource)
	hl, ok := m.handlers[location]
	if !ok {
		return
	}
	hl.deregister(id)
	if hl.head == nil {
		delete(m.handlers, location)
	}
}

// handle handles a request.
func (m *mapping) handle(ctx *context) error {
	// Find handler.
	location := m.location(ctx.Domain(), ctx.Resource())
	hl, ok := m.handlers[location]
	if !ok {
		defaultLocation := m.location(ctx.DefaultDomain(), ctx.DefaultResource())
		hl, ok = m.handlers[defaultLocation]
		if !ok {
			return errors.New(ErrNoHandler, errorMessages, location, defaultLocation)
		}
		location = defaultLocation
	}
	// Dispatch by method.
	logger.Infof("handling %s", ctx)
	return hl.handle(ctx)
}

// location builds the map key for domain and resource.
func (m *mapping) location(domain, resource string) string {
	return strings.ToLower(domain + "/" + resource)
}

//--------------------
// REGISTRATIONS
//--------------------

// Registration encapsulates one handler registration.
type Registration struct {
	Domain   string
	Resource string
	Handler  ResourceHandler
}

// Registrations is a number handler registratons.
type Registrations []Registration

//--------------------
// MULTIPLEXER
//--------------------

// multiplexer implements the Multiplexer interface.
type multiplexer struct {
	mutex           sync.RWMutex
	mapping         *mapping
	sceneManager    SceneManager
	basePath        string
	defaultDomain   string
	defaultResource string
	templateCache   *templateCache
}

// newMultiplexer returns a multiplexer without set options for usage
// inside the package server.
func newMultiplexer() Multiplexer {
	return &multiplexer{
		mapping:         newMapping(),
		basePath:        defaultBasePath,
		defaultDomain:   defaultDefaultDomain,
		defaultResource: defaultDefaultResource,
		templateCache:   newTemplateCache(),
	}
}

// NewMultiplexer creates a new HTTP multiplexer.
func NewMultiplexer(options ...Option) (Multiplexer, error) {
	mux := newMultiplexer()
	for _, option := range options {
		if err := option(mux); err != nil {
			return nil, err
		}
	}
	return mux, nil
}

// ParseTemplate is specified on the Multiplexer interface.
func (mux *multiplexer) ParseTemplate(templateID, template, contentType string) {
	mux.templateCache.parse(templateID, template, contentType)
}

// Register is specified on the Multiplexer interface.
func (mux *multiplexer) Register(domain, resource string, handler ResourceHandler) error {
	mux.mutex.Lock()
	defer mux.mutex.Unlock()
	err := handler.Init(mux, domain, resource)
	if err != nil {
		return err
	}
	return mux.mapping.register(domain, resource, handler)
}

// RegisterAll is specified on the Multiplexer interface.
func (mux *multiplexer) RegisterAll(registrations Registrations) error {
	for _, registration := range registrations {
		err := mux.Register(registration.Domain, registration.Resource, registration.Handler)
		if err != nil {
			return err
		}
	}
	return nil
}

// Deregister is specified on the Multiplexer interface.
func (mux *multiplexer) Deregister(domain, resource, id string) {
	mux.mutex.Lock()
	defer mux.mutex.Unlock()
	mux.mapping.deregister(domain, resource, id)
}

// ServeHTTP is specified on the http.Handler interface.
func (mux *multiplexer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.mutex.RLock()
	defer mux.mutex.RUnlock()
	ctx, err := newContext(mux, r, w)
	if err != nil {
		mux.internalServerError("error preparing request", ctx, err)
		return
	}
	if mux.sceneManager != nil {
		scene, err := mux.sceneManager.Scene(ctx)
		if err != nil {
			mux.internalServerError("error retrieving scene for request", ctx, err)
			return
		}
		ctx.scene = scene
	}
	measuring := monitoring.BeginMeasuring(ctx.String())
	defer measuring.EndMeasuring()
	err = mux.mapping.handle(ctx)
	if err != nil {
		mux.internalServerError("error handling request", ctx, err)
	}
}

// internalServerError logs an internal error and returns it to the user.
func (mux *multiplexer) internalServerError(format string, ctx Context, err error) {
	msg := fmt.Sprintf(format+" %q: %v", ctx, err)
	logger.Errorf(msg)
	http.Error(ctx.ResponseWriter(), msg, http.StatusInternalServerError)
}

// EOF
