// Tideland Go Library - Time Extensions
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package timex

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
	ErrCrontabCannotBeRecovered = iota + 1
	ErrRetriedTooLong
	ErrRetriedTooOften
)

var errorMessages = errors.Messages{
	ErrCrontabCannotBeRecovered: "crontab cannot be recovered: %v",
	ErrRetriedTooLong:           "retried longer than %v",
	ErrRetriedTooOften:          "retried more than %d times",
}

// EOF
