// Tideland Go Library - String Extensions - Unit Tests
//
// Copyright (C) 2015-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package stringex_test

//--------------------
// IMPORTS
//--------------------

import (
	"strings"
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/stringex"
)

//--------------------
// TESTS
//--------------------

// TestWrapping tests wrapping a standard function to a processor.
func TestWrapping(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	upperCaser := stringex.WrapProcessorFunc(strings.ToUpper)

	value, ok := upperCaser("test")
	assert.True(ok)
	assert.Equal(value, "TEST")
}

// TestSplitMapProcessor tests the splitting and mapping.
func TestSplitMapProcessor(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	sep := "the"
	upperCaser := stringex.WrapProcessorFunc(strings.ToUpper)
	splitMapper := stringex.NewSplitMapProcessor(sep, upperCaser)

	value, ok := splitMapper("the quick brown fox jumps over the lazy dog")
	assert.True(ok)
	assert.Equal(value, "the QUICK BROWN FOX JUMPS OVER the LAZY DOG")
}

// TestSubstringProcessor tests retrieving substrings.
func TestSubstringProcessor(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tests := []struct {
		index  int
		length int
		in     string
		out    string
		ok     bool
	}{
		{0, 5, "yadda", "yadda", true},
		{0, 3, "yadda", "yad", true},
		{2, 3, "yadda", "dda", true},
		{2, 5, "yadda", "dda", true},
		{-1, 5, "yadda", "yadda", true},
		{-1, 10, "yadda", "yadda", true},
		{0, 0, "yadda", "", false},
	}

	for _, test := range tests {
		substringer := stringex.NewSubstringProcessor(test.index, test.length)
		out, ok := substringer(test.in)
		assert.Equal(ok, test.ok)
		assert.Equal(out, test.out)
	}
}

// TestTrimmingProcessors tests the trimming.
func TestTrimmingProcessors(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	in := "+++++foo+++"

	// Prefix.
	plusPreTrimmer := stringex.NewTrimPrefixProcessor("+")
	plusPlusPreTrimmer := stringex.NewTrimPrefixProcessor("++")

	value, ok := plusPreTrimmer(in)
	assert.True(ok)
	assert.Equal(value, "foo+++")
	value, ok = plusPlusPreTrimmer(in)
	assert.True(ok)
	assert.Equal(value, "+foo+++")

	// Suffix.
	plusSufTrimmer := stringex.NewTrimSuffixProcessor("+")
	plusPlusSufTrimmer := stringex.NewTrimSuffixProcessor("++")

	value, ok = plusSufTrimmer(in)
	assert.True(ok)
	assert.Equal(value, "+++++foo")
	value, ok = plusPlusSufTrimmer(in)
	assert.True(ok)
	assert.Equal(value, "+++++foo+")

	// Chaining.
	plusTrimmer := stringex.NewChainProcessor(plusPreTrimmer, plusSufTrimmer)
	plusPlusTrimmer := stringex.NewChainProcessor(plusPlusPreTrimmer, plusPlusSufTrimmer)

	value, ok = plusTrimmer(in)
	assert.True(ok)
	assert.Equal(value, "foo")
	value, ok = plusPlusTrimmer(in)
	assert.True(ok)
	assert.Equal(value, "+foo+")
}

// TestScenario tests the combination of multiple processors.
func TestScenario(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	in := "+++++yadda+++/-----foobar--/+-+-testing-+-+/out"
	trimmer := stringex.NewTrimFuncProcessor(func(r rune) bool {
		return r == '+' || r == '-'
	})
	substringer := stringex.NewSubstringProcessor(0, 4)
	omatcher := stringex.NewMatchProcessor("o+")
	uppercaser := stringex.WrapProcessorFunc(strings.ToUpper)
	bracer := stringex.ProcessorFunc(func(in string) (string, bool) {
		return "(" + in + ")", true
	})
	updater := stringex.NewChainProcessor(trimmer, substringer, omatcher, uppercaser, bracer)
	allUpdater := stringex.NewSplitMapProcessor("/", updater)
	replacer := stringex.ProcessorFunc(func(in string) (string, bool) {
		return strings.Replace(in, "/", "::", -1), true
	})
	fullUpdater := stringex.NewChainProcessor(allUpdater, replacer)

	value, ok := fullUpdater(in)
	assert.True(ok)
	assert.Equal(value, "(FOOB)::(OUT)")
}

// EOF
