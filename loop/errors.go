// Tideland Go Library - Loop - Errors
//
// Copyright (C) 2013-2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package loop

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
	ErrLoopPanicked = iota + 1
	ErrSentinelRecovers
)

var errorMessages = errors.Messages{
	ErrLoopPanicked:     "loop panicked: %v",
	ErrSentinelRecovers: "sentinel recovers due to error: %v",
}

//--------------------
// ERROR
//--------------------

// IsSentinelRecoversError checks for a recovering error.
func IsSentinelRecoversError(err error) bool {
	return errors.IsError(err, ErrSentinelRecovers)
}

// EOF
