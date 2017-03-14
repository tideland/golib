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
	"io/ioutil"
	"testing"

	"path/filepath"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/cache"
)

//--------------------
// TESTS
//--------------------

// TestFileLoader tests the loading of files.
func TestFileLoader(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	td := audit.NewTempDir(assert)
	defer td.Restore()

	createFile(assert, td.String(), "fa", 1000)
	createFile(assert, td.String(), "fb", 1500)
	createFile(assert, td.String(), "fc", 1999)
	createFile(assert, td.String(), "fd", 2000)
	createFile(assert, td.String(), "fe", 4000)

	loader := cache.NewFileLoader(td.String(), 2000)

	ca, err := loader("fa")
	assert.Nil(err)
	assert.Equal(ca.ID(), "fa")
	fca, ok := ca.(cache.FileCacheable)
	assert.True(ok)
	da := make([]byte, 3000)
	n, err := fca.Read(da)
	assert.Nil(err)
	assert.Equal(n, 1000)
}

//--------------------
// HEKPERS
//--------------------

// createFile creates a file for loader tests.
func createFile(assert audit.Assertion, dir, name string, size int) string {
	fn := filepath.Join(dir, name)
	data := []byte{}
	for i := 0; i < size; i++ {
		data = append(data, 'X')
	}
	err := ioutil.WriteFile(fn, []byte(data), 0644)
	assert.Nil(err)
	return fn
}

// EOF
