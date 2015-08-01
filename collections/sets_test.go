// Tideland Go Library - Collections - Sets - Unit Tests
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
	"errors"
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/collections"
)

//--------------------
// TESTS
//--------------------

// TestSetsAddRemove tests the core set methods.
func TestSetsAddRemove(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	set := collections.NewSet("foo", 42, true)
	assert.Length(set, 3)
	set.Add("foo", "bar", 123)
	assert.Length(set, 5)
	all := set.All()
	assert.Length(all, 5)
	set.Remove("yadda")
	assert.Length(set, 5)
	set.Remove("bar", 42)
	assert.Length(set, 3)
	set.Remove(false, "foo")
	assert.Length(set, 2)
	set.Deflate()
	assert.Length(set, 0)
}

// TestSetsFindAll tests the finding of set values.
func TestSetsFindAll(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	set := collections.NewSet("foo", "bar", 42, true, "yadda", 12345)
	vs, err := set.FindAll(func(v interface{}) (bool, error) {
		switch v.(type) {
		case string:
			return true, nil
		default:
			return false, nil
		}
	})
	assert.Nil(err)
	assert.Length(vs, 3)

	vs, err = set.FindAll(func(v interface{}) (bool, error) {
		return false, errors.New("ouch")
	})
	assert.ErrorMatch(err, ".* cannot find all matching values: ouch")
	assert.Length(vs, 0)
}

// TestSetsDoAll tests the iteration over set values.
func TestSetsDoAll(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	set := collections.NewSet("foo", "bar", 42, true, "yadda", 12345)
	sl := 0
	err := set.DoAll(func(v interface{}) error {
		if s, ok := v.(string); ok {
			sl += len(s)
		}
		return nil
	})
	assert.Nil(err)
	assert.Equal(sl, 11)

	err = set.DoAll(func(v interface{}) error {
		return errors.New("ouch")
	})
	assert.ErrorMatch(err, ".* cannot perform function on all values: ouch")
}

// TestStringSetsAddRemove tests the core set methods.
func TestStringSetsAddRemove(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	set := collections.NewStringSet("foo", "42", "true")
	assert.Length(set, 3)
	set.Add("foo", "bar", "123")
	assert.Length(set, 5)
	all := set.All()
	assert.Length(all, 5)
	set.Remove("yadda")
	assert.Length(set, 5)
	set.Remove("bar", "42")
	assert.Length(set, 3)
	set.Remove("false", "foo")
	assert.Length(set, 2)
	set.Deflate()
	assert.Length(set, 0)
}

// TestStringSetsFindAll tests the finding of set values.
func TestStringSetsFindAll(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	set := collections.NewStringSet("foo", "bar", "42", "true", "yadda", "12345")
	vs, err := set.FindAll(func(v string) (bool, error) {
		return len(v) == 3, nil
	})
	assert.Nil(err)
	assert.Length(vs, 2)

	vs, err = set.FindAll(func(v string) (bool, error) {
		return false, errors.New("ouch")
	})
	assert.ErrorMatch(err, ".* cannot find all matching values: ouch")
	assert.Length(vs, 0)
}

// TestStringSetsDoAll tests the iteration over set values.
func TestStringSetsDoAll(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	set := collections.NewStringSet("foo", "bar", "42", "true", "yadda", "12345")
	sl := 0
	err := set.DoAll(func(v string) error {
		sl += len(v)
		return nil
	})
	assert.Nil(err)
	assert.Equal(sl, 22)

	err = set.DoAll(func(v string) error {
		return errors.New("ouch")
	})
	assert.ErrorMatch(err, ".* cannot perform function on all values: ouch")
}

// EOF
