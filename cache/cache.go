// Tideland Go Library - Cache
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cache

//--------------------
// IMPORTS
//--------------------

import (
	"sync"
	"time"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/loop"
)

//--------------------
// CONSTANTS
//--------------------

const (
	ErrCannotRetrieve = iota + 1
)

var errorMessages = errors.Messages{
	ErrCannotRetrieve: "cannot retrieve cached value: %v",
}

//--------------------
// CACHE MANAGER
//--------------------

type cacheManagerChange struct {
	register bool
	value    *cachedValue
}

// cacheManager stores references and sets the values
// to nil periodically.
type cacheManager struct {
	values  []*cachedValue
	changec chan *cacheManagerChange
	loop    loop.Loop
}

// newCacheManager creates a new manager.
func newCacheManager() *cacheManager {
	m := &cacheManager{
		values:  []*cachedValue{},
		changec: make(chan *cacheManagerChange),
	}
	m.loop = loop.Go(m.backendLoop)
	return m
}

// register a cached value.
func (m *cacheManager) register(v *cachedValue) {
	m.changec <- &cacheManagerChange{true, v}
}

// unRegister a cached value.
func (m *cacheManager) unregister(v *cachedValue) {
	m.changec <- &cacheManagerChange{false, v}
}

// backendLoop processing the changes and the cleanings.
func (m *cacheManager) backendLoop(l loop.Loop) error {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-l.ShallStop():
			return nil
		case c := <-m.changec:
			if c.register {
				m.doRegister(c.value)
			} else {
				m.doUnregister(c.value)
			}
		case <-ticker.C:
			m.doCleaning()
		}
	}
}

// doRegister performs the registration.
func (m *cacheManager) doRegister(v *cachedValue) {
	// Look for a free space.
	for i := range m.values {
		if m.values[i] == nil {
			v.id = i
			m.values[i] = v
			return
		}
	}
	// None found, append.
	i := len(m.values)
	v.id = i
	m.values = append(m.values, v)
}

// doUnregister performs the unregistration.
func (m *cacheManager) doUnregister(v *cachedValue) {
	m.values[v.id] = nil
}

// doCleaning cleans the cached values.
func (m *cacheManager) doCleaning() {
	now := time.Now()
	for _, v := range m.values {
		if v != nil {
			v.checkCleaning(now)
		}
	}
}

// cache is the only instance of the cache manager.
var cache = newCacheManager()

//--------------------
// CACHED VALUE
//--------------------

type CachedValue interface {
	// Value returns the cached value. If an error occurred
	// during retrieval that will be returned too.
	Value() (v interface{}, err error)

	// Clear clears the cached value so that it will be
	// retrieved again when Value() is called the next time.
	Clear()

	// Remove removes this cached value from the cache.
	Remove()
}

// RetrievalFunc is the signature of a function responsible for the retrieval
// of the cached value from somewhere else in the system, e.g. a database.
type RetrievalFunc func() (interface{}, error)

// cachedValue implements the CachedValue interface.
type cachedValue struct {
	mux           sync.Mutex
	id            int
	value         interface{}
	retrievalFunc RetrievalFunc
	ttl           time.Duration
	lastAccess    time.Time
}

// NewCachedValue creates a new cache. The retrieval func is
// responsible for the retrieval of the value while ttl defines
// how long the value is valid.
func NewCachedValue(r RetrievalFunc, ttl time.Duration) CachedValue {
	v := &cachedValue{
		retrievalFunc: r,
		ttl:           ttl,
		lastAccess:    time.Now(),
	}
	cache.register(v)
	return v
}

// Value implements the CachedValue interface.
func (v *cachedValue) Value() (value interface{}, err error) {
	v.mux.Lock()
	defer v.mux.Unlock()
	defer func() {
		if r := recover(); r != nil {
			value = nil
			err = errors.New(ErrCannotRetrieve, errorMessages, r)
		}
	}()
	if v.value != nil {
		if time.Now().Sub(v.lastAccess) > v.ttl {
			v.value = nil
		}
	}
	if v.value == nil {
		if v.value, err = v.retrievalFunc(); err != nil {
			v.value = nil
			return nil, err
		}
	}
	v.lastAccess = time.Now()
	return v.value, nil
}

// Clear implements the CachedValue interface.
func (v *cachedValue) Clear() {
	v.mux.Lock()
	defer v.mux.Unlock()
	v.value = nil
}

// Remove implements the CachedValue interface.
func (v *cachedValue) Remove() {
	v.mux.Lock()
	defer v.mux.Unlock()
	cache.unregister(v)
	v.value = nil
	v.retrievalFunc = nil
}

// checkCleaning checks if the timespan between now and the last
// access is largen than the time to live. In this case the value
// is cleared.
func (v *cachedValue) checkCleaning(now time.Time) {
	v.mux.Lock()
	defer v.mux.Unlock()
	if now.Sub(v.lastAccess) > v.ttl {
		v.value = nil
	}
}

// EOF
