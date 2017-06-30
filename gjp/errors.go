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
	ErrInvalidDocument
	ErrSetting
	ErrLeafToNode
	ErrNodeToLeaf
	ErrUnsupportedType
	ErrCorrupting
	ErrProcessing
)

var errorMessages = errors.Messages{
	ErrUnmarshalling:   "cannot unmarshal document",
	ErrInvalidDocument: "invalid %s document, no internal implementation",
	ErrSetting:         "failed setting the node '%s'",
	ErrLeafToNode:      "cannot convert leaf to node",
	ErrNodeToLeaf:      "cannot convert node to leaf",
	ErrUnsupportedType: "builder does not support type: %v",
	ErrCorrupting:      "setting a value on a node would corrupt the document",
	ErrProcessing:      "failed processing the node '%s'",
}

// EOF
