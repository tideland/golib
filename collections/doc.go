// Tideland Go Library - Collections
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library collections package provides some typical and
// often used collection types like stacks and trees. They are implemented
// as generic collections managing empty interfaces as well as typed
// ones, e.g. for strings. They are not synchronized, so this has to
// be done by the user.
package collections

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
	return version.New(2, 0, 0, "alpha", "2015-06-10")
}

// EOF
