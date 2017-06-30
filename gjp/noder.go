// Tideland Go Library - Generic JSON Parser - Noder
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
// NODER
//--------------------

// noder defines common interface of node and leaf.
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

//--------------------
// LEAF
//--------------------

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

//--------------------
// NODE
//--------------------

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
