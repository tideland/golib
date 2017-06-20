// Tideland Go Library - Generic JSON Parser - Builder
//
// Copyright (C) 2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package gjp

//--------------------
// IMPORTS
//--------------------

import ()

//--------------------
// BUILDER
//--------------------

// Builder can be used to build JSON documents step by step.
type Builder interface {
	// SetValue sets the value at the given path to
	SetValue(path string, value interface{}) error

	// Clear removes the so far build document data.
	Clear()
}

// builder implements Builder.
type builder struct {
	root *root
}

// NewBuilder creates a new builder instance.
func NewBuilder(separator string) Builder {
	return &builder{
		root: newRoot(separator, nil),
	}
}

// setValue implements Builder.
func (b *builder) SetValue(path string, value interface{}) error {
	return b.root.setValueAt(path, value)
}

// Clear implements Builder.
func (b *builder) Clear() {
	b.root = newRoot(b.root.separator, nil)
}

// EOF
