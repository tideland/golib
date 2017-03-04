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
	"sync"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/cache"
	"github.com/tideland/golib/errors"
)

//--------------------
// CONSTANTS
//--------------------

const (
	idErrorDuringLoading    = "/error/during/loading"
	idErrorDuringReloading  = "/error/during/reloading"
	idErrorDuringCheck      = "/error/during/check"
	idErrorDuringDiscarding = "/error/during/discarding"
	idTimeout               = "/timeout"
	idIsOutdated            = "/is/outdated"
	idValidCacheable        = "/valid/cacheable"
	idSuccessfulDiscarding  = "/successful/discarding"
	idConcurrent            = "/concurrent"
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

	cacheable, err := c.Load(idErrorDuringLoading, time.Second)
	assert.Nil(cacheable)
	assert.ErrorMatch(err, ".*error during loading.*")

	cacheable, err = c.Load(idTimeout, 10*time.Millisecond)
	assert.Nil(cacheable)
	assert.ErrorMatch(err, ".*timeout.*")

	now := time.Now()
	cacheable, err = c.Load(idValidCacheable, time.Second)
	first := time.Now().Sub(now)
	assert.Nil(err)
	assert.Equal(cacheable.ID(), idValidCacheable)
	now = time.Now()
	cacheable, err = c.Load(idValidCacheable, time.Second)
	second := time.Now().Sub(now)
	assert.Nil(err)
	assert.Equal(cacheable.ID(), idValidCacheable)
	assert.True(second < first)
}

// TestConcurrentLoading tests the concurrent loading.
func TestConcurrentLoading(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	c, err := cache.New(cache.ID("concurrent-loading"), cache.Loader(testCacheableLoader))
	assert.Nil(err)
	defer c.Stop()

	var wg sync.WaitGroup

	for i := 9; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cacheable, err := c.Load(idConcurrent, time.Second)
			assert.Nil(err)
			assert.Equal(cacheable.ID(), idConcurrent)
			assert.Logf("Goroutine %d loaded %q", n, cacheable.ID())
		}(i)
	}
	wg.Wait()
}

// TestOutdating tests the outdating of Cacheables.
func TestOutdating(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	c, err := cache.New(cache.ID("outdating"), cache.Loader(testCacheableLoader))
	assert.Nil(err)
	defer c.Stop()

	// Test if outdate check fails.
	cacheable, err := c.Load(idErrorDuringCheck, time.Second)
	assert.Nil(err)
	assert.Equal(cacheable.ID(), idErrorDuringCheck)

	cacheable, err = c.Load(idErrorDuringCheck, time.Second)
	assert.Nil(cacheable)
	assert.ErrorMatch(err, ".*error during check if '/error/during/check' is outdated.*")

	// Test reload when outdated.
	for i := 1; i < 6; i++ {
		_, err := c.Load(idIsOutdated, time.Second)
		assert.Nil(err)
		assert.Equal(loaded[idIsOutdated], i)
	}
	_, err = c.Load(idIsOutdated, time.Second)
	assert.Nil(err)
	assert.Equal(loaded[idIsOutdated], 1)

	// Test error during reload.
	for i := 1; i < 6; i++ {
		_, err := c.Load(idErrorDuringReloading, time.Second)
		assert.Nil(err)
		assert.Equal(loaded[idErrorDuringReloading], i)
	}
	_, err = c.Load(idErrorDuringReloading, time.Second)
	assert.ErrorMatch(err, ".*error during loading.*")
}

// TestDiscarding tests the discarding of Cacheables.
func TestDiscarding(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	c, err := cache.New(cache.ID("discarding"), cache.Loader(testCacheableLoader))
	assert.Nil(err)
	defer c.Stop()

	// Test successful discarding, multiple times ok.
	cacheable, err := c.Load(idSuccessfulDiscarding, time.Second)
	assert.Nil(err)
	assert.Equal(cacheable.ID(), idSuccessfulDiscarding)

	err = c.Discard(idSuccessfulDiscarding)
	assert.Nil(err)

	err = c.Discard(idSuccessfulDiscarding)
	assert.Nil(err)

	// And now discarding with error.
	cacheable, err = c.Load(idErrorDuringDiscarding, time.Second)
	assert.Nil(err)
	assert.NotNil(cacheable)

	err = c.Discard(idErrorDuringDiscarding)
	assert.True(errors.IsError(err, cache.ErrDiscard))

	// Discard while several ones are waiting.
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cacheable, err := c.Load(idTimeout, time.Second)
			if err == nil {
				assert.Equal(cacheable.ID(), idTimeout)
			} else {
				assert.True(errors.IsError(err, cache.ErrDiscardedWhileLoading))
			}
		}()
	}
	err = c.Discard(idTimeout)
	assert.Nil(err)
	wg.Wait()
}

//--------------------
// HELPERS
//--------------------

const (
	errLoading = iota + 1
	errIsOutdated
	errDiscarding
	errDoubleDiscarding
)

var errorMessages = errors.Messages{
	errLoading:          "error during loading",
	errIsOutdated:       "error during check if '%s' is outdated",
	errDiscarding:       "error during discarding of '%s'",
	errDoubleDiscarding: "cacheable '%s' double discarded",
}

var (
	mutex    sync.Mutex
	loaded   = map[string]int{}
	reloaded = map[string]bool{}
)

// testCacheableLoader loads the testCacheable.
func testCacheableLoader(id string) (cache.Cacheable, error) {
	switch id {
	case idErrorDuringLoading:
		return nil, errors.New(errLoading, errorMessages)
	case idTimeout:
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(50 * time.Millisecond)
	mutex.Lock()
	loaded[id] = 1
	mutex.Unlock()
	if id == idErrorDuringReloading && reloaded[id] {
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
	case idErrorDuringCheck:
		return false, errors.New(errIsOutdated, errorMessages, tc.id)
	case idErrorDuringReloading:
		mutex.Lock()
		loaded[tc.id]++
		mutex.Unlock()
		if loaded[tc.id] == 5 {
			mutex.Lock()
			reloaded[tc.id] = true
			mutex.Unlock()
			return true, nil
		}
	case idIsOutdated:
		mutex.Lock()
		loaded[tc.id]++
		mutex.Unlock()
		if loaded[tc.id] == 5 {
			return true, nil
		}
	}
	return false, nil
}

// Discard tells the Cacheable to clean up itself.
func (tc *testCacheable) Discard() error {
	if tc.id == idErrorDuringDiscarding {
		return errors.New(errDiscarding, errorMessages, tc.id)
	}
	if tc.discarded {
		return errors.New(errDoubleDiscarding, errorMessages, tc.id)
	}
	tc.discarded = true
	return nil
}

// EOF
