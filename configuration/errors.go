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
	ErrInvalidConfigurationPath
	ErrInvalidFormat
)

var errorMessages = errors.Messages{
	ErrIllegalSourceFormat:      "illegal source format",
	ErrIllegalConfigSource:      "illegal source for configuration: %v",
	ErrInvalidConfigurationPath: "invalid configuration path %q",
	ErrInvalidFormat:            "invalid value format of %q",
}

// EOF
