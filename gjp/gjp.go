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
	root      noder
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
		separator: separator,
		root:      rawToNoder(raw),
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
	nr, ok := d.noderAt(path)
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

// SetValueAt implements Document.
func (d *document) SetValueAt(path string, value interface{}) error {
	return d.setValueAt(path, value)
}

// ValueAt implements Document.
func (d *document) ValueAt(path string) Value {
	raw, ok := d.valueAt(path)
	return &value{raw, ok && raw != nil}
}

// Clear implements Document.
func (d *document) Clear() {
	d.root = nil
}

// Query implements Document.
func (d *document) Query(pattern string) (PathValues, error) {
	return nil, nil
}

// Process implements Document.
func (d *document) Process(processor ValueProcessor) error {
	return d.root.process([]string{}, d.separator, processor)
}

// MarshalJSON implements json.Marshaler.
func (d *document) MarshalJSON() ([]byte, error) {
	raw := noderToRaw(d.root)
	return json.Marshal(raw)
}

// setValueAt sets a value at a given path. If needed it's created.
func (d *document) setValueAt(path string, value interface{}) error {
	nr, err := d.ensureNoderAt(path)
	if err != nil {
		return errors.Annotate(err, ErrSetting, errorMessages, path)
	}
	err = nr.setValue(value)
	if err != nil {
		return errors.Annotate(err, ErrSetting, errorMessages, path)
	}
	return nil
}

// valueAt retrieves the data at a given path.
func (d *document) valueAt(path string) (interface{}, bool) {
	nr, ok := d.noderAt(path)
	if !ok {
		// Noder not found.
		return nil, false
	}
	_, ok = nr.isNode()
	if ok {
		// We need a value.
		return nil, false
	}
	// Found our value.
	return nr.value(), true
}

// ensureNoderAt ensures and returns a noder at a given path.
func (d *document) ensureNoderAt(path string) (noder, error) {
	parts := stringex.SplitMap(path, d.separator, func(p string) (string, bool) {
		if p == "" {
			return "", false
		}
		return p, true
	})
	// Check if data has been initialized.
	if d.root == nil {
		if len(parts) == 0 {
			d.root = &leaf{}
		} else {
			d.root = node{}
		}
	}
	// Now let the root handle the parts.
	return d.root.ensureNoderAt(parts...)
}

// noderAt retrieves the noder at a given path.
func (d *document) noderAt(path string) (noder, bool) {
	if d.root == nil {
		// No data yet.
		return nil, false
	}
	parts := stringex.SplitMap(path, d.separator, func(p string) (string, bool) {
		if p == "" {
			return "", false
		}
		return p, true
	})
	if len(parts) == 0 {
		// Root is all we need.
		return d.root, true
	}
	n, ok := d.root.isNode()
	if !ok {
		// No node, but path needs it.
		return nil, false
	}
	nr, ok := n.at(parts)
	if !ok {
		// Not found.
		return nil, false
	}
	// Found noder.
	return nr, true
}

// EOF
