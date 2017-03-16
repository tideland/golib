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
	"path/filepath"
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/cache"
)

//--------------------
// CONSTANTS
//--------------------

const multiplier = 1024 * 1024

//--------------------
// TESTS
//--------------------

// TestFileLoader tests the loading of files.
func TestFileLoader(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	td := audit.NewTempDir(assert)
	defer td.Restore()

	createFile(assert, td.String(), "fa", 1)
	createFile(assert, td.String(), "fb", 2)
	createFile(assert, td.String(), "fc", 3)
	createFile(assert, td.String(), "fd", 4)
	createFile(assert, td.String(), "fe", 5)

	loader := cache.NewFileLoader(td.String(), int64(3*multiplier))
	tests := []struct {
		name string
		size int
	}{
		{"fa", 1},
		{"fb", 2},
		{"fc", 3},
		{"fd", 4},
		{"fe", 5},
	}
	for i, test := range tests {
		assert.Logf("test #%d: %s with size %d mb", i, test.name, test.size)
		for j := 0; j < 10; j++ {
			c, err := loader(test.name)
			assert.Nil(err)
			assert.Equal(c.ID(), test.name)
			fc, ok := c.(cache.FileCacheable)
			assert.True(ok)
			p := make([]byte, test.size*multiplier)
			rc, err := fc.ReadCloser()
			assert.Nil(err)
			n, err := rc.Read(p)
			assert.Nil(err)
			assert.Equal(n, test.size*multiplier)
			err = rc.Close()
			assert.Nil(err)
		}
	}
}

//--------------------
// HEKPERS
//--------------------

// createFile creates a file for loader tests.
func createFile(assert audit.Assertion, dir, name string, size int) string {
	fn := filepath.Join(dir, name)
	data := []byte{}
	for i := 0; i < size*multiplier; i++ {
		data = append(data, 'X')
	}
	err := ioutil.WriteFile(fn, []byte(data), 0644)
	assert.Nil(err)
	return fn
}

// EOF
