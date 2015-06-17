// Tideland Go Library - Cells
//
// Copyright (C) 2010-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"time"
)

//--------------------
// CELL OPTIONS
//--------------------

// Option allows to set an option of a cell.
type Option func(c *cell) error

// EventBufferSize is the option to set the event buffer size
// for each cell.
func EventBufferSize(size int) Option {
	return func(c *cell) error {
		if size < defaultEventBufferSize {
			size = defaultEventBufferSize
		}
		c.eventc = make(chan Event, size)
		return nil
	}
}

// CellRecoveryFrequency is the option to control number of
// accepted cell recoverings in a given time frame before
// a cell is terminated after an error.
func CellRecoveryFrequency(number int, duration time.Duration) Option {
	return func(c *cell) error {
		if number < defaultRecoveringNumber {
			number = defaultRecoveringNumber
		}
		if duration < defaultRecoveringDuration {
			duration = defaultRecoveringDuration
		}
		c.recoveringNumber = number
		c.recoveringDuration = duration
		return nil
	}
}

// EOF
