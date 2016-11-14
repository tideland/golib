// Tideland Go Library
//
// Copyright (C) 2014-2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package library

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/version"
)

//--------------------
// VERSION
//--------------------

// Version returns the Go Library version.
func Version() Version {
	return New(4, 13, 0, "alpha", "2016-11-14")
}

// EOF
