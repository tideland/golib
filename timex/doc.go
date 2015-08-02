// Tideland Go Library - Time Extensions
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library timex package helps when working with times and dates.
// Beside tests it contains a crontab for chronological jobs and a retry function
// to let code blocks be retried under well defined conditions regarding time and
// count.
package timex

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
	return version.New(3, 2, 0)
}

// EOF
