// Tideland Go Library - Collections - Stacks - Unit Tests
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

// TestStackPushPop tests the core stack methods.
func TestStackPushPop(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	// Start with an empty stack.
	sa := collections.NewStack()
	assert.Length(sa, 0)
	sa.Push("foo")
	sa.Push(4711)
	assert.Length(sa, 2)
	sa.Push()
	assert.Length(sa, 2)
	sa.Push(false, 8.15)
	assert.Length(sa, 4)
	v, err := sa.Peek()
	assert.Nil(err)
	assert.Equal(v, 8.15)
	v, err = sa.Pop()
	assert.Nil(err)
	assert.Equal(v, 8.15)
	assert.Length(sa, 3)

	// Start with a filled stack.
	sb := collections.NewStack("a", true, 4711)
	assert.Length(sb, 3)
	v, err = sb.Pop()
	assert.Nil(err)
	assert.Equal(v, 4711)
	assert.Length(sb, 2)
	v, err = sb.Pop()
	assert.Nil(err)
	assert.Equal(v, true)
	assert.Length(sb, 1)
	v, err = sb.Pop()
	assert.Nil(err)
	assert.Equal(v, "a")
	assert.Length(sb, 0)

	// Popping the last one returns an error.
	v, err = sb.Pop()
	assert.ErrorMatch(err, ".*collection is empty")

	// And now deflate the first one.
	sa.Deflate()
	assert.Length(sa, 0)
}

// TestStackAll tests the retrieval of all values.
func TestStackAll(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	s := collections.NewStack(1, "b", 3.0, true)
	all := s.All()
	assert.Equal(all, []interface{}{1, "b", 3.0, true})
	all = s.AllReverse()
	assert.Equal(all, []interface{}{true, 3.0, "b", 1})
}

// TestStringStackPushPop tests the core string stack methods.
func TestStringStackPushPop(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	// Start with an empty stack.
	sa := collections.NewStringStack()
	assert.Length(sa, 0)
	sa.Push("foo")
	sa.Push("bar")
	assert.Length(sa, 2)
	sa.Push()
	assert.Length(sa, 2)
	sa.Push("baz", "yadda")
	assert.Length(sa, 4)
	v, err := sa.Peek()
	assert.Nil(err)
	assert.Equal(v, "yadda")
	v, err = sa.Pop()
	assert.Nil(err)
	assert.Equal(v, "yadda")
	assert.Length(sa, 3)

	// Start with a filled stack.
	sb := collections.NewStringStack("a", "b", "c")
	assert.Length(sb, 3)
	v, err = sb.Pop()
	assert.Nil(err)
	assert.Equal(v, "c")
	assert.Length(sb, 2)
	v, err = sb.Pop()
	assert.Nil(err)
	assert.Equal(v, "b")
	assert.Length(sb, 1)
	v, err = sb.Pop()
	assert.Nil(err)
	assert.Equal(v, "a")
	assert.Length(sb, 0)

	// Popping the last one returns an error.
	v, err = sb.Pop()
	assert.ErrorMatch(err, ".*collection is empty")

	// And now deflate the first one.
	sa.Deflate()
	assert.Length(sa, 0)
}

// EOF
