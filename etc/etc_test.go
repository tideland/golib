// Tideland Go Library - Etc - Unit Tests
//
// Copyright (C) 2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package etc_test

//--------------------
// IMPORTS
//--------------------

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/etc"
)

//--------------------
// TESTS
//--------------------

// TestRead tests reading a configuration out of a reader.
func TestRead(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{etc {foo 42}{bar 24}}"
	cfg, err := etc.Read(strings.NewReader(source))
	assert.Nil(err)

	source = "{something {gnagnagna}}"
	cfg, err = etc.Read(strings.NewReader(source))
	assert.Nil(cfg)
	assert.ErrorMatch(err, `*. illegal source format: .* node not found`)

	source = "{etc {gna 1}{gna 2}}"
	cfg, err = etc.Read(strings.NewReader(source))
	assert.Nil(cfg)
	assert.ErrorMatch(err, `*. illegal source format: .* cannot build node structure: node has multiple values`)

	source = "{etc {gna 1 {foo x} 2}}"
	cfg, err = etc.Read(strings.NewReader(source))
	assert.Nil(cfg)
	assert.ErrorMatch(err, `*. illegal source format: .* cannot build node structure: node has multiple values`)

	source = "{etc {foo/bar 1}{bar/foo 2}}"
	cfg, err = etc.Read(strings.NewReader(source))
	assert.Nil(cfg)
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

	cfg, err := etc.ReadFile(etcFilename)
	assert.Nil(err)
	v := cfg.ValueAsString("foo", "X")
	assert.Equal(v, "42")
	v = cfg.ValueAsString("bar", "Y")
	assert.Equal(v, "24")

	_, err = etc.ReadFile("some-not-existing-configuration-file-due-to-wierd-name")
	assert.ErrorMatch(err, `.* cannot read configuration file .*`)
}

// TestTemplates tests the substitution of templates.
func TestTemplates(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := `{etc
	{a foo}
	{tests
		{valid-a x[a]x}
		{valid-b x[sub/b]x}
		{invalid-c x[c||123]x}
		{invalid-d x[sub/d||456]x}
		{invalid-e x[unknown]x}
		{invalid-f x[]x}
		{invalid-g x[||]x}
	}
	{sub {b bar}}
	}`
	cfg, err := etc.Read(strings.NewReader(source))
	assert.Nil(err)

	// First test regular ones, then those with templates.
	vs := cfg.ValueAsString("a", "xxx")
	assert.Equal(vs, "foo")
	vs = cfg.ValueAsString("sub/b", "xxx")
	assert.Equal(vs, "bar")

	vs = cfg.ValueAsString("tests/valid-a", "xxx")
	assert.Equal(vs, "xfoox")
	vs = cfg.ValueAsString("tests/valid-b", "xxx")
	assert.Equal(vs, "xbarx")
	vs = cfg.ValueAsString("tests/invalid-c", "xxx")
	assert.Equal(vs, "x123x")
	vs = cfg.ValueAsString("tests/invalid-d", "xxx")
	assert.Equal(vs, "x456x")
	vs = cfg.ValueAsString("tests/invalid-e", "xxx")
	assert.Equal(vs, "x[unknown]x")
	vs = cfg.ValueAsString("tests/invalid-f", "xxx")
	assert.Equal(vs, "x[]x")
	vs = cfg.ValueAsString("tests/invalid-g", "xxx")
	assert.Equal(vs, "xx")
}

// TestHasPath tests the checking of paths.
func TestHasPath(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{etc {a Hello}{sub {a World}}}"
	cfg, err := etc.Read(strings.NewReader(source))
	assert.Nil(err)

	assert.True(cfg.HasPath("a"))
	assert.True(cfg.HasPath("sub"))
	assert.True(cfg.HasPath("sub/a"))

	assert.False(cfg.HasPath("b"))
	assert.False(cfg.HasPath("sub/b"))
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
			42}
	}}`
	cfg, err := etc.Read(strings.NewReader(source))
	assert.Nil(err)

	vs := cfg.ValueAsString("a", "foo")
	assert.Equal(vs, "Hello")
	vb := cfg.ValueAsBool("b", false)
	assert.Equal(vb, true)
	vi := cfg.ValueAsInt("c", 1)
	assert.Equal(vi, -1)
	vf := cfg.ValueAsFloat64("d", 1.0)
	assert.Equal(vf, 47.11)
	vs = cfg.ValueAsString("sub/a", "bar")
	assert.Equal(vs, "World")
	vi = cfg.ValueAsInt("sub/b", 12345)
	assert.Equal(vi, 42)
}

// TestGetDefault tests the retrieval of default values.
func TestGetFail(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{etc {a Hello}{sub {a World}}}"
	cfg, err := etc.Read(strings.NewReader(source))
	assert.Nil(err)

	vs := cfg.ValueAsString("b", "foo")
	assert.Equal(vs, "foo")
	vb := cfg.ValueAsBool("b", false)
	assert.Equal(vb, false)
	vi := cfg.ValueAsInt("c", 1)
	assert.Equal(vi, 1)
	vf := cfg.ValueAsFloat64("d", 1.0)
	assert.Equal(vf, 1.0)
	vb = cfg.ValueAsBool("sub/a", false)
	assert.Equal(vb, false)
	vi = cfg.ValueAsInt("sub/b", 12345)
	assert.Equal(vi, 12345)
}

// TestSplit tests the splitting of configurations.
func TestSplit(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{etc {a Hello}{sub {a World}{b Friend}}}"
	cfg, err := etc.ReadString(source)
	assert.Nil(err)

	// Test the splitting.
	subcfg, err := cfg.Split("sub")
	assert.Nil(err)
	va := subcfg.ValueAsString("a", "Foo")
	assert.Equal(va, "World")
	vb := subcfg.ValueAsString("b", "Bar")
	assert.Equal(vb, "Friend")

	// Changing the sub configuration must not
	// change the original configuration.
	applied, err := subcfg.Apply(etc.Application{
		"c": "Darling",
	})
	ac := applied.ValueAsString("c", "A1")
	assert.Equal(ac, "Darling")
	ac = subcfg.ValueAsString("c", "A2")
	assert.Equal(ac, "A2")
	ac = cfg.ValueAsString("c", "A3")
	assert.Equal(ac, "A3")
}

// TestDump tests the dumping of a configuration.
func TestDump(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{etc {a Hello}{sub {a World}}}"
	cfg, err := etc.ReadString(source)
	assert.Nil(err)

	dump, err := cfg.Dump()
	assert.Nil(err)
	assert.Length(dump, 3)

	source = "{etc}"
	cfg, err = etc.ReadString(source)
	assert.Nil(err)

	applied, err := cfg.Apply(dump)
	assert.Nil(err)
	vs := applied.ValueAsString("a", "foo")
	assert.Equal(vs, "Hello")
	vs = applied.ValueAsString("sub", "bar")
	assert.Equal(vs, "")
	vs = applied.ValueAsString("sub/a", "baz")
	assert.Equal(vs, "World")
}

// TestApply tests the applying of values.
func TestApply(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{etc {a Hello}{sub {a World}}}"
	cfg, err := etc.ReadString(source)
	assert.Nil(err)

	applied, err := cfg.Apply(etc.Application{
		"sub/a": "Tester",
		"B":     "42",
	})
	assert.Nil(err)
	vs := applied.ValueAsString("a", "foo")
	assert.Equal(vs, "Hello")
	vs = applied.ValueAsString("sub/a", "bar")
	assert.Equal(vs, "Tester")
	vi := applied.ValueAsInt("b", -1)
	assert.Equal(vi, 42)
}

// TestContext tests adding a configuration to a context
// an retrieve it again.
func TestContext(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	source := "{etc {a Hello}{sub {a World}}}"
	cfg, err := etc.ReadString(source)
	assert.Nil(err)

	ctx := context.Background()
	noCfg, ok := etc.FromContext(ctx)
	assert.False(ok)
	assert.Nil(noCfg)

	cfgCtx := etc.NewContext(ctx, cfg)
	yesCfg, ok := etc.FromContext(cfgCtx)
	assert.True(ok)
	vs := yesCfg.ValueAsString("a", "foo")
	assert.Equal(vs, "Hello")
	vs = yesCfg.ValueAsString("sub/a", "bar")
	assert.Equal(vs, "World")
}

// EOF
