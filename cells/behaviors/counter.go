// Tideland Go Library - Cell Behaviors - Counter
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
// COUNTER BEHAVIOR
//--------------------

// Counters is a set of named counters and their values.
type Counters map[string]int64

// CounterFunc is the signature of a function which analyzis
// an event and returns, which counters shall be incremented.
type CounterFunc func(id string, event cells.Event) []string

// counterBehavior counts events based on the counter function.
type counterBehavior struct {
	ctx         cells.Context
	counterFunc CounterFunc
	counters    Counters
}

// NewCounterBehavior creates a counter behavior based on the passed
// function. It increments and emits those counters named by the result
// of the counter function. The counters can be retrieved with the
// request "counters?" and reset with "reset!".
func NewCounterBehavior(cf CounterFunc) cells.Behavior {
	return &counterBehavior{nil, cf, make(Counters)}
}

// Init the behavior.
func (b *counterBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	return nil
}

// Terminate the behavior.
func (b *counterBehavior) Terminate() error {
	return nil
}

// ProcessEvent counts the event for the return value of the counter func
// and emits this value.
func (b *counterBehavior) ProcessEvent(event cells.Event) error {
	switch event.Topic() {
	case cells.CountersTopic:
		response := b.copyCounters()
		if err := event.Respond(response); err != nil {
			return err
		}
	case cells.ResetTopic:
		b.counters = make(map[string]int64)
	default:
		cids := b.counterFunc(b.ctx.ID(), event)
		if cids != nil {
			for _, cid := range cids {
				v, ok := b.counters[cid]
				if ok {
					b.counters[cid] = v + 1
				} else {
					b.counters[cid] = 1
				}
				topic := "counter:" + cid
				b.ctx.EmitNew(topic, b.counters[cid], event.Scene())
			}
		}
	}
	return nil
}

// Recover from an error.
func (b *counterBehavior) Recover(err interface{}) error {
	return nil
}

// copyCounters copies the counters for a request.
func (b *counterBehavior) copyCounters() Counters {
	copiedCounters := make(Counters)
	for key, value := range b.counters {
		copiedCounters[key] = value
	}
	return copiedCounters
}

// EOF
