// Tideland Go Library - Generic JSON Parser - Errors
//
// Copyright (C) 2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package gjp

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/errors"
)

//--------------------
// CONSTANTS
//--------------------

// Error codes of the etc package.
const (
	ErrUnmarshalling = iota + 1
)

var errorMessages = errors.Messages{
	ErrUnmarshalling: "cannot unmarshal document",
}

// EOF
