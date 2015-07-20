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
	"github.com/tideland/golib/stringex"
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
	assert.ErrorMatch(err, `*. illegal source format: .* node not found`)

	source = "{config {gna 1}{gna 2}}"
	config, err = configuration.Read(strings.NewReader(source))
	assert.Nil(config)
	assert.ErrorMatch(err, `*. illegal source format: .* cannot build node structure: node has multiple values`)

	source = "{config {gna 1 {foo x} 2}}"
	config, err = configuration.Read(strings.NewReader(source))
	assert.Nil(config)
	assert.ErrorMatch(err, `*. illegal source format: .* cannot build node structure: node has multiple values`)

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

	keys, err := config.At().List()
	assert.Nil(err)
	assert.Length(keys, 4)
	assert.Equal(keys, []string{"a", "b", "c", "sub"})

	keys, err = config.At("sub").List()
	assert.Nil(err)
	assert.Length(keys, 2)
	assert.Equal(keys, []string{"a", "b"})

	keys, err = config.At("sub", "a").List()
	assert.Nil(err)
	assert.Length(keys, 0)

	_, err = config.At("x").List()
	assert.ErrorMatch(err, `.* invalid configuration path "/config/x"`)
}

// TestValueSuccess tests the successful retrieval of values.
func TestValueSuccess(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := `{config
	{a  Hello}
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

	v, err := config.At("a").Value()
	assert.Nil(err)
	assert.Equal(v, "Hello")
	v, err = config.At("b").Value()
	assert.Nil(err)
	assert.Equal(v, "true")
	v, err = config.At("c").Value()
	assert.Nil(err)
	assert.Equal(v, "-1")
	v, err = config.At("d").Value()
	assert.Nil(err)
	assert.Equal(v, "47.11")
	v, err = config.At("sub", "a").Value()
	assert.Nil(err)
	assert.Equal(v, "World")
	v, err = config.At("sub", "b").Value()
	assert.Nil(err)
	assert.Equal(v, "42")

	d := stringex.NewDefaulter("config", true)
	vi := d.AsInt(config.At("c"), 42)
	assert.Equal(vi, -1)
	vf := d.AsFloat64(config.At("d"), 12.34)
	assert.Equal(vf, 47.11)
}

// TestGetFail tests the failing retrieval of values.
func TestGetFail(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{config {a Hello}{sub {a World}}}"
	config, err := configuration.Read(strings.NewReader(source))
	assert.Nil(err)

	_, err = config.At("x").Value()
	assert.ErrorMatch(err, `.* invalid configuration path "/config/x"`)
	_, err = config.At("sub", "x").Value()
	assert.ErrorMatch(err, `.* invalid configuration path "/config/sub/x"`)
}

// TestApply tests the applying of values.
func TestApply(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{config {a Hello}{sub {a World}}}"
	config, err := configuration.ReadString(source)
	assert.Nil(err)

	applied, err := config.Apply(map[string]string{
		"sub/a": "Tester",
		"b":     "42",
	})
	assert.Nil(err)
	v, err := applied.At("sub", "a").Value()
	assert.Nil(err)
	assert.Equal(v, "Tester")
	v, err = applied.At("a").Value()
	assert.Nil(err)
	assert.Equal(v, "Hello")
	v, err = applied.At("b").Value()
	assert.Nil(err)
	assert.Equal(v, "42")

}

// EOF
