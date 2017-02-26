// Tideland Go Library - Cache
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cache

//--------------------
// IMPORTS
//--------------------

import (
	"time"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/loop"
)

//--------------------
// CONSTANTS
//--------------------

// bucketStatus defines the different statuses of a bucket.
type bucketStatus int

const (
	statusLoading bucketStatus = iota + 1
	statusLoaded
)

//--------------------
// CACHEABLE
//--------------------

// Cacheable defines the interface for all cacheable information.
type Cacheable interface {
	// ID returns the identifier of the information.
	ID() string

	// IsOutdated checks if their's a newer version of the Cacheable.
	IsOutdated() (bool, error)

	// Discard tells the Cacheable to clean up itself.
	Discard() error
}

//--------------------
// LOADER
//--------------------

// CacheableLoader allows the user to define a function for
// loading/reloading of cacheable instances.
type CacheableLoader func(id string) (Cacheable, error)

//--------------------
// OPTIONS
//--------------------

// Option allows to configure a Cache.
type Option func(c Cache) error

// Loader returns the option to set the loader function.
func Loader(l CacheableLoader) Option {
	return func(c Cache) error {
		switch oc := c.(type) {
		case *cache:
			oc.load = l
			return nil
		default:
			return errors.New(ErrIllegalCache, errorMessages)
		}
	}
}

//--------------------
// CACHE
//--------------------

// Cache loads and returns instances by ID and caches them in memory.
type Cache interface {
	// Load returns a Cacheable from memory or source.
	Load(id string, timeout time.Duration) (Cacheable, error)

	// Discard explicitely removes a Cacheable from Cache. Normally
	// done automatically.
	Discard(id string) error

	// Stop tells the Cache to stop working.
	Stop() error
}

// responser descibes a channel for functions returning
// the result of a call.
type responser chan func() (Cacheable, error)

// bucket contains a Cacheable and the data needed to manage it.
type bucket struct {
	cacheable Cacheable
	status    bucketStatus
	loaded    time.Time
	lastUsed  time.Time
	waiters   []responser
}

// cache implements the Cache interface.
type cache struct {
	load     CacheableLoader
	interval time.Duration
	ttl      time.Duration
	buckets  map[string]*bucket
	taskc    chan task
	backend  loop.Loop
}

// New creates a new cache.
func New(options ...Option) (Cache, error) {
	c := &cache{
		interval: time.Minute,
		ttl:      10 * time.Minute,
		buckets:  make(map[string]*bucket),
		taskc:    make(chan task),
	}
	for _, option := range options {
		if err := option(c); err != nil {
			return nil, errors.Annotate(err, ErrSettingOptions, errorMessages)
		}
	}
	if c.load == nil {
		return nil, errors.New(ErrNoLoader, errorMessages)
	}
	c.backend = loop.Go(c.backendLoop)
	return c, nil
}

// Load implements the Cache interface.
func (c *cache) Load(id string, timeout time.Duration) (Cacheable, error) {
	// Send lookup task.
	responsec := make(responser, 1)
	c.taskc <- lookupTask(id, responsec)
	// Receive response.
	select {
	case response := <-responsec:
		return response()
	case <-time.After(timeout):
		return nil, errors.New(ErrTimeout, errorMessages)
	}
}

// Discard implements the Cache interface.
func (c *cache) Discard(id string) error {
	return nil
}

// Stop implements the Cache interface.
func (c *cache) Stop() error {
	return c.backend.Stop()
}

// backendLoop runs the cache.
func (c *cache) backendLoop(l loop.Loop) error {
	// Prepare ticker for lifetime check.
	checker := time.NewTicker(c.interval)
	defer checker.Stop()
	// Run loop.
	for {
		select {
		case <-l.ShallStop():
			return nil
		case do := <-c.taskc:
			if err := do(c); err != nil {
				return err
			}
		case <-checker.C:
			if err := c.checkLifetime(); err != nil {
				return err
			}
		}
	}
}

// checkLifetime looks for too long unused Cacheables
// and removes them.
func (c *cache) checkLifetime() error {
	unused := []string{}
	now := time.Now()
	// First find old ones.
	for id, bucket := range c.buckets {
		if bucket.lastUsed.Add(c.ttl).Before(now) {
			unused = append(unused, id)
		}
	}
	// Now delete found ones.
	for _, id := range unused {
		cacheable := c.buckets[id].cacheable
		delete(c.buckets, id)
		if err := cacheable.Discard(); err != nil {
			return errors.Annotate(err, ErrDiscard, errorMessages, id)
		}
	}
	return nil
}

// EOF
