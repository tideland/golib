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
	"github.com/tideland/golib/stringex"
)

//--------------------
// PROCESSING FUNCTIONS
//--------------------

// splitPath splits and cleans the path into parts.
func splitPath(path, separator string) []string {
	return stringex.SplitFilter(path, separator, func(part string) bool {
		return part != ""
	})
}

// isValue checks if the raw is a value and returns it
// type-safe. Otherwise nil and false are returned.
func isValue(raw interface{}) (interface{}, bool) {
	v, ok := raw.(interface{})
	return v, ok
}

// isObject checks if the raw is an object and returns it
// type-safe. Otherwise nil and false are returned.
func isObject(raw interface{}) (map[string]interface{}, bool) {
	o, ok := raw.(map[string]interface{})
	return o, ok
}

// isArray checks if the raw is an array and returns it
// type-safe. Otherwise nil and false are returned.
func isArray(raw interface{}) ([]interface{}, bool) {
	a, ok := raw.([]interface{})
	return a, ok
}

// valueAt returns the value at the path parts.
func valueAt(node interface{}, parts []string) (interface{}, error) {
	length := len(parts)
	if length == 0 {
		// End of the parts.
		return node, nil
	}
	// Further access depends on part content and type.
	part := parts[0]
	if part == "" {
		return node, nil
	}
	if o, ok := isObject(node); ok {
		// JSON object.
		field, ok := o[part]
		if !ok {
			return nil, errors.New(ErrInvalidPart, errorMessages, part)
		}
		return valueAt(field, parts[1:])
	}
	if a, ok := isArray(node); ok {
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
func setValueAt(root, value interface{}, parts []string) error {
	h, t := ht(parts)
	return setNodeValueAt(root, nil, value, h, t)
}

// ht retrieves head and tail from parts.
func ht(parts []string) (string, []string) {
	switch len(parts) {
	case 0:
		return "", []string{}
	case 1:
		return parts[0], []string{}
	default:
		return parts[0], parts[1:]
	}
}

// setNodeValueAt is used recursively by setValueAt().
func setNodeValueAt(node, parent, value interface{}, head string, tail []string) error {
	if head == "" {
		// End of the game.
		return errors.New(ErrInvalidPath, errorMessages)
	}
	if o, ok := isObject(node); ok {
		// JSON object.
		_, ok := isValue(o[head])
		if len(tail) == 0 && ok {
			o[head] = value
			return nil
		}
		h, t := ht(tail)
		return setNodeValueAt(o[head], o, value, h, t)
	}
	if a, ok := isArray(node); ok {
		// JSON array.
		index, err := strconv.Atoi(head)
		// TODO Mue 2017-07-18 Mue Extend array if too small.
		if err != nil || index >= len(a) {
			return errors.New(ErrInvalidPart, errorMessages, head)
		}
		_, ok := isValue(a[index])
		if len(tail) == 0 && ok {
			a[index] = value
			return nil
		}
		h, t := ht(tail)
		return setNodeValueAt(a[index], a, value, h, t)
	}
	return errors.New(ErrInvalidPath, errorMessages)
}

// process processes one leaf or node.
func process(raw interface{}, path []string, separator string, processor ValueProcessor) error {
	return nil
}

// EOF
