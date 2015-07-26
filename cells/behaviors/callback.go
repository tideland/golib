// Tideland Go Library - Cell Behaviors - Callback
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
// CALLBACK BEHAVIOR
//--------------------

// CallbackFunc is a function called by the behavior when it recieves an event.
type CallbackFunc func(topic string, payload cells.Payload) error

// callbackBehavior is an event processor calling all stored functions
// if it receives an event.
type callbackBehavior struct {
	ctx           cells.Context
	callbackFuncs []CallbackFunc
}

// NewCallbackBehavior creates a behavior with a number of callback functions.
// Each time an event is received those functions are called in the same order
// they have been passed.
func NewCallbackBehavior(cbfs ...CallbackFunc) cells.Behavior {
	if len(cbfs) == 0 {
		logger.Errorf("callback created without callback functions")
	}
	return &callbackBehavior{nil, cbfs}
}

// Init the behavior.
func (b *callbackBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	return nil
}

// Terminate the behavior.
func (b *callbackBehavior) Terminate() error {
	return nil
}

// ProcessEvent calls a callback functions with the event data.
func (b *callbackBehavior) ProcessEvent(event cells.Event) error {
	for _, callbackFunc := range b.callbackFuncs {
		if err := callbackFunc(event.Topic(), event.Payload()); err != nil {
			return err
		}
	}
	return nil
}

// Recover from an error.
func (b *callbackBehavior) Recover(err interface{}) error {
	return nil
}

// EOF
