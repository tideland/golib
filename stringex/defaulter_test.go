// Tideland Go Library - String Extensions - Unit Tests
//
// Copyright (C) 2015-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package stringex_test

//--------------------
// IMPORTS
//--------------------

import (
	"errors"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/stringex"
)

//--------------------
// CONSTANTS
//--------------------

const (
	maxUint = ^uint(0)
	minUint = 0
	maxInt  = int(maxUint >> 1)
	minInt  = -maxInt - 1
)

//--------------------
// TESTS
//--------------------

// TestAsString checks the access of string values.
func TestAsString(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	d := stringex.NewDefaulter("AsString", true)
	tests := []struct {
		v        valuer
		dv       string
		expected string
	}{
		{valuer{"foo", false}, "bar", "foo"},
		{valuer{"foo", true}, "bar", "bar"},
	}
	for i, test := range tests {
		assert.Logf("test %v %d: %v and default %v", d, i, test.v, test.dv)
		sv := d.AsString(test.v, test.dv)
		assert.Equal(sv, test.expected)
	}
}

// TestAsStringSlice checks the access of string slice values.
func TestAsStringSlice(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	d := stringex.NewDefaulter("AsStringSlice", true)
	tests := []struct {
		v        valuer
		sep      string
		dv       []string
		expected []string
	}{
		{valuer{"a/b/c", false}, "/", []string{"a"}, []string{"a", "b", "c"}},
		{valuer{"a/b/c", true}, "/", []string{"a"}, []string{"a"}},
		{valuer{"a/b/c", false}, "/", nil, []string{"a", "b", "c"}},
		{valuer{"a/b/c", true}, "/", nil, nil},
		{valuer{"", false}, "/", nil, []string{""}},
		{valuer{"", true}, "/", []string{"foo"}, []string{"foo"}},
		{valuer{"a/b/c", false}, ":", []string{"a"}, []string{"a/b/c"}},
	}
	for i, test := range tests {
		assert.Logf("test %v %d: %v and default %v", d, i, test.v, test.dv)
		sv := d.AsStringSlice(test.v, test.sep, test.dv)
		assert.Equal(sv, test.expected)
	}
}

// TestAsStringMap checks the access of string map values.
func TestAsStringMap(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	d := stringex.NewDefaulter("AsStringMap", true)
	tests := []struct {
		v        valuer
		rsep     string
		kvsep    string
		dv       map[string]string
		expected map[string]string
	}{
		{valuer{"a:1/b:2", false}, "/", ":", map[string]string{"a": "1"}, map[string]string{"a": "1", "b": "2"}},
		{valuer{"a:1/b:2", true}, "/", ":", map[string]string{"a": "1"}, map[string]string{"a": "1"}},
		{valuer{"a:1/b:2", false}, "/", ":", nil, map[string]string{"a": "1", "b": "2"}},
		{valuer{"a:1/b:2", true}, "/", ":", nil, nil},
		{valuer{"", false}, "/", ":", nil, map[string]string{"": ""}},
		{valuer{"", true}, "/", ":", map[string]string{"a": "1"}, map[string]string{"a": "1"}},
		{valuer{"a:1/b:2", false}, "|", "=", nil, map[string]string{"a:1/b:2": "a:1/b:2"}},
	}
	for i, test := range tests {
		assert.Logf("test %v %d: %v and default %v", d, i, test.v, test.dv)
		sv := d.AsStringMap(test.v, test.rsep, test.kvsep, test.dv)
		assert.Equal(sv, test.expected)
	}
}

// TestAsBool checks the access of bool values.
func TestAsBool(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	d := stringex.NewDefaulter("AsBool", true)
	tests := []struct {
		v        valuer
		dv       bool
		expected bool
	}{
		{valuer{"1", false}, false, true},
		{valuer{"t", false}, false, true},
		{valuer{"T", false}, false, true},
		{valuer{"TRUE", false}, false, true},
		{valuer{"true", false}, false, true},
		{valuer{"True", false}, false, true},
		{valuer{"wahr", false}, true, true},
		{valuer{"", true}, true, true},
		{valuer{"0", false}, true, false},
		{valuer{"f", false}, true, false},
		{valuer{"F", false}, true, false},
		{valuer{"FALSE", false}, true, false},
		{valuer{"false", false}, true, false},
		{valuer{"False", false}, true, false},
		{valuer{"falsch", false}, false, false},
		{valuer{"", true}, false, false},
	}
	for i, test := range tests {
		assert.Logf("test %v %d: %v and default %v", d, i, test.v, test.dv)
		bv := d.AsBool(test.v, test.dv)
		assert.Equal(bv, test.expected)
	}
}

// TestAsInt checks the access of int values.
func TestAsInt(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	maxIntS := strconv.FormatInt(int64(maxInt), 10)
	minIntS := strconv.FormatInt(int64(minInt), 10)
	d := stringex.NewDefaulter("AsInt", true)
	tests := []struct {
		v        valuer
		dv       int
		expected int
	}{
		{valuer{"0", false}, 0, 0},
		{valuer{"1", false}, 0, 1},
		{valuer{"-1", false}, 0, -1},
		{valuer{maxIntS, false}, 0, maxInt},
		{valuer{minIntS, false}, 0, minInt},
		{valuer{"999999999999999999999", false}, 1, 1},
		{valuer{"-999999999999999999999", false}, 1, 1},
		{valuer{"one two three", false}, 1, 1},
		{valuer{"1", true}, 2, 2},
		{valuer{"-1", true}, -2, -2},
	}
	for i, test := range tests {
		assert.Logf("test %v %d: %v and default %v", d, i, test.v, test.dv)
		bv := d.AsInt(test.v, test.dv)
		assert.Equal(bv, test.expected)
	}
}

// TestAsInt64 checks the access of int64 values.
func TestAsInt64(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	maxInt64S := strconv.FormatInt(math.MaxInt64, 10)
	minInt64S := strconv.FormatInt(math.MinInt64, 10)
	d := stringex.NewDefaulter("AsInt64", true)
	tests := []struct {
		v        valuer
		dv       int64
		expected int64
	}{
		{valuer{"0", false}, 0, 0},
		{valuer{"1", false}, 0, 1},
		{valuer{"-1", false}, 0, -1},
		{valuer{maxInt64S, false}, 0, math.MaxInt64},
		{valuer{minInt64S, false}, 0, math.MinInt64},
		{valuer{"999999999999999999999", false}, 1, 1},
		{valuer{"-999999999999999999999", false}, 1, 1},
		{valuer{"one two three", false}, 1, 1},
		{valuer{"1", true}, 2, 2},
		{valuer{"-1", true}, -2, -2},
	}
	for i, test := range tests {
		assert.Logf("test %v %d: %v and default %v", d, i, test.v, test.dv)
		bv := d.AsInt64(test.v, test.dv)
		assert.Equal(bv, test.expected)
	}
}

// TestAsUint checks the access of uint values.
func TestAsUint(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	maxUintS := strconv.FormatUint(uint64(maxUint), 10)
	d := stringex.NewDefaulter("AsUint", true)
	tests := []struct {
		v        valuer
		dv       uint
		expected uint
	}{
		{valuer{"0", false}, 0, 0},
		{valuer{"1", false}, 0, 1},
		{valuer{maxUintS, false}, 0, maxUint},
		{valuer{"999999999999999999999", false}, 1, 1},
		{valuer{"one two three", false}, 1, 1},
		{valuer{"-1", true}, 1, 1},
	}
	for i, test := range tests {
		assert.Logf("test %v %d: %v and default %v", d, i, test.v, test.dv)
		bv := d.AsUint(test.v, test.dv)
		assert.Equal(bv, test.expected)
	}
}

// TestAsUInt64 checks the access of uint64 values.
func TestAsUInt64(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	maxUInt64S := strconv.FormatUint(math.MaxUint64, 10)
	d := stringex.NewDefaulter("AsUInt64", true)
	tests := []struct {
		v        valuer
		dv       uint64
		expected uint64
	}{
		{valuer{"0", false}, 0, 0},
		{valuer{"1", false}, 0, 1},
		{valuer{maxUInt64S, false}, 0, math.MaxUint64},
		{valuer{"999999999999999999999", false}, 1, 1},
		{valuer{"one two three", false}, 1, 1},
		{valuer{"-1", true}, 1, 1},
	}
	for i, test := range tests {
		assert.Logf("test %v %d: %v and default %v", d, i, test.v, test.dv)
		bv := d.AsUint64(test.v, test.dv)
		assert.Equal(bv, test.expected)
	}
}

// TestAsFloat64 checks the access of float64 values.
func TestAsFloat64(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	maxFloat64S := strconv.FormatFloat(math.MaxFloat64, 'e', -1, 64)
	minFloat64S := strconv.FormatFloat(-1*math.MaxFloat64, 'e', -1, 64)
	d := stringex.NewDefaulter("AsFloat4", true)
	tests := []struct {
		v        valuer
		dv       float64
		expected float64
	}{
		{valuer{"0.0", false}, 0.0, 0.0},
		{valuer{"1.0", false}, 0.0, 1.0},
		{valuer{"-1.0", false}, 0.0, -1.0},
		{valuer{maxFloat64S, false}, 0.0, math.MaxFloat64},
		{valuer{minFloat64S, false}, 0.0, math.MaxFloat64 * -1.0},
		{valuer{"9e+999", false}, 1.0, 1.0},
		{valuer{"-9e+999", false}, 1.0, 1.0},
		{valuer{"one.two", false}, 1.0, 1.0},
		{valuer{"1.0", true}, 2.0, 2.0},
		{valuer{"-1.0", true}, -2.0, -2.0},
	}
	for i, test := range tests {
		assert.Logf("test %v %d: %v and default %v", d, i, test.v, test.dv)
		bv := d.AsFloat64(test.v, test.dv)
		assert.Equal(bv, test.expected)
	}
}

// TestAsTime checks the access of time values.
func TestAsTime(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	y2k := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	nowStr, now := audit.BuildTime(time.RFC3339Nano, 0)

	d := stringex.NewDefaulter("AsTime", true)
	tests := []struct {
		v        valuer
		layout   string
		dv       time.Time
		expected time.Time
	}{
		{valuer{nowStr, false}, time.RFC3339Nano, y2k, now},
		{valuer{nowStr, true}, time.RFC3339Nano, y2k, y2k},
		{valuer{nowStr, false}, "any false layout", y2k, y2k},
		{valuer{"", false}, time.RFC3339Nano, y2k, y2k},
	}
	for i, test := range tests {
		assert.Logf("test %v %d: %v and default %v", d, i, test.v, test.dv)
		bv := d.AsTime(test.v, test.layout, test.dv)
		assert.Equal(bv, test.expected)
	}
}

// TestAsDuration checks the access of duration values.
func TestAsDuration(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	d := stringex.NewDefaulter("AsDuration", true)
	tests := []struct {
		v        valuer
		dv       time.Duration
		expected time.Duration
	}{
		{valuer{"1s", false}, time.Second, time.Second},
		{valuer{"1s", true}, time.Minute, time.Minute},
		{valuer{"2", false}, time.Minute, time.Minute},
		{valuer{"1 hour", false}, time.Minute, time.Minute},
		{valuer{"4711h", false}, time.Minute, 4711 * time.Hour},
	}
	for i, test := range tests {
		assert.Logf("test %v %d: %v and default %v", d, i, test.v, test.dv)
		bv := d.AsDuration(test.v, test.dv)
		assert.Equal(bv, test.expected)
	}
}

// TestDefaulterString checks the output of the defaulter as string.
func TestDefaulterString(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	d := stringex.NewDefaulter("my-id", true)
	s := d.String()

	assert.Equal(s, "Defaulter{my-id}")
}

// TestStringValuer checks the simple valuer for plain strings.
func StringValuer(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	d := stringex.NewDefaulter("StringValuer", false)
	sv := stringex.StringValuer("4711")

	assert.Equal(d.AsString(sv, "12345"), "4711")
	assert.Equal(d.AsInt(sv, 12345), 4711)

	sv = stringex.StringValuer("")

	assert.Equal(d.AsString(sv, "12345"), "12345")
	assert.Equal(d.AsInt(sv, 12345), 12345)
}

//--------------------
// HELPER
//--------------------

type valuer struct {
	value string
	err   bool
}

func (v valuer) Value() (string, error) {
	if v.err {
		return "", errors.New(v.value)
	}
	return v.value, nil
}

func (v valuer) String() string {
	if v.err {
		return "value '" + v.value + "' with error"
	}
	return "value '" + v.value + "' without error"
}

// EOF
