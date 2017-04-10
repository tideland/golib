// Tideland Go Library - String Extensions - Defaulter
//
// Copyright (C) 2015-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package stringex

//--------------------
// IMPORTS
//--------------------

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tideland/golib/logger"
)

//--------------------
// DEFAULTER
//--------------------

// Defaulter provides an access to valuers interpreting the strings
// as types. In case of access or conversion errors these errors
// will be logged and the passed default values are returned.
type Defaulter interface {
	fmt.Stringer

	// AsString returns the value itself, only an error will
	// return the default.
	AsString(v Valuer, dv string) string

	// AsStringSlice returns the value as slice of strings
	// separated by the passed separator.
	AsStringSlice(v Valuer, sep string, dv []string) []string

	// AsStringMap returns the value as map of strings to strings.
	// The rows are separated by the rsep, the key/values per
	// row with kvsep.
	AsStringMap(v Valuer, rsep, kvsep string, dv map[string]string) map[string]string

	// AsBool returns the value interpreted as bool. Here the
	// strings 1, t, T, TRUE, true, True are interpreted as
	// true, the strings 0, f, F, FALSE, false, False as false.
	AsBool(v Valuer, dv bool) bool

	// AsInt returns the value interpreted as int.
	AsInt(v Valuer, dv int) int

	// AsInt64 returns the value interpreted as int64.
	AsInt64(v Valuer, dv int64) int64

	// AsUint returns the value interpreted as uint.
	AsUint(v Valuer, dv uint) uint

	// AsUint64 returns the value interpreted as uint64.
	AsUint64(v Valuer, dv uint64) uint64

	// AsFloat64 returns the value interpreted as float64.
	AsFloat64(v Valuer, dv float64) float64

	// AsTime returns the value interpreted as time in
	// the given layout.
	AsTime(v Valuer, layout string, dv time.Time) time.Time

	// AsDuration returns the value interpreted as duration.
	AsDuration(v Valuer, dv time.Duration) time.Duration
}

// defaulter implements the Defaulter.
type defaulter struct {
	id  string
	log bool
}

// NewDefaulter creates a defaulter with the given settings.
func NewDefaulter(id string, log bool) Defaulter {
	return &defaulter{
		id:  id,
		log: log,
	}
}

// AsString implements Defaulter.
func (d *defaulter) AsString(v Valuer, dv string) string {
	value, err := v.Value()
	if err != nil {
		d.logValuerError(err)
		return dv
	}
	return value
}

// AsStringSlice implements Defaulter.
func (d *defaulter) AsStringSlice(v Valuer, sep string, dv []string) []string {
	value, err := v.Value()
	if err != nil {
		d.logValuerError(err)
		return dv
	}
	return strings.Split(value, sep)
}

// AsStringMap implements Defaulter.
func (d *defaulter) AsStringMap(v Valuer, rsep, kvsep string, dv map[string]string) map[string]string {
	value, err := v.Value()
	if err != nil {
		d.logValuerError(err)
		return dv
	}
	rows := strings.Split(value, rsep)
	mvalue := make(map[string]string, len(rows))
	for _, row := range rows {
		kv := strings.SplitN(row, kvsep, 2)
		if len(kv) == 2 {
			mvalue[kv[0]] = kv[1]
		} else {
			mvalue[kv[0]] = kv[0]
		}
	}
	return mvalue
}

// AsBool implements Defaulter.
func (d *defaulter) AsBool(v Valuer, dv bool) bool {
	value, err := v.Value()
	if err != nil {
		d.logValuerError(err)
		return dv
	}
	bvalue, err := strconv.ParseBool(value)
	if err != nil {
		d.logFormatError("bool", err)
		return dv
	}
	return bvalue
}

// AsInt implements Defaulter.
func (d *defaulter) AsInt(v Valuer, dv int) int {
	value, err := v.Value()
	if err != nil {
		d.logValuerError(err)
		return dv
	}
	ivalue, err := strconv.ParseInt(value, 10, 0)
	if err != nil {
		d.logFormatError("int", err)
		return dv
	}
	return int(ivalue)
}

// AsInt64 implements Defaulter.
func (d *defaulter) AsInt64(v Valuer, dv int64) int64 {
	value, err := v.Value()
	if err != nil {
		d.logValuerError(err)
		return dv
	}
	ivalue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		d.logFormatError("int64", err)
		return dv
	}
	return int64(ivalue)
}

// AsUint implements Defaulter.
func (d *defaulter) AsUint(v Valuer, dv uint) uint {
	value, err := v.Value()
	if err != nil {
		d.logValuerError(err)
		return dv
	}
	uivalue, err := strconv.ParseUint(value, 10, 0)
	if err != nil {
		d.logFormatError("uint", err)
		return dv
	}
	return uint(uivalue)
}

// AsUint64 implements Defaulter.
func (d *defaulter) AsUint64(v Valuer, dv uint64) uint64 {
	value, err := v.Value()
	if err != nil {
		d.logValuerError(err)
		return dv
	}
	uivalue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		d.logFormatError("uint64", err)
		return dv
	}
	return uint64(uivalue)
}

// AsFloat64 implements Defaulter.
func (d *defaulter) AsFloat64(v Valuer, dv float64) float64 {
	value, err := v.Value()
	if err != nil {
		d.logValuerError(err)
		return dv
	}
	fvalue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		d.logFormatError("float64", err)
		return dv
	}
	return fvalue
}

// AsTime implements Defaulter.
func (d *defaulter) AsTime(v Valuer, layout string, dv time.Time) time.Time {
	value, err := v.Value()
	if err != nil {
		d.logValuerError(err)
		return dv
	}
	tvalue, err := time.Parse(layout, value)
	if err != nil {
		d.logFormatError("time", err)
		return dv
	}
	return tvalue
}

// AsDuration implements Defaulter.
func (d *defaulter) AsDuration(v Valuer, dv time.Duration) time.Duration {
	value, err := v.Value()
	if err != nil {
		d.logValuerError(err)
		return dv
	}
	dvalue, err := time.ParseDuration(value)
	if err != nil {
		d.logFormatError("duration", err)
		return dv
	}
	return dvalue
}

// String implements fmt.Stringer.
func (d *defaulter) String() string {
	return fmt.Sprintf("Defaulter{%s}", d.id)
}

// logValuerError logs the passed valuer error if configured.
func (d *defaulter) logValuerError(err error) {
	d.logError("value returned with error", err)
}

// logFormatError logs the passed format error if configured.
func (d *defaulter) logFormatError(t string, err error) {
	d.logError(fmt.Sprintf("value has illegal format for %q", t), err)
}

// logError finally checks logging and formatting before logging an error.
func (d *defaulter) logError(format string, err error) {
	if !d.log {
		return
	}
	format += ": %v"
	if len(d.id) > 0 {
		logger.Errorf("(%s) "+format, d.id, err)
	} else {
		logger.Errorf(format, err)
	}
}

//--------------------
// STRING VALUER
//--------------------

// StringValuer implements the Valuer interface for simple strings.
type StringValuer string

// Value implements the Valuer interface.
func (sv StringValuer) Value() (string, error) {
	v := string(sv)
	if len(v) == 0 {
		return "", errors.New("[-empty-]")
	}
	return v, nil
}

// EOF
