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

// NewChainProcessor creates a processor chaning the passed processors.
func NewChainProcessor(processors ...Processor) ProcessorFunc {
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

// NewSwitchProcessor creates a processor taking the first processor function
// for creating a temporary result and the decision if to pass it to the
// trueBranch or the falseBranch.
func NewSwitchProcessor(decider, trueBranch, falseBranch ProcessorFunc) ProcessorFunc {
	return func(in string) (string, bool) {
		temp, ok := decider.Process(in)
		if ok {
			return trueBranch.Process(temp)
		}
		return falseBranch.Process(temp)
	}
}

// NewLoopProcessor creates a processor letting the processor function
// work on the input until it returns false (aka while true). Itself then
// will return the processed sting and always true.
func NewLoopProcessor(processor Processor) ProcessorFunc {
	return func(in string) (string, bool) {
		temp, ok := processor.Process(in)
		for ok {
			temp, ok = processor.Process(temp)
		}
		return temp, true
	}
}

// NewSplitMapProcessor creates a processor splitting the input and
// mapping the parts.
func NewSplitMapProcessor(sep string, mapper Processor) ProcessorFunc {
	return func(in string) (string, bool) {
		parts := strings.Split(in, sep)
		out := []string{}
		for _, part := range parts {
			if mp, ok := mapper.Process(part); ok {
				out = append(out, mp)
			}
		}
		return strings.Join(out, sep), true
	}
}

// NewTrimPrefixProcessor returns a processor trimming a prefix of
// the input as long as it can find it.
func NewTrimPrefixProcessor(prefix string) ProcessorFunc {
	prefixTrimmer := func(in string) (string, bool) {
		out := strings.TrimPrefix(in, prefix)
		return out, out != in
	}
	return NewLoopProcessor(ProcessorFunc(prefixTrimmer))
}

// NewTrimSuffixProcessor returns a processor trimming a prefix of
// the input as long as it can find it.
func NewTrimSuffixProcessor(prefix string) ProcessorFunc {
	suffixTrimmer := func(in string) (string, bool) {
		out := strings.TrimSuffix(in, prefix)
		return out, out != in
	}
	return NewLoopProcessor(ProcessorFunc(suffixTrimmer))
}

// EOF
