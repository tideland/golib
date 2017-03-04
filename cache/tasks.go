// Tideland Go Library - Cache - Tasks
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
)

//--------------------
// TASKS
//--------------------

// task contains any task a cache shall do.
type task func(c *cache) error

// failedTask notifies the cache that a loading failed.
func failedTask(id string, err error) task {
	return func(c *cache) error {
		// Check for discarded Cacheable first.
		if c.buckets[id] == nil {
			return nil
		}
		// Notify all waiters.
		for _, waiter := range c.buckets[id].waiters {
			waiter <- func() (Cacheable, *Info, error) {
				return nil, nil, errors.Annotate(err, ErrLoading, errorMessages, id)
			}
		}
		delete(c.buckets, id)
		return nil
	}
}

// successTask notifies the cache that a loading succeeded.
func successTask(id string, cacheable Cacheable) task {
	return func(c *cache) error {
		// Check for discarded Cacheable first.
		if c.buckets[id] == nil {
			return nil
		}
		// Set bucket values.
		b := c.buckets[id]
		b.cacheable = cacheable
		b.status = statusLoaded
		b.loaded = time.Now()
		b.lastUsed = b.loaded
		// Notify all waiters.
		for _, waiter := range c.buckets[id].waiters {
			waiter <- func() (Cacheable, *Info, error) {
				return cacheable, nil, nil
			}
		}
		b.waiters = nil
		return nil
	}
}

// loading is the asynchronous loading function.
func loading(c *cache, id string) {
	cacheable, err := c.load(id)
	if err != nil {
		c.taskc <- failedTask(id, err)
	} else {
		c.taskc <- successTask(id, cacheable)
	}
}

// lookupTask returns the task for looking up the cache.
func lookupTask(id string, responsec responder) task {
	return func(c *cache) error {
		b, ok := c.buckets[id]
		switch {
		case !ok:
			// ID is unknown.
			c.buckets[id] = &bucket{
				status:  statusLoading,
				waiters: []responder{responsec},
			}
			go loading(c, id)
		case ok && b.status == statusLoading:
			// ID is known but Cacheable is not yet retrieved.
			b.waiters = append(b.waiters, responsec)
		case ok && b.status == statusLoaded:
			// ID is known and Cacheable is loaded.
			outdated, err := b.cacheable.IsOutdated()
			if err != nil {
				// Error during check if outdated.
				responsec <- func() (Cacheable, *Info, error) {
					return nil, nil, errors.Annotate(err, ErrCheckOutdated, errorMessages, id)
				}
				delete(c.buckets, id)
				return nil
			}
			if outdated {
				// Outdated, so reload.
				c.buckets[id].status = statusLoading
				c.buckets[id].waiters = []responder{responsec}
				go loading(c, id)
			}
			// Everything fine.
			b.lastUsed = time.Now()
			responsec <- func() (Cacheable, *Info, error) {
				return b.cacheable, nil, nil
			}
		}
		return nil
	}
}

// discardTask returns the task for discarding a Cacheable.
func discardTask(id string, responsec responder) task {
	return func(c *cache) error {
		b, ok := c.buckets[id]
		if !ok {
			// Not found, so nothing to discard.
			responsec <- func() (Cacheable, *Info, error) {
				return nil, nil, nil
			}
			return nil
		}
		// Discard Cacheable, notify possible waiters,
		// delete bucket, and notify caller.
		var err error
		if b.cacheable != nil {
			err = b.cacheable.Discard()
		}
		for _, waiter := range b.waiters {
			waiter <- func() (Cacheable, *Info, error) {
				return nil, nil, errors.New(ErrDiscardedWhileLoading, errorMessages, id)
			}
		}
		delete(c.buckets, id)
		if err != nil {
			err = errors.Annotate(err, ErrDiscard, errorMessages, id)
		}
		responsec <- func() (Cacheable, *Info, error) {
			return nil, nil, err
		}
		return nil
	}
}

// infoTask returns the task for gathering statistical information.
func infoTask(responsec responder) task {
	return func(c *cache) error {
	}
}

// EOF
