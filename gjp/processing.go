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
func setValueAt(raw, value interface{}, parts ...string) error {
	length := len(parts)
	if length == 0 {
		// End of the parts.
		return errors.New(ErrSetting, errorMessages)
	}
	part := parts[length-1]
	parent, err := valueAt(raw, parts[:length-1]...)
	if err != nil {
		return errors.Annotate(err, ErrSetting, errorMessages, part)
	}
	// Check type of parent.
	if o, ok := isObject(parent); ok {
		// JSON object.
		o[part] = value
		return nil
	}
	if a, ok := isArray(parent); ok {
		// JSON array.
		index, err := strconv.Atoi(part)
		if err != nil {
			return errors.Annotate(err, ErrInvalidPart, errorMessages, part)
		}
		if index >= len(a) {
			// TODO 2017-07-15 Mue Extend the array.
		}
		a[index] = value
		return nil
	}
	return errors.New(ErrSetting, errorMessages, part)
}

// process processes one leaf or node.
func process(raw interface{}, path []string, separator string, processor ValueProcessor) error {
	return nil
}

//--------------------
// LEAF
//--------------------

// leaf represents a leaf in a JSON document tree.It
// contains the value.
type leaf struct {
	raw interface{}
}

// isObject implements noder.
func (l *leaf) isObject() (object, bool) {
	switch rt := l.raw.(type) {
	case object:
		return rt, true
	case map[string]interface{}:
		o := object{}
		for k, v := range rt {
			o[k] = rawToNoder(v)
		}
		return o, true
	case []interface{}:
		a := array{}
		for _, v := range rt {
			a = append(a, rawToNoder(v))
		}
		return a, true
	default:
		return nil, false
	}
}

// isArray implements noder.
func (l *leaf) isArray() (array, bool) {
	return nil, false
}

// ensureNoderAt ensures the existing of a leaf noder
// based on the path parts. It has to be the last one,
// so the length of parts has to be 0 for a positive
// answer.
func (l *leaf) ensureNoderAt(parts ...string) (noder, error) {
	if len(parts) == 0 {
		return l, nil
	}
	return nil, errors.New(ErrLeafToNode, errorMessages)
}

// setValue implements noder.
func (l *leaf) setValue(value interface{}) error {
	if value == nil {
		l.raw = nil
		return nil
	}
	switch tv := value.(type) {
	case string, int, float64, bool:
		l.raw = tv
	default:
		return errors.New(ErrUnsupportedType, errorMessages, value)
	}
	return nil
}

// value implements noder.
func (l *leaf) value() interface{} {
	return l.raw
}

// rawValue implements noder.
func (l *leaf) rawValue() interface{} {
	return l.raw
}

// process implements noder.
func (l *leaf) process(path []string, separator string, processor ValueProcessor) error {
	return processor(strings.Join(path, separator), &value{l.raw, l.raw != nil})
}

//--------------------
// OBJECT
//--------------------

// object represents one JSON object.
type object map[string]noder

// isObject implements noder.
func (o object) isObject() (object, bool) {
	return o, true
}

// isArray implements noder.
func (o object) isArray() (array, bool) {
	return nil, false
}

// ensureNoderAt ensures the existing of a leaf noder
// based on the path parts.
func (o object) ensureNoderAt(parts ...string) (noder, error) {
	switch len(parts) {
	case 0:
		// Addressing this object.
		return nil, errors.New(ErrNodeToLeaf, errorMessages)
	case 1:
		// Last part.
		pnoder := o[parts[0]]
		if pnoder == nil {
			o[parts[0]] = &leaf{}
		}
		return o[parts[0]], nil
	default:
		// More to come.
		pnoder := o[parts[0]]
		if pnoder == nil {
			o[parts[0]] = object{}
		}
		return o[parts[0]].ensureNoderAt(parts[1:]...)
	}
}

// setValue implements noder.
func (o object) setValue(value interface{}) error {
	return errors.New(ErrCorrupting, errorMessages)
}

// value implements noder.
func (o object) value() interface{} {
	return o
}

// rawValue implements noder.
func (o object) rawValue() interface{} {
	raw := map[string]interface{}{}
	for key, value := range o {
		raw[key] = value.rawValue()
	}
	return raw
}

// process implements noder.
func (o object) process(path []string, separator string, processor ValueProcessor) error {
	for ok, on := range o {
		op := append(path, ok)
		err := on.process(op, separator, processor)
		if err != nil {
			ps := strings.Join(op, separator)
			return errors.Annotate(err, ErrProcessing, errorMessages, ps)
		}
	}
	return nil
}

// at returns the noder at the given path or
// nil and false.
func (o object) at(path []string) (noder, bool) {
	lp := len(path)
	if lp == 0 {
		// End of path.
		return o, true
	}
	pzero, ok := o[path[0]]
	if !ok {
		// Path part not found.
		return nil, false
	}
	if ozero, ok := pzero.isObject(); ok {
		// Object, continue recursively.
		return ozero.at(path[1:])
	} else if azero, ok := pzero.isArray(); ok {
		// Array, continue recursively.
		return azero.at(path[1:])
	}
	if lp > 1 {
		// Reached value before end of path.
		return nil, false
	}
	// We're done.
	return pzero, true
}

//--------------------
// ARRAY
//--------------------

// array represents one JSON array.
type array []noder

// isObject implements noder.
func (a array) isObject() (object, bool) {
	return nil, false
}

// isArray implements noder.
func (a array) isArray() (array, bool) {
	return a, true
}

// ensureNoderAt ensures the existing of a leaf noder
// based on the path parts.
func (a array) ensureNoderAt(parts ...string) (noder, error) {
	plen := len(parts)
	if plen == 0 {
		// Addressing this array.
		return nil, errors.New(ErrNodeToLeaf, errorMessages)
	}
	index, err := strconv.Atoi(parts[0])
	if err != nil {
		// TODO 2017-07-14 Mue Need different error
		return nil, errors.Annotate(err, ErrNodeToLeaf, errorMessages)
	}
	if index >= len(a) {
		// TODO 2017-07-14 Mue Enhance array.
	}
	if plen == 1 {
		pnoder := a[index]
		if pnoder == nil {
			a[index] = &leaf{}
		}
		return a[index], nil
	}
	// More to come.
	pnoder := a[index]
	if pnoder == nil {
		a[index] = object{}
	}
	return a[index].ensureNoderAt(parts[1:]...)
}

// setValue implements noder.
func (a array) setValue(value interface{}) error {
	return errors.New(ErrCorrupting, errorMessages)
}

// value implements noder.
func (a array) value() interface{} {
	return a
}

// rawValue implements noder.
func (a array) rawValue() interface{} {
	raw := []interface{}{}
	for _, value := range a {
		raw = append(raw, value.rawValue())
	}
	return raw
}

// process implements noder.
func (a array) process(path []string, separator string, processor ValueProcessor) error {
	for ok, on := range o {
		op := append(path, ok)
		err := on.process(op, separator, processor)
		if err != nil {
			ps := strings.Join(op, separator)
			return errors.Annotate(err, ErrProcessing, errorMessages, ps)
		}
	}
	return nil
}

// at returns the noder at the given path or
// nil and false.
func (o object) at(path []string) (noder, bool) {
	lp := len(path)
	if lp == 0 {
		// End of path.
		return o, true
	}
	pzero, ok := o[path[0]]
	if !ok {
		// Path part not found.
		return nil, false
	}
	if ozero, ok := pzero.isObject(); ok {
		// Object, continue recursively.
		return ozero.at(path[1:])
	} else if azero, ok := pzero.isArray(); ok {
		// Array, continue recursively.
		return azero.at(path[1:])
	}
	if lp > 1 {
		// Reached value before end of path.
		return nil, false
	}
	// We're done.
	return pzero, true
}

//--------------------
// CONVERTING
//--------------------

// rawToNoder conerts the raw interface into a
// noder which may be a node or a value.
func rawToNoder(raw interface{}) noder {
	l := &leaf{raw}
	if n, ok := l.isNode(); ok {
		return n
	}
	return l
}

// noderToRaw creates a marshable structure
// out of a noder.
func noderToRaw(nr noder) interface{} {
	return nr.rawValue()
}

// EOF
