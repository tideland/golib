// Tideland Go Library - Web - Context
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
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/scene"
)

//--------------------
// CONST
//--------------------

const (
	ContentTypePlain = "text/plain"
	ContentTypeHTML  = "text/html"
	ContentTypeXML   = "application/xml"
	ContentTypeJSON  = "application/json"
	ContentTypeGOB   = "application/vnd.tideland.gob"
)

//--------------------
// ENVELOPE
//--------------------

// Envelope is a helper to give a qualified feedback in RESTful requests.
// It contains wether the request has been successful, in case of an
// error an additional message and the payload.
type Envelope struct {
	Success bool
	Message string
	Payload interface{}
}

//--------------------
// LANGUAGE
//--------------------

// Language is the valued language a request accepts as response.
type Language struct {
	Locale string
	Value  float64
}

// Languages is the ordered set of accepted languages.
type Languages []Language

// Len returns the number of languages to fulfill the sort interface.
func (ls Languages) Len() int {
	return len(ls)
}

// Less returns if the language with the index i has a smaller
// value than the one with index j to fulfill the sort interface.
func (ls Languages) Less(i, j int) bool {
	return ls[i].Value < ls[j].Value
}

// Swap swaps the languages with the indexes i and j.
func (ls Languages) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}

//--------------------
// CONTEXT
//--------------------

// context implements the Context interface.
type context struct {
	multiplexer    *multiplexer
	request        *http.Request
	responseWriter http.ResponseWriter
	scene          scene.Scene
	domain         string
	resource       string
	resourceID     string
}

// newContext creates a new context and parse the URL path.
func newContext(m *multiplexer, r *http.Request, rw http.ResponseWriter) (*context, error) {
	// Init the context.
	ctx := &context{
		multiplexer:    m,
		request:        r,
		responseWriter: rw,
	}
	// Split path for REST identifiers.
	parts := strings.Split(r.URL.Path[len(ctx.BasePath()):], "/")
	switch len(parts) {
	case 3:
		ctx.resourceID = parts[2]
		ctx.resource = parts[1]
		ctx.domain = parts[0]
	case 2:
		ctx.resource = parts[1]
		ctx.domain = parts[0]
	case 1:
		ctx.resource = ctx.DefaultResource()
		ctx.domain = parts[0]
	case 0:
		ctx.resource = ctx.DefaultResource()
		ctx.domain = ctx.DefaultDomain()
	default:
		ctx.resourceID = strings.Join(parts[2:], "/")
		ctx.resource = parts[1]
		ctx.domain = parts[0]
	}
	return ctx, nil
}

// String returns method, domain, resource and resource ID of the context
// in a readable way.
func (ctx *context) String() string {
	if ctx.resourceID == "" {
		return fmt.Sprintf("%s /%s/%s", ctx.request.Method, ctx.domain, ctx.resource)
	}
	return fmt.Sprintf("%s /%s/%s/%s", ctx.request.Method, ctx.domain, ctx.resource, ctx.resourceID)
}

// BasePath is specified on the Context interface.
func (ctx *context) BasePath() string {
	return ctx.multiplexer.basePath
}

// DefaultDomain is specified on the Context interface.
func (ctx *context) DefaultDomain() string {
	return ctx.multiplexer.defaultDomain
}

// DefaultResource is specified on the Context interface.
func (ctx *context) DefaultResource() string {
	return ctx.multiplexer.defaultResource
}

// Request is specified on the Context interface.
func (ctx *context) Request() *http.Request {
	return ctx.request
}

// ResponseWriter is specified on the Context interface.
func (ctx *context) ResponseWriter() http.ResponseWriter {
	return ctx.responseWriter
}

// Domain is specified on the Context interface.
func (ctx *context) Domain() string {
	return ctx.domain
}

// Resource is specified on the Context interface.
func (ctx *context) Resource() string {
	return ctx.resource
}

// ResourceID is specified on the Context interface.
func (ctx *context) ResourceID() string {
	return ctx.resourceID
}

// Scene is specified on the Context interface.
func (ctx *context) Scene() scene.Scene {
	return ctx.scene
}

// AcceptsContentType is specified on the Context interface.
func (ctx *context) AcceptsContentType(contentType string) bool {
	return strings.Contains(ctx.request.Header.Get("Accept"), contentType)
}

// HasContentType is specified on the Context interface.
func (ctx *context) HasContentType(contentType string) bool {
	return strings.Contains(ctx.request.Header.Get("Content-Type"), contentType)
}

// Languages is specified on the Context interface.
func (ctx *context) Languages() Languages {
	accept := ctx.request.Header.Get("Accept-Language")
	languages := Languages{}
	for _, part := range strings.Split(accept, ",") {
		lv := strings.Split(part, ";")
		if len(lv) == 1 {
			languages = append(languages, Language{lv[0], 1.0})
		} else {
			value, err := strconv.ParseFloat(lv[1], 64)
			if err != nil {
				value = 0.0
			}
			languages = append(languages, Language{lv[0], value})
		}
	}
	sort.Reverse(languages)
	return languages
}

// createPath creates a path out of the major URL parts.
func (ctx *context) createPath(domain, resource, resourceID string) string {
	path := ctx.BasePath() + domain + "/" + resource
	if resourceID != "" {
		path = path + "/" + resourceID
	}
	return path
}

// InternalPath is specified on the Context interface.
func (ctx *context) InternalPath(domain, resource, resourceID string, query ...KeyValue) string {
	path := ctx.createPath(domain, resource, resourceID)
	if len(query) > 0 {
		path += "?" + KeyValues(query).String()
	}
	return path
}

// Redirect is specified on the Context interface.
func (ctx *context) Redirect(domain, resource, resourceID string) {
	path := ctx.createPath(domain, resource, resourceID)
	http.Redirect(ctx.responseWriter, ctx.request, path, http.StatusTemporaryRedirect)
}

// RenderTemplate is specified on the Context interface.
func (ctx *context) RenderTemplate(templateID string, data interface{}) {
	ctx.multiplexer.templateCache.render(ctx.responseWriter, templateID, data)
}

// WriteGOB is specified on the Context interface.
func (ctx *context) WriteGOB(data interface{}) {
	enc := gob.NewEncoder(ctx.responseWriter)
	ctx.responseWriter.Header().Set("Content-Type", ContentTypeGOB)
	enc.Encode(data)
}

// ReadGOB is specified on the Context interface.
func (ctx *context) ReadGOB(data interface{}) error {
	if !ctx.HasContentType(ContentTypeGOB) {
		return errors.New(ErrInvalidContentType, errorMessages, ContentTypeGOB)
	}
	dec := gob.NewDecoder(ctx.request.Body)
	err := dec.Decode(data)
	ctx.request.Body.Close()
	return err
}

// WriteJSON is specified on the Context interface.
func (ctx *context) WriteJSON(data interface{}, html bool) {
	body, err := json.Marshal(data)
	if err != nil {
		http.Error(ctx.responseWriter, err.Error(), http.StatusInternalServerError)
	}
	if html {
		var buf bytes.Buffer
		json.HTMLEscape(&buf, body)
		body = buf.Bytes()
	}
	ctx.responseWriter.Header().Set("Content-Type", ContentTypeJSON)
	ctx.responseWriter.Write(body)
}

// PositiveJSONFeedback is specified on the Context interface.
func (ctx *context) PositiveJSONFeedback(msg string, p interface{}, args ...interface{}) {
	amsg := fmt.Sprintf(msg, args...)
	ctx.WriteJSON(&Envelope{true, amsg, p}, true)
}

// NegativeJSONFeedback is specified on the Context interface.
func (ctx *context) NegativeJSONFeedback(msg string, args ...interface{}) {
	amsg := fmt.Sprintf(msg, args...)
	ctx.WriteJSON(&Envelope{false, amsg, nil}, true)
}

// ReadJSON is specified on the Context interface.
func (ctx *context) ReadJSON(data interface{}) error {
	if !ctx.HasContentType(ContentTypeJSON) {
		return errors.New(ErrInvalidContentType, errorMessages, ContentTypeJSON)
	}
	body, err := ioutil.ReadAll(ctx.request.Body)
	ctx.request.Body.Close()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &data)
}

// ReadGenericJSON is specified on the Context interface.
func (ctx *context) ReadGenericJSON() (map[string]interface{}, error) {
	if !ctx.HasContentType(ContentTypeJSON) {
		return nil, errors.New(ErrInvalidContentType, errorMessages, ContentTypeJSON)
	}
	data := map[string]interface{}{}
	err := ctx.ReadJSON(&data)
	return data, err
}

// WriteXML is specified on the Context interface.
func (ctx *context) WriteXML(data interface{}) {
	body, err := xml.Marshal(data)
	if err != nil {
		http.Error(ctx.responseWriter, err.Error(), http.StatusInternalServerError)
	}
	ctx.responseWriter.Header().Set("Content-Type", ContentTypeXML)
	ctx.responseWriter.Write(body)
}

// ReadXML is specified on the Context interface.
func (ctx *context) ReadXML(data interface{}) error {
	if !ctx.HasContentType(ContentTypeXML) {
		return errors.New(ErrInvalidContentType, errorMessages, ContentTypeXML)
	}
	body, err := ioutil.ReadAll(ctx.request.Body)
	ctx.request.Body.Close()
	if err != nil {
		return err
	}
	return xml.Unmarshal(body, &data)
}

// EOF
