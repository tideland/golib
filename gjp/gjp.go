// Tideland Go Library - Generic JSON Parser
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

	"github.com/tideland/golib/errors"
)

//--------------------
// DOCUMENT
//--------------------

// ValueProcessor describes a function for the processing of
// values while iterating over a document.
type ValueProcessor func(path string, value Value) error

// Document represents one JSON document.
type Document interface {
	json.Marshaler

	// Length returns the number of elements for the given path.
	Length(path string) int

	// ValueAt returns the addressed value.
	ValueAt(path string) Value

	// Process iterates over a document and processes its values.
	// There's no order, so nesting into an embedded document or
	// list may come earlier than higher level paths.
	Process(processor ValueProcessor) error
}

// document implements Document.
type document struct {
	root *root
}

// Parse reads a raw document and returns it as
// accessible document.
func Parse(data []byte, separator string) (Document, error) {
	var raw interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, errors.Annotate(err, ErrUnmarshalling, errorMessages)
	}
	return &document{
		root: newRoot(separator, rawToNoder(raw)),
	}, nil
}

// Length implements Document.
func (d *document) Length(path string) int {
	nr, ok := d.root.noderAt(path)
	if !ok {
		// Noder not found.
		return -1
	}
	// Check if node or value.
	n, ok := nr.isNode()
	if ok {
		return len(n)
	}
	return 1
}

// ValueAt implements Document.
func (d *document) ValueAt(path string) Value {
	raw, ok := d.root.valueAt(path)
	return &value{raw, ok && raw != nil}
}

// Process implements Document.
func (d *document) Process(processor ValueProcessor) error {
	return d.root.process([]string{}, processor)
}

// MarshalJSON implements json.Marshaler.
func (d *document) MarshalJSON() ([]byte, error) {
	raw := noderToRaw(d.root.data)
	return json.Marshal(raw)
}

// EOF
