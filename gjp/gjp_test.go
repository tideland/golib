// Tideland Go Library - Generic JSON Parser - Unit Tests
//
// Copyright (C) 2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package gjp_test

//--------------------
// IMPORTS
//--------------------

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/gjp"
)

//--------------------
// TESTS
//--------------------

// TestLength tests retrieving values as strings.
func TestLength(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	bs, _ := createDocument(assert)

	doc, err := gjp.Parse(bs, "/")
	assert.Nil(err)
	l := doc.Length("X")
	assert.Equal(l, -1)
	l = doc.Length("")
	assert.Equal(l, 4)
	l = doc.Length("B")
	assert.Equal(l, 3)
	l = doc.Length("B/2")
	assert.Equal(l, 5)
	l = doc.Length("/B/2/D")
	assert.Equal(l, 2)
	l = doc.Length("/B/1/S")
	assert.Equal(l, 3)
	l = doc.Length("/B/1/S/0")
	assert.Equal(l, 1)
}

// TestSeparator tests using different separators.
func TestSeparator(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	bs, lo := createDocument(assert)

	// Slash as separator, once even starting with it.
	doc, err := gjp.Parse(bs, "/")
	assert.Nil(err)
	sv := doc.Value("A").AsString("illegal")
	assert.Equal(sv, lo.A)
	sv = doc.Value("B/0/A").AsString("illegal")
	assert.Equal(sv, lo.B[0].A)
	sv = doc.Value("/B/1/D/A").AsString("illegal")
	assert.Equal(sv, lo.B[1].D.A)
	sv = doc.Value("/B/2/S").AsString("illegal")
	assert.Equal(sv, "illegal")

	// Now two colons.
	doc, err = gjp.Parse(bs, "::")
	assert.Nil(err)
	sv = doc.Value("A").AsString("illegal")
	assert.Equal(sv, lo.A)
	sv = doc.Value("B::0::A").AsString("illegal")
	assert.Equal(sv, lo.B[0].A)
	sv = doc.Value("B::1::D::A").AsString("illegal")
	assert.Equal(sv, lo.B[1].D.A)

	// Check if is undefined.
	v := doc.Value("you-wont-find-me")
	assert.True(v.IsUndefined())
}

// TestString tests retrieving values as strings.
func TestString(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	bs, _ := createDocument(assert)

	doc, err := gjp.Parse(bs, "/")
	assert.Nil(err)
	sv := doc.Value("A").AsString("illegal")
	assert.Equal(sv, "Level One")
	sv = doc.Value("B/0/B").AsString("illegal")
	assert.Equal(sv, "100")
	sv = doc.Value("B/0/C").AsString("illegal")
	assert.Equal(sv, "true")
	sv = doc.Value("B/0/D/B").AsString("illegal")
	assert.Equal(sv, "10.1")
}

// TestInt tests retrieving values as ints.
func TestInt(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	bs, _ := createDocument(assert)

	doc, err := gjp.Parse(bs, "/")
	assert.Nil(err)
	iv := doc.Value("A").AsInt(-1)
	assert.Equal(iv, -1)
	iv = doc.Value("B/0/B").AsInt(-1)
	assert.Equal(iv, 100)
	iv = doc.Value("B/0/C").AsInt(-1)
	assert.Equal(iv, 1)
	iv = doc.Value("B/0/S/2").AsInt(-1)
	assert.Equal(iv, 1)
	iv = doc.Value("B/0/D/B").AsInt(-1)
	assert.Equal(iv, 10)
}

// TestFloat64 tests retrieving values as float64.
func TestFloat64(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	bs, _ := createDocument(assert)

	doc, err := gjp.Parse(bs, "/")
	assert.Nil(err)
	fv := doc.Value("A").AsFloat64(-1.0)
	assert.Equal(fv, -1.0)
	fv = doc.Value("B/1/B").AsFloat64(-1.0)
	assert.Equal(fv, 200.0)
	fv = doc.Value("B/0/C").AsFloat64(-99)
	assert.Equal(fv, 1.0)
	fv = doc.Value("B/0/S/3").AsFloat64(-1.0)
	assert.Equal(fv, 2.2)
	fv = doc.Value("B/1/D/B").AsFloat64(-1.0)
	assert.Equal(fv, 20.2)
}

// TestBool tests retrieving values as bool.
func TestBool(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	bs, _ := createDocument(assert)

	doc, err := gjp.Parse(bs, "/")
	assert.Nil(err)
	bv := doc.Value("A").AsBool(false)
	assert.Equal(bv, false)
	bv = doc.Value("B/0/C").AsBool(false)
	assert.Equal(bv, true)
	bv = doc.Value("B/0/S/0").AsBool(false)
	assert.Equal(bv, false)
	bv = doc.Value("B/0/S/2").AsBool(false)
	assert.Equal(bv, true)
	bv = doc.Value("B/0/S/4").AsBool(false)
	assert.Equal(bv, true)
}

//--------------------
// HELPERS
//--------------------

type levelThree struct {
	A string
	B float64
}

type levelTwo struct {
	A string
	B int
	C bool
	D *levelThree
	S []string
}

type levelOne struct {
	A string
	B []*levelTwo
	D time.Duration
	T time.Time
}

func createDocument(assert audit.Assertion) ([]byte, *levelOne) {
	lo := &levelOne{
		A: "Level One",
		B: []*levelTwo{
			&levelTwo{
				A: "Level Two - A",
				B: 100,
				C: true,
				D: &levelThree{
					A: "Level Three",
					B: 10.1,
				},
				S: []string{
					"red",
					"green",
					"1",
					"2.2",
					"true",
				},
			},
			&levelTwo{
				A: "Level Two - B",
				B: 200,
				C: false,
				D: &levelThree{
					A: "Level Three",
					B: 20.2,
				},
				S: []string{
					"orange",
					"blue",
					"white",
				},
			},
			&levelTwo{
				A: "Level Two - C",
				B: 300,
				C: true,
				D: &levelThree{
					A: "Level Three",
					B: 30.3,
				},
			},
		},
		D: 5 * time.Second,
		T: time.Date(2017, time.April, 29, 20, 30, 0, 0, time.UTC),
	}
	bs, err := json.Marshal(lo)
	assert.Nil(err)
	return bs, lo
}

// EOF
