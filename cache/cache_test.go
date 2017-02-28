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
	"time"

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
	err = c.Stop()
	assert.Nil(err)
}

// TestLoading tests the load method of a Cache.
func TestLoading(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	c, err := cache.New(cache.ID("loading"), cache.Loader(testCacheableLoader))
	assert.Nil(err)
	defer c.Stop()

	cacheable, err := c.Load("error-during-loading", time.Second)
	assert.Nil(cacheable)
	assert.ErrorMatch(err, ".*error during loading.*")

	cacheable, err = c.Load("timeout", 10*time.Millisecond)
	assert.Nil(cacheable)
	assert.ErrorMatch(err, ".*timeout.*")

	now := time.Now()
	cacheable, err = c.Load("valid-cacheable", time.Second)
	first := time.Now().Sub(now)
	assert.Nil(err)
	assert.Equal(cacheable.ID(), "valid-cacheable")
	now = time.Now()
	cacheable, err = c.Load("valid-cacheable", time.Second)
	second := time.Now().Sub(now)
	assert.Nil(err)
	assert.Equal(cacheable.ID(), "valid-cacheable")
	assert.True(second < first)
}

// TestOutdating tests the outdating of Cacheables.
func TestOutdating(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	c, err := cache.New(cache.ID("outdating"), cache.Loader(testCacheableLoader))
	assert.Nil(err)
	defer c.Stop()

	// Test if outdate check fails.
	cacheable, err := c.Load("error-during-check", time.Second)
	assert.Nil(err)
	assert.Equal(cacheable.ID(), "error-during-check")

	cacheable, err = c.Load("error-during-check", time.Second)
	assert.Nil(cacheable)
	assert.ErrorMatch(err, ".*error during check if 'error-during-check' is obtained.*")

	// Test reload when outdated.
	for i := 1; i < 6; i++ {
		_, err := c.Load("is-outdated", time.Second)
		assert.Nil(err)
		assert.Equal(loaded["is-outdated"], i)
	}
	_, err = c.Load("is-outdated", time.Second)
	assert.Nil(err)
	assert.Equal(loaded["is-outdated"], 1)

	// Test error during reload.
	for i := 1; i < 6; i++ {
		_, err := c.Load("error-during-reloading", time.Second)
		assert.Nil(err)
		assert.Equal(loaded["error-during-reloading"], i)
	}
	_, err = c.Load("error-during-reloading", time.Second)
	assert.ErrorMatch(err, ".*error during loading.*")
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

var loaded = map[string]int{}

var reloaded = map[string]bool{}

// testCacheableLoader loads the testCacheable.
func testCacheableLoader(id string) (cache.Cacheable, error) {
	switch id {
	case "error-during-loading":
		return nil, errors.New(errLoading, errorMessages)
	case "timeout":
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(50 * time.Millisecond)
	loaded[id] = 1
	if id == "error-during-reloading" && reloaded[id] {
		return nil, errors.New(errLoading, errorMessages)
	}
	return &testCacheable{
		id:        id,
		discarded: false,
	}, nil
}

// testCacheable implements Cacheable for testing.
type testCacheable struct {
	id        string
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
		loaded[tc.id]++
		if loaded[tc.id] == 5 {
			reloaded[tc.id] = true
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
