// Tideland Go Library - Errors - Unit Tests
//
// Copyright (C) 2013-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package errors_test

//--------------------
// IMPORTS
//--------------------

import (
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/errors"
)

//--------------------
// TESTS
//--------------------

// TestIsError tests the creation and checking of errors.
func TestIsError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	ec := 42
	messages := errors.Messages{ec: "test error %d"}
	err := errors.New(ec, messages, 1)

	assert.ErrorMatch(err, `\[ERRORS_TEST:042\] test error 1`)
	assert.True(errors.IsError(err, ec))
	assert.False(errors.IsError(err, 0))

	err = testError("test error 2")

	assert.ErrorMatch(err, "test error 2")
	assert.False(errors.IsError(err, ec))
	assert.False(errors.IsError(err, 0))
}

// TestValidation checks the validation of errors and
// the retrieval of details.
func TestValidation(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	ec := 1
	messages := errors.Messages{ec: "valid"}
	err := errors.New(ec, messages)
	packageName, fileName, line, lerr := errors.Location(err)

	assert.True(errors.Valid(err))
	assert.Nil(lerr)
	assert.Equal(packageName, "github.com/tideland/golib/errors_test")
	assert.Equal(fileName, "errors_test.go")
	assert.Equal(line, 51)
}

// TestAnnotation the annotation of errors with new errors.
func TestAnnotation(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	ec := 123
	messages := errors.Messages{ec: "annotated"}
	aerr := testError("wrapped")
	err := errors.Annotate(aerr, ec, messages)

	assert.ErrorMatch(err, `\[ERRORS_TEST:123\] annotated: wrapped`)
	assert.Equal(errors.Annotated(err), aerr)
	assert.True(errors.IsInvalidTypeError(errors.Annotated(aerr)))
	assert.Length(errors.Stack(err), 2)
}

// TestCollection tests the collection of multiple errors to one.
func TestCollection(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	errA := testError("foo")
	errB := testError("bar")
	errC := testError("baz")
	errD := testError("yadda")
	cerr := errors.Collect(errA, errB, errC, errD)

	assert.ErrorMatch(cerr, "foo\nbar\nbaz\nyadda")
}

// TestDoAll tests the iteration over errors.
func TestDoAll(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	msgs := []string{}
	f := func(err error) {
		msgs = append(msgs, err.Error())
	}

	// Test it on annotated errors.
	messages := errors.Messages{
		1: "foo",
		2: "bar",
		3: "baz",
		4: "yadda",
	}
	errX := testError("xxx")
	errA := errors.Annotate(errX, 1, messages)
	errB := errors.Annotate(errA, 2, messages)
	errC := errors.Annotate(errB, 3, messages)
	errD := errors.Annotate(errC, 4, messages)

	errors.DoAll(errD, f)

	assert.Length(msgs, 5)

	// Test it on collected errors.
	msgs = []string{}
	errA = testError("foo")
	errB = testError("bar")
	errC = testError("baz")
	errD = testError("yadda")
	cerr := errors.Collect(errA, errB, errC, errD)

	errors.DoAll(cerr, f)

	assert.Equal(msgs, []string{"foo", "bar", "baz", "yadda"})

	// Test it on a single error.
	msgs = []string{}
	errA = testError("foo")

	errors.DoAll(errA, f)

	assert.Equal(msgs, []string{"foo"})
}

//--------------------
// HELPERS
//--------------------

type testError string

func (e testError) Error() string {
	return string(e)
}

// EOF
