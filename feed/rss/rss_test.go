// Tideland Go Library - RSS Feed - Unit Tests
//
// Copyright (C) 2012-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package rss_test

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"net/url"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/feed/rss"
)

//--------------------
// TESTS
//--------------------

// Test parsing and composing of date/times.
func TestParseComposeTime(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	nowOne := time.Now()
	nowStr := rss.ComposeTime(nowOne)

	assert.Logf("now as string: %s", nowStr)

	year, month, day := nowOne.Date()
	hour, min, _ := nowOne.Clock()
	loc := nowOne.Location()
	nowCmp := time.Date(year, month, day, hour, min, 0, 0, loc)
	nowTwo, err := rss.ParseTime(nowStr)

	assert.Nil(err, "No error during time parsing.")
	assert.Equal(nowCmp, nowTwo, "Both times have to be equal.")

	// Now some tests with different date formats.
	_, err = rss.ParseTime("21 Jun 2012 23:00 CEST")
	assert.Nil(err, "No error during time parsing.")
	_, err = rss.ParseTime("Thu, 21 Jun 2012 23:00 CEST")
	assert.Nil(err, "No error during time parsing.")
	_, err = rss.ParseTime("Thu, 21 Jun 2012 23:00 +0100")
	assert.Nil(err, "No error during time parsing.")
}

// Test encoding and decoding a doc.
func TestEncodeDecode(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	r1 := &rss.RSS{
		Version: rss.Version,
		Channel: rss.Channel{
			Title:       "Test Encode/Decode",
			Link:        "http://www.tideland.biz/rss",
			Description: "A test document.",
			Categories: []*rss.Category{
				{"foo", ""},
				{"bar", "baz"},
			},
			Items: []*rss.Item{
				{
					Title:       "Item 1",
					Description: "This is item 1",
					GUID:        &rss.GUID{"http://www.tideland.biz/rss/item-1", false},
				},
				{
					Title:       "Item 2",
					Description: "This is item 2",
					GUID:        &rss.GUID{"http://www.tideland.biz/rss/item-2", true},
				},
			},
		},
	}
	b := &bytes.Buffer{}

	err := rss.Encode(b, r1)
	assert.Nil(err, "Encoding returns no error.")
	assert.Substring("<title>Test Encode/Decode</title>", b.String(), "Title has been encoded correctly.")

	r2, err := rss.Decode(b)
	assert.Nil(err, "Decoding returns no error.")
	assert.Equal(r2.Channel.Title, "Test Encode/Decode", "Title has been decoded correctly.")
	assert.Length(r2.Channel.Items, 2, "Decoded document has the right number of items.")
}

// Test validating a doc.
func TestValidate(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	r := &rss.RSS{
		Version: "1.2.3",
	}
	err := r.Validate()
	assert.ErrorMatch(err, `.* invalid RSS document version: "1.2.3"`)
	r = &rss.RSS{
		Version: rss.Version,
	}
	err = r.Validate()
	assert.ErrorMatch(err, `.* channel title must not be empty`)
	r = &rss.RSS{
		Version: rss.Version,
		Channel: rss.Channel{
			Title: "Test Title",
		},
	}
	err = r.Validate()
	assert.ErrorMatch(err, `.* channel description must not be empty`)
	r = &rss.RSS{
		Version: rss.Version,
		Channel: rss.Channel{
			Title:       "Test Title",
			Description: "Test Description",
		},
	}
	err = r.Validate()
	assert.ErrorMatch(err, `.* channel link must not be empty`)
	r = &rss.RSS{
		Version: rss.Version,
		Channel: rss.Channel{
			Title:       "Test Title",
			Description: "Test Description",
			Link:        "p://%FF/foo#bar",
		},
	}
	err = r.Validate()
	assert.ErrorMatch(err, `.* cannot parse channel link: parse p://%FF/foo: hexadecimal escape in host`)
}

// Test getting a doc.
func TestGet(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	u, _ := url.Parse("http://www.rssboard.org/files/sample-rss-2.xml")
	r, err := rss.Get(u)
	assert.Nil(err, "Getting the RSS document returns no error.")
	err = r.Validate()
	assert.Nil(err, "Validating returns no error.")
	b := &bytes.Buffer{}
	err = rss.Encode(b, r)
	assert.Nil(err, "Encoding returns no error.")
	assert.Logf("--- RSS ---\n%s", b)
}

// EOF
