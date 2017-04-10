// Tideland Go Library - Etc - Errors
//
// Copyright (C) 2016-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package etc

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/errors"
)

//--------------------
// CONSTANTS
//--------------------

// Error codes of the etc package.
const (
	ErrIllegalSourceFormat = iota + 1
	ErrIllegalConfigSource
	ErrCannotReadFile
	ErrCannotPostProcess
	ErrInvalidPath
	ErrCannotSplit
	ErrCannotApply
)

var errorMessages = errors.Messages{
	ErrIllegalSourceFormat: "illegal source format",
	ErrIllegalConfigSource: "illegal source for configuration: %v",
	ErrCannotReadFile:      "cannot read configuration file %q",
	ErrCannotPostProcess:   "cannot post-process configuration: %q",
	ErrInvalidPath:         "invalid configuration path %q",
	ErrCannotSplit:         "cannot split configuration",
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
