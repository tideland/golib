// Tideland Go Library - Audit
//
// Copyright (C) 2012-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package audit

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"strings"
	"time"
)

//--------------------
// TEST
//--------------------

// Test represents the test inside an assert.
type Test uint

// Tests provided by the assertion.
const (
	Invalid Test = iota + 1
	True
	False
	Nil
	NotNil
	Equal
	Different
	Contents
	About
	Range
	Substring
	Case
	Match
	ErrorMatch
	Implementor
	Assignable
	Unassignable
	Empty
	NotEmpty
	Length
	Panics
	PathExists
	Wait
	Retry
	Fail
)

//--------------------
// ASSERTION
//--------------------

// MakeSigChan is a simple one-liner to create the buffered signal channel
// for the wait assertion.
func MakeSigChan() chan interface{} {
	return make(chan interface{}, 1)
}

// Assertion defines the available test methods.
type Assertion interface {
	// IncrCallstackOffset allows test libraries using the audit
	// package internally to adjust the callstack offset. This
	// way test output shows the correct location. Deferring
	// the returned function restores the former offset.
	IncrCallstackOffset() func()

	// Logf can be used to display useful information during testing.
	Logf(format string, args ...interface{})

	// True tests if obtained is true.
	True(obtained bool, msgs ...string) bool

	// False tests if obtained is false.
	False(obtained bool, msgs ...string) bool

	// Nil tests if obtained is nil.
	Nil(obtained interface{}, msgs ...string) bool

	// NotNil tests if obtained is not nil.
	NotNil(obtained interface{}, msgs ...string) bool

	// Equal tests if obtained and expected are equal.
	Equal(obtained, expected interface{}, msgs ...string) bool

	// Different tests if obtained and expected are different.
	Different(obtained, expected interface{}, msgs ...string) bool

	// Contents tests if the obtained data is part of the expected
	// string, array, or slice.
	Contents(obtained, full interface{}, msgs ...string) bool

	// About tests if obtained and expected are near to each other
	// (within the given extent).
	About(obtained, expected, extent float64, msgs ...string) bool

	// Range tests if obtained is larger or equal low and lower or
	// equal high. Allowed are byte, int and float64 for numbers, runes,
	// strings, times, and duration. In case of obtained arrays,
	// slices, and maps low and high have to be ints for testing
	// the length.
	Range(obtained, low, high interface{}, msgs ...string) bool

	// Substring tests if obtained is a substring of the full string.
	Substring(obtained, full string, msgs ...string) bool

	// Case tests if obtained string is uppercase or lowercase.
	Case(obtained string, upperCase bool, msgs ...string) bool

	// Match tests if the obtained string matches a regular expression.
	Match(obtained, regex string, msgs ...string) bool

	// ErrorMatch tests if the obtained error as string matches a
	// regular expression.
	ErrorMatch(obtained error, regex string, msgs ...string) bool

	// Implementor tests if obtained implements the expected
	// interface variable pointer.
	Implementor(obtained, expected interface{}, msgs ...string) bool

	// Assignable tests if the types of expected and obtained are assignable.
	Assignable(obtained, expected interface{}, msgs ...string) bool

	// Unassignable tests if the types of expected and obtained are
	// not assignable.
	Unassignable(obtained, expected interface{}, msgs ...string) bool

	// Empty tests if the len of the obtained string, array, slice
	// map, or channel is 0.
	Empty(obtained interface{}, msgs ...string) bool

	// NotEmpty tests if the len of the obtained string, array, slice
	// map, or channel is greater than 0.
	NotEmpty(obtained interface{}, msgs ...string) bool

	// Length tests if the len of the obtained string, array, slice
	// map, or channel is equal to the expected one.
	Length(obtained interface{}, expected int, msgs ...string) bool

	// Panics checks if the passed function panics.
	Panics(pf func(), msgs ...string) bool

	// PathExists checks if the passed path or file exists.
	PathExists(path string, msgs ...string) bool

	// Wait until a received signal or a timeout. The signal has
	// to be the expected value.
	Wait(sigc <-chan interface{}, expected interface{}, timeout time.Duration, msgs ...string) bool

	// Retry calls the passed function and expects it to return true. Otherwise
	// it pauses for the given duration and retries the call the defined number.
	Retry(rf func() bool, retries int, pause time.Duration, msgs ...string) bool

	// Fail always fails.
	Fail(msgs ...string) bool
}

// NewAssertion creates a new Assertion instance.
func NewAssertion(f Failer) Assertion {
	return &assertion{
		failer: f,
	}
}

// NewPanicAssertion creates a new Assertion instance which panics if a test fails.
func NewPanicAssertion() Assertion {
	return NewAssertion(&panicFailer{})
}

// NewValidationAssertion creates a new Assertion instance which collections
// validation failures. The returned Failures instance allows to test an access
// them.
func NewValidationAssertion() (Assertion, Failures) {
	vf := &validationFailer{}
	return NewAssertion(vf), vf
}

// NewTestingAssertion creates a new Assertion instance for use with the testing
// package. The *testing.T has to be passed as failable, the first argument.
// shallFail controls if a failing assertion also lets fail the Go test.
func NewTestingAssertion(f Failable, shallFail bool) Assertion {
	return NewAssertion(&testingFailer{
		failable:  f,
		offset:    2,
		shallFail: shallFail,
	})
}

// assertion implements the assertion interface.
type assertion struct {
	Tester
	failer Failer
}

// Logf implements the Assertion interface.
func (a *assertion) IncrCallstackOffset() func() {
	return a.failer.IncrCallstackOffset()
}

// Logf implements the Assertion interface.
func (a *assertion) Logf(format string, args ...interface{}) {
	a.failer.Logf(format, args...)
}

// True implements the Assertion interface.
func (a *assertion) True(obtained bool, msgs ...string) bool {
	if !a.IsTrue(obtained) {
		return a.failer.Fail(True, obtained, true, msgs...)
	}
	return true
}

// False implements the Assertion interface.
func (a *assertion) False(obtained bool, msgs ...string) bool {
	if a.IsTrue(obtained) {
		return a.failer.Fail(False, obtained, false, msgs...)
	}
	return true
}

// Nil implements the Assertion interface.
func (a *assertion) Nil(obtained interface{}, msgs ...string) bool {
	if !a.IsNil(obtained) {
		return a.failer.Fail(Nil, obtained, nil, msgs...)
	}
	return true
}

// NotNil implements the Assertion interface.
func (a *assertion) NotNil(obtained interface{}, msgs ...string) bool {
	if a.IsNil(obtained) {
		return a.failer.Fail(NotNil, obtained, nil, msgs...)
	}
	return true
}

// Equal implements the Assertion interface.
func (a *assertion) Equal(obtained, expected interface{}, msgs ...string) bool {
	if !a.IsEqual(obtained, expected) {
		return a.failer.Fail(Equal, obtained, expected, msgs...)
	}
	return true
}

// Different implements the Assertion interface.
func (a *assertion) Different(obtained, expected interface{}, msgs ...string) bool {
	if a.IsEqual(obtained, expected) {
		return a.failer.Fail(Different, obtained, expected, msgs...)
	}
	return true
}

// Contents implements the Assertion interface.
func (a *assertion) Contents(part, full interface{}, msgs ...string) bool {
	contains, err := a.Contains(part, full)
	if err != nil {
		return a.failer.Fail(Contents, part, full, "type missmatch: "+err.Error())
	}
	if !contains {
		return a.failer.Fail(Contents, part, full, msgs...)
	}
	return true
}

// About implements the Assertion interface.
func (a *assertion) About(obtained, expected, extent float64, msgs ...string) bool {
	if !a.IsAbout(obtained, expected, extent) {
		return a.failer.Fail(About, obtained, expected, msgs...)
	}
	return true
}

// Range implements the Assertion interface.
func (a *assertion) Range(obtained, low, high interface{}, msgs ...string) bool {
	expected := &lowHigh{low, high}
	inRange, err := a.IsInRange(obtained, low, high)
	if err != nil {
		return a.failer.Fail(Range, obtained, expected, "type missmatch: "+err.Error())
	}
	if !inRange {
		return a.failer.Fail(Range, obtained, expected, msgs...)
	}
	return true
}

// Substring implements the Assertion interface.
func (a *assertion) Substring(obtained, full string, msgs ...string) bool {
	if !a.IsSubstring(obtained, full) {
		return a.failer.Fail(Substring, obtained, full, msgs...)
	}
	return true
}

// Case implements the Assertion interface.
func (a *assertion) Case(obtained string, upperCase bool, msgs ...string) bool {
	if !a.IsCase(obtained, upperCase) {
		if upperCase {
			return a.failer.Fail(Case, obtained, strings.ToUpper(obtained), msgs...)
		}
		return a.failer.Fail(Case, obtained, strings.ToLower(obtained), msgs...)
	}
	return true
}

// Match implements the Assertion interface.
func (a *assertion) Match(obtained, regex string, msgs ...string) bool {
	matches, err := a.IsMatching(obtained, regex)
	if err != nil {
		return a.failer.Fail(Match, obtained, regex, "can't compile regex: "+err.Error())
	}
	if !matches {
		return a.failer.Fail(Match, obtained, regex, msgs...)
	}
	return true
}

// ErrorMatch implements the Assertion interface.
func (a *assertion) ErrorMatch(obtained error, regex string, msgs ...string) bool {
	if obtained == nil {
		return a.failer.Fail(ErrorMatch, nil, regex, "error is nil")
	}
	matches, err := a.IsMatching(obtained.Error(), regex)
	if err != nil {
		return a.failer.Fail(ErrorMatch, obtained, regex, "can't compile regex: "+err.Error())
	}
	if !matches {
		return a.failer.Fail(ErrorMatch, obtained, regex, msgs...)
	}
	return true
}

// Implementor implements the Assertion interface.
func (a *assertion) Implementor(obtained, expected interface{}, msgs ...string) bool {
	implements, err := a.IsImplementor(obtained, expected)
	if err != nil {
		return a.failer.Fail(Implementor, obtained, expected, err.Error())
	}
	if !implements {
		return a.failer.Fail(Implementor, obtained, expected, msgs...)
	}
	return implements
}

// Assignable implements the Assertion interface.
func (a *assertion) Assignable(obtained, expected interface{}, msgs ...string) bool {
	if !a.IsAssignable(obtained, expected) {
		return a.failer.Fail(Assignable, obtained, expected, msgs...)
	}
	return true
}

// Unassignable implements the Assertion interface.
func (a *assertion) Unassignable(obtained, expected interface{}, msgs ...string) bool {
	if a.IsAssignable(obtained, expected) {
		return a.failer.Fail(Unassignable, obtained, expected, msgs...)
	}
	return true
}

// Empty implements the Assertion interface.
func (a *assertion) Empty(obtained interface{}, msgs ...string) bool {
	length, err := a.Len(obtained)
	if err != nil {
		return a.failer.Fail(Empty, ValueDescription(obtained), 0, err.Error())
	}
	if length > 0 {
		return a.failer.Fail(Empty, length, 0, msgs...)

	}
	return true
}

// NotEmpty implements the Assertion interface.
func (a *assertion) NotEmpty(obtained interface{}, msgs ...string) bool {
	length, err := a.Len(obtained)
	if err != nil {
		return a.failer.Fail(NotEmpty, ValueDescription(obtained), 0, err.Error())
	}
	if length == 0 {
		return a.failer.Fail(NotEmpty, length, 0, msgs...)

	}
	return true
}

// Length implements the Assertion interface.
func (a *assertion) Length(obtained interface{}, expected int, msgs ...string) bool {
	length, err := a.Len(obtained)
	if err != nil {
		return a.failer.Fail(Length, ValueDescription(obtained), expected, err.Error())
	}
	if length != expected {
		return a.failer.Fail(Length, length, expected, msgs...)
	}
	return true
}

// Panics implements the Assertion interface.
func (a *assertion) Panics(pf func(), msgs ...string) bool {
	if !a.HasPanic(pf) {
		return a.failer.Fail(Panics, ValueDescription(pf), nil, msgs...)
	}
	return true
}

// PathExists implements the Assertion interface.
func (a *assertion) PathExists(obtained string, msgs ...string) bool {
	valid, err := a.IsValidPath(obtained)
	if err != nil {
		return a.failer.Fail(PathExists, obtained, true, err.Error())
	}
	if !valid {
		return a.failer.Fail(PathExists, obtained, true, msgs...)
	}
	return true
}

// Wait implements the Assertion interface.
func (a *assertion) Wait(sigc <-chan interface{}, expected interface{}, timeout time.Duration, msgs ...string) bool {
	select {
	case obtained := <-sigc:
		if !a.IsEqual(obtained, expected) {
			return a.failer.Fail(Wait, obtained, expected, msgs...)
		}
		return true
	case <-time.After(timeout):
		return a.failer.Fail(Wait, "timeout "+timeout.String(), "signal true", msgs...)
	}
}

// Retry implements the Assertion interface.
func (a *assertion) Retry(rf func() bool, retries int, pause time.Duration, msgs ...string) bool {
	start := time.Now()
	for r := 0; r < retries; r++ {
		if rf() {
			return true
		}
		time.Sleep(pause)
	}
	needed := time.Now().Sub(start)
	info := fmt.Sprintf("timeout after %v and %d retries", needed, retries)
	return a.failer.Fail(Retry, info, "successful call", msgs...)
}

// Fail implements the Assertion interface.
func (a *assertion) Fail(msgs ...string) bool {
	return a.failer.Fail(Fail, nil, nil, msgs...)
}

//--------------------
// HELPER
//--------------------

// lowHigh transports the expected borders of a range test.
type lowHigh struct {
	low  interface{}
	high interface{}
}

// lenable is an interface for the Len() mehod.
type lenable interface {
	Len() int
}

// obexString constructs a descriptive sting matching
// to test, obtained, and expected value.
func obexString(test Test, obtained, expected interface{}) string {
	switch test {
	case True, False, Nil, NotNil, Empty, NotEmpty:
		return fmt.Sprintf("'%v'", obtained)
	case Implementor, Assignable, Unassignable:
		return fmt.Sprintf("'%v' <> '%v'", ValueDescription(obtained), ValueDescription(expected))
	case Range:
		lh := expected.(*lowHigh)
		return fmt.Sprintf("not '%v' <= '%v' <= '%v'", lh.low, obtained, lh.high)
	case Fail:
		return "fail intended"
	default:
		return fmt.Sprintf("'%v' <> '%v'", obtained, expected)
	}
}

// failString constructs a fail string for panics or
// validition errors.
func failString(test Test, obex string, msgs ...string) string {
	var out string
	if test == Fail {
		out = fmt.Sprintf("assert failed: %s", obex)
	} else {
		out = fmt.Sprintf("assert '%s' failed: %s", test, obex)
	}
	jmsgs := strings.Join(msgs, " ")
	if len(jmsgs) > 0 {
		out += " (" + jmsgs + ")"
	}
	return out
}

// EOF
