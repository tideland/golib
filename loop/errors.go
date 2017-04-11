// Tideland Go Library - Loop - Errors
//
// Copyright (C) 2013-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package loop

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/errors"
)

//--------------------
// CONSTANTS
//--------------------

// Error codes of the loop package.
const (
	ErrLoopPanicked = iota + 1
	ErrHandlingFailed
	ErrRestartNonStopped
	ErrKilledBySentinel
)

var errorMessages = errors.Messages{
	ErrLoopPanicked:      "loop panicked: %v",
	ErrHandlingFailed:    "error handling for %q failed",
	ErrRestartNonStopped: "cannot restart unstopped %q",
	ErrKilledBySentinel:  "%q killed by sentinel",
}

//--------------------
// TESTING
//--------------------

// IsKilledBySentinelError allows to check, if a loop or
// sentinel has been stopped due to internal reasons or
// after the error of another loop or sentinel.
func IsKilledBySentinelError(err error) bool {
	return errors.IsError(err, ErrKilledBySentinel)
}

// EOF
