// Tideland Go Library - Monitoring
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package monitoring

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
	ErrMonitorPanicked = iota + 1
	ErrMonitorCannotBeRecovered
	ErrMeasuringPointNotExists
	ErrStaySetVariableNotExists
	ErrDynamicStatusNotExists
)

var errorMessages = errors.Messages{
	ErrMonitorPanicked:          "monitor backend panicked",
	ErrMonitorCannotBeRecovered: "monitor cannot be recovered: %v",
	ErrMeasuringPointNotExists:  "measuring point %q does not exist",
	ErrStaySetVariableNotExists: "stay-set variable %q does not exist",
	ErrDynamicStatusNotExists:   "dynamic status %q does not exist",
}

//--------------------
// TESTING
//--------------------

// IsMonitorPanickedError returns true, if the error signals that
// the monitor backend panicked.
func IsMonitorPanickedError(err error) bool {
	return errors.IsError(err, ErrMonitorPanicked)
}

// IsMonitorCannotBeRecoveredError returns true, if the error signals that
// the monitor backend has panicked to often and cannot be recovered.
func IsMonitorCannotBeRecoveredError(err error) bool {
	return errors.IsError(err, ErrMonitorCannotBeRecovered)
}

// IsMeasuringPointNotExistsError returns true, if the error signals that
// a wanted measuring point cannot be retrieved because it doesn't exists.
func IsMeasuringPointNotExistsError(err error) bool {
	return errors.IsError(err, ErrMeasuringPointNotExists)
}

// IsStaySetVariableNotExistsError returns true, if the error signals that
// a wanted stay-set variable cannot be retrieved because it doesn't exists.
func IsStaySetVariableNotExistsError(err error) bool {
	return errors.IsError(err, ErrStaySetVariableNotExists)
}

// IsDynamicStatusNotExistsError returns true, if the error signals that
// a wanted dynamic status cannot be retrieved because it doesn't exists.
func IsDynamicStatusNotExistsError(err error) bool {
	return errors.IsError(err, ErrDynamicStatusNotExists)
}

// EOF
