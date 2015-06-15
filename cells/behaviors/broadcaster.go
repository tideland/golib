// Tideland Go Library - Cell Behaviors - Broadcaster
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
// BROADCASTER BEHAVIOR
//--------------------

// broadcasterBehavior is a simple repeater.
type broadcasterBehavior struct {
	ctx cells.Context
}

// NewBroadcasterBehavior creates a broadcasting behavior that just emits every
// received event. It's intended to work as an entry point for events, which
// shall be immediately processed by several subscribers.
func NewBroadcasterBehavior() cells.Behavior {
	return &broadcasterBehavior{}
}

// Init the behavior.
func (b *broadcasterBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	return nil
}

// Terminate the behavior.
func (b *broadcasterBehavior) Terminate() error {
	return nil
}

// ProcessEvent emits the event to all subscribers.
func (b *broadcasterBehavior) ProcessEvent(event cells.Event) error {
	b.ctx.Emit(event)
	return nil
}

// Recover from an error.
func (b *broadcasterBehavior) Recover(err interface{}) error {
	return nil
}

// EOF
