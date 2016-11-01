// Tideland Go Library - Logger
//
// Copyright (C) 2012-2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library logger package provides a flexible way
// to log information with different levels and on different backends.
package logger

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/version"
)

//--------------------
// VERSION
//--------------------

// Version returns the version of the logger package.
func Version() version.Version {
	return version.New(4, 3, 0)
}

// EOF
