// Tideland Go Library - Loop - Errors
//
// Copyright (C) 2013-2016 Frank Mueller / Tideland / Oldenburg / Germany
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

const (
	ErrLoopPanicked = iota + 1
	ErrHandlingFailed
)

var errorMessages = errors.Messages{
	ErrLoopPanicked:   "loop panicked: %v",
	ErrHandlingFailed: "nadling of error notification for %q failed",
}

// EOF
