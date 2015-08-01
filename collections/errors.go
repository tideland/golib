// Tideland Go Library - Collections
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package collections

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
	ErrEmpty = iota + 1
	ErrNilValue
	ErrDuplicate
	ErrIllegalPath
	ErrNodeAddChild
	ErrCannotRemoveRoot
	ErrNodeNotFound
	ErrNodeFindFirst
	ErrNodeFindAll
	ErrNodeDoAll
	ErrNodeDoChildren
	ErrFindAll
	ErrDoAll
)

var errorMessages = errors.Messages{
	ErrEmpty:            "collection is empty",
	ErrNilValue:         "cannot add nil value",
	ErrDuplicate:        "duplicates are not allowed",
	ErrIllegalPath:      "cannot naviragte to the wanted node",
	ErrNodeAddChild:     "cannot add child node",
	ErrCannotRemoveRoot: "cannot remove root",
	ErrNodeNotFound:     "node not found",
	ErrNodeFindFirst:    "cannot find first node",
	ErrNodeFindAll:      "cannot find all matching nodes",
	ErrNodeDoAll:        "cannot perform function on all nodes",
	ErrNodeDoChildren:   "cannot perform function on child nodes",
	ErrFindAll:          "cannot find all matching values",
	ErrDoAll:            "cannot perform function on all values",
}

//--------------------
// CHECKERS
//--------------------

// IsNodeNotFoundError checks if the error signals that a node
// cannot be found.
func IsNodeNotFoundError(err error) bool {
	return errors.IsError(err, ErrNodeNotFound)
}

// EOF
