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
	"time"

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

// TestGetSuccess tests the successful retrieval of values.
func TestGetSuccess(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := `{config
	{a  Hello}
	{b true}
	{c -1}
	{d     47.11     }
	{e 2015-06-25T23:45:00+02:00}
	{f 2h15m30s}
	{sub
		{a
			World}
		{b
			42}}}`
	config, err := configuration.Read(strings.NewReader(source))
	assert.Nil(err)

	vs, err := config.At("a").Get()
	assert.Nil(err)
	assert.Equal(vs, "Hello")
	vb, err := config.At("b").GetBool()
	assert.Nil(err)
	assert.Equal(vb, true)
	vi, err := config.At("c").GetInt()
	assert.Nil(err)
	assert.Equal(vi, -1)
	vd, err := config.At("d").GetFloat64()
	assert.Nil(err)
	assert.Equal(vd, 47.11)
	vtim, err := config.At("e").GetTime()
	assert.Nil(err)
	loc, err := time.LoadLocation("CET")
	assert.Nil(err)
	assert.Equal(vtim.String(), time.Date(2015, time.June, 25, 23, 45, 00, 0, loc).String())
	vdur, err := config.At("f").GetDuration()
	assert.Nil(err)
	assert.Equal(vdur, 8130*time.Second)

	vs, err = config.At("sub", "a").Get()
	assert.Nil(err)
	assert.Equal(vs, "World")
	vi, err = config.At("sub", "b").GetInt()
	assert.Nil(err)
	assert.Equal(vi, 42)
}

// TestGetFail tests the failing retrieval of values.
func TestGetFail(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{config {a Hello}{sub {a World}}}"
	config, err := configuration.Read(strings.NewReader(source))
	assert.Nil(err)

	_, err = config.At("x").Get()
	assert.ErrorMatch(err, `.* invalid configuration path "/config/x"`)
	_, err = config.At("sub", "x").Get()
	assert.ErrorMatch(err, `.* invalid configuration path "/config/sub/x"`)

	vb, err := config.At("a").GetBool()
	assert.Equal(vb, false)
	assert.ErrorMatch(err, `.* invalid value format of "Hello": strconv.ParseBool: .*`)
	vi, err := config.At("a").GetInt()
	assert.Equal(vi, 0)
	assert.ErrorMatch(err, `.* invalid value format of "Hello": strconv.ParseInt: .*`)
	vf, err := config.At("a").GetFloat64()
	assert.Equal(vf, 0.0)
	assert.ErrorMatch(err, `.* invalid value format of "Hello": strconv.ParseFloat: .*`)
}

// TestGetDefault tests the retrieval of values with defaults.
func TestGetDefault(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{config {a Hello}}"
	config, err := configuration.ReadString(source)
	assert.Nil(err)

	vs := config.At("a").GetDefault("default")
	assert.Equal(vs, "Hello")

	vs = config.At("foo").GetDefault("default")
	assert.Equal(vs, "default")
	vb := config.At("foo").GetBoolDefault(true)
	assert.True(vb)
	vi := config.At("foo").GetIntDefault(42)
	assert.Equal(vi, 42)
	vf := config.At("foo").GetFloat64Default(47.11)
	assert.Equal(vf, 47.11)
	now := time.Now()
	vt := config.At("foo").GetTimeDefault(now)
	assert.Equal(vt, now)
	dur := 5 * time.Second
	vd := config.At("foo").GetDurationDefault(dur)
	assert.Equal(vd, dur)
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
	vs, err := applied.At("sub", "a").Get()
	assert.Nil(err)
	assert.Equal(vs, "Tester")
	vi, err := applied.At("b").GetInt()
	assert.Nil(err)
	assert.Equal(vi, 42)

}

// EOF
