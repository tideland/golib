// Tideland Go Libray - Cell Behaviors - Ticker
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
	"time"

	"github.com/tideland/golib/cells"
	"github.com/tideland/golib/loop"
)

//--------------------
// TICKER BEHAVIOR
//--------------------

// tickerBehavior emits events in chronological order.
type tickerBehavior struct {
	ctx      cells.Context
	duration time.Duration
	loop     loop.Loop
}

// NewTickerBehavior creates a ticker behavior.
func NewTickerBehavior(duration time.Duration) cells.Behavior {
	return &tickerBehavior{
		duration: duration,
	}
}

// Init the behavior.
func (b *tickerBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	b.loop = loop.Go(b.tickerLoop)
	return nil
}

// Terminate the behavior.
func (b *tickerBehavior) Terminate() error {
	return b.loop.Stop()
}

// PrecessEvent does nothing here.
func (b *tickerBehavior) ProcessEvent(event cells.Event) error {
	if event.Topic() == TickerTopic {
		pvs := cells.PayloadValues{
			TickerIDPayload:   b.ctx.ID(),
			TickerTimePayload: time.Now(),
		}
		b.ctx.EmitNew(TickerTopic, pvs, nil)
	}
	return nil
}

// Recover from an error. Counter will be set back to the initial counter.
func (b *tickerBehavior) Recover(err interface{}) error {
	return nil
}

// tickerLoop sends ticker events to its own process method.
func (b *tickerBehavior) tickerLoop(l loop.Loop) error {
	for {
		select {
		case <-l.ShallStop():
			return nil
		case now := <-time.After(b.duration):
			// Notify myself, action there to avoid
			// race when subscribers are updated.
			b.ctx.Environment().EmitNew(b.ctx.ID(), TickerTopic, now, nil)
		}
	}
}

// EOF
