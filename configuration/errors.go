// Tideland Go Library - Configuration - Errors
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package configuration

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
	ErrIllegalSourceFormat = iota + 1
	ErrIllegalConfigSource
	ErrInvalidPath
	ErrInvalidFormat
	ErrCannotApply
)

var errorMessages = errors.Messages{
	ErrIllegalSourceFormat: "illegal source format",
	ErrIllegalConfigSource: "illegal source for configuration: %v",
	ErrInvalidPath:         "invalid configuration path %q",
	ErrInvalidFormat:       "invalid value format of %q",
	ErrCannotApply:         "cannot apply values to configuration",
}

//--------------------
// ERROR CHECKING
//--------------------

// IsInvalidPathError checks if a path cannot be found.
func IsInvalidPathError(err error) bool {
	return errors.IsError(err, ErrInvalidPath)
}

// IsInvalidFormatError checks if a value hasn't the
// expected format.
func IsInvalidFormatError(err error) bool {
	return errors.IsError(err, ErrInvalidFormat)
}

// EOF
