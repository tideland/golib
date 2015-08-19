// Tideland Go Library - Monitoring
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library monitoring package supports three kinds of
// system monitoring. They are helpful to understand what's happening
// inside a system during runtime. So execution times can be measured
// and analyzed, stay-set variables integrated and dynamic control
// value retrieval provided. The backend is exchangeable and the whole
// monitoring can be turned off.
package monitoring

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
	return version.New(4, 0, 0)
}

// EOF
