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
	assert.Equal(l, 4)
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
	sv := doc.ValueAsString("A", "illegal")
	assert.Equal(sv, lo.A)
	sv = doc.ValueAsString("B/0/A", "illegal")
	assert.Equal(sv, lo.B[0].A)
	sv = doc.ValueAsString("/B/1/D/A", "illegal")
	assert.Equal(sv, lo.B[1].D.A)

	// Now two colons.
	doc, err = gjp.Parse(bs, "::")
	assert.Nil(err)
	sv = doc.ValueAsString("A", "illegal")
	assert.Equal(sv, lo.A)
	sv = doc.ValueAsString("B::0::A", "illegal")
	assert.Equal(sv, lo.B[0].A)
	sv = doc.ValueAsString("B::1::D::A", "illegal")
	assert.Equal(sv, lo.B[1].D.A)
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
