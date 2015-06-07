// Tideland Go Library - Simple Markup Language - Errors
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package sml

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
	ErrBuilder = iota
	ErrReader
	ErrNoRootProcessor
	ErrRegisteredPlugin
)

var errorMessages = errors.Messages{
	ErrBuilder:          "cannot build node structure: %v",
	ErrReader:           "cannot read SML document: %v",
	ErrNoRootProcessor:  "no root processor registered",
	ErrRegisteredPlugin: "plugin processor with tag %q is already registered",
}

//--------------------
// ERROR
//--------------------

// IsBuilderError checks for an error during node building.
func IsBuilderError(err error) bool {
	return errors.IsError(err, ErrBuilder)
}

// IsReaderError checks for an error during SML text reading.
func IsReaderError(err error) bool {
	return errors.IsError(err, ErrBuilder)
}

// IsNoRootProcessorError checks for an unregistered root
// processor.
func IsNoRootProcessorError(err error) bool {
	return errors.IsError(err, ErrNoRootProcessor)
}

// IsRegisteredPluginError checks for the error of an already
// registered plugin.
func IsRegisteredPluginError(err error) bool {
	return errors.IsError(err, ErrRegisteredPlugin)
}

// EOF
