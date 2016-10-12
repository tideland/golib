// Tideland Go Library - Etc
//
// Copyright (C) 2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library etc configuration package provides the reading,
// parsing, and accessing of configuration data. Different readers
// can be passed as sources for the SML formatted input. If values
// contain templates formatted [path||default] the configuration tries
// to read the value out of the given path and replace the template.
// The defaiult value is optional. It will be used, if the path cannot
// be found.
package etc

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
func PackageVersion() version.Version {
	return version.New(1, 6, 0)
}

// EOF
