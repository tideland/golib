// Tideland Go Library - Collections - Ring Buffer - Unit Tests
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package collections_test

//--------------------
// IMPORTS
//--------------------

import (
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/collections"
)

//--------------------
// TESTS
//--------------------

// TestRingBufferPush tests the pushing of values.
func TestRingBufferPush(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	rb := collections.NewRingBuffer(0)
	assert.Equal(rb.Cap(), 2)
	assert.Length(rb, 0)

	rb = collections.NewRingBuffer(10)
	assert.Equal(rb.Cap(), 10)
	assert.Length(rb, 0)

	rb.Push(1, "alpha", nil, true)
	assert.Length(rb, 4)
	assert.Equal(rb.String(), "[1]->[alpha]->[<nil>]->[true]")
}

// TestRingBufferPop tests the popping of values.
func TestRingBufferPop(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	rb := collections.NewRingBuffer(10)
	assert.Equal(rb.Cap(), 10)
	assert.Length(rb, 0)

	rb.Push(1, "alpha", nil, true)
	assert.Length(rb, 4)
	assert.Equal(rb.String(), "[1]->[alpha]->[<nil>]->[true]")

	tests := []struct{
		value interface{}
		ok bool
		length int
	}{
		{1, true, 3},
		{"alpha", true, 2},
		{nil, true, 1},
		{true, true, 0},
		{nil, false, 0},
	}
	for _, test := range tests {
		v, ok := rb.Pop()
		assert.Equal(v, test.value)
		assert.Equal(ok, test.ok)
		assert.Length(rb, test.length)
	}
	rb.Push(2, "beta")
	assert.Equal(rb.Cap(), 10)
	assert.Length(rb, 2)
}

// TestRingBufferGrow tests the growing of the ring buffer.
func TestRingBufferGrow(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	rb := collections.NewRingBuffer(4)
	assert.Equal(rb.Cap(), 4)
	assert.Length(rb, 0)

	rb.Push(1, 2, 3, 4, 5, 6, 7, 8)
	assert.Equal(rb.Cap(), 8)
	assert.Length(rb, 8)

	rb.Pop()
	rb.Pop()

	assert.Equal(rb.Cap(), 8)
	assert.Length(rb, 6)
}

// EOF