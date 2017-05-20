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
	upperProcessor := stringex.WrapProcessorFunc(strings.ToUpper)

	value, ok := upperProcessor("test")
	assert.True(ok)
	assert.Equal(value, "TEST")
}

// TestSplitMapProcessor tests the splitting and mapping.
func TestSplitMapProcessor(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	sep := "the"
	upperProcessor := stringex.WrapProcessorFunc(strings.ToUpper)
	splitMapProcessor := stringex.NewSplitMapProcessor(sep, upperProcessor)

	value, ok := splitMapProcessor("the quick brown fox jumps over the lazy dog")
	assert.True(ok)
	assert.Equal(value, "the QUICK BROWN FOX JUMPS OVER the LAZY DOG")
}

//--------------------
// HELPERS
//--------------------

// EOF
