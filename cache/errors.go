// Tideland Go Library - Cache
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cache

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/errors"
)

//--------------------
// CONSTANTS
//--------------------

// Errors of the Cache.
const (
	ErrSettingOptions = iota + 1
	ErrIllegalCache
	ErrNoLoader
	ErrTimeout
)

var errorMessages = errors.Messages{
	ErrSettingOptions: "cannot set option",
	ErrIllegalCache:   "illegal cache type for option",
	ErrNoLoader:       "no loader configured",
	ErrTimeout:        "timeout",
}

// EOF
