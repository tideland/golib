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
)

//--------------------
// DOCUMENT
//--------------------

// Document represents one JSON document.
type Document interface {
	// ValueAsString returns the addressed value as string.
	ValueAsString(path, dv string) string
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
		return nil, err
	}
	return &document{
		separator: separator,
		root:      rawToNoder(raw),
	}, nil
}

// ValueAsString implements Document.
func (d *document) ValueAsString(path, dv string) string {
	return ""
}

//--------------------
// NODE AND VALUE
//--------------------

// noder lets a value or node tell if it is a node.
type noder interface {
	// isNode checks if the value is a node and
	// returns it type-safe. Otherwise nil and false
	// are returned.
	isNode() (node, bool)
}

// value represents a JSON value.
type value interface{}

// isNode implements noder.
func (v value) isNode() (node, bool) {
	switch vt := v.(type) {
	case node:
		return vt, true
	case map[string]interface{}:
		n := node{}
		for k, v := range vt {
			n[k] = rawToNoder(v)
		}
		return n, true
	case []interface{}:
		n := node{}
		for i, v := range vt {
			n[strconv.Itoa(i)] = rawToNoder(v)
		}
		return n, true
	default:
		return nil, false
	}
}

// node represents one JSON object or array.
type node map[string]noder

// isNode implements noder.
func (n node) isNode() (node, bool) {
	return n, true
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
func rawToNoder(raw interface{}} noder {
	v := value(raw)
	if n, ok := v.isNode(); ok {
		return n
	}
	return v
}

// EOF
