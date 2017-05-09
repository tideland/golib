// Tideland Go Library - Generic JSON Parser - Value
//
// Copyright (C) 2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package gjp

//--------------------
// IMPORTS
//--------------------

import (
	"strconv"
)

//--------------------
// VALUE
//--------------------

// Value contains one JSON value.
type Value interface {
	// AsString returns the value as string.
	AsString(dv string) string

	// AsInt returns the value as int.
	AsInt(dv int) int

	// AsFloat64 returns the value as float64.
	AsFloat64(dv float64) float64

	// AsBool returns the value as bool.
	AsBool(dv bool) bool
}

// value implements Value.
type value struct {
	raw interface{}
	ok  bool
}

// AsString implements Value.
func (v *value) AsString(dv string) string {
	if !v.ok {
		return dv
	}
	switch tv := v.raw.(type) {
	case string:
		return tv
	case int:
		return strconv.Itoa(tv)
	case float64:
		return strconv.FormatFloat(tv, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(tv)
	}
	return dv
}

// AsInt implements Value.
func (v *value) AsInt(dv int) int {
	if !v.ok {
		return dv
	}
	switch tv := v.raw.(type) {
	case string:
		i, err := strconv.Atoi(tv)
		if err != nil {
			return dv
		}
		return i
	case int:
		return tv
	case float64:
		return int(tv)
	case bool:
		if tv {
			return 1
		}
		return 0
	}
	return dv
}

// AsFloat64 implements Value.
func (v *value) AsFloat64(dv float64) float64 {
	if !v.ok {
		return dv
	}
	switch tv := v.raw.(type) {
	case string:
		f, err := strconv.ParseFloat(tv, 64)
		if err != nil {
			return dv
		}
		return f
	case int:
		return float64(tv)
	case float64:
		return tv
	case bool:
		if tv {
			return 1.0
		}
		return 0.0
	}
	return dv
}

// AsBool implements Value.
func (v *value) AsBool(dv bool) bool {
	if !v.ok {
		return dv
	}
	switch tv := v.raw.(type) {
	case string:
		b, err := strconv.ParseBool(tv)
		if err != nil {
			return dv
		}
		return b
	case int:
		return tv == 1
	case float64:
		return tv == 1.0
	case bool:
		return tv
	}
	return dv
}

// EOF
