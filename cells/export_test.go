// Tideland Go Library - Cells - Unit Tests - Export
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
// CONST
//--------------------

const (
	MinEventBufferSize    = minEventBufferSize
	MinRecoveringNumber   = minRecoveringNumber
	MinRecoveringDuration = minRecoveringDuration
	MinEmitTimeout        = minEmitTimeout
	MaxEmitTimeout        = maxEmitTimeout
)

//--------------------
// CELL INSIGHT
//--------------------

// CellInsight allows to access internal information
// of a cell.
type CellInsight struct {
	c *cell
}

func InspectCell(env Environment, id string) *CellInsight {
	e := env.(*environment)
	cs, err := e.cells.cells(id)
	if err != nil {
		panic(err)
	}
	return &CellInsight{cs[0]}
}

func (ci *CellInsight) ID() string {
	return ci.c.id
}

func (ci *CellInsight) EventBufferSize() int {
	return cap(ci.c.eventc)
}

func (ci *CellInsight) RecoveringNumber() int {
	return ci.c.recoveringNumber
}

func (ci *CellInsight) RecoveringDuration() time.Duration {
	return ci.c.recoveringDuration
}

func (ci *CellInsight) EmitTimeout() int {
	return ci.c.emitTimeout
}

// EOF
