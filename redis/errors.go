// Tideland Go Library - Redis Client - Errors
//
// Copyright (C) 2009-2017 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package redis

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/errors"
)

//--------------------
// CONSTANTS
//--------------------

// Error codes.
const (
	ErrInvalidConfiguration = iota
	ErrPoolLimitReached
	ErrConnectionEstablishing
	ErrConnectionBroken
	ErrInvalidResponse
	ErrServerResponse
	ErrTimeout
	ErrAuthenticate
	ErrSelectDatabase
	ErrUseSubscription
	ErrInvalidType
	ErrInvalidKey
	ErrIllegalItemIndex
	ErrIllegalItemType
)

var errorMessages = errors.Messages{
	ErrInvalidConfiguration:   "invalid configuration value in field %q: %v",
	ErrPoolLimitReached:       "connection pool limit (%d) reached",
	ErrConnectionEstablishing: "cannot establish connection",
	ErrConnectionBroken:       "cannot %s, connection is broken",
	ErrInvalidResponse:        "invalid server response: %q",
	ErrServerResponse:         "server responded error: %v",
	ErrTimeout:                "timeout waiting for response",
	ErrAuthenticate:           "cannot authenticate",
	ErrSelectDatabase:         "cannot select database",
	ErrUseSubscription:        "use subscription type for subscriptions",
	ErrInvalidType:            "invalid type conversion of \"%v\" to %q",
	ErrInvalidKey:             "invalid key %q",
	ErrIllegalItemIndex:       "item index %d is illegal for result set size %d",
	ErrIllegalItemType:        "item at index %d is no %s",
}

// EOF
