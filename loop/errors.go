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
	ErrInvalidLoop = iota + 1
	ErrInvalidSentinel
	ErrAlreadyManaged
)

var errorMessages = errors.Messages{
	ErrInvalidLoop:     "invalid implementation of loop, sentinel needs own",
	ErrInvalidSentinel: "loop not managed by this sentinel",
	ErrAlreadyManaged:  "loop is already managed by another sentinel",
}

// EOF
