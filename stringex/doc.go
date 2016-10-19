// Tideland Go Library - String Extensions
//
// Copyright (C) 2015-2016 Frank Mueller / Tideland / Oldenburg / Germay
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library stringex package helps when working with strings.
package stringex

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/version"
)

//--------------------
// VERSION
//--------------------

// Version returns the version of the stringex package.
func Version() version.Version {
	return version.New(1, 2, 0)
}

// EOF
