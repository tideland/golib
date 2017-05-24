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
	"strconv"
	"strings"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/stringex"
)

//--------------------
// DOCUMENT
//--------------------

// ValueProcessor describes a function for the processing of
// values while iterating over a document.
type ValueProcessor func(path string, value Value) error

// Document represents one JSON document.
type Document interface {
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
	separator string
	root      noder
}

// Parse reads a raw document and returns it as
// accessable document.
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

// ValueAt implements Document.
func (d *document) ValueAt(path string) Value {
	raw, ok := d.valueAt(path)
	return &value{raw, ok && raw != nil}
}

// Process implements Document.
func (d *document) Process(processor ValueProcessor) error {
	return d.root.process([]string{}, d.separator, processor)
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

// noderAt retrieves the noder at a given path.
func (d *document) noderAt(path string) (noder, bool) {
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

//--------------------
// NODE AND LEAF
//--------------------

// noder lets a value or node tell if it is a node.
type noder interface {
	// isNode checks if the value is a node and
	// returns it type-safe. Otherwise nil and false
	// are returned.
	isNode() (node, bool)

	// value returns the raw value of a node.
	value() interface{}

	// process processes one leaf or node.
	process(path []string, separator string, processor ValueProcessor) error
}

// leaf represents a leaf in a JSON document tree.It
// contains the value.
type leaf struct {
	raw interface{}
}

// isNode implements noder.
func (l leaf) isNode() (node, bool) {
	switch rt := l.raw.(type) {
	case node:
		return rt, true
	case map[string]interface{}:
		n := node{}
		for k, v := range rt {
			n[k] = rawToNoder(v)
		}
		return n, true
	case []interface{}:
		n := node{}
		for i, v := range rt {
			n[strconv.Itoa(i)] = rawToNoder(v)
		}
		return n, true
	default:
		return nil, false
	}
}

// value implements noder.
func (l leaf) value() interface{} {
	return l.raw
}

// process implements noder.
func (l leaf) process(path []string, separator string, processor ValueProcessor) error {
	return processor(strings.Join(path, separator), &value{l.raw, l.raw != nil})
}

// node represents one JSON object or array.
type node map[string]noder

// isNode implements noder.
func (n node) isNode() (node, bool) {
	return n, true
}

// value implements noder.
func (n node) value() interface{} {
	return n
}

// process implements noder.
func (n node) process(path []string, separator string, processor ValueProcessor) error {
	for nk, nn := range n {
		np := append(path, nk)
		err := nn.process(np, separator, processor)
		if err != nil {
			ps := strings.Join(np, separator)
			return errors.Annotate(err, ErrProcessing, errorMessages, ps)
		}
	}
	return nil
}

// at returns the noder at the given path or
// nil and false.
func (n node) at(path []string) (noder, bool) {
	lp := len(path)
	if lp == 0 {
		// End of path.
		return n, true
	}
	nr, ok := n[path[0]]
	if !ok {
		// Path part not found.
		return nil, false
	}
	nn, ok := nr.isNode()
	if ok {
		// Continue recursively.
		return nn.at(path[1:])
	}
	if lp > 1 {
		// Reached value before end of path.
		return nil, false
	}
	// We're done.
	return nr, true
}

// rawToNoder conerts the raw interface into a
// noder which may be a node or a value.
func rawToNoder(raw interface{}) noder {
	l := leaf{raw}
	if n, ok := l.isNode(); ok {
		return n
	}
	return l
}

// EOF
