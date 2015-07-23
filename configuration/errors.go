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
	ErrCannotReadFile
	ErrInvalidPath
	ErrCannotApply
)

var errorMessages = errors.Messages{
	ErrIllegalSourceFormat: "illegal source format",
	ErrIllegalConfigSource: "illegal source for configuration: %v",
	ErrCannotReadFile:      "cannot read configuration file %q",
	ErrInvalidPath:         "invalid configuration path %q",
	ErrCannotApply:         "cannot apply values to configuration",
}

//--------------------
// ERROR CHECKING
//--------------------

// IsInvalidPathError checks if a path cannot be found.
func IsInvalidPathError(err error) bool {
	return errors.IsError(err, ErrInvalidPath)
}

// EOF
