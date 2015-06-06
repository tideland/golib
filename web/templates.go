// Tideland Go Library - Web - Templates
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
	"io/ioutil"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/tideland/golib/errors"
)

//--------------------
// TEMPLATE CACHE
//--------------------

// templateCacheEntry stores the parsed template and the
// content type.
type templateCacheEntry struct {
	id             string
	filename       string
	timestamp      time.Time
	parsedTemplate *template.Template
	contentType    string
}

// load reads the raw template from a file and parses it.
func (tce *templateCacheEntry) load(filename string) error {
	rawTemplate, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return tce.parse(string(rawTemplate))
}

// parse parses a raw template.
func (tce *templateCacheEntry) parse(rawTemplate string) error {
	parsedTemplate, err := template.New(tce.id).Parse(rawTemplate)
	if err != nil {
		return err
	}
	tce.timestamp = time.Now()
	tce.parsedTemplate = parsedTemplate
	return nil
}

// isValid checks if the the entry is younger than the
// passed validity period.
func (tce *templateCacheEntry) isValid(validityPeriod time.Duration) bool {
	return tce.timestamp.Add(validityPeriod).After(time.Now())
}

// templateCache stores preparsed templates.
type templateCache struct {
	mux   sync.RWMutex
	cache map[string]*templateCacheEntry
}

// newTemplateCache creates a new cache.
func newTemplateCache() *templateCache {
	return &templateCache{
		cache: make(map[string]*templateCacheEntry),
	}
}

// parse parses a template an stores it.
func (tc *templateCache) parse(id, t, ct string) error {
	tc.mux.Lock()
	defer tc.mux.Unlock()
	tmpl, err := template.New(id).Parse(t)
	if err != nil {
		return err
	}
	tc.cache[id] = &templateCacheEntry{id, "", time.Now(), tmpl, ct}
	return nil
}

// loadAndParse loads a template out of the filesystem, parses and stores it.
func (tc *templateCache) loadAndParse(id, fn, ct string) error {
	t, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}
	return tc.parse(id, string(t), ct)
}

// render executes the pre-parsed template with the data. It also sets
// the content type header.
func (tc *templateCache) render(rw http.ResponseWriter, id string, data interface{}) error {
	tc.mux.RLock()
	defer tc.mux.RUnlock()
	entry, ok := tc.cache[id]
	if !ok {
		return errors.New(ErrNoCachedTemplate, errorMessages, id)
	}
	rw.Header().Set("Content-Type", entry.contentType)
	err := entry.parsedTemplate.Execute(rw, data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// EOF
