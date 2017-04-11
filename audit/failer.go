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
	"bytes"
	"errors"
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

//--------------------
// FAILER
//--------------------

// Failer describes a type controlling how an assert
// reacts after a failure.
type Failer interface {
	// IncrCallstackOffset increases the callstack offset for
	// the assertion output (see Assertion) and returns a function
	// for restoring.
	IncrCallstackOffset() func()

	// Logf can be used to display useful information during testing.
	Logf(format string, args ...interface{})

	// Fail will be called if an assert fails.
	Fail(test Test, obtained, expected interface{}, msgs ...string) bool
}

// FailureDetail contains detailed information of a failure.
type FailureDetail interface {
	// TImestamp tells when the failure has happened.
	Timestamp() time.Time

	// Locations returns file name, line number, and
	// function name of the failure.
	Location() (string, int, string)

	// Test tells which kind of test has failed.
	Test() Test

	// Error returns the failure as error.
	Error() error

	// Message return the optional test message.
	Message() string
}

// failureDetail implements the FailureDetail interface.
type failureDetail struct {
	timestamp  time.Time
	fileName   string
	lineNumber int
	funcName   string
	test       Test
	err        error
	message    string
}

// TImestamp implements the FailureDetail interface.
func (d *failureDetail) Timestamp() time.Time {
	return d.timestamp
}

// Locations implements the FailureDetail interface.
func (d *failureDetail) Location() (string, int, string) {
	return d.fileName, d.lineNumber, d.funcName
}

// Test implements the FailureDetail interface.
func (d *failureDetail) Test() Test {
	return d.test
}

// Error implements the FailureDetail interface.
func (d *failureDetail) Error() error {
	return d.err
}

// Message implements the FailureDetail interface.
func (d *failureDetail) Message() string {
	return d.message
}

// Failures collects the collected failures
// of a validation assertion.
type Failures interface {
	// HasErrors returns true, if assertion failures happened.
	HasErrors() bool

	// Details returns the collected details.
	Details() []FailureDetail

	// Errors returns the so far collected errors.
	Errors() []error

	// Error returns the collected errors as one error.
	Error() error
}

//--------------------
// PANIC FAILER
//--------------------

// panicFailer reacts with a panic.
type panicFailer struct{}

// IncrCallstackOffset implements the Failer interface.
func (f panicFailer) IncrCallstackOffset() func() {
	return func() {}
}

// Logf implements the Failer interface.
func (f panicFailer) Logf(format string, args ...interface{}) {
	backendPrintf(format+"\n", args...)
}

// Fail implements the Failer interface.
func (f panicFailer) Fail(test Test, obtained, expected interface{}, msgs ...string) bool {
	obex := obexString(test, obtained, expected)
	panic(failString(test, obex, msgs...))
}

// NewPanicAssertion creates a new Assertion instance which panics if a test fails.
func NewPanicAssertion() Assertion {
	return NewAssertion(&panicFailer{})
}

//--------------------
// VALIDATION FAILER
//--------------------

// validationFailer collects validation errors, e.g. when
// validating form input data.
type validationFailer struct {
	mux     sync.Mutex
	offset  int
	details []FailureDetail
	errs    []error
}

// HasErrors implements the Failures interface.
func (f *validationFailer) HasErrors() bool {
	f.mux.Lock()
	defer f.mux.Unlock()
	return len(f.errs) > 0
}

// Details implements the Failures interface.
func (f *validationFailer) Details() []FailureDetail {
	f.mux.Lock()
	defer f.mux.Unlock()
	return f.details
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

// IncrCallstackOffset implements the Failer interface.
func (f *validationFailer) IncrCallstackOffset() func() {
	f.mux.Lock()
	defer f.mux.Unlock()
	offset := f.offset
	f.offset++
	return func() {
		f.mux.Lock()
		defer f.mux.Unlock()
		f.offset = offset
	}
}

// Logf implements the Failer interface.
func (f *validationFailer) Logf(format string, args ...interface{}) {
	f.mux.Lock()
	defer f.mux.Unlock()
	pc, file, line, _ := runtime.Caller(f.offset)
	_, fileName := path.Split(file)
	funcNameParts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	funcNamePartsIdx := len(funcNameParts) - 1
	funcName := funcNameParts[funcNamePartsIdx]
	prefix := fmt.Sprintf("%s:%d %s(): ", fileName, line, funcName)
	backendPrintf(prefix+format+"\n", args...)
}

// Fail implements the Failer interface.
func (f *validationFailer) Fail(test Test, obtained, expected interface{}, msgs ...string) bool {
	f.mux.Lock()
	defer f.mux.Unlock()
	pc, file, line, _ := runtime.Caller(f.offset)
	_, fileName := path.Split(file)
	funcNameParts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	funcNamePartsIdx := len(funcNameParts) - 1
	funcName := funcNameParts[funcNamePartsIdx]
	obex := obexString(test, obtained, expected)
	err := errors.New(failString(test, obex, msgs...))
	detail := &failureDetail{
		timestamp:  time.Now(),
		fileName:   fileName,
		lineNumber: line,
		funcName:   funcName,
		test:       test,
		err:        err,
		message:    strings.Join(msgs, " "),
	}
	f.details = append(f.details, detail)
	f.errs = append(f.errs, err)
	return false
}

// NewValidationAssertion creates a new Assertion instance which collections
// validation failures. The returned Failures instance allows to test an access
// them.
func NewValidationAssertion() (Assertion, Failures) {
	vf := &validationFailer{
		offset:  2,
		details: []FailureDetail{},
		errs:    []error{},
	}
	return NewAssertion(vf), vf
}

//--------------------
// TESTING FAILER
//--------------------

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
	offset    int
	shallFail bool
}

// IncrCallstackOffset implements the failer interface.
func (f *testingFailer) IncrCallstackOffset() func() {
	f.mux.Lock()
	defer f.mux.Unlock()
	offset := f.offset
	f.offset++
	return func() {
		f.mux.Lock()
		defer f.mux.Unlock()
		f.offset = offset
	}
}

// Logf implements the Failer interface.
func (f *testingFailer) Logf(format string, args ...interface{}) {
	f.mux.Lock()
	defer f.mux.Unlock()
	pc, file, line, _ := runtime.Caller(f.offset)
	_, fileName := path.Split(file)
	funcNameParts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	funcNamePartsIdx := len(funcNameParts) - 1
	funcName := funcNameParts[funcNamePartsIdx]
	prefix := fmt.Sprintf("%s:%d %s(): ", fileName, line, funcName)
	backendPrintf(prefix+format+"\n", args...)
}

// Fail implements the Failer interface.
func (f *testingFailer) Fail(test Test, obtained, expected interface{}, msgs ...string) bool {
	f.mux.Lock()
	defer f.mux.Unlock()
	pc, file, line, _ := runtime.Caller(f.offset)
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
	case Contents:
		switch typedObtained := obtained.(type) {
		case string:
			fmt.Fprintf(buffer, "Part.......: %s\n", typedObtained)
			fmt.Fprintf(buffer, "Full.......: %s\n", expected)
		default:
			fmt.Fprintf(buffer, "Part.......: %v\n", obtained)
			fmt.Fprintf(buffer, "Full.......: %v\n", expected)
		}
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

// EOF
