// Tideland Go Library - String Extensions
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package stringex

//--------------------
// IMPORTS
//--------------------

import (
	"strings"
)

//--------------------
// VALUER
//--------------------

// Valuer describes returning a string value or an error
// if it does not exist are another access error happened.
type Valuer interface {
	// Value returns a string or a potential error during access.
	Value() (string, error)
}

//--------------------
// SPLITTER
//--------------------

// SplitFilter splits the string s by the separator
// sep and then filters the parts. Only those where f
// returns true will be part of the result. So it even
// cout be empty.
func SplitFilter(s, sep string, f func(p string) bool) []string {
	parts := strings.Split(s, sep)
	out := []string{}
	for _, part := range parts {
		if f(part) {
			out = append(out, part)
		}
	}
	return out
}

// SplitMap splits the string s by the separator
// sep and then maps the parts by the function m.
// Only those where m also returns true will be part
// of the result. So it even could be empty.
func SplitMap(s, sep string, m func(p string) (string, bool)) []string {
	parts := strings.Split(s, sep)
	out := []string{}
	for _, part := range parts {
		if mp, ok := m(part); ok {
			out = append(out, mp)
		}
	}
	return out
}

// EOF
