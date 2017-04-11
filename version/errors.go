// Tideland Go Library - Version - Errors
//
// Copyright (C) 2016-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package version

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/errors"
)

//--------------------
// CONSTANTS
//--------------------

// Error codes of the version package.
const (
	ErrIllegalVersionFormat = iota + 1
)

var errorMessages = errors.Messages{
	ErrIllegalVersionFormat: "illegal version format: %s",
}

// EOF
