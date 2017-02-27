// Tideland Go Library - Cache - Unit Tests
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cache_test

//--------------------
// IMPORTS
//--------------------

import (
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/cache"
	"github.com/tideland/golib/errors"
)

//--------------------
// TESTS
//--------------------

// TestNoLoader tests the creation of a cache without
// a loader. Must lead to an error.
func TestNoLoader(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	c, err := cache.New()
	assert.Nil(c)
	assert.True(errors.IsError(err, cache.ErrNoLoader))
}

// TestLoader tests the creation of a cache with
// a loader.
func TestLoader(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	c, err := cache.New(cache.ID("loader"), cache.Loader(testCacheableLoader))
	assert.Nil(err)
	assert.NotNil(c)
}

//--------------------
// HELPERS
//--------------------

const (
	errLoading = iota + 1
	errIsObtained
	errDoubleDiscarding
)

var errorMessages = errors.Messages{
	errLoading:          "error during loading",
	errIsObtained:       "error during check if '%s' is obtained",
	errDoubleDiscarding: "cacheable '%s' double discarded",
}

// testCacheableLoader loads the testCacheable.
func testCacheableLoader(id string) (cache.Cacheable, error) {
	if id == "error-during-loading" {
		return nil, errors.New(errLoading, errorMessages)
	}
	return &testCacheable{
		id:        id,
		retrieved: 1,
		discarded: false,
	}, nil
}

// testCacheable implements Cacheable for testing.
type testCacheable struct {
	id        string
	retrieved int
	discarded bool
}

func (tc *testCacheable) ID() string {
	return tc.id
}

// IsOutdated checks if their's a newer version of the Cacheable.
func (tc *testCacheable) IsOutdated() (bool, error) {
	switch tc.id {
	case "error-during-check":
		return false, errors.New(errIsObtained, errorMessages, tc.id)
	case "is-outdated":
		tc.retrieved++
		if tc.retrieved == 5 {
			return true, nil
		}
	}
	return false, nil
}

// Discard tells the Cacheable to clean up itself.
func (tc *testCacheable) Discard() error {
	if tc.discarded {
		return errors.New(errDoubleDiscarding, errorMessages, tc.id)
	}
	tc.discarded = true
	return nil
}

// EOF
