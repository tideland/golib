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
	content   interface{}
}

// Parse reads a raw document and returns it as
// accessable document.
func Parse(data []byte, separator string) (Document, error) {
	var content interface{}
	err := json.Unmarshal(data, &content)
	if err != nil {
		return nil, err
	}
	return &document{
		separator: separator,
		content:   content,
	}, nil
}

// ValueAsString implements Document.
func (d *document) ValueAsString(path, dv string) string {
	return ""
}

// splitPath splits the path by the separator.
func (d *document) splitPath(path string) []string {
	return strings.Split(path, d.separator)
}

// resolvePath tries to resolve the given path and
// returns the found object.
func (d *document) resolvePath(pathParts []string) (interface{}, error) {
	return nil, nil
}

// EOF
