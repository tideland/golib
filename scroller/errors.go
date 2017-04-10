// Tideland Go Library - Scroller
//
// Copyright (C) 2014-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package scroller

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/errors"
)

//--------------------
// CONSTANTS
//--------------------

// Error codes of the scroller package.
const (
	ErrNoSource = iota + 1
	ErrNoTarget
	ErrNegativeLines
)

var errorMessages = errors.Messages{
	ErrNoSource:      "cannot start scroller: no source",
	ErrNoTarget:      "cannot start scroller: no target",
	ErrNegativeLines: "negative number of lines not allowed: %d",
}

//--------------------
// TESTING
//--------------------

// IsNoSourceError returns true, if the error signals that
// no source has been passed.
func IsNoSourceError(err error) bool {
	return errors.IsError(err, ErrNoSource)
}

// IsNoTargetError returns true, if the error signals that
// no target has been passed.
func IsNoTargetError(err error) bool {
	return errors.IsError(err, ErrNoTarget)
}

// IsNegativeLinesError returns true, if the error shows the
// setting of a negative number of lines to start with.
func IsNegativeLinesError(err error) bool {
	return errors.IsError(err, ErrNegativeLines)
}

// EOF
