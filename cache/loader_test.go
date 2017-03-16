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

	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/cache"
	"github.com/tideland/golib/monitoring"
)

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

	mega := 1024 * 1024
	loader := cache.NewFileLoader(td.String(), int64(3*mega))
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
			m := monitoring.BeginMeasuring(test.name)
			c, err := loader(test.name)
			assert.Nil(err)
			assert.Equal(c.ID(), test.name)
			fc, ok := c.(cache.FileCacheable)
			assert.True(ok)
			p := make([]byte, test.size*mega)
			rc, err := fc.ReadCloser()
			assert.Nil(err)
			n, err := rc.Read(p)
			assert.Nil(err)
			assert.Equal(n, test.size*mega)
			err = rc.Close()
			assert.Nil(err)
			m.EndMeasuring()
		}
	}
	time.Sleep(5 * time.Second)
	monitoring.MeasuringPointsPrintAll()
}

//--------------------
// HEKPERS
//--------------------

// createFile creates a file for loader tests.
func createFile(assert audit.Assertion, dir, name string, size int) string {
	fn := filepath.Join(dir, name)
	mega := 1024 * 1024
	data := []byte{}
	for i := 0; i < size*mega; i++ {
		data = append(data, 'X')
	}
	err := ioutil.WriteFile(fn, []byte(data), 0644)
	assert.Nil(err)
	return fn
}

// EOF
