// Tideland Go Library - Audit - Unit Tests
//
// Copyright (C) 2012-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package audit_test

//--------------------
// IMPORTS
//--------------------

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
)

//--------------------
// TESTS
//--------------------

// TestAssertTrue tests the True() assertion.
func TestAssertTrue(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.True(true, "should not fail")
	failingAssert.True(false, "should fail and be logged")
}

// TestAssertFalse tests the False() assertion.
func TestAssertFalse(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.False(false, "should not fail")
	failingAssert.False(true, "should fail and be logged")
}

// TestAssertNil tests the Nil() assertion.
func TestAssertNil(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.Nil(nil, "should not fail")
	failingAssert.Nil("not nil", "should fail and be logged")
}

// TestAssertNotNil tests the NotNil() assertion.
func TestAssertNotNil(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.NotNil("not nil", "should not fail")
	failingAssert.NotNil(nil, "should fail and be logged")
}

// TestAssertEqual tests the Equal() assertion.
func TestAssertEqual(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	m := map[string]int{"one": 1, "two": 2, "three": 3}
	now := time.Now()
	nowStr := now.Format(time.RFC3339Nano)
	nowParsedA, err := time.Parse(time.RFC3339Nano, nowStr)
	nowParsedB, err := time.Parse(time.RFC3339Nano, nowStr)

	successfulAssert.Nil(err, "should not fail")
	successfulAssert.Equal(nowParsedA, nowParsedB, "should not fail")
	successfulAssert.Equal(nil, nil, "should not fail")
	successfulAssert.Equal(true, true, "should not fail")
	successfulAssert.Equal(1, 1, "should not fail")
	successfulAssert.Equal("foo", "foo", "should not fail")
	successfulAssert.Equal(map[string]int{"one": 1, "three": 3, "two": 2}, m, "should not fail")
	failingAssert.Equal("one", 1, "should fail and be logged")
	failingAssert.Equal("two", "2", "should fail and be logged")
}

// TestAssertDifferent tests the Different() assertion.
func TestAssertDifferent(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	m := map[string]int{"one": 1, "two": 2, "three": 3}

	successfulAssert.Different(nil, "nil", "should not fail")
	successfulAssert.Different("true", true, "should not fail")
	successfulAssert.Different(1, 2, "should not fail")
	successfulAssert.Different("foo", "bar", "should not fail")
	successfulAssert.Different(map[string]int{"three": 3, "two": 2}, m, "should not fail")
	failingAssert.Different("one", "one", "should fail and be logged")
	failingAssert.Different(2, 2, "should fail and be logged")
}

// TestAssertAbout tests the About() assertion.
func TestAssertAbout(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.About(1.0, 1.0, 0.0, "equal, no extend")
	successfulAssert.About(1.0, 1.0, 0.1, "equal, little extend")
	successfulAssert.About(0.9, 1.0, 0.1, "different, within bounds of extent")
	successfulAssert.About(1.1, 1.0, 0.1, "different, within bounds of extent")
	failingAssert.About(0.8, 1.0, 0.1, "different, out of bounds of extent")
	failingAssert.About(1.2, 1.0, 0.1, "different, out of bounds of extent")
}

// TestAssertRange tests the Range() assertion.
func TestAssertRange(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.Range(byte(9), byte(1), byte(22), "byte in range")
	successfulAssert.Range(9, 1, 22, "int in range")
	successfulAssert.Range(9.0, 1.0, 22.0, "float64 in range")
	successfulAssert.Range('f', 'a', 'z', "rune in range")
	successfulAssert.Range("foo", "a", "zzzzz", "string in range")
	successfulAssert.Range([]int{1, 2, 3}, 1, 10, "slice length in range")
	successfulAssert.Range([3]int{1, 2, 3}, 1, 10, "array length in range")
	successfulAssert.Range(map[int]int{3: 1, 2: 2, 1: 3}, 1, 10, "map length in range")
	failingAssert.Range(byte(1), byte(10), byte(20), "byte out of range")
	failingAssert.Range(1, 10, 20, "int out of range")
	failingAssert.Range(1.0, 10.0, 20.0, "float64 out of range")
	failingAssert.Range('a', 'x', 'z', "rune out of range")
	failingAssert.Range("aaa", "uuuuu", "zzzzz", "string out of range")
	failingAssert.Range([]int{1, 2, 3}, 5, 10, "slice length out of range")
	failingAssert.Range([3]int{1, 2, 3}, 5, 10, "array length out of range")
	failingAssert.Range(map[int]int{3: 1, 2: 2, 1: 3}, 5, 10, "map length out of range")
}

// TestAssertContents tests the Contents() assertion.
func TestAssertContents(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.Contents("bar", "foobarbaz")
	successfulAssert.Contents(4711, []int{1, 2, 3, 4711, 5, 6, 7, 8, 9})
	failingAssert.Contents(4711, "12345-4711-67890")
	failingAssert.Contents(4711, "foo")
	failingAssert.Contents(4711, []interface{}{1, "2", 3, "4711", 5, 6, 7, 8, 9})
	successfulAssert.Contents("4711", []interface{}{1, "2", 3, "4711", 5, 6, 7, 8, 9})
}

// TestAssertSubstring tests the Substring() assertion.
func TestAssertSubstring(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.Substring("is assert", "this is assert test", "should not fail")
	successfulAssert.Substring("test", "this is 1 test", "should not fail")
	failingAssert.Substring("foo", "this is assert test", "should fail and be logged")
	failingAssert.Substring("this  is  assert  test", "this is assert test", "should fail and be logged")
}

// TestAssertCase tests the Case() assertion.
func TestAssertCase(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.Case("FOO", true, "is all uppercase")
	successfulAssert.Case("foo", false, "is all lowercase")
	failingAssert.Case("Foo", true, "is mixed case")
	failingAssert.Case("Foo", false, "is mixed case")
}

// TestAssertMatch tests the Match() assertion.
func TestAssertMatch(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.Match("this is assert test", "this.*test", "should not fail")
	successfulAssert.Match("this is 1 test", "this is [0-9] test", "should not fail")
	failingAssert.Match("this is assert test", "foo", "should fail and be logged")
	failingAssert.Match("this is assert test", "this*test", "should fail and be logged")
}

// TestAssertErrorMatch tests the ErrorMatch() assertion.
func TestAssertErrorMatch(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	err := errors.New("oops, an error")

	successfulAssert.ErrorMatch(err, "oops, an error", "should not fail")
	successfulAssert.ErrorMatch(err, "oops,.*", "should not fail")
	failingAssert.ErrorMatch(err, "foo", "should fail and be logged")
}

// TestAssertImplementor tests the Implementor() assertion.
func TestAssertImplementor(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	var err error
	var w io.Writer

	successfulAssert.Implementor(errors.New("error test"), &err, "should not fail")
	failingAssert.Implementor("string test", &err, "should fail and be logged")
	failingAssert.Implementor(errors.New("error test"), &w, "should fail and be logged")
}

// TestAssertAssignable tests the Assignable() assertion.
func TestAssertAssignable(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.Assignable(1, 5, "should not fail")
	failingAssert.Assignable("one", 5, "should fail and be logged")
}

// TestAssertUnassignable tests the Unassignable() assertion.
func TestAssertUnassignable(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.Unassignable("one", 5, "should not fail")
	failingAssert.Unassignable(1, 5, "should fail and be logged")
}

// TestAssertEmpty tests the Empty() assertion.
func TestAssertEmpty(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.Empty("", "should not fail")
	successfulAssert.Empty([]bool{}, "should also not fail")
	failingAssert.Empty("not empty", "should fail and be logged")
	failingAssert.Empty([3]int{1, 2, 3}, "should also fail and be logged")
	failingAssert.Empty(true, "illegal type has to fail")
}

// TestAssertNotEmpty tests the NotEmpty() assertion.
func TestAsserNotEmpty(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.NotEmpty("not empty", "should not fail")
	successfulAssert.NotEmpty([3]int{1, 2, 3}, "should also not fail")
	failingAssert.NotEmpty("", "should fail and be logged")
	failingAssert.NotEmpty([]int{}, "should also fail and be logged")
	failingAssert.NotEmpty(true, "illegal type has to fail")
}

// TestAssertLength tests the Length() assertion.
func TestAssertLength(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.Length("", 0, "should not fail")
	successfulAssert.Length([]bool{true, false}, 2, "should also not fail")
	failingAssert.Length("not empty", 0, "should fail and be logged")
	failingAssert.Length([3]int{1, 2, 3}, 10, "should also fail and be logged")
	failingAssert.Length(true, 1, "illegal type has to fail")
}

// TestAssertPanics tests the Panics() assertion.
func TestAssertPanics(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	successfulAssert.Panics(func() { panic("ouch") }, "should panic")
	failingAssert.Panics(func() { _ = 1 + 1 }, "should not panic")
}

// TestAssertPathExists tests the PathExists() assertion.
func TestAssertPathExists(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	td := audit.NewTempDir(successfulAssert)
	successfulAssert.NotNil(td)
	defer td.Restore()

	successfulAssert.PathExists(td.String(), "temporary directory exists")
	failingAssert.PathExists("/this/path/will/hopefully/not/exist", "illegal path")
}

// TestAssertWait tests the wait testing.
func TestAssertWait(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	sigc := audit.MakeSigChan()
	go func() {
		time.Sleep(50 * time.Millisecond)
		sigc <- true
	}()
	successfulAssert.Wait(sigc, true, 100*time.Millisecond, "should be true")

	go func() {
		time.Sleep(50 * time.Millisecond)
		sigc <- false
	}()
	failingAssert.Wait(sigc, true, 100*time.Millisecond, "should be false")

	go func() {
		time.Sleep(200 * time.Millisecond)
		sigc <- true
	}()
	failingAssert.Wait(sigc, true, 100*time.Millisecond, "should timeout")
}

// TestAssertRetry tests the retry testing.
func TestAssertRetry(t *testing.T) {
	successfulAssert := successfulAssertion(t)
	failingAssert := failingAssertion(t)

	i := 0
	successfulAssert.Retry(func() bool {
		i++
		return i == 5
	}, 10, 10*time.Millisecond, "should succeed")

	failingAssert.Retry(func() bool { return false }, 10, 10*time.Millisecond, "should fail")
}

// TestAssertFail tests the fail testing.
func TestAssertFail(t *testing.T) {
	failingAssert := failingAssertion(t)

	failingAssert.Fail("this should fail")
}

// TestTestingAssertion tests the testing assertion.
func TestTestingAssertion(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	foo := func() {}
	bar := 4711

	assert.Assignable(47, 11, "should not fail")
	assert.Assignable(foo, bar, "should fail (but not the test)")
	assert.Assignable(foo, bar)
	assert.Assignable(foo, bar, "this", "should", "fail", "too")
}

// TestPanicAssertion tests if the panic assertions panic when they fail.
func TestPanicAssert(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Logf("panic worked: '%v'", err)
		}
	}()

	assert := audit.NewPanicAssertion()
	foo := func() {}

	assert.Assignable(47, 11, "should not fail")
	assert.Assignable(47, foo, "should fail")

	t.Errorf("should not be reached")
}

// TestValidationAssertion test the validation of data.
func TestValidationAssertion(t *testing.T) {
	assert, failures := audit.NewValidationAssertion()

	assert.True(true, "should not fail")
	assert.True(false, "sould fail")
	assert.Equal(1, 2, "should fail")

	if !failures.HasErrors() {
		t.Errorf("should have errors")
	}
	if len(failures.Errors()) != 2 {
		t.Errorf("wrong number of errors")
	}
	t.Log(failures.Error())
}

//--------------------
// META FAILER
//--------------------

type metaFailer struct {
	t    *testing.T
	fail bool
}

func (f *metaFailer) Logf(format string, args ...interface{}) {
	f.t.Logf(format, args...)
}

func (f *metaFailer) Fail(test audit.Test, obtained, expected interface{}, msgs ...string) bool {
	msg := strings.Join(msgs, " ")
	if msg != "" {
		msg = " [" + msg + "]"
	}
	format := "testing assert %q failed: '%v' (%v) <> '%v' (%v)" + msg
	obtainedVD := audit.ValueDescription(obtained)
	expectedVD := audit.ValueDescription(expected)
	f.Logf(format, test, obtained, obtainedVD, expected, expectedVD)
	if f.fail {
		f.t.FailNow()
	}
	return f.fail
}

// successfulAssertion returns an assertion which doesn't expect a failing.
func successfulAssertion(t *testing.T) audit.Assertion {
	return audit.NewAssertion(&metaFailer{t, true})
}

// failingAssertion returns an assertion which only logs a failing but doesn't fail.
func failingAssertion(t *testing.T) audit.Assertion {
	return audit.NewAssertion(&metaFailer{t, false})
}

// EOF
