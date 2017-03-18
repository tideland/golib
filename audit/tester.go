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
	"os"
	"reflect"
	"regexp"
	"strings"
)

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
	}
	// Some types have to be tested via reflection.
	value := reflect.ValueOf(obtained)
	kind := value.Kind()
	switch kind {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
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

// Contains checks if the part type is matching to the full type and
// if the full data containes the part data.
func (t Tester) Contains(part, full interface{}) (bool, error) {
	switch fullValue := full.(type) {
	case string:
		// Content of a string.
		switch partValue := part.(type) {
		case string:
			return strings.Contains(fullValue, partValue), nil
		case []byte:
			return strings.Contains(fullValue, string(partValue)), nil
		default:
			partString := fmt.Sprintf("%v", partValue)
			return strings.Contains(fullValue, partString), nil
		}
	case []byte:
		// Content of a byte slice.
		switch partValue := part.(type) {
		case string:
			return bytes.Contains(fullValue, []byte(partValue)), nil
		case []byte:
			return bytes.Contains(fullValue, partValue), nil
		default:
			partBytes := []byte(fmt.Sprintf("%v", partValue))
			return bytes.Contains(fullValue, partBytes), nil
		}
	default:
		// Content of any array or slice, use reflection.
		value := reflect.ValueOf(full)
		kind := value.Kind()
		if kind == reflect.Array || kind == reflect.Slice {
			length := value.Len()
			for i := 0; i < length; i++ {
				current := value.Index(i)
				if reflect.DeepEqual(part, current.Interface()) {
					return true, nil
				}
			}
			return false, nil
		}
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

// IsValidPath checks if the given directory or
// file path exists.
func (t Tester) IsValidPath(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// EOF
