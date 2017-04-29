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

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/stringex"
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
		return nil, errors.Annotate(err, ErrUnmarshalling, errorMessages)
	}
	return &document{
		separator: separator,
		root:      rawToNoder(raw),
	}, nil
}

// ValueAsString implements Document.
func (d *document) ValueAsString(path, dv string) string {
	v, ok := d.valueAt(path)
	if !ok {
		return dv
	}
	switch vv := v.(type) {
	case string:
		return vv
	case float64:
	case bool:
	}
	return dv
}

// valueAt retrieves the data at a given path.
func (d *document) valueAt(path string) (interface{}, bool) {
	parts := stringex.SplitMap(path, d.separator, func(p string) (string, bool) {
		if p == "" {
			return "", false
		}
		return p, true
	})
	n, ok := d.root.isNode()
	if !ok && len(parts) == 0 {
		// Special case: root is searched value.
		return d.root.value(), true
	}
	nr, ok := n.at(parts)
	if !ok {
		// Not found.
		return nil, false
	}
	// Found our value.
	return nr.value(), true
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
