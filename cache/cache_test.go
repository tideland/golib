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

// EOF
