// Tideland Go Library - Cache
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// Package cache lazily loads information on demand and caches them.
// The data inside a cache has to implement the Cacheable interface
// which also contains methods for checking, if the information is
// outdated, and for discarding the cached instance. It is loaded
// by a user defined CacheableLoader function.
package cache

// EOF
