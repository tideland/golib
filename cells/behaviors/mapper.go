// Tideland Go Library - Cell Behaviors - Mapper
//
// Copyright (C) 2010-2015 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package behaviors

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/cells"
)

//--------------------
// MAPPER BEHAVIOR
//--------------------

// MapFunc is a function type mapping an event to another one.
type MapFunc func(id string, event cells.Event) (cells.Event, error)

// mapperBehavior maps the received event to a new event.
type mapperBehavior struct {
	ctx     cells.Context
	mapFunc MapFunc
}

// NewMapperBehavior creates a map behavior based on the passed function.
// It emits the mapped events.
func NewMapperBehavior(mf MapFunc) cells.Behavior {
	return &mapperBehavior{nil, mf}
}

// Init the behavior.
func (b *mapperBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	return nil
}

// Terminate the behavior.
func (b *mapperBehavior) Terminate() error {
	return nil
}

// ProcessEvent maps the received event to a new one and emits it.
func (b *mapperBehavior) ProcessEvent(event cells.Event) error {
	mappedEvent, err := b.mapFunc(b.ctx.ID(), event)
	if err != nil {
		return err
	}
	if mappedEvent != nil {
		b.ctx.Emit(mappedEvent)
	}
	return nil
}

// Recover from an error.
func (b *mapperBehavior) Recover(err interface{}) error {
	return nil
}

// EOF
