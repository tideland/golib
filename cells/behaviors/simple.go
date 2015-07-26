// Tideland Go Library - Cell Behaviors - Simple Processor
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
	"github.com/tideland/golib/logger"
)

//--------------------
// SIMPLE BEHAVIOR
//--------------------

// SimpleProcessorFunc is a function type doing the event processing.
type SimpleProcessorFunc func(ctx cells.Context, event cells.Event) error

// simpleBehavior is a simple event processor using the processor
// function for its own logic.
type simpleBehavior struct {
	ctx           cells.Context
	processorFunc SimpleProcessorFunc
}

// NewSimpleProcessorBehavior creates a filter behavior based on the passed function.
// Instead of an own logic and an own state it uses the passed simple processor
// function for the event processing.
func NewSimpleProcessorBehavior(spf SimpleProcessorFunc) cells.Behavior {
	if spf == nil {
		spf = func(ctx cells.Context, event cells.Event) error {
			logger.Errorf("simple processor %q used without function to handle event %v", ctx.ID(), event)
			return nil
		}
	}
	return &simpleBehavior{nil, spf}
}

// Init the behavior.
func (b *simpleBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	return nil
}

// Terminate the behavior.
func (b *simpleBehavior) Terminate() error {
	return nil
}

// ProcessEvent calls the simple processor function.
func (b *simpleBehavior) ProcessEvent(event cells.Event) error {
	return b.processorFunc(b.ctx, event)
}

// Recover from an error.
func (b *simpleBehavior) Recover(err interface{}) error {
	return nil
}

// EOF
