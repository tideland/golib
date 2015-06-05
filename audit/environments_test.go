// Tideland Go Library - Audit - Unit Tests
//
// Copyright (C) 2013-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package audit_test

//--------------------
// IMPORTS
//--------------------

import (
	"os"
	"testing"

	"github.com/tideland/golib/audit"
)

//--------------------
// TESTS
//--------------------

// TestTempDirCreate tests the creation of temporary directories.
func TestTempDirCreate(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	testDir := func(dir string) {
		fi, err := os.Stat(dir)
		assert.Nil(err)
		assert.True(fi.IsDir())
		assert.Equal(fi.Mode().Perm(), os.FileMode(0700))
	}

	td := audit.NewTempDir(assert)
	assert.NotNil(td)
	defer td.Restore()

	tds := td.String()
	assert.NotEmpty(tds)
	testDir(tds)

	sda := td.Mkdir("subdir", "foo")
	assert.NotEmpty(sda)
	testDir(sda)
	sdb := td.Mkdir("subdir", "bar")
	assert.NotEmpty(sdb)
	testDir(sdb)
}

// TestTempDirRestore tests the restoring of temporary created
// directories.
func TestTempDirRestore(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	td := audit.NewTempDir(assert)
	assert.NotNil(td)
	tds := td.String()
	fi, err := os.Stat(tds)
	assert.Nil(err)
	assert.True(fi.IsDir())

	td.Restore()
	fi, err = os.Stat(tds)
	assert.ErrorMatch(err, "stat .* no such file or directory")
}

// TestEnvVarsSet tests the setting of temporary environment variables.
func TestEnvVarsSet(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	testEnv := func(key, value string) {
		v := os.Getenv(key)
		assert.Equal(v, value)
	}

	ev := audit.NewEnvVars(assert)
	assert.NotNil(ev)
	defer ev.Restore()

	ev.Set("TESTING_ENV_A", "FOO")
	testEnv("TESTING_ENV_A", "FOO")
	ev.Set("TESTING_ENV_B", "BAR")
	testEnv("TESTING_ENV_B", "BAR")

	ev.Unset("TESTING_ENV_A")
	testEnv("TESTING_ENV_A", "")
}

// TestEnvVarsREstore tests the restoring of temporary set environment
// variables.
func TestEnvVarsRestore(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	testEnv := func(key, value string) {
		v := os.Getenv(key)
		assert.Equal(v, value)
	}

	ev := audit.NewEnvVars(assert)
	assert.NotNil(ev)

	path := os.Getenv("PATH")
	assert.NotEmpty(path)

	ev.Set("PATH", "/foo:/bar/bin")
	testEnv("PATH", "/foo:/bar/bin")
	ev.Set("PATH", "/bar:/foo:/yadda/bin")
	testEnv("PATH", "/bar:/foo:/yadda/bin")

	ev.Restore()

	testEnv("PATH", path)
}

// EOF
