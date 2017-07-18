// Tideland Go Library - Generic JSON Processor
//
// Copyright (C) 2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package gjp

//--------------------
// IMPORTS
//--------------------

import (
	"encoding/json"

	"strings"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/stringex"
)

//--------------------
// DOCUMENT
//--------------------

// PathValue is the combination of path and value.
type PathValue struct {
	Path  string
	Value Value
}

// PathValues contains a number of path/value combinations.
type PathValues []PathValue

// ValueProcessor describes a function for the processing of
// values while iterating over a document.
type ValueProcessor func(path string, value Value) error

// Document represents one JSON document.
type Document interface {
	json.Marshaler

	// Length returns the number of elements for the given path.
	Length(path string) int

	// SetValueAt sets the value at the given path.
	SetValueAt(path string, value interface{}) error

	// ValueAt returns the addressed value.
	ValueAt(path string) Value

	// Clear removes the so far build document data.
	Clear()

	// Query allows to find pathes matching a given pattern.
	Query(pattern string) (PathValues, error)

	// Process iterates over a document and processes its values.
	// There's no order, so nesting into an embedded document or
	// list may come earlier than higher level paths.
	Process(processor ValueProcessor) error
}

// document implements Document.
type document struct {
	separator string
	root      interface{}
}

// Parse reads a raw document and returns it as
// accessible document.
func Parse(data []byte, separator string) (Document, error) {
	var root interface{}
	err := json.Unmarshal(data, &root)
	if err != nil {
		return nil, errors.Annotate(err, ErrUnmarshalling, errorMessages)
	}
	return &document{
		separator: separator,
		root:      root,
	}, nil
}

// NewDocument creates a new empty document.
func NewDocument(separator string) Document {
	return &document{
		separator: separator,
	}
}

// Length implements Document.
func (d *document) Length(path string) int {
	n, err := valueAt(d.root, splitPath(path, d.separator))
	if err != nil {
		return -1
	}
	// Check if object or array.
	o, ok := isObject(n)
	if ok {
		return len(o)
	}
	a, ok := isArray(n)
	if ok {
		return len(a)
	}
	return 1
}

// SetValueAt implements Document.
func (d *document) SetValueAt(path string, value interface{}) error {
	return d.setValueAt(path, value)
}

// ValueAt implements Document.
func (d *document) ValueAt(path string) Value {
	n, err := valueAt(d.root, splitPath(path, d.separator))
	return &value{n, err == nil}
}

// Clear implements Document.
func (d *document) Clear() {
	d.root = nil
}

// Query implements Document.
func (d *document) Query(pattern string) (PathValues, error) {
	pvs := PathValues{}
	err := d.Process(func(path string, value Value) error {
		if stringex.Matches(pattern, path, false) {
			pvs = append(pvs, PathValue{
				Path:  path,
				Value: value,
			})
		}
		return nil
	})
	return pvs, err
}

// Process implements Document.
func (d *document) Process(processor ValueProcessor) error {
	return process(d.root, []string{}, d.separator, processor)
}

// MarshalJSON implements json.Marshaler.
func (d *document) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.root)
}

// setValueAt sets a value at a given path. If needed it's created.
func (d *document) setValueAt(path string, value interface{}) error {
	parts := strings.Split(path, d.separator)
	return setValueAt(d.root, value, parts)
}

// valueAt retrieves the data at a given path.
func (d *document) valueAt(path string) (interface{}, bool) {
	parts := strings.Split(path, d.separator)
	n, err := valueAt(d.root, parts)
	if err != nil {
		return nil, false
	}
	return n, true
}

// EOF
