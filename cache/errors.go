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
	ErrLoading
	ErrCheckOutdated
	ErrDiscard
	ErrTimeout
)

var errorMessages = errors.Messages{
	ErrSettingOptions: "cannot set option",
	ErrIllegalCache:   "illegal cache type for option",
	ErrNoLoader:       "no loader configured",
	ErrLoading:        "cannot load cacheable '%s'",
	ErrCheckOutdated:  "cannot check if '%s' is outdated",
	ErrDiscard:        "cannot discard '%s'",
	ErrTimeout:        "timeout",
}

// EOF
