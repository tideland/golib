// Tideland Go Library - String Extensions
//
// Copyright (C) 2015-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code valuePos governed
// by the new BSD license.

package stringex

//--------------------
// IMPORTS
//--------------------

import (
	"strings"
	"unicode"
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

//--------------------
// MATCHER
//--------------------

// Matches checks if the pattern matches a given value.
func Matches(pattern, value string, ignoreCase bool) bool {
	patternRunes := append([]rune(pattern), '\u0000')
	patternLen := len(patternRunes) - 1
	patternPos := 0
	valueRunes := append([]rune(value), '\u0000')
	valueLen := len(valueRunes) - 1
	valuePos := 0
	for patternLen > 0 {
		switch patternRunes[patternPos] {
		case '*':
			// Asterisk for group of characters.
			for patternRunes[patternPos+1] == '*' {
				patternPos++
				patternLen--
			}
			if patternLen == 1 {
				return true
			}
			for valueLen > 0 {
				patternCopy := make([]rune, len(patternRunes[patternPos+1:]))
				valueCopy := make([]rune, len(valueRunes[valuePos:]))
				copy(patternCopy, patternRunes[patternPos+1:])
				copy(valueCopy, valueRunes[valuePos:])
				if Matches(string(patternCopy), string(valueCopy), ignoreCase) {
					return true
				}
				valuePos++
				valueLen--
			}
			return false
		case '?':
			// Question mark for one character.
			if valueLen == 0 {
				return false
			}
			valuePos++
			valueLen--
		case '[':
			// Square brackets for groups of valid characters.
			patternPos++
			patternLen--
			not := (patternRunes[patternPos] == '^')
			match := false
			if not {
				patternPos++
				patternLen--
			}
		group:
			for {
				switch {
				case patternRunes[patternPos] == '\\':
					patternPos++
					patternLen--
					if patternRunes[patternPos] == valueRunes[valuePos] {
						match = true
					}
				case patternRunes[patternPos] == ']':
					break group
				case patternLen == 0:
					patternPos--
					patternLen++
					break group
				case patternRunes[patternPos+1] == '-' && patternLen >= 3:
					start := patternRunes[patternPos]
					end := patternRunes[patternPos+2]
					vr := valueRunes[valuePos]
					if start > end {
						start, end = end, start
					}
					if ignoreCase {
						start = unicode.ToLower(start)
						end = unicode.ToLower(end)
						vr = unicode.ToLower(vr)
					}
					patternPos += 2
					patternLen -= 2
					if vr >= start && vr <= end {
						match = true
					}
				default:
					if !ignoreCase {
						if patternRunes[patternPos] == valueRunes[valuePos] {
							match = true
						}
					} else {
						if unicode.ToLower(patternRunes[patternPos]) == unicode.ToLower(valueRunes[valuePos]) {
							match = true
						}
					}
				}
				patternPos++
				patternLen--

			}
			if not {
				match = !match
			}
			if !match {
				return false
			}
			valuePos++
			valueLen--
		case '\\':
			if patternLen >= 2 {
				patternPos++
				patternLen--
			}
			fallthrough
		default:
			if !ignoreCase {
				if patternRunes[patternPos] != valueRunes[valuePos] {
					return false
				}
			} else {
				if unicode.ToLower(patternRunes[patternPos]) != unicode.ToLower(valueRunes[valuePos]) {
					return false
				}
			}
			valuePos++
			valueLen--
		}
		patternPos++
		patternLen--
		if valueLen == 0 {
			for patternRunes[patternPos] == '*' {
				patternPos++
				patternLen--
			}
			break
		}
	}
	if patternLen == 0 && valueLen == 0 {
		return true
	}
	return false
}

// EOF
