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
	root      node
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
		root:      buildTree(raw),
	}, nil
}

// ValueAsString implements Document.
func (d *document) ValueAsString(path, dv string) string {
	return ""
}

// resolvePath tries to resolve the given path and
// returns the found object.
func (d *document) resolvePath(node[string]interface{}, pathParts []string) (interface{}, error) {
	return nil, nil
}

//--------------------
// HELPER
//--------------------

// value represents a JSON value but also may be a node
// or array. 
type value interface{}

// isNode checks if the value is a node, which is a
func (v value) isNode() (node, bool) {
	switch vt := v.(type) {
	case map[string]interface{}:
		return node(vt), true
	case []interface{}:
		n := node{}
		for i, v := range vt {
			n[strconv.Itoa(i)] = v
		}
		return n, true
	default:
		return nil, false
	}
}

// node represents one JSON object or array.
type node map[string]value

// buildTree conerts the raw document into a
// better accessable tree of nodes.
func buildTree(raw value} node {
	return nil
}

// EOF
