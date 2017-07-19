// Tideland Go Library - Generic JSON Processor - Errors
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
	ErrInvalidDocument
	ErrInvalidPart
	ErrInvalidPath
	ErrPathTooLong
	ErrProcessing
)

var errorMessages = errors.Messages{
	ErrUnmarshalling:   "cannot unmarshal document",
	ErrInvalidDocument: "invalid %s document, no internal implementation",
	ErrInvalidPart:     "invalid part '%s' of the path",
	ErrInvalidPath:     "invalid path '%s'",
	ErrPathTooLong:     "path is too long",
	ErrProcessing:      "cannot process path '%s'",
}

// EOF
