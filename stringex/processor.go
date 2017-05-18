// Tideland Go Library - String Extensions
//
// Copyright (C) 2015-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package stringex

//--------------------
// IMPORTS
//--------------------

import (
	"strings"
)

//--------------------
// PROCESSOR
//--------------------

// Processor defines a type able to process strings.
type Processor interface {
	// Process takes a string and processes it. If the result is to
	// be ignored the bool has to be false.
	Process(in string) (string, bool)
}

//--------------------
// PROCESSOR FUNCTIONS
//--------------------

// ProcessorFunc describes functions processing a string and returning
// the new one. A returned false means to ignore the result.
type ProcessorFunc func(in string) (string, bool)

// Process implements Processor.
func (pf ProcessorFunc) Process(in string) (string, bool) {
	return pf(in)
}

// WrapProcessorFunc takes a standard string processing function and
// returns it as a ProcessorFunc.
func WrapProcessorFunc(f func(fin string) string) ProcessorFunc {
	return func(in string) (string, bool) {
		return f(in), true
	}
}

//--------------------
// FACTORIES
//--------------------

// NewProcessorChain returns a function chaning the passed processors.
func NewProcessorChain(processors ...Processor) ProcessorFunc {
	return func(in string) (string, bool) {
		out := in
		ok := true
		for _, processor := range processors {
			out, ok = processor.Process(out)
			if !ok {
				return "", false
			}
		}
		return out, ok
	}
}

// NewSplitMapProcessor creates a processor splitting the input and
// mapping the parts.
func NewSplitMapProcessor(sep string, m ProcessorFunc) Processor {
	pf := func(in string) (string, bool) {
		parts := strings.Split(in, sep)
		out := []string{}
		for _, part := range parts {
			if mp, ok := m(part); ok {
				out = append(out, mp)
			}
		}
		return strings.Join(out, sep), true
	}
	return ProcessorFunc(pf)
}

// EOF
