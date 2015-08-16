// Tideland Go Library - Cells - Unit Tests - Behaviors
//
// Copyright (C) 2010-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cells_test

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"time"

	"github.com/tideland/golib/cells"
	"github.com/tideland/golib/logger"
)

//--------------------
// TOPICS
//--------------------

const (
	// iterateTopic lets the test behavior iterate over its subscribers.
	iterateTopic = "iterate!!"

	// panicTopic lets the test behavior panic to check recovering.
	panicTopic = "panic!"

	// subscribersTopic returns the current subscribers.
	subscribersTopic = "subscribers?"

	// emitTopic tells the cell to emit a test event.
	emitTopic = "emit!"

	// sleepTopic lets the cell sleep for a longer time so the queue gets full.
	sleepTopic = "sleep!"
)

//--------------------
// TEST BEHAVIOR
//--------------------

// testBehavior implements a simple behavior used in the tests.
// It collects and re-emits all events, returns them with the
// topic "processed" and delets all collected with the
// topic "reset".
type testBehavior struct {
	ctx         cells.Context
	processed   []string
	recoverings int
}

var _ cells.Behavior = &testBehavior{}

func newTestBehavior() *testBehavior {
	return &testBehavior{nil, []string{}, 0}
}

func (b *testBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	return nil
}

func (b *testBehavior) Terminate() error {
	return nil
}

func (b *testBehavior) ProcessEvent(event cells.Event) error {
	switch event.Topic() {
	case cells.ProcessedTopic:
		processed := make([]string, len(b.processed))
		copy(processed, b.processed)
		err := event.Respond(processed)
		if err != nil {
			return err
		}
	case cells.ResetTopic:
		b.processed = []string{}
	case cells.PingTopic:
		err := event.Respond(cells.PongResponse)
		if err != nil {
			return err
		}
	case iterateTopic:
		err := b.ctx.SubscribersDo(func(s cells.Subscriber) error {
			return s.ProcessNewEvent("love", b.ctx.ID()+" loves "+s.ID(), event.Scene())
		})
		if err != nil {
			return err
		}
	case emitTopic:
		return b.ctx.EmitNew(sleepTopic, event.Payload(), event.Scene())
	case sleepTopic:
		logger.Debugf("BEHAVIOR %q SLEEPS NOW!", b.ctx.ID())
		time.Sleep(2 * time.Second)
	case panicTopic:
		panic("Ouch!")
	case subscribersTopic:
		var ids []string
		b.ctx.SubscribersDo(func(s cells.Subscriber) error {
			ids = append(ids, s.ID())
			return nil
		})
		err := event.Respond(ids)
		if err != nil {
			return err
		}
	default:
		b.processed = append(b.processed, fmt.Sprintf("%v", event))
		return b.ctx.Emit(event)
	}
	return nil
}

func (b *testBehavior) Recover(r interface{}) error {
	b.recoverings++
	if b.recoverings > 5 {
		return cells.NewCannotRecoverError(b.ctx.ID(), r)
	}
	return nil
}

// testEventBufferBehavior allows testing the setting
// of the event buffer size.
type testEventBufferBehavior struct {
	*testBehavior

	size int
}

var _ cells.BehaviorEventBufferSize = (*testEventBufferBehavior)(nil)

func newTestEventBufferBehavior(size int) cells.Behavior {
	return &testEventBufferBehavior{
		testBehavior: newTestBehavior(),
		size:         size,
	}
}

func (b *testEventBufferBehavior) EventBufferSize() int {
	return b.size
}

// testRecoveringFrequencyBehavior allows testing the setting
// of the recovering frequency.
type testRecoveringFrequencyBehavior struct {
	*testBehavior

	number   int
	duration time.Duration
}

var _ cells.BehaviorRecoveringFrequency = (*testRecoveringFrequencyBehavior)(nil)

func newTestRecoveringFrequencyBehavior(number int, duration time.Duration) cells.Behavior {
	return &testRecoveringFrequencyBehavior{
		testBehavior: newTestBehavior(),
		number:       number,
		duration:     duration,
	}
}

func (b *testRecoveringFrequencyBehavior) RecoveringFrequency() (int, time.Duration) {
	return b.number, b.duration
}

// testEmitTimeoutBehavior allows testing the setting
// of the event buffer size.
type testEmitTimeoutBehavior struct {
	*testBehavior

	timeout time.Duration
}

var _ cells.BehaviorEmitTimeout = (*testEmitTimeoutBehavior)(nil)

func newTestEmitTimeoutBehavior(timeout time.Duration) cells.Behavior {
	return &testEmitTimeoutBehavior{
		testBehavior: newTestBehavior(),
		timeout:      timeout,
	}
}

func (b *testEmitTimeoutBehavior) EmitTimeout() time.Duration {
	return b.timeout
}

//--------------------
// BEHAVIORS
//--------------------

// emitBehavior simply emits the sleep topic to its subscribers.
type emitBehavior struct {
	ctx cells.Context
}

var _ cells.Behavior = &emitBehavior{}

func newEmitBehavior() *emitBehavior {
	return &emitBehavior{}
}

func (b *emitBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	return nil
}

func (b *emitBehavior) Terminate() error {
	return nil
}

func (b *emitBehavior) ProcessEvent(event cells.Event) error {
	logger.Infof("cell %q emits sleep event", b.ctx.ID())
	return b.ctx.EmitNew(sleepTopic, event.Payload(), nil)
}

func (b *emitBehavior) Recover(r interface{}) error {
	return nil
}

func (b *emitBehavior) EmitTimeout() time.Duration {
	return 2500 * time.Millisecond
}

// sleepBehavior simply emits the sleep topic to its subscribers.
type sleepBehavior struct {
	ctx cells.Context
}

var _ cells.Behavior = &sleepBehavior{}

func newSleepBehavior() *sleepBehavior {
	return &sleepBehavior{}
}

func (b *sleepBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	return nil
}

func (b *sleepBehavior) Terminate() error {
	return nil
}

func (b *sleepBehavior) ProcessEvent(event cells.Event) error {
	logger.Infof("cell %q sleeps a bit, payload is %v", b.ctx.ID(), event.Payload())
	time.Sleep(4 * time.Second)
	return nil
}

func (b *sleepBehavior) Recover(r interface{}) error {
	return nil
}

func (b *sleepBehavior) EmitTimeout() time.Duration {
	return 2500 * time.Millisecond
}

// EOF
