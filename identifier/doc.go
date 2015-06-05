// Tideland Go Library - Identifier
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library identifier package provides different ways
// to produce identifiers like e.g. UUIDs.
//
// The UUID generation can be done according the versions 1, 3, 4, and 5.
// Other identifier types are based on passed data or types. Here
// the individual parts are harmonized and concatenated by the
// passed seperators. It is the users responsibility to check if
// the identifier is unique in its context.
package identifier

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
