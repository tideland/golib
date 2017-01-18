// Tideland Go Library - Audit
//
// Copyright (C) 2013-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package audit

//--------------------
// IMPORTS
//--------------------

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
)

//--------------------
// TEMPDIR
//--------------------

// TempDir represents a temporary directory and possible subdirectories
// for testing purposes. It simply is created with
//
//     assert := audit.NewTestingAssertion(t, false)
//     td := audit.NewTempDir(assert)
//     defer td.Restore()
//
//     tdName := td.String()
//     subName:= td.Mkdir("my", "sub", "directory")
//
// The deferred Restore() removes the temporary directory with all
// contents.
type TempDir struct {
	assert Assertion
	dir    string
}

// NewTempDir creates a new temporary directory usable for direct
// usage or further subdirectories.
func NewTempDir(assert Assertion) *TempDir {
	id := make([]byte, 8)
	td := &TempDir{
		assert: assert,
	}
	for i := 0; i < 256; i++ {
		_, err := rand.Read(id[:])
		td.assert.Nil(err)
		dir := filepath.Join(os.TempDir(), fmt.Sprintf("gots-%x", id))
		if err = os.Mkdir(dir, 0700); err == nil {
			td.dir = dir
			break
		}
		if td.dir == "" {
			msg := fmt.Sprintf("cannot create temporary directory %q: %v", td.dir, err)
			td.assert.Fail(msg)
			return nil
		}
	}
	return td
}

// Restore deletes the temporary directory and all contents.
func (td *TempDir) Restore() {
	err := os.RemoveAll(td.dir)
	if err != nil {
		msg := fmt.Sprintf("cannot remove temporary directory %q: %v", td.dir, err)
		td.assert.Fail(msg)
	}
}

// Mkdir creates a potentially nested directory inside the
// temporary directory.
func (td *TempDir) Mkdir(name ...string) string {
	innerName := filepath.Join(name...)
	fullName := filepath.Join(td.dir, innerName)
	if err := os.MkdirAll(fullName, 0700); err != nil {
		msg := fmt.Sprintf("cannot create nested temporary directory %q: %v", fullName, err)
		td.assert.Fail(msg)
	}
	return fullName
}

// String returns the temporary directory.
func (td *TempDir) String() string {
	return td.dir
}

//--------------------
// ENVVARS
//--------------------

// EnvVars allows to change and restore environment variables. The
// same variable can be set multiple times. Simply do
//
//     assert := audit.NewTestingAssertion(t, false)
//     ev := audit.NewEnvVars(assert)
//     defer ev.Restore()
//
//     ev.Set("MY_VAR", myValue)
//
//     ...
//
//     ev.Set("MY_VAR", anotherValue)
//
// The deferred Restore() resets to the original values.
type EnvVars struct {
	assert Assertion
	vars   map[string]string
}

// NewEnvVars creates
func NewEnvVars(assert Assertion) *EnvVars {
	ev := &EnvVars{
		assert: assert,
		vars:   make(map[string]string),
	}
	return ev
}

// Restore resets all changed environment variables
func (ev *EnvVars) Restore() {
	for key, value := range ev.vars {
		if err := os.Setenv(key, value); err != nil {
			msg := fmt.Sprintf("cannot reset environment variable %q: %v", key, err)
			ev.assert.Fail(msg)
		}
	}
}

// Set sets an environment variable to a new value.
func (ev *EnvVars) Set(key, value string) {
	v := os.Getenv(key)
	_, ok := ev.vars[key]
	if !ok {
		ev.vars[key] = v
	}
	if err := os.Setenv(key, value); err != nil {
		msg := fmt.Sprintf("cannot set environment variable %q: %v", key, err)
		ev.assert.Fail(msg)
	}
}

// Unset unsets an environment variable.
func (ev *EnvVars) Unset(key string) {
	v := os.Getenv(key)
	_, ok := ev.vars[key]
	if !ok {
		ev.vars[key] = v
	}
	if err := os.Unsetenv(key); err != nil {
		msg := fmt.Sprintf("cannot unset environment variable %q: %v", key, err)
		ev.assert.Fail(msg)
	}
}

// EOF
