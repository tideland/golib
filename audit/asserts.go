// Tideland Go Library - Audit
//
// Copyright (C) 2012-2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package audit

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"errors"
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
	Wait
	Retry
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
	Range:        "range",
	Substring:    "substring",
	Case:         "case",
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
	Retry:        "retry",
	Fail:         "fail",
}

func (t Test) String() string {
	if int(t) < len(testNames) {
		return testNames[t]
	}
	return "invalid"
}

//--------------------
// PRINTER
//--------------------

// Printer allows to switch between different outputs.
type Printer interface {
	// Printf prints a formatted information.
	Printf(format string, args ...interface{})
}

// printerBackend is the globally printer used during
// the assertions.
var (
	printerBackend Printer = &fmtPrinter{}
	mux            sync.Mutex
)

// fmtPrinter uses the standard fmt package for printing.
type fmtPrinter struct{}

// Printf implements the Printer interface.
func (p *fmtPrinter) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// SetPrinter sets a new global printer and returns the
// current one.
func SetPrinter(p Printer) Printer {
	mux.Lock()
	defer mux.Unlock()
	cp := printerBackend
	printerBackend = p
	return cp
}

// backendPrintf uses the printer backend for output. It is used
// in the types below.
func backendPrintf(format string, args ...interface{}) {
	mux.Lock()
	defer mux.Unlock()
	printerBackend.Printf(format, args...)
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

// Failures collects the collected failures
// of a validation assertion.
type Failures interface {
	// HasErrors returns true, if assertion failures happened.
	HasErrors() bool

	// Errors returns the so far collected errors.
	Errors() []error

	// Error returns the collected errors as one error.
	Error() error
}

// panicFailer reacts with a panic.
type panicFailer struct{}

// Logf implements the Failer interface.
func (f panicFailer) Logf(format string, args ...interface{}) {
	backendPrintf(format+"\n", args...)
}

// Fail implements the Failer interface.
func (f panicFailer) Fail(test Test, obtained, expected interface{}, msgs ...string) bool {
	obex := obexString(test, obtained, expected)
	panic(failString(test, obex, msgs...))
}

// validationFailer collects validation errors and additionally.
type validationFailer struct {
	mux  sync.Mutex
	errs []error
}

// HasErrors implements the Failures interface.
func (f *validationFailer) HasErrors() bool {
	f.mux.Lock()
	defer f.mux.Unlock()
	return len(f.errs) > 0
}

// Errors implements the Failures interface.
func (f *validationFailer) Errors() []error {
	f.mux.Lock()
	defer f.mux.Unlock()
	return f.errs
}

// Error implements the Failures interface.
func (f *validationFailer) Error() error {
	f.mux.Lock()
	defer f.mux.Unlock()
	strs := []string{}
	for i, err := range f.errs {
		strs = append(strs, fmt.Sprintf("[%d] %v", i, err))
	}
	return errors.New(strings.Join(strs, " / "))
}

// Logf implements the Failer interface.
func (f *validationFailer) Logf(format string, args ...interface{}) {
	f.mux.Lock()
	defer f.mux.Unlock()
	backendPrintf(format+"\n", args...)
}

// Fail implements the Failer interface.
func (f *validationFailer) Fail(test Test, obtained, expected interface{}, msgs ...string) bool {
	f.mux.Lock()
	defer f.mux.Unlock()
	obex := obexString(test, obtained, expected)
	f.errs = append(f.errs, errors.New(failString(test, obex, msgs...)))
	return false
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

// Logf implements the Failer interface.
func (f testingFailer) Logf(format string, args ...interface{}) {
	f.mux.Lock()
	defer f.mux.Unlock()
	_, file, line, _ := runtime.Caller(3)
	_, fileName := path.Split(file)
	prefix := fmt.Sprintf("%s:%d: ", fileName, line)
	backendPrintf(prefix+format+"\n", args...)
}

// Fail implements the Failer interface.
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
	backendPrintf(buffer.String())
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
	// (within the given extent).
	About(obtained, expected, extent float64, msgs ...string) bool

	// Range tests if obtained is larger or equal low and lower or
	// equal high. Allowed are byte, int and float64 for numbers, runes
	// and strings, or as a length test array, slices, and maps.
	Range(obtained, low, hight interface{}, msgs ...string) bool

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
		shallFail: shallFail,
	})
}

// assertion implements the assertion interface.
type assertion struct {
	Tester
	failer Failer
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

// IsAbout checks if obtained and expected are to a given extent almost equal.
func (t Tester) IsAbout(obtained, expected, extent float64) bool {
	if extent < 0.0 {
		extent = extent * (-1)
	}
	low := expected - extent
	high := expected + extent
	return low <= obtained && obtained <= high
}

// IsInRange checks, if obtained is inside the given range. In case of a
// slice, array, or map it will check agains the length.
func (t Tester) IsInRange(obtained, low, high interface{}) (bool, error) {
	// First standard types.
	switch o := obtained.(type) {
	case byte:
		l, lok := low.(byte)
		h, hok := high.(byte)
		if !lok && !hok {
			return false, errors.New("low and/or high are no byte")
		}
		return l <= o && o <= h, nil
	case int:
		l, lok := low.(int)
		h, hok := high.(int)
		if !lok && !hok {
			return false, errors.New("low and/or high are no int")
		}
		return l <= o && o <= h, nil
	case float64:
		l, lok := low.(float64)
		h, hok := high.(float64)
		if !lok && !hok {
			return false, errors.New("low and/or high are no float64")
		}
		return l <= o && o <= h, nil
	case rune:
		l, lok := low.(rune)
		h, hok := high.(rune)
		if !lok && !hok {
			return false, errors.New("low and/or high are no rune")
		}
		return l <= o && o <= h, nil
	case string:
		l, lok := low.(string)
		h, hok := high.(string)
		if !lok && !hok {
			return false, errors.New("low and/or high are no string")
		}
		return l <= o && o <= h, nil
	}
	// Now check the collection types.
	ol, err := t.Len(obtained)
	if err != nil {
		return false, errors.New("no valid type with a length")
	}
	l, lok := low.(int)
	h, hok := high.(int)
	if !lok && !hok {
		return false, errors.New("low and/or high are no int")
	}
	return l <= ol && ol <= h, nil
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
	return false, errors.New("full value is no string, array, or slice")
}

// IsSubstring checks if obtained is a substring of the full string.
func (t Tester) IsSubstring(obtained, full string) bool {
	return strings.Contains(full, obtained)
}

// IsCase checks if the obtained string is uppercase or lowercase.
func (t Tester) IsCase(obtained string, upperCase bool) bool {
	if upperCase {
		return obtained == strings.ToUpper(obtained)
	}
	return obtained == strings.ToLower(obtained)
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

// lowHigh transports the expected borders of a range test.
type lowHigh struct {
	low  interface{}
	high interface{}
}

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
	return make(chan interface{}, 16)
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
