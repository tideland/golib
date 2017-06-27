// Tideland Go Library - String Extensions
//
// Copyright (C) 2015-2017 Frank Mueller / Tideland / Oldenburg / Germay
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// Package stringex of the Tideland Go Library helps when working with
// strings. So SplitFilter() and SplitMap() split given strings by a
// separator and user defined functions are called for each part to
// filter or map those.
//
// Matches() provides a more simple string matching than regular
// expressions. Patterns are ? for one char, * for multiple chars,
// and [aeiou] or [0-9] for group or ranges of chars. Both latter
// can be negotiated with [^abc] while the pattern chars also can
// be escaped with \.
//
// While the Valuer defines the interface to anything that may
// return a value as string the Default helps to interpret these
// strings as other data types. In case they don't match a default
// value will be returned.
//
// Processor defines an interface for the processing of strings.
// Those easily can be chained or used for stream splitting again
// working with processor chains. Some processors are already
// pre-defined.
package stringex

// EOF
