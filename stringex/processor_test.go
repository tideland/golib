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
	in := "/+++++one+++/-----two--/+-+-three-+-+/four"
	trimmer := stringex.NewTrimFuncProcessor(func(r rune) bool {
		return r == "+" || r == "-"
	})
	bracer := stringex.ProcessorFunc(func(in string) (string, bool) {
		if in == "" {
			return "", true
		}
		return "(" + in + ")", true
	})
	updater := stringex.NewChainProcessor(trimmer, bracer)
	fullUpdater := stringex.NewSplitMapProcessor("/", updater)

	value, ok := fullUpdater(in)
	assert.True(ok)
	assert.Equal(value, "/(one)/(two)/(three)/(four)")
}

// EOF
