// Tideland Go Library - Simple Markup Language
//
// Copyright (C) 2009-2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library sml package provides a very simple markup language
// using a kind of LISP like notation with curly braces.
//
// The tag only consists out of the chars 'a' to 'z', '0' to '9'
// and '-'. Also several parts of the tag can be seperated by colons.
// The package contains a kind of DOM as well as a parser and a
// processor. The latter is used e.g. for printing SML documents.
package sml

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/version"
)

//--------------------
// VERSION
//--------------------

// Version returns the version of the SML package.
func Version() version.Version {
	return version.New(3, 1, 1)
}

// EOF
