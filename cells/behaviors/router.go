// Tideland Go Libray - Cell Behaviors - Router
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
// ROUTER BEHAVIOR
//--------------------

// RouterFunc is a function type determinig which subscribed
// cells shall receive the event.
type RouterFunc func(emitterID, subscriberID string, event cells.Event) (bool, error)

// routerBehavior check for each received event which subscriber will
// get it based on the router function.
type routerBehavior struct {
	ctx        cells.Context
	routerFunc RouterFunc
}

// NewRouterBehavior creates a router behavior using the passed function
// to determine to which subscriber the received event will be emitted.
func NewRouterBehavior(rf RouterFunc) cells.Behavior {
	return &routerBehavior{nil, rf}
}

// Init the behavior.
func (b *routerBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	return nil
}

// Terminate the behavior.
func (b *routerBehavior) Terminate() error {
	return nil
}

// ProcessEvent emits the event to those ids returned by the router
// function.
func (b *routerBehavior) ProcessEvent(event cells.Event) error {
	return b.ctx.SubscribersDo(func(s cells.Subscriber) error {
		ok, err := b.routerFunc(b.ctx.ID(), s.ID(), event)
		if err != nil {
			return err
		}
		if ok {
			return s.ProcessEvent(event)
		}
		return nil
	})
}

// Recover from an error.
func (b *routerBehavior) Recover(err interface{}) error {
	return nil
}

// EOF
