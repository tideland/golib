// Tideland Go Library - Identifier
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package identifier

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

//--------------------
// IDENTIFIER GENERATORS
//--------------------

// LimitedSepIdentifier builds an identifier out of multiple parts,
// all as lowercase strings and concatenated with the separator
// Non letters and digits are exchanged with dashes and
// reduced to a maximum of one each. If limit is true only
// 'a' to 'z' and '0' to '9' are allowed.
func LimitedSepIdentifier(sep string, limit bool, parts ...interface{}) string {
	iparts := make([]string, 0)
	for _, p := range parts {
		tmp := strings.Map(func(r rune) rune {
			// Check letter and digit.
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				lcr := unicode.ToLower(r)
				if limit {
					// Only 'a' to 'z' and '0' to '9'.
					if lcr <= unicode.MaxASCII {
						return lcr
					} else {
						return ' '
					}
				} else {
					// Every char is allowed.
					return lcr
				}
			}
			return ' '
		}, fmt.Sprintf("%v", p))
		// Only use non-empty identifier parts.
		if ipart := strings.Join(strings.Fields(tmp), "-"); len(ipart) > 0 {
			iparts = append(iparts, ipart)
		}
	}
	return strings.Join(iparts, sep)
}

// SepIdentifier builds an identifier out of multiple parts, all
// as lowercase strings and concatenated with the separator
// Non letters and digits are exchanged with dashes and
// reduced to a maximum of one each.
func SepIdentifier(sep string, parts ...interface{}) string {
	return LimitedSepIdentifier(sep, false, parts...)
}

// Identifier works like SepIdentifier but the seperator
// is set to be a colon.
func Identifier(parts ...interface{}) string {
	return SepIdentifier(":", parts...)
}

// JoinedIdentifier builds a new identifier, joinded with the
// colon as the seperator.
func JoinedIdentifier(identifiers ...string) string {
	return strings.Join(identifiers, ":")
}

// TypeAsIdentifierPart transforms the name of the arguments type into
// a part for identifiers. It's splitted at each uppercase char,
// concatenated with dashes and transferred to lowercase.
func TypeAsIdentifierPart(i interface{}) string {
	var buf bytes.Buffer
	fullTypeName := reflect.TypeOf(i).String()
	lastDot := strings.LastIndex(fullTypeName, ".")
	typeName := fullTypeName[lastDot+1:]
	for i, r := range typeName {
		if unicode.IsUpper(r) {
			if i > 0 {
				buf.WriteRune('-')
			}
		}
		buf.WriteRune(r)
	}
	return strings.ToLower(buf.String())
}

// EOF
