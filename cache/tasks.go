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
)

//--------------------
// TASKS
//--------------------

// task contains any task a cache shall do.
type task func(c *cache) error

// cancelTask notifies the cache that a loading failed.
func cancelTask(id string, err error) task {
	return func(c *cache) error {
		for _, waiter := range c.buckets[id].waiters {
			waiter <- func() (Cacheable, error) {
				return nil, err
			}
		}
		delete(c.buckets, id)
		return nil
	}
}

// addTask notifies the cache that a loading succeeded.
func addTask(id string, cacheable Cacheable) task {
	return func(c *cache) error {
		b := c.buckets[id]
		b.cacheable = cacheable
		b.status = statusLoaded
		b.loaded = time.Now()
		b.lastUsed = b.loaded
		for _, waiter := range c.buckets[id].waiters {
			waiter <- func() (Cacheable, error) {
				return cacheable, nil
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
		c.taskc <- cancelTask(id, err)
	} else {
		c.taskc <- addTask(id, cacheable)
	}
}

// lookupTask returns the task for looking up the cache.
func lookupTask(id string, responsec responser) task {
	return func(c *cache) error {
		b, ok := c.buckets[id]
		switch {
		case !ok:
			c.buckets[id] = &bucket{
				status:  statusNew,
				waiters: []responser{responsec},
			}
			go loading(c, id)
		case ok && b.status == statusNew:
		}
		return nil
	}
}

// EOF
