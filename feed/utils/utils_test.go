// Tideland Go Libray - Feed Utils
//
// Copyright (C) 2012-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package utils_test

//--------------------
// IMPORTS
//--------------------

import (
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/feed/utils"
)

//--------------------
// TESTS
//--------------------

func TestStripTags(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	in := "<p>The quick brown <b>fox</b> jumps over the lazy <em>dog</em>.</p>"
	out, err := utils.StripTags(in, true, false)
	assert.Nil(err, "No error during stripping.")
	assert.Equal(out, "The quick brown fox jumps over the lazy dog .", "Tags have been removed.")

	in = "<p>The quick brown <b>fox</b> jumps over the lazy <em>dog.</p>"
	out, err = utils.StripTags(in, true, false)
	assert.ErrorMatch(err, `XML syntax error on line 1.*`, "Error in document detected.")

	in = "<p>The quick brown <b>fox</b> jumps over the lazy <em>dog.</p>"
	out, err = utils.StripTags(in, false, false)
	assert.Nil(err, "No error during stripping.")
	assert.Equal(out, "The quick brown fox jumps over the lazy dog.", "Tags have been removed.")

	in = "<p>The quick brown <b>fox &amp; goose</b> jump over the lazy &lt;em&gt;dog&lt;/em&gt;.</p>"
	out, err = utils.StripTags(in, true, false)
	assert.Nil(err, "No error during stripping.")
	assert.Equal(out, "The quick brown fox & goose jump over the lazy <em>dog</em>.", "Tags have been removed.")

	in = "<p>The quick brown <b>fox &amp;amp; goose</b> jump over the lazy &lt;em&gt;dog&lt;/em&gt;.</p>"
	out, err = utils.StripTags(in, true, true)
	assert.Nil(err, "No error during stripping.")
	assert.Equal(out, "The quick brown fox & goose jump over the lazy dog .", "Tags have been removed.")
}

// EOF
