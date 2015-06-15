// Tideland Go Library - Cells
//
// Copyright (C) 2010-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/identifier"
)

//--------------------
// CONST
//--------------------

// defaultBufferSize is the minimum size of the event buffer per cell.
const defaultBufferSize = 256

//--------------------
// OPTIONS
//--------------------

// Option allows to set an option of the environment.
type Option func(env Environment)

// Options is a set of options.
type Options []Option

// ID is the option to set the ID of the environment.
func ID(id string) Option {
	return func(env Environment) {
		e := env.(*environment)
		if id == "" {
			e.id = identifier.NewUUID().String()
		} else {
			e.id = id
		}
	}
}

// BufferSize is the option to set the event buffer size for each cell.
func BufferSize(size int) Option {
	return func(env Environment) {
		e := env.(*environment)
		if size < defaultBufferSize {
			size = defaultBufferSize
		}
		e.bufferSize = size
	}
}

// EOF
