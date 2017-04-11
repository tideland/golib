// Tideland Go Library - RSS Feed - Errors
//
// Copyright (C) 2012-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package rss

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/errors"
)

//--------------------
// CONSTANTS
//--------------------

// Error codes of the RSS package.
const (
	ErrValidation = iota + 1
	ErrParsing
)

var errorMessages = map[int]string{
	ErrParsing: "cannot parse %s",
}

//--------------------
// ERROR CHECKING
//--------------------

// IsValidationError checks if the error signals an invalid feed.
func IsValidationError(err error) bool {
	return errors.IsError(err, ErrValidation)
}

// IsParsingError checks if the error signals a bad formatted value.
func IsParsingError(err error) bool {
	return errors.IsError(err, ErrParsing)
}

// EOF
