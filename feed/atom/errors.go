// Tideland Go Library - Atom Feed
//
// Copyright (C) 2012-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package atom

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
	ErrValidation = iota + 1
	ErrParsing
	ErrNoPlainText
)

var errorMessages = errors.Messages{
	ErrParsing:     "cannot parse %s",
	ErrNoPlainText: "cannot convert text element %q to plain text",
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

// IsNoPlainTextError checks if the error signals no plain content
// inside a text element.
func IsNoPlainTextError(err error) bool {
	return errors.IsError(err, ErrNoPlainText)
}

// EOF
