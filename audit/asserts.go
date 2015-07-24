// Tideland Go Library - Audit
//
// Copyright (C) 2012-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package audit

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

//--------------------
// TEST
//--------------------

// Test represents the test inside an assert.
type Test uint

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
	Substring
	Match
	ErrorMatch
	Implementor
	Assignable
	Unassignable
	Empty
	NotEmpty
	Length
	Panics
	Wait
	Fail
)

var testNames = []string{
	Invalid:      "invalid",
	True:         "true",
	False:        "false",
	Nil:          "nil",
	NotNil:       "not nil",
	Equal:        "equal",
	Different:    "different",
	Contents:     "contents",
	About:        "about",
	Substring:    "substring",
	Match:        "match",
	ErrorMatch:   "error match",
	Implementor:  "implementor",
	Assignable:   "assignable",
	Unassignable: "unassignable",
	Empty:        "empty",
	NotEmpty:     "not empty",
	Length:       "length",
	Panics:       "panics",
	Wait:         "wait",
	Fail:         "fail",
}

func (t Test) String() string {
	if int(t) < len(testNames) {
		return testNames[t]
	}
	return "invalid"
}

//--------------------
// FAILER
//--------------------

// Failer describes a type controlling how an assert
// reacts after a failure.
type Failer interface {
	// Logf can be used to display useful information during testing.
	Logf(format string, args ...interface{})

	// Fail will be called if an assert fails.
	Fail(test Test, obtained, expected interface{}, msgs ...string) bool
}

// panicFailer reacts with a panic.
type panicFailer struct{}

// Logf is specified on the Failer interface.
func (f panicFailer) Logf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// Fail is specified on the Failer interface.
func (f panicFailer) Fail(test Test, obtained, expected interface{}, msgs ...string) bool {
	var obex string
	switch test {
	case True, False, Nil, NotNil, Empty, NotEmpty:
		obex = fmt.Sprintf("'%v'", obtained)
	case Implementor, Assignable, Unassignable:
		obex = fmt.Sprintf("'%v' <> '%v'", ValueDescription(obtained), ValueDescription(expected))
	case Fail:
		obex = "fail intended"
	default:
		obex = fmt.Sprintf("'%v' <> '%v'", obtained, expected)
	}
	if len(msgs) > 0 {
		jmsgs := strings.Join(msgs, " ")
		if test == Fail {
			panic(fmt.Sprintf("assert failed: %s (%s)", obex, jmsgs))
		} else {
			panic(fmt.Sprintf("assert '%s' failed: %s (%s)", test, obex, jmsgs))
		}
	} else {
		if test == Fail {
			panic(fmt.Sprintf("assert failed: %s", obex))
		} else {
			panic(fmt.Sprintf("assert '%s' failed: %s", test, obex))
		}
	}
}

// Failable allows an assertion to signal a fail to an external instance
// like testing.T or testing.B.
type Failable interface {
	FailNow()
}

// testingFailer works together with the testing package of Go and
// may signal the fail to it.
type testingFailer struct {
	mux       sync.Mutex
	failable  Failable
	shallFail bool
}

// Logf is specified on the Failer interface.
func (f testingFailer) Logf(format string, args ...interface{}) {
	f.mux.Lock()
	defer f.mux.Unlock()
	_, file, line, _ := runtime.Caller(3)
	_, fileName := path.Split(file)
	prefix := fmt.Sprintf("%s:%d: ", fileName, line)
	fmt.Printf(prefix+format+"\n", args...)
}

// Fail is specified on the Failer interface.
func (f testingFailer) Fail(test Test, obtained, expected interface{}, msgs ...string) bool {
	f.mux.Lock()
	defer f.mux.Unlock()
	pc, file, line, _ := runtime.Caller(3)
	_, fileName := path.Split(file)
	funcNameParts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	funcNamePartsIdx := len(funcNameParts) - 1
	funcName := funcNameParts[funcNamePartsIdx]
	buffer := &bytes.Buffer{}
	fmt.Fprintf(buffer, "--------------------------------------------------------------------------------\n")
	if test == Fail {
		fmt.Fprintf(buffer, "%s:%d: Assert failed!\n\n", fileName, line)
	} else {
		fmt.Fprintf(buffer, "%s:%d: Assert '%s' failed!\n\n", fileName, line, test)
	}
	fmt.Fprintf(buffer, "Function...: %s()\n", funcName)
	switch test {
	case True, False, Nil, NotNil, Empty, NotEmpty, Panics:
		fmt.Fprintf(buffer, "Obtained...: %v\n", obtained)
	case Implementor, Assignable, Unassignable:
		fmt.Fprintf(buffer, "Obtained...: %v\n", ValueDescription(obtained))
		fmt.Fprintf(buffer, "Expected...: %v\n", ValueDescription(expected))
	case Fail:
	default:
		fmt.Fprintf(buffer, "Obtained...: %v\n", obtained)
		fmt.Fprintf(buffer, "Expected...: %v\n", expected)
	}
	if len(msgs) > 0 {
		fmt.Fprintf(buffer, "Description: %s\n", strings.Join(msgs, "\n             "))
	}
	fmt.Fprintf(buffer, "--------------------------------------------------------------------------------\n")
	fmt.Print(buffer)
	if f.shallFail {
		f.failable.FailNow()
	}
	return false
}

//--------------------
// ASSERTION
//--------------------

// Assertion defines the available test methods.
type Assertion interface {
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
	// (within the given extend).
	About(obtained, expected, extend float64, msgs ...string) bool

	// Substring tests if obtained is a substring of the full string.
	Substring(obtained, full string, msgs ...string) bool

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

	// Wait until a received signal or a timeout. The signal has
	// to be the expected value.
	Wait(sigc <-chan interface{}, expected interface{}, timeout time.Duration, msgs ...string) bool

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

// NewTestingAssertion creates a new Assertion instance for use with the testing
// package. The *testing.T has to be passed as failable, the first argument.
// shallFail controls if a failing assertion also lets fail the Go test.
func NewTestingAssertion(f Failable, shallFail bool) Assertion {
	return NewAssertion(&testingFailer{
		failable:  f,
		shallFail: shallFail,
	})
}

// assertion implements the assertion interface.
type assertion struct {
	Tester
	failer Failer
}

// Logf is specified on the Assertion interface.
func (a *assertion) Logf(format string, args ...interface{}) {
	a.failer.Logf(format, args...)
}

// True is specified on the Assertion interface.
func (a *assertion) True(obtained bool, msgs ...string) bool {
	if !a.IsTrue(obtained) {
		return a.failer.Fail(True, obtained, true, msgs...)
	}
	return true
}

// False is specified on the Assertion interface.
func (a *assertion) False(obtained bool, msgs ...string) bool {
	if a.IsTrue(obtained) {
		return a.failer.Fail(False, obtained, false, msgs...)
	}
	return true
}

// Nil is specified on the Assertion interface.
func (a *assertion) Nil(obtained interface{}, msgs ...string) bool {
	if !a.IsNil(obtained) {
		return a.failer.Fail(Nil, obtained, nil, msgs...)
	}
	return true
}

// NotNil is specified on the Assertion interface.
func (a *assertion) NotNil(obtained interface{}, msgs ...string) bool {
	if a.IsNil(obtained) {
		return a.failer.Fail(NotNil, obtained, nil, msgs...)
	}
	return true
}

// Equal is specified on the Assertion interface.
func (a *assertion) Equal(obtained, expected interface{}, msgs ...string) bool {
	if !a.IsEqual(obtained, expected) {
		return a.failer.Fail(Equal, obtained, expected, msgs...)
	}
	return true
}

// Different is specified on the Assertion interface.
func (a *assertion) Different(obtained, expected interface{}, msgs ...string) bool {
	if a.IsEqual(obtained, expected) {
		return a.failer.Fail(Different, obtained, expected, msgs...)
	}
	return true
}

// Contents is specified on the Assertion interface.
func (a *assertion) Contents(obtained, full interface{}, msgs ...string) bool {
	contains, err := a.Contains(obtained, full)
	if err != nil {
		return a.failer.Fail(Contents, obtained, full, "type missmatch: "+err.Error())
	}
	if !contains {
		return a.failer.Fail(Contents, obtained, full, msgs...)
	}
	return true
}

// About is specified on the Assertion interface.
func (a *assertion) About(obtained, expected, extend float64, msgs ...string) bool {
	if !a.IsAbout(obtained, expected, extend) {
		return a.failer.Fail(About, obtained, expected, msgs...)
	}
	return true
}

// Substring is specified on the Assertion interface.
func (a *assertion) Substring(obtained, full string, msgs ...string) bool {
	if !a.IsSubstring(obtained, full) {
		return a.failer.Fail(Substring, obtained, full, msgs...)
	}
	return true
}

// Match is specified on the Assertion interface.
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

// ErrorMatch is specified on the Assertion interface.
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

// Implementor is specified on the Assertion interface.
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

// Assignable is specified on the Assertion interface.
func (a *assertion) Assignable(obtained, expected interface{}, msgs ...string) bool {
	if !a.IsAssignable(obtained, expected) {
		return a.failer.Fail(Assignable, obtained, expected, msgs...)
	}
	return true
}

// Unassignable is specified on the Assertion interface.
func (a *assertion) Unassignable(obtained, expected interface{}, msgs ...string) bool {
	if a.IsAssignable(obtained, expected) {
		return a.failer.Fail(Unassignable, obtained, expected, msgs...)
	}
	return true
}

// Empty is specified on the Assertion interface.
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

// NotEmpty is specified on the Assertion interface.
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

// Length is specified on the Assertion interface.
func (a *assertion) Length(obtained interface{}, expected int, msgs ...string) bool {
	length, err := a.Len(obtained)
	if err != nil {
		return a.failer.Fail(NotEmpty, ValueDescription(obtained), expected, err.Error())
	}
	if length != expected {
		return a.failer.Fail(Length, length, expected, msgs...)

	}
	return true
}

// Panics is specified on the Assertion interface.
func (a *assertion) Panics(pf func(), msgs ...string) bool {
	if !a.HasPanic(pf) {
		return a.failer.Fail(Panics, ValueDescription(pf), nil, msgs...)
	}
	return true
}

// Wait is specified on the Assertion interface.
func (a *assertion) Wait(sigc <-chan interface{}, expected interface{}, timeout time.Duration, msgs ...string) bool {
	select {
	case obtained := <-sigc:
		if !a.IsEqual(obtained, expected) {
			return a.failer.Fail(Wait, obtained, expected, msgs...)
		}
		return true
	case <-time.After(timeout):
		return a.failer.Fail(Wait, "timeout " + timeout.String(), "signal true", msgs...)
	}
}

// Fail is specified on the Assertion interface.
func (a *assertion) Fail(msgs ...string) bool {
	return a.failer.Fail(Fail, nil, nil, msgs...)
}

//--------------------
// TESTER
//--------------------

// Tester is a helper which can be used in own Assertion implementations.
type Tester struct{}

// IsTrue checks if obtained is true.
func (t Tester) IsTrue(obtained bool) bool {
	return obtained == true
}

// IsNil checks if obtained is nil in a safe way.
func (t Tester) IsNil(obtained interface{}) bool {
	if obtained == nil {
		// Standard test.
		return true
	} else {
		// Some types have to be tested via reflection.
		value := reflect.ValueOf(obtained)
		kind := value.Kind()
		switch kind {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			return value.IsNil()
		}
	}
	return false
}

// IsEqual checks if obtained and expected are equal.
func (t Tester) IsEqual(obtained, expected interface{}) bool {
	return reflect.DeepEqual(obtained, expected)
}

// IsAbout checks if obtained and expected are to a given extend almost equal.
func (t Tester) IsAbout(obtained, expected, extend float64) bool {
	if extend < 0.0 {
		extend = extend * (-1)
	}
	expectedMin := expected - extend
	expectedMax := expected + extend
	return obtained >= expectedMin && obtained <= expectedMax
}

// Contains checks if the obtained type is matching to the full type and
// if that containes the obtained data.
func (t Tester) Contains(obtained, full interface{}) (bool, error) {
	obtainedValue := reflect.ValueOf(obtained)
	fullValue := reflect.ValueOf(full)
	fullKind := fullValue.Kind()
	switch fullKind {
	case reflect.String:
		obtainedString := obtainedValue.String()
		fulltString := fullValue.String()
		return strings.Contains(fulltString, obtainedString), nil
	case reflect.Array, reflect.Slice:
		length := fullValue.Len()
		for i := 0; i < length; i++ {
			currentValue := fullValue.Index(i)
			if reflect.DeepEqual(obtained, currentValue.Interface()) {
				return true, nil
			}
		}
		return false, nil
	}
	return false, fmt.Errorf("full value is no string, array, or slice")
}

// IsSubstring checks if obtained is a substring of the full string.
func (t Tester) IsSubstring(obtained, full string) bool {
	return strings.Contains(full, obtained)
}

// IsMatching checks if the obtained string matches a regular expression.
func (t Tester) IsMatching(obtained, regex string) (bool, error) {
	return regexp.MatchString("^"+regex+"$", obtained)
}

// IsImplementor checks if obtained implements the expected interface variable pointer.
func (t Tester) IsImplementor(obtained, expected interface{}) (bool, error) {
	obtainedValue := reflect.ValueOf(obtained)
	expectedValue := reflect.ValueOf(expected)
	if !obtainedValue.IsValid() {
		return false, fmt.Errorf("obtained value is invalid: %v", obtained)
	}
	if !expectedValue.IsValid() || expectedValue.Kind() != reflect.Ptr || expectedValue.Elem().Kind() != reflect.Interface {
		return false, fmt.Errorf("expected value is no interface variable pointer: %v", expected)
	}
	return obtainedValue.Type().Implements(expectedValue.Elem().Type()), nil
}

// IsAssignable checks if the types of obtained and expected are assignable.
func (t Tester) IsAssignable(obtained, expected interface{}) bool {
	obtainedValue := reflect.ValueOf(obtained)
	expectedValue := reflect.ValueOf(expected)
	return obtainedValue.Type().AssignableTo(expectedValue.Type())
}

// Length checks the len of the obtained string, array, slice, map or channel.
func (t Tester) Len(obtained interface{}) (int, error) {
	// Check using the lenable interface.
	if l, ok := obtained.(lenable); ok {
		return l.Len(), nil
	}
	// Check the standard types.
	obtainedValue := reflect.ValueOf(obtained)
	obtainedKind := obtainedValue.Kind()
	switch obtainedKind {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return obtainedValue.Len(), nil
	default:
		descr := ValueDescription(obtained)
		return 0, fmt.Errorf("obtained %s is no array, chan, map, slice, string or understands Len()", descr)
	}
}

// HasPanic checks if the passed function panics.
func (t Tester) HasPanic(pf func()) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			// Panic, that's ok!
			ok = true
		}
	}()
	pf()
	return false
}

//--------------------
// HELPER
//--------------------

// ValueDescription returns a description of a value as string.
func ValueDescription(value interface{}) string {
	rvalue := reflect.ValueOf(value)
	kind := rvalue.Kind()
	switch kind {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return kind.String() + " of " + rvalue.Type().Elem().String()
	case reflect.Func:
		return kind.String() + " " + rvalue.Type().Name() + "()"
	case reflect.Interface, reflect.Struct:
		return kind.String() + " " + rvalue.Type().Name()
	case reflect.Ptr:
		return kind.String() + " to " + rvalue.Type().Elem().String()
	}
	// Default.
	return kind.String()
}

// MakeSigChan is a simple one-liner to create the buffered signal channel
// for the wait assertion.
func MakeSigChan() chan interface{} {
	return make(chan interface{}, 1)
}

// lenable is an interface for the Len() mehod.
type lenable interface {
	Len() int
}

// EOF
