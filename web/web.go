// Tideland Go Library - Web
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

	"github.com/tideland/golib/scene"
)

//--------------------
// SCENE MANAGER
//--------------------

// SceneManager can be configured to retrieve or create a scene based
// on formation inside the context, e.g. a scene ID as part of a request.
type SceneManager interface {
	// Scene returns the matching scene for the passed context.
	Scene(ctx Context) (scene.Scene, error)

	// Stop tells the scene manager to stop working.
	Stop() error
}

//--------------------
// CONTEXT INTERFACE
//--------------------

// Context encapsulates all the needed information for handling
// a request.
type Context interface {
	fmt.Stringer

	// BasePath returns the configured base path.
	BasePath() string

	// DefaultDomain returns the configured default domain.
	DefaultDomain() string

	// DefaultResource returns the configured default resource.
	DefaultResource() string

	// Request returns the used Go HTTP request.
	Request() *http.Request

	// ResponseWriter returns the used Go HTTP response writer.
	ResponseWriter() http.ResponseWriter

	// Domain returns the requests domain.
	Domain() string

	// Resource returns the requests resource.
	Resource() string

	// ResourceID return the requests resource ID.
	ResourceID() string

	// Scene returns the current scene.
	Scene() scene.Scene

	// AcceptsContentType checks if the requestor accepts a given content type.
	AcceptsContentType(contentType string) bool

	// HasContentType checks if the sent content has the given content type.
	HasContentType(contentType string) bool

	// Languages returns the accepted language with the quality values.
	Languages() Languages

	// InternalPath builds an internal path out of the passed parts.
	InternalPath(domain, resource, resourceID string, query ...KeyValue) string

	// Redirect to a domain, resource and resource ID (optional).
	Redirect(domain, resource, resourceID string)

	// RenderTemplate renders a template with the passed data to the response writer.
	RenderTemplate(templateID string, data interface{})

	// WriteGOB encodes the passed data to GOB and writes it to the response writer.
	WriteGOB(data interface{})

	// ReadGOB checks if the request content type is GOB, reads its body
	// and decodes it to the value pointed to by data.
	ReadGOB(data interface{}) error

	// WriteJSON marshals the passed data to JSON and writes it to the response writer.
	// The HTML flag controls the data encoding.
	WriteJSON(data interface{}, html bool)

	// PositiveJSONFeedback produces a positive feedback envelope
	// encoded in JSON.
	PositiveJSONFeedback(msg string, p interface{}, args ...interface{})

	// NegativeJSONFeedback produces a negative feedback envelope
	// encoded in JSON.
	NegativeJSONFeedback(msg string, args ...interface{})

	// ReadJSON checks if the request content type is JSON, reads its body
	// and unmarshals it to the value pointed to by data.
	ReadJSON(data interface{}) error

	// ReadGenericJSON works like ReadJSON but can be used if the transmitted
	// type is unknown or has no Go representation. It will a mapping according to
	// http://golang.org/pkg/json/#Unmarshal.
	ReadGenericJSON() (map[string]interface{}, error)

	// WriteXML marshals the passed data to XML and writes it to the response writer.
	WriteXML(data interface{})

	// ReadXML checks if the request content type is XML, reads its body
	// and unmarshals it to the value pointed to by data.
	ReadXML(data interface{}) error
}

//--------------------
// RESOURCE HANDLER INTERFACE
//--------------------

// ResourceHandler is the base interface for all resource
// handlers understanding the REST verbs. It allows the
// initialization and returns an id that should be unique
// for the combination of domain and resource. So it can
// later be removed again.
type ResourceHandler interface {
	// ID returns the deployment ID of the handler.
	ID() string

	// Init initializes the resource handler after registrations.
	Init(mux Multiplexer, domain, resource string) error
}

// GetResourceHandler is the additional interface for
// handlers understanding the verb GET.
type GetResourceHandler interface {
	Get(ctx Context) (bool, error)
}

// HeadResourceHandler is the additional interface for
// handlers understanding the verb HEAD.
type HeadResourceHandler interface {
	Head(ctx Context) (bool, error)
}

// PutResourceHandler is the additional interface for
// handlers understanding the verb PUT.
type PutResourceHandler interface {
	Put(ctx Context) (bool, error)
}

// PostResourceHandler is the additional interface for
// handlers understanding the verb POST.
type PostResourceHandler interface {
	Post(ctx Context) (bool, error)
}

// DeleteResourceHandler is the additional interface for
// handlers understanding the verb DELETE.
type DeleteResourceHandler interface {
	Delete(ctx Context) (bool, error)
}

// OptionsResourceHandler is the additional interface for
// handlers understanding the verb OPTION.
type OptionsResourceHandler interface {
	Options(ctx Context) (bool, error)
}

//--------------------
// MULTIPLEXER INTERFACE
//--------------------

// Multiplexer maps the domain and resource parts of a URL to
// their registered handlers. It implements the http.Handler
// interface.
type Multiplexer interface {
	http.Handler

	// ParseTemplate parses a raw template into the cache.
	ParseTemplate(templateID, template, contentType string)

	// Register adds a resource handler for a given domain and resource.
	Register(domain, resource string, handler ResourceHandler) error

	// RegisterAll allows to register multiple handler in one run.
	RegisterAll(registrations Registrations) error

	// Deregister removes a resource handler for a given domain and resource.
	Deregister(domain, resource, id string)
}

// EOF
