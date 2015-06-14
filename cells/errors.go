// Tideland Go Library - Cells - Errors
//
// Copyright (C) 2010-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cells

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
	ErrCellInit = iota + 1
	ErrCannotRecover
	ErrDuplicateID
	ErrInvalidID
	ErrExecuteID
	ErrEventRecovering
	ErrRecoveredTooOften
	ErrNoTopic
	ErrNoRequest
	ErrInactive
	ErrStopping
	ErrTimeout
	ErrMissingScene
	ErrInvalidResponseEvent
	ErrInvalidResponse
)

var errorMessages = map[int]string{
	ErrCellInit:             "cell %q cannot initialize",
	ErrCannotRecover:        "cannot recover cell %q: %v",
	ErrDuplicateID:          "cell with ID %q is already registered",
	ErrInvalidID:            "cell with ID %q does not exist",
	ErrExecuteID:            "cannot %s with cell %q",
	ErrEventRecovering:      "cell cannot recover after error %v",
	ErrRecoveredTooOften:    "cell needs too much recoverings, last error",
	ErrNoTopic:              "event has no topic",
	ErrNoRequest:            "cannot respond, event is no request",
	ErrInactive:             "cell %q is inactive",
	ErrStopping:             "%s is stopping",
	ErrTimeout:              "operation needed too long with %v",
	ErrMissingScene:         "missing scene for request",
	ErrInvalidResponseEvent: "event not valid for a response: %v",
	ErrInvalidResponse:      "request returned invalid response: %v",
}

//--------------------
// ERROR CHECKING
//--------------------

// IsCellInitError checks if an error is a cell init error.
func IsCellInitError(err error) bool {
	return errors.IsError(err, ErrCellInit)
}

// NewCannotRecoverError returns an error showing that a cell cannot
// recover from errors.
func NewCannotRecoverError(id string, err interface{}) error {
	return errors.New(ErrCannotRecover, errorMessages, id, err)
}

// IsCannotRecoverError checks if an error shows a cell that cannot
// recover.
func IsCannotRecoverError(err error) bool {
	return errors.IsError(err, ErrCannotRecover)
}

// IsDuplicateIDError checks if an error is a cell already exists error.
func IsDuplicateIDError(err error) bool {
	return errors.IsError(err, ErrDuplicateID)
}

// IsInvalidIDError checks if an error is a cell does not exist error.
func IsInvalidIDError(err error) bool {
	return errors.IsError(err, ErrInvalidID)
}

// IsEventRecoveringError checks if an error is an error recovering error.
func IsEventRecoveringError(err error) bool {
	return errors.IsError(err, ErrEventRecovering)
}

// IsRecoveredTooOftenError checks if an error is an illegal query error.
func IsRecoveredTooOftenError(err error) bool {
	return errors.IsError(err, ErrRecoveredTooOften)
}

// IsNoTopicError checks if an error shows that an event has no topic..
func IsNoTopicError(err error) bool {
	return errors.IsError(err, ErrNoTopic)
}

// IsNoRequestError checks if an error signals that an event is no request.
func IsNoRequestError(err error) bool {
	return errors.IsError(err, ErrNoRequest)
}

// IsInactiveError checks if an error is a cell inactive error.
func IsInactiveError(err error) bool {
	return errors.IsError(err, ErrInactive)
}

// IsStoppingError checks if the error shows a stopping entity.
func IsStoppingError(err error) bool {
	return errors.IsError(err, ErrStopping)
}

// IsTimeoutError checks if an error is a timeout error.
func IsTimeoutError(err error) bool {
	return errors.IsError(err, ErrTimeout)
}

// IsMissingSceneError checks if an error signals a request
// without a scene.
func IsMissingSceneError(err error) bool {
	return errors.IsError(err, ErrMissingScene)
}

// IsInvalidResponseEventError checks if an error signals an event
// used for a response but containing no storeID as payload and/or
// no scene.
func IsInvalidResponseEventError(err error) bool {
	return errors.IsError(err, ErrInvalidResponseEvent)
}

// NewInvalidResponseError returns an error showing that a
// response to a request has an illegal type.
func NewInvalidResponseError(response interface{}) error {
	return errors.New(ErrInvalidResponse, errorMessages, response)
}

// IsInvalidResponseError checks if an error signals an
// invalid response.
func IsInvalidResponseError(err error) bool {
	return errors.IsError(err, ErrInvalidResponse)
}

// EOF
