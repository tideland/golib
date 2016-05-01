// Tideland Go Library - Atom Feed
//
// Copyright (C) 2012-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package atom_test

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"net/url"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/feed/atom"
)

//--------------------
// TESTS
//--------------------

// Test parsing and composing of date/times.
func TestParseComposeTime(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	nowOne := time.Now()
	strOne := atom.ComposeTime(nowOne)

	nowTwo, err := atom.ParseTime(strOne)
	strTwo := atom.ComposeTime(nowTwo)

	assert.Nil(err)
	assert.Equal(strOne, strTwo)
}

// Test encoding and decoding a doc.
func TestEncodeDecode(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	a1 := &atom.Feed{
		XMLNS:   atom.XMLNS,
		Id:      "http://tideland.biz/feed/atom",
		Title:   &atom.Text{"Test Encode/Decode", "", "text"},
		Updated: atom.ComposeTime(time.Now()),
		Entries: []*atom.Entry{
			{
				Id:      "http://tideland.biz/feed/atom/1",
				Title:   &atom.Text{"Entry 1", "", "text"},
				Updated: atom.ComposeTime(time.Now()),
			},
			{
				Id:      "http://tideland.biz/feed/atom/2",
				Title:   &atom.Text{"Entry 2", "", "text"},
				Updated: atom.ComposeTime(time.Now()),
			},
		},
	}
	b := &bytes.Buffer{}

	err := atom.Encode(b, a1)
	assert.Nil(err, "Encoding returns no error.")
	assert.Substring(`<title type="text">Test Encode/Decode</title>`, b.String(), "Title has been encoded correctly.")

	a2, err := atom.Decode(b)
	assert.Nil(err, "Decoding returns no error.")
	assert.Equal(a2.Title.Text, "Test Encode/Decode", "Title has been decoded correctly.")
	assert.Length(a2.Entries, 2, "Decoded feed has the right number of items.")
}

// Test getting a feed.
func TestGet(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	u, _ := url.Parse("http://rss.golem.de/rss.php?feed=ATOM1.0")
	f, err := atom.Get(u)
	assert.Nil(err, "Getting the Atom document returns no error.")
	err = f.Validate()
	assert.Nil(err, "Validating returns no error.")
	b := &bytes.Buffer{}
	err = atom.Encode(b, f)
	assert.Nil(err, "Encoding returns no error.")
	assert.Logf("--- Atom ---\n%s", b)
}

// EOF
