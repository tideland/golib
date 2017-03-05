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

	"fmt"

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
	idCleanup               = "/cleanup/%d"
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
	te := initEnvironment()

	c, err := cache.New(cache.ID("loader"), cache.Loader(te.loader))
	assert.Nil(err)
	assert.NotNil(c)
	err = c.Stop()
	assert.Nil(err)
}

// TestLoading tests the load method of a Cache.
func TestLoading(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	te := initEnvironment()

	c, err := cache.New(cache.ID("loading"), cache.Loader(te.loader))
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
	te := initEnvironment()

	c, err := cache.New(cache.ID("concurrent-loading"), cache.Loader(te.loader))
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

// TestOutdatingFail tests the failing outdating of Cacheables.
func TestOutdatingFail(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	te := initEnvironment()

	c, err := cache.New(cache.ID("outdating-fail"), cache.Loader(te.loader))
	assert.Nil(err)
	defer c.Stop()

	cacheable, err := c.Load(idErrorDuringCheck, time.Second)
	assert.Nil(err)
	assert.Equal(cacheable.ID(), idErrorDuringCheck)

	cacheable, err = c.Load(idErrorDuringCheck, time.Second)
	assert.Nil(cacheable)
	assert.ErrorMatch(err, ".*error during check if '/error/during/check' is outdated.*")
}

// TestOutdatingReload tests the reload of outdated Cacheables.
func TestOutdatingReload(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	te := initEnvironment()

	c, err := cache.New(cache.ID("outdating-reload"), cache.Loader(te.loader))
	assert.Nil(err)
	defer c.Stop()

	for i := 1; i < 6; i++ {
		_, err := c.Load(idIsOutdated, time.Minute)
		assert.Nil(err)
		assert.Equal(te.loaded[idIsOutdated], i)
	}
	_, err = c.Load(idIsOutdated, time.Second)
	assert.Nil(err)
	assert.Equal(te.loaded[idIsOutdated], 1)
}

// TestOutdatingReloadError tests an error during reload of
// outdated Cacheables.
func TestOutdatingReloadError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	te := initEnvironment()

	c, err := cache.New(cache.ID("outdating-reload-error"), cache.Loader(te.loader))
	assert.Nil(err)
	defer c.Stop()

	for i := 1; i < 6; i++ {
		_, err := c.Load(idErrorDuringReloading, time.Second)
		assert.Nil(err)
		assert.Equal(te.loaded[idErrorDuringReloading], i)
	}
	_, err = c.Load(idErrorDuringReloading, time.Second)
	assert.ErrorMatch(err, ".*error during loading.*")
}

// TestDiscarding tests the discarding of Cacheables.
func TestDiscarding(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	te := initEnvironment()

	c, err := cache.New(cache.ID("discarding"), cache.Loader(te.loader))
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

// TestCleanup tests the cleanup of unused Cacheables.
func TestCleanup(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	te := initEnvironment()

	c, err := cache.New(cache.ID("cleanup"), cache.Loader(te.loader),
		cache.Interval(250*time.Millisecond), cache.TTL(250*time.Millisecond))
	assert.Nil(err)
	defer c.Stop()

	// Fill the cache.
	for i := 0; i < 50; i++ {
		id := fmt.Sprintf(idCleanup, i)
		cacheable, err := c.Load(id, time.Second)
		assert.Nil(err)
		assert.Equal(cacheable.ID(), id)
	}
	firstLen := c.Len()
	time.Sleep(500 * time.Millisecond)
	secondLen := c.Len()
	assert.Logf("1st: %d > 2nd: %d", firstLen, secondLen)
	assert.True(firstLen > secondLen)
}

// TestClear tests the clearing of a Cache.
func TestClear(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	te := initEnvironment()

	c, err := cache.New(cache.ID("clear"), cache.Loader(te.loader))
	assert.Nil(err)
	defer c.Stop()

	// Fill the cache.
	for i := 0; i < 50; i++ {
		id := fmt.Sprintf(idCleanup, i)
		cacheable, err := c.Load(id, time.Second)
		assert.Nil(err)
		assert.Equal(cacheable.ID(), id)
	}
	firstLen := c.Len()
	err = c.Clear()
	assert.Nil(err)
	secondLen := c.Len()
	assert.Logf("1st: %d > 2nd: %d", firstLen, secondLen)
	assert.True(firstLen > secondLen)
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

type testEnvironment struct {
	mutex    sync.Mutex
	loaded   map[string]int
	reloaded map[string]bool
}

// initEnvironment creates a new test environment.
func initEnvironment() *testEnvironment {
	return &testEnvironment{
		loaded:   make(map[string]int),
		reloaded: make(map[string]bool),
	}
}

// loader loads the testCacheable.
func (te *testEnvironment) loader(id string) (cache.Cacheable, error) {
	switch id {
	case idErrorDuringLoading:
		return nil, errors.New(errLoading, errorMessages)
	case idTimeout:
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(50 * time.Millisecond)
	te.mutex.Lock()
	te.loaded[id] = 1
	te.mutex.Unlock()
	if id == idErrorDuringReloading && te.reloaded[id] {
		return nil, errors.New(errLoading, errorMessages)
	}
	return &testCacheable{
		te:        te,
		id:        id,
		discarded: false,
	}, nil
}

// testCacheable implements Cacheable for testing.
type testCacheable struct {
	te        *testEnvironment
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
		tc.te.mutex.Lock()
		tc.te.loaded[tc.id]++
		tc.te.mutex.Unlock()
		if tc.te.loaded[tc.id] == 5 {
			tc.te.mutex.Lock()
			tc.te.reloaded[tc.id] = true
			tc.te.mutex.Unlock()
			return true, nil
		}
	case idIsOutdated:
		tc.te.mutex.Lock()
		tc.te.loaded[tc.id]++
		tc.te.mutex.Unlock()
		if tc.te.loaded[tc.id] == 5 {
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
