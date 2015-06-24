// Tideland Go Library - Configuration - Unit Tests
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package configuration_test

//--------------------
// IMPORTS
//--------------------

import (
	"strings"
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/configuration"
)

//--------------------
// CONFIG
//--------------------

// TestRead tests reading a configuration out of a reader.
func TestRead(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{config {foo 42}{bar 24}}"
	config, err := configuration.Read(strings.NewReader(source))
	assert.Nil(err)

	source = "{something {gnagnagna}}"
	config, err = configuration.Read(strings.NewReader(source))
	assert.Nil(config)
	assert.ErrorMatch(err, `*. illegal source for configuration: does not start with "config" node`)

	source = "{config {gna 1}{gna 2}}"
	config, err = configuration.Read(strings.NewReader(source))
	assert.Nil(config)
	assert.ErrorMatch(err, `*. illegal source for configuration: node has multiple values`)

	source = "{config {gna 1 {foo x} 2}}"
	config, err = configuration.Read(strings.NewReader(source))
	assert.Nil(config)
	assert.ErrorMatch(err, `*. illegal source for configuration: node has multiple values`)

	source = "{config {foo/bar 1}{bar/foo 2}}"
	config, err = configuration.Read(strings.NewReader(source))
	assert.Nil(config)
	assert.ErrorMatch(err, `*. illegal source format: .*`)
}

// TestList tests the listing of configuration keys.
func TestList(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := `{config {a 1}{b 2}{c 3}{sub {a 4.1}{b 4.2}}}`
	config, err := configuration.ReadString(source)
	assert.Nil(err)

	keys, err := config.List()
	assert.Nil(err)
	assert.Length(keys, 4)
	assert.Equal(keys, []string{"a", "b", "c", "sub"})

	keys, err = config.List("sub")
	assert.Nil(err)
	assert.Length(keys, 2)
	assert.Equal(keys, []string{"a", "b"})

	keys, err = config.List("sub", "a")
	assert.Nil(err)
	assert.Length(keys, 0)

	_, err = config.List("x")
	assert.ErrorMatch(err, `.* invalid configuration path "/config/x"`)
}

// TestGetSuccess tests the successful retrieval of values.
func TestGetSuccess(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := `{config
	{a Hello}
	{b true}
	{c -1}
	{d     47.11     }
	{sub
		{a
			World}
		{b
			42}}}`
	config, err := configuration.Read(strings.NewReader(source))
	assert.Nil(err)

	vs, err := config.Get("a")
	assert.Nil(err)
	assert.Equal(vs, "Hello")
	vb, err := config.GetBool("b")
	assert.Nil(err)
	assert.Equal(vb, true)
	vi, err := config.GetInt("c")
	assert.Nil(err)
	assert.Equal(vi, -1)
	vd, err := config.GetFloat64("d")
	assert.Nil(err)
	assert.Equal(vd, 47.11)

	vs, err = config.Get("sub", "a")
	assert.Nil(err)
	assert.Equal(vs, "World")
	vi, err = config.GetInt("sub", "b")
	assert.Nil(err)
	assert.Equal(vi, 42)
}

// TestGetFail tests the failing retrieval of values.
func TestGetFail(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{config {a Hello}{sub {a World}}}"
	config, err := configuration.Read(strings.NewReader(source))
	assert.Nil(err)

	_, err = config.Get("x")
	assert.ErrorMatch(err, `.* invalid configuration path "/config/x"`)
	_, err = config.Get("sub", "x")
	assert.ErrorMatch(err, `.* invalid configuration path "/config/sub/x"`)

	vb, err := config.GetBool("a")
	assert.Equal(vb, false)
	assert.ErrorMatch(err, `.* invalid value format of "Hello": strconv.ParseBool: .*`)
	vi, err := config.GetInt("a")
	assert.Equal(vi, 0)
	assert.ErrorMatch(err, `.* invalid value format of "Hello": strconv.ParseInt: .*`)
	vf, err := config.GetFloat64("a")
	assert.Equal(vf, 0.0)
	assert.ErrorMatch(err, `.* invalid value format of "Hello": strconv.ParseFloat: .*`)
}

// EOF
