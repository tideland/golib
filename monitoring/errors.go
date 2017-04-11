// Tideland Go Library - Monitoring
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
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

// Error codes of the monitoring package.
const (
	ErrMonitoringPanicked = iota + 1
	ErrMonitoringCannotBeRecovered
	ErrMeasuringPointNotExists
	ErrStaySetVariableNotExists
	ErrDynamicStatusNotExists
)

var errorMessages = errors.Messages{
	ErrMonitoringPanicked:          "monitoring backend panicked",
	ErrMonitoringCannotBeRecovered: "monitoring backend cannot be recovered: %v",
	ErrMeasuringPointNotExists:     "measuring point %q does not exist",
	ErrStaySetVariableNotExists:    "stay-set variable %q does not exist",
	ErrDynamicStatusNotExists:      "dynamic status %q does not exist",
}

//--------------------
// TESTING
//--------------------

// IsMonitoringPanickedError returns true, if the error signals that
// the monitoring backend panicked.
func IsMonitoringPanickedError(err error) bool {
	return errors.IsError(err, ErrMonitoringPanicked)
}

// IsMonitoringCannotBeRecoveredError returns true, if the error signals that
// the monitoring backend has panicked to often and cannot be recovered.
func IsMonitoringCannotBeRecoveredError(err error) bool {
	return errors.IsError(err, ErrMonitoringCannotBeRecovered)
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
