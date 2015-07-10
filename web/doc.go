// Tideland Go Library - Web
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library web package provides a framework for a component based web
// development, especially following the principles of REST. Internally it uses the
// standard http, template, json and xml packages. The business logic has to be
// implemented in components that fullfill the individual handler interfaces. They
// work on a context with some helpers but also have got access to the original
// Request and ResponseWriter arguments.
package web

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
	return version.New(4, 0, 1)
}

// EOF
