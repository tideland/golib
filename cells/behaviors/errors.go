// Tideland Go Library - Cell Behaviors - Errors
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package behaviors

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/errors"
)

//--------------------
// CONSTANTS
//--------------------

// Error codes.
const (
	ErrCannotReadConfiguration = iota + 1
	ErrCannotValidateConfiguration
)

var errorMessages = errors.Messages{
	ErrCannotReadConfiguration:     "cannot read configuration",
	ErrCannotValidateConfiguration: "configuration validation failed",
}

// EOF
