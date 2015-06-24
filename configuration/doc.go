// Tideland Go Library - Configuration
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library configuration package provides the reading,
// parsing, and accessing of configuration data. Different readers
// can be passed as sources for the SML formatted input.
package configuration

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
	return version.New(1, 0, 0, "alpha", "2015-06-22")
}

// EOF
