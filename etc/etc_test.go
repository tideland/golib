// Tideland Go Library - Etc - Unit Tests
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package etc_test

//--------------------
// IMPORTS
//--------------------

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/etc"
)

//--------------------
// CONFIG
//--------------------

// TestRead tests reading a configuration out of a reader.
func TestRead(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{etc {foo 42}{bar 24}}"
	config, err := etc.Read(strings.NewReader(source))
	assert.Nil(err)

	source = "{something {gnagnagna}}"
	config, err = etc.Read(strings.NewReader(source))
	assert.Nil(config)
	assert.ErrorMatch(err, `*. illegal source format: .* node not found`)

	source = "{etc {gna 1}{gna 2}}"
	config, err = etc.Read(strings.NewReader(source))
	assert.Nil(config)
	assert.ErrorMatch(err, `*. illegal source format: .* cannot build node structure: node has multiple values`)

	source = "{etc {gna 1 {foo x} 2}}"
	config, err = etc.Read(strings.NewReader(source))
	assert.Nil(config)
	assert.ErrorMatch(err, `*. illegal source format: .* cannot build node structure: node has multiple values`)

	source = "{etc {foo/bar 1}{bar/foo 2}}"
	config, err = etc.Read(strings.NewReader(source))
	assert.Nil(config)
	assert.ErrorMatch(err, `*. illegal source format: .*`)
}

// TestReadFile tests reading a configuration out of a file.
func TestReadFile(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tempDir := audit.NewTempDir(assert)
	defer tempDir.Restore()
	etcFile, err := ioutil.TempFile(tempDir.String(), "etc")
	assert.Nil(err)
	etcFilename := etcFile.Name()
	_, err = etcFile.WriteString("{etc {foo 42}{bar 24}}")
	assert.Nil(err)
	etcFile.Close()

	config, err := etc.ReadFile(etcFilename)
	assert.Nil(err)
	v := config.ValueAsString("foo", "X")
	assert.Equal(v, "42")
	v = config.ValueAsString("bar", "Y")
	assert.Equal(v, "24")

	_, err = etc.ReadFile("some-not-existing-configuration-file-due-to-wierd-name")
	assert.ErrorMatch(err, `.* cannot read configuration file .*`)
}

// TestValueSuccess tests the successful retrieval of values.
func TestValueSuccess(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := `{etc
	{a  Hello}
	{b true}
	{c -1}
	{d     47.11     }
	{sub
		{a
			World}
		{b
			42}}}`
	config, err := etc.Read(strings.NewReader(source))
	assert.Nil(err)

	vs := config.ValueAsString("a", "foo")
	assert.Equal(vs, "Hello")
	vb := config.ValueAsBool("b", false)
	assert.Equal(vb, true)
	vi := config.ValueAsInt("c", 1)
	assert.Equal(vi, -1)
	vf := config.ValueAsFloat64("d", 1.0)
	assert.Equal(vf, 47.11)
	vs = config.ValueAsString("sub/a", "bar")
	assert.Equal(vs, "World")
	vi = config.ValueAsInt("sub/b", 12345)
	assert.Equal(vi, 42)
}

// TestGetDefault tests the retrieval of default values.
func TestGetFail(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{etc {a Hello}{sub {a World}}}"
	config, err := etc.Read(strings.NewReader(source))
	assert.Nil(err)

	vs := config.ValueAsString("b", "foo")
	assert.Equal(vs, "foo")
	vb := config.ValueAsBool("b", false)
	assert.Equal(vb, false)
	vi := config.ValueAsInt("c", 1)
	assert.Equal(vi, 1)
	vf := config.ValueAsFloat64("d", 1.0)
	assert.Equal(vf, 1.0)
	vb = config.ValueAsBool("sub/a", false)
	assert.Equal(vb, false)
	vi = config.ValueAsInt("sub/b", 12345)
	assert.Equal(vi, 12345)
}

// TestApply tests the applying of values.
func TestApply(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{etc {a Hello}{sub {a World}}}"
	config, err := etc.ReadString(source)
	assert.Nil(err)

	applied, err := config.Apply(map[string]string{
		"sub/a": "Tester",
		"b":     "42",
	})
	assert.Nil(err)
	vs := applied.ValueAsString("a", "foo")
	assert.Equal(vs, "Hello")
	vs = applied.ValueAsString("sub/a", "bar")
	assert.Equal(vs, "Tester")
	vi := applied.ValueAsInt("b", -1)
	assert.Equal(vi, 42)

}

// EOF
