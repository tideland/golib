// Tideland Go Library - String Extensions
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package stringex

//--------------------
// IMPORTS
//--------------------

import ()

//--------------------
// VALUER
//--------------------

// Valuer describes returning a string value or an error
// if it does not exist are another access error happened.
type Valuer interface {
	// Value returns a string or a potential error during access.
	Value() (string, error)
}

// EOF
