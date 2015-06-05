// Tideland Go Library - Loop
//
// Copyright (C) 2013-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library loop package is intended to support
// the developer implementing the typical Go idiom for
// concurrent applications running in a loop in the background
// and doing a select on one or more channels. Stopping those
// loops or getting aware of internal errors requires extra
// efforts. The loop package helps to control this kind of
// goroutines.
package loop

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/version"
)

//--------------------
// VERSION
//--------------------

// PackageVersion returns the version of the version package.
func PackageVersion() version.Version {
	return version.New(3, 0, 0)
}

// EOF
