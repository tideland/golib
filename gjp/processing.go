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
	"strings"

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
	if raw == nil {
		return raw, true
	}
	_, ook := isObject(raw)
	_, aok := isArray(raw)
	if ook || aok {
		return nil, false
	}
	return raw, true
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
	// Further access depends on part content node and type.
	head, tail := ht(parts)
	if head == "" {
		return node, nil
	}
	if o, ok := isObject(node); ok {
		// JSON object.
		field, ok := o[head]
		if !ok {
			return nil, errors.New(ErrInvalidPart, errorMessages, head)
		}
		return valueAt(field, tail)
	}
	if a, ok := isArray(node); ok {
		// JSON array.
		index, err := strconv.Atoi(head)
		if err != nil || index >= len(a) {
			return nil, errors.Annotate(err, ErrInvalidPart, errorMessages, head)
		}
		return valueAt(a[index], tail)
	}
	// Parts left but field value.
	return nil, errors.New(ErrPathTooLong, errorMessages)
}

// setValueAt sets the value at the path parts.
func setValueAt(root, value interface{}, parts []string) (interface{}, error) {
	h, t := ht(parts)
	return setNodeValueAt(root, value, h, t)
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
func setNodeValueAt(node, value interface{}, head string, tail []string) (interface{}, error) {
	// Check for nil node first.
	if node == nil {
		return addNodeValueAt(value, head, tail)
	}
	// Otherwise it should be an object or an array.
	if o, ok := isObject(node); ok {
		// JSON object.
		_, ok := isValue(o[head])
		switch {
		case !ok && len(tail) == 0:
			return nil, errors.New(ErrCorruptingDocument, errorMessages)
		case ok && o[head] != nil && len(tail) > 0:
			return nil, errors.New(ErrCorruptingDocument, errorMessages)
		case ok && len(tail) == 0:
			o[head] = value
		default:
			h, t := ht(tail)
			subnode, err := setNodeValueAt(o[head], value, h, t)
			if err != nil {
				return nil, err
			}
			o[head] = subnode
		}
		return o, nil
	}
	if a, ok := isArray(node); ok {
		// JSON array.
		index, err := strconv.Atoi(head)
		if err != nil {
			return nil, errors.New(ErrInvalidPart, errorMessages, head)
		}
		a = ensureArray(a, index+1)
		_, ok := isValue(a[index])
		switch {
		case !ok && len(tail) == 0:
			return nil, errors.New(ErrCorruptingDocument, errorMessages)
		case ok && a[index] != nil && len(tail) > 0:
			return nil, errors.New(ErrCorruptingDocument, errorMessages)
		case ok && len(tail) == 0:
			a[index] = value
		default:
			h, t := ht(tail)
			subnode, err := setNodeValueAt(a[index], value, h, t)
			if err != nil {
				return nil, err
			}
			a[index] = subnode
		}
		return a, nil
	}
	return nil, errors.New(ErrInvalidPath, errorMessages, head)
}

// addNodeValueAt is used recursively by setValueAt().
func addNodeValueAt(value interface{}, head string, tail []string) (interface{}, error) {
	// JSON value.
	if head == "" {
		return value, nil
	}
	index, err := strconv.Atoi(head)
	if err != nil {
		// JSON object.
		o := map[string]interface{}{}
		if len(tail) == 0 {
			o[head] = value
			return o, nil
		}
		h, t := ht(tail)
		subnode, err := addNodeValueAt(value, h, t)
		if err != nil {
			return nil, err
		}
		o[head] = subnode
		return o, nil
	}
	// JSON array.
	a := ensureArray([]interface{}{}, index+1)
	if len(tail) == 0 {
		a[index] = value
		return a, nil
	}
	h, t := ht(tail)
	subnode, err := addNodeValueAt(value, h, t)
	if err != nil {
		return nil, err
	}
	a[index] = subnode
	return a, nil
}

// ensureArray ensures the right len of an array.
func ensureArray(a []interface{}, l int) []interface{} {
	if len(a) >= l {
		return a
	}
	b := make([]interface{}, l)
	copy(b, a)
	return b
}

// process processes node recursively.
func process(node interface{}, parts []string, separator string, processor ValueProcessor) error {
	mkerr := func(err error, ps []string) error {
		return errors.Annotate(err, ErrProcessing, errorMessages, pathify(ps, separator))
	}
	// First check objects and arrays.
	if o, ok := isObject(node); ok {
		if len(o) == 0 {
			// Empty object.
			return processor(pathify(parts, separator), &value{o, nil})
		}
		for field, subnode := range o {
			fieldparts := append(parts, field)
			if err := process(subnode, fieldparts, separator, processor); err != nil {
				return mkerr(err, parts)
			}
		}
		return nil
	}
	if a, ok := isArray(node); ok {
		if len(a) == 0 {
			// Empty array.
			return processor(pathify(parts, separator), &value{a, nil})
		}
		for index, subnode := range a {
			indexparts := append(parts, strconv.Itoa(index))
			if err := process(subnode, indexparts, separator, processor); err != nil {
				return mkerr(err, parts)
			}
		}
		return nil
	}
	// Reached a value at the end.
	return processor(pathify(parts, separator), &value{node, nil})
}

// pathify creates a path out of parts and separator.
func pathify(parts []string, separator string) string {
	return separator + strings.Join(parts, separator)
}

// EOF
