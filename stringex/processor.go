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
	"fmt"
	"regexp"
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

// errorProcessorFunc returns a processor func used in case of failing
// preparation steps.
func errorProcessorFunc(err error) ProcessorFunc {
	return func(in string) (string, bool) {
		return fmt.Sprintf("error processing '%s': %v", in, err), false
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

// NewConditionProcessor creates a processor taking the first processor
// for creating a temporary result and a decision. Based on the decision
// the temporary result is passed to an affirmer or a negater.
func NewConditionProcessor(decider, affirmer, negater Processor) ProcessorFunc {
	return func(in string) (string, bool) {
		temp, ok := decider.Process(in)
		if ok {
			return affirmer.Process(temp)
		}
		return negater.Process(temp)
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
// mapping the parts. It will only contain those where the mapper
// returns true. So it can be used as a filter too. Afterwards the
// collected mapped parts are joined again.
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

// NewSubstringProcessor returns a processor slicing the input
// based on the index and length.
func NewSubstringProcessor(index, length int) ProcessorFunc {
	return func(in string) (string, bool) {
		if length < 1 {
			return "", false
		}
		if index < 0 {
			index = 0
		}
		if index >= len(in) {
			return "", true
		}
		out := in[index:]
		if length > len(out) {
			length = len(out)
		}
		return out[:length], true
	}
}

// NewMatchProcessor returns a processor evaluating the input
// against a given pattern and returns the input and true
// when it is matching.
func NewMatchProcessor(pattern string) ProcessorFunc {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return errorProcessorFunc(err)
	}
	return func(in string) (string, bool) {
		return in, r.MatchString(in)
	}
}

// NewTrimFuncProcessor returns a processor trimming prefix and
// suffix of the input based on the return value of the passed
// function checking each rune.
func NewTrimFuncProcessor(f func(r rune) bool) ProcessorFunc {
	return func(in string) (string, bool) {
		return strings.TrimFunc(in, f), true
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

// NewUpperProcessor returns a processor converting the
// input to upper-case.
func NewUpperProcessor() ProcessorFunc {
	return WrapProcessorFunc(strings.ToUpper)
}

// NewLowerProcessor returns a processor converting the
// input to lower-case.
func NewLowerProcessor() ProcessorFunc {
	return WrapProcessorFunc(strings.ToLower)
}

// EOF
