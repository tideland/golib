// Tideland Go Library - Generic JSON Parser - Root
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
// ROOT
//--------------------

// root manages the document structure starting at
// its root.
type root struct {
	separator string
	data      noder
}

// newRoot creates a document root.
func newRoot(separator string, data noder) *root {
	return &root{
		separator: separator,
		data:      data,
	}
}

// setValueAt sets a value at a given path. If needed it's created.
func (r *root) setValueAt(path string, value interface{}) error {
	nr, err := r.ensureNoderAt(path)
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
func (r *root) valueAt(path string) (interface{}, bool) {
	nr, ok := r.noderAt(path)
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
func (r *root) ensureNoderAt(path string) (noder, error) {
	parts := stringex.SplitMap(path, r.separator, func(p string) (string, bool) {
		if p == "" {
			return "", false
		}
		return p, true
	})
	// Check if data has been initialized.
	if r.data == nil {
		if len(parts) == 0 {
			r.data = &leaf{}
		} else {
			r.data = node{}
		}
	}
	// Now let the data handle the parts.
	return r.data.ensureNoderAt(parts...)
}

// noderAt retrieves the noder at a given path.
func (r *root) noderAt(path string) (noder, bool) {
	if r.data == nil {
		// No data yet.
		return nil, false
	}
	parts := stringex.SplitMap(path, r.separator, func(p string) (string, bool) {
		if p == "" {
			return "", false
		}
		return p, true
	})
	if len(parts) == 0 {
		// Root is all we need.
		return r.data, true
	}
	n, ok := r.data.isNode()
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

// process tells the root node to start processing.
func (r *root) process(path []string, processor ValueProcessor) error {
	return r.data.process(path, r.separator, processor)
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

	// ensureNoderAt ensures that the passed parts
	// exist as noder.
	ensureNoderAt(parts ...string) (noder, error)

	// setValue sets the value of the node.
	setValue(value interface{}) error

	// value returns the value of a node.
	value() interface{}

	// rawValue returns the raw value of a node for marshalling.
	rawValue() interface{}

	// process processes one leaf or node.
	process(path []string, separator string, processor ValueProcessor) error
}

// leaf represents a leaf in a JSON document tree.It
// contains the value.
type leaf struct {
	raw interface{}
}

// isNode implements noder.
func (l *leaf) isNode() (node, bool) {
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

// node represents one JSON object or array.
type node map[string]noder

// isNode implements noder.
func (n node) isNode() (node, bool) {
	return n, true
}

// ensureNoderAt ensures the existing of a leaf noder
// based on the path parts.
func (n node) ensureNoderAt(parts ...string) (noder, error) {
	switch len(parts) {
	case 0:
		// Addressing this node.
		return nil, errors.New(ErrNodeToLeaf, errorMessages)
	case 1:
		// Last part.
		pnoder := n[parts[0]]
		if pnoder == nil {
			n[parts[0]] = &leaf{}
		}
		return n[parts[0]], nil
	default:
		// More to come.
		pnoder := n[parts[0]]
		if pnoder == nil {
			n[parts[0]] = node{}
		}
		return n[parts[0]].ensureNoderAt(parts[1:]...)
	}
}

// setValue implements noder.
func (n node) setValue(value interface{}) error {
	return errors.New(ErrCorrupting, errorMessages)
}

// value implements noder.
func (n node) value() interface{} {
	return n
}

// rawValue implements noder.
func (n node) rawValue() interface{} {
	isArray := func() (int, bool) {
		max := -1
		for key := range n {
			idx, err := strconv.Atoi(key)
			if err != nil {
				return 0, false
			}
			if idx > max {
				max = idx
			}
		}
		return max, true
	}
	// Check for array or object.
	max, ok := isArray()
	if ok {
		raw := []interface{}{}
		for i := 0; i <= max; i++ {
			key := strconv.Itoa(i)
			value, ok := n[key]
			if ok {
				// Value is set.
				raw = append(raw, value.rawValue())
			} else {
				// Value is unset.
				raw = append(raw, nil)
			}
		}
		return raw
	}
	// Standard json object.
	raw := map[string]interface{}{}
	for key, value := range n {
		raw[key] = value.rawValue()
	}
	return raw
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
