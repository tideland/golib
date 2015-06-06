// Tideland Go Library - Web - Errors
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package web

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
	ErrDuplicateHandler = iota + 1
	ErrInitHandler
	ErrIllegalRequest
	ErrNoHandler
	ErrNoGetHandler
	ErrNoHeadHandler
	ErrNoPutHandler
	ErrNoPostHandler
	ErrNoDeleteHandler
	ErrNoOptionsHandler
	ErrMethodNotSupported
	ErrInvalidContentType
	ErrNoCachedTemplate
	ErrSceneManagement
	ErrUploadingFile
)

var errorMessages = errors.Messages{
	ErrDuplicateHandler:   "cannot register handler %q, it is already registered",
	ErrInitHandler:        "error during initialization of handler %q",
	ErrIllegalRequest:     "illegal request containing too many parts",
	ErrNoHandler:          "found no handler with ID %q, also no configured default handler with ID %q",
	ErrNoGetHandler:       "handler %q is no handler for GET requests",
	ErrNoHeadHandler:      "handler %q is no handler for HEAD requests",
	ErrNoPutHandler:       "handler %q is no handler for PUT requests",
	ErrNoPostHandler:      "handler %q is no handler for POST requests",
	ErrNoDeleteHandler:    "handler %q is no handler for DELETE requests",
	ErrNoOptionsHandler:   "handler %q is no handler for OPTIONS requests",
	ErrMethodNotSupported: "method %q is not supported",
	ErrInvalidContentType: "content type is not %q",
	ErrNoCachedTemplate:   "template %q is not cached",
	ErrSceneManagement:    "scene cannot be managed: %v",
	ErrUploadingFile:      "uploaded file cannot be handled by %q",
}

// EOF
