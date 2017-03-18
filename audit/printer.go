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
	"reflect"
	"sync"
)

//--------------------
// PRINTER
//--------------------

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
	PathExists:   "path exists",
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

// EOF
