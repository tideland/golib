// Tideland Go Library - Scene
//
// Copyright (C) 2014-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package scene

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
	ErrSceneEnded = iota + 1
	ErrTimeout
	ErrPropAlreadyExist
	ErrPropNotFound
	ErrCleanupFailed
	ErrWaitedTooLong
)

var errorMessages = errors.Messages{
	ErrSceneEnded:       "scene already ended",
	ErrTimeout:          "scene %s timeout reached at %v",
	ErrPropAlreadyExist: "prop %q already exist",
	ErrPropNotFound:     "prop %q does not exist",
	ErrCleanupFailed:    "cleanup of prop %q failed",
	ErrWaitedTooLong:    "waiting for signal %q timed out",
}

//--------------------
// TESTING
//--------------------

// IsSceneEndedError returns true, if the error signals that
// the scene isn't active anymore.
func IsSceneEndedError(err error) bool {
	return errors.IsError(err, ErrSceneEnded)
}

// IsTimeoutError returns true, if the error signals that
// the scene end after an absolute timeout.
func IsTimeoutError(err error) bool {
	return errors.IsError(err, ErrTimeout)
}

// IsPropAlreadyExistError returns true, if the error signals a
// double prop key.
func IsPropAlreadyExistError(err error) bool {
	return errors.IsError(err, ErrPropAlreadyExist)
}

// IsPropNotFoundError returns true, if the error signals a
// non-existing prop.
func IsPropNotFoundError(err error) bool {
	return errors.IsError(err, ErrPropNotFound)
}

// IsCleanupFaildError returns true, if the error signals the
// failing of a prop error.
func IsCleanupFailedError(err error) bool {
	return errors.IsError(err, ErrCleanupFailed)
}

// IsWaitedTooLongError returns true, if the error signals a
// timeout when waiting for a signal.
func IsWaitedTooLongError(err error) bool {
	return errors.IsError(err, ErrWaitedTooLong)
}

// EOF
