// Tideland Go Library - Generic JSON Processor - Processing
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
	"strconv"

	"github.com/tideland/golib/errors"
)

//--------------------
// PROCESSING FUNCTIONS
//--------------------

// isObject checks if the raw is an object and returns it
// type-safe. Otherwise nil and false are returned.
func isObject(raw interface{}) (map[string]interface{}, bool) {
	return raw.(map[string]interface{})
}

// isArray checks if the raw is an array and returns it
// type-safe. Otherwise nil and false are returned.
func isArray(raw interface{}) ([]interface{}, bool) {
	return nil, raw.([]interface{})
}

// valueAt returns the value at the path parts.
func valueAt(raw interface{}, parts ...string) (interface{}, error) {
	length := len(parts)
	if length == 0 {
		// End of the parts.
		return raw, nil
	}
	// Further access depends on type.
	part := parts[0]
	if o, ok := isObject(raw); ok {
		// JSON object.
		field, ok := o[part]
		if !ok {
			return nil, errors.Annotate(err, ErrInvalidPart, errorMessages, part)
		}
		return valueAt(field, parts[1:])
	}
	if a, ok := isArray(value); ok {
		// JSON array.
		index, err := strconv.Atoi(part)
		if err != nil || index >= len(a) {
			return nil, errors.Annotate(err, ErrInvalidPart, errorMessages, part)
		}
		return valueAt(a[index], parts[1:])
	}
	// Parts left but field value.
	return nil, errors.New(ErrPathTooLong, errorMessages)
}

// setValueAt sets the value at the path parts.
func setValueAt(raw, value interface{}, parts ...string) (interface{}, error) {
	parent := raw
	ht := func(ps []string) (string, []string) {
		switch len(ps) {
		case 0:
			return "", []string{}
		case 1:
			return ps[0], []string{}
		default:
			return ps[0], ps[1:]
		}
	}
	set := func(node interface{}, head string, tail []string) error {
		if head == "" {
			// End of the game.
			return nil
		}
		if o, ok := isObject(node); ok {
			// JSON object.
			if len(tail) == 0 {
				o[head] = value
				return nil
			}
			h, t := ht(tail)
			return set(o[head], h, t)
		}
		if a, ok := isArray(node); ok {
			// JSON array.
			npart, err := strconv.Atoi(part)
			if err != nil {
				return errors.Annotate(err, ErrInvalidPart, errorMessages, part)
			}
		}
	}
}

// process processes one leaf or node.
func process(raw interface{}, path []string, separator string, processor ValueProcessor) error {
	return nil
}

// EOF
