// Tideland Go Library - Cells - Unit Tests
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
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/cells"
	"github.com/tideland/golib/errors"
)

//--------------------
// TESTS
//--------------------

// TestEvent tests the event construction.
func TestEvent(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	event, err := cells.NewEvent("foo", "bar", nil)
	assert.Nil(err)
	assert.Equal(event.Topic(), "foo")
	assert.Equal(event.String(), "<event: \"foo\" / payload: <\"default\": bar>>")

	bar, ok := event.Payload().Get(cells.DefaultPayload)
	assert.True(ok)
	assert.Equal(bar, "bar")

	_, err = cells.NewEvent("", nil, nil)
	assert.True(cells.IsNoTopicError(err))

	_, err = cells.NewEvent("yadda", nil, nil)
	assert.Nil(err)
}

// TestEnvironment tests general environment methods.
func TestEnvironment(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	envOne := cells.NewEnvironment("part", 1, "of", "env", "ONE")
	defer envOne.Stop()

	id := envOne.ID()
	assert.Equal(id, "part:1:of:env:one")

	envTwo := cells.NewEnvironment("environment TWO")
	defer envTwo.Stop()

	id = envTwo.ID()
	assert.Equal(id, "environment-two")
}

// TestEnvironmentStartStopCell tests starting, checking and
// stopping of cells.
func TestEnvironmentStartStopCell(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	env := cells.NewEnvironment("start-stop")
	defer env.Stop()

	err := env.StartCell("foo", newTestBehavior())
	assert.Nil(err)

	hasFoo := env.HasCell("foo")
	assert.True(hasFoo)

	err = env.StopCell("foo")
	assert.Nil(err)
	hasFoo = env.HasCell("foo")
	assert.False(hasFoo)

	hasBar := env.HasCell("bar")
	assert.False(hasBar)
	err = env.StopCell("bar")
	assert.True(errors.IsError(err, cells.ErrInvalidID))
	hasBar = env.HasCell("bar")
	assert.False(hasBar)
}

// TestBehaviorEventBufferSize tests the setting of
// the event buffer size.
func TestBehaviorEventBufferSize(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	env := cells.NewEnvironment("event-buffer")
	defer env.Stop()

	err := env.StartCell("negative", newTestEventBufferBehavior(-8))
	assert.Nil(err)
	ci := cells.InspectCell(env, "negative")
	assert.Equal(ci.EventBufferSize(), cells.MinEventBufferSize)

	err = env.StartCell("low", newTestEventBufferBehavior(1))
	assert.Nil(err)
	ci = cells.InspectCell(env, "low")
	assert.Equal(ci.EventBufferSize(), cells.MinEventBufferSize)

	err = env.StartCell("high", newTestEventBufferBehavior(2*cells.MinEventBufferSize))
	assert.Nil(err)
	ci = cells.InspectCell(env, "high")
	assert.Equal(ci.EventBufferSize(), 2*cells.MinEventBufferSize)
}

// TestBehaviorRecoveringFrequency tests the setting of
// the recovering frequency.
func TestBehaviorRecoveringFrequency(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	env := cells.NewEnvironment("recovering-frequency")
	defer env.Stop()

	err := env.StartCell("negative", newTestRecoveringFrequencyBehavior(-1, time.Second))
	assert.Nil(err)
	ci := cells.InspectCell(env, "negative")
	assert.Equal(ci.RecoveringNumber(), cells.MinRecoveringNumber)
	assert.Equal(ci.RecoveringDuration(), cells.MinRecoveringDuration)

	err = env.StartCell("low", newTestRecoveringFrequencyBehavior(10, time.Millisecond))
	assert.Nil(err)
	ci = cells.InspectCell(env, "low")
	assert.Equal(ci.RecoveringNumber(), cells.MinRecoveringNumber)
	assert.Equal(ci.RecoveringDuration(), cells.MinRecoveringDuration)

	err = env.StartCell("high", newTestRecoveringFrequencyBehavior(12, time.Minute))
	assert.Nil(err)
	ci = cells.InspectCell(env, "high")
	assert.Equal(ci.RecoveringNumber(), 12)
	assert.Equal(ci.RecoveringDuration(), time.Minute)
}

// TestBehaviorEmitTimeout tests the setting of
// the emit timeout.
func TestBehaviorEmitTimeout(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	env := cells.NewEnvironment("emit-timeout")
	defer env.Stop()

	err := env.StartCell("low", newTestEmitTimeoutBehavior(time.Millisecond))
	assert.Nil(err)
	ci := cells.InspectCell(env, "low")
	assert.Equal(ci.EmitTimeout(), cells.MinEmitTimeout)

	err = env.StartCell("correct", newTestEmitTimeoutBehavior(10*time.Second))
	assert.Nil(err)
	ci = cells.InspectCell(env, "correct")
	assert.Equal(ci.EmitTimeout(), 10*time.Second)

	err = env.StartCell("high", newTestEmitTimeoutBehavior(2*cells.MaxEmitTimeout))
	assert.Nil(err)
	ci = cells.InspectCell(env, "high")
	assert.Equal(ci.EmitTimeout(), cells.MaxEmitTimeout)
}

// TestEnvironmentSubscribeUnsubscribe tests subscribing,
// checking and unsubscribing of cells.
func TestEnvironmentSubscribeUnsubscribe(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	env := cells.NewEnvironment("subscribe-unsubscribe")
	defer env.Stop()

	err := env.StartCell("foo", newTestBehavior())
	assert.Nil(err)
	err = env.StartCell("bar", newTestBehavior())
	assert.Nil(err)
	err = env.StartCell("baz", newTestBehavior())
	assert.Nil(err)
	err = env.StartCell("yadda", newTestBehavior())
	assert.Nil(err)

	err = env.Subscribe("humpf", "foo")
	assert.True(errors.IsError(err, cells.ErrInvalidID))
	err = env.Subscribe("foo", "humpf")
	assert.True(errors.IsError(err, cells.ErrInvalidID))

	err = env.Subscribe("foo", "bar", "baz")
	assert.Nil(err)
	subs, err := env.Subscribers("foo")
	assert.Nil(err)
	assert.Contents("bar", subs)
	assert.Contents("baz", subs)

	err = env.Unsubscribe("foo", "bar")
	assert.Nil(err)
	subs, err = env.Subscribers("foo")
	assert.Nil(err)
	assert.Contents("baz", subs)

	err = env.Unsubscribe("foo", "baz")
	assert.Nil(err)
	subs, err = env.Subscribers("foo")
	assert.Nil(err)
	assert.Empty(subs)
}

// TestEnvironmentStopUnsubscribe tests the unsubscribe of a cell when
// it is stopped.
func TestEnvironmentStopUnsubscribe(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	env := cells.NewEnvironment("stop-unsubscribe")
	defer env.Stop()

	err := env.StartCell("foo", newTestBehavior())
	assert.Nil(err)
	err = env.StartCell("bar", newTestBehavior())
	assert.Nil(err)
	err = env.StartCell("baz", newTestBehavior())
	assert.Nil(err)

	err = env.Subscribe("foo", "bar", "baz")
	assert.Nil(err)

	err = env.StopCell("bar")
	assert.Nil(err)

	// Expect only baz because bar is stopped.
	response, err := env.Request("foo", subscribersTopic, nil, nil, cells.DefaultTimeout)
	assert.Nil(err)
	assert.Equal(response, []string{"baz"})
}

// TestEnvironmentSubscribersDo tests the iteration over
// the subscribers.
func TestEnvironmentSubscribersDo(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	env := cells.NewEnvironment("subscribers-do")
	defer env.Stop()

	err := env.StartCell("foo", newTestBehavior())
	assert.Nil(err)
	err = env.StartCell("bar", newTestBehavior())
	assert.Nil(err)
	err = env.StartCell("baz", newTestBehavior())
	assert.Nil(err)

	err = env.Subscribe("foo", "bar", "baz")
	assert.Nil(err)
	err = env.EmitNew("foo", iterateTopic, nil, nil)
	assert.Nil(err)

	time.Sleep(200 * time.Millisecond)

	collected, err := env.Request("bar", cells.ProcessedTopic, nil, nil, cells.DefaultTimeout)
	assert.Nil(err)
	assert.Length(collected, 1)
	assert.Contents(`<event: "love" / payload: <"default": foo loves bar>>`, collected)
	collected, err = env.Request("baz", cells.ProcessedTopic, nil, nil, cells.DefaultTimeout)
	assert.Nil(err)
	assert.Length(collected, 1)
	assert.Contents(`<event: "love" / payload: <"default": foo loves baz>>`, collected)
}

// TestEnvironmentScenario tests creating and using the
// environment in a simple way.
func TestEnvironmentScenario(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	env := cells.NewEnvironment("scenario")
	defer env.Stop()

	err := env.StartCell("foo", newTestBehavior())
	assert.Nil(err)
	err = env.StartCell("bar", newTestBehavior())
	assert.Nil(err)
	err = env.StartCell("collector", newTestBehavior())
	assert.Nil(err)

	err = env.Subscribe("foo", "bar")
	assert.Nil(err)
	err = env.Subscribe("bar", "collector")
	assert.Nil(err)

	err = env.EmitNew("foo", "lorem", 4711, nil)
	assert.Nil(err)
	err = env.EmitNew("foo", "ipsum", 1234, nil)
	assert.Nil(err)
	response, err := env.Request("foo", cells.PingTopic, nil, nil, cells.DefaultTimeout)
	assert.Nil(err)
	assert.Equal(response, cells.PongResponse)

	time.Sleep(200 * time.Millisecond)

	collected, err := env.Request("collector", cells.ProcessedTopic, nil, nil, cells.DefaultTimeout)
	assert.Nil(err)
	assert.Length(collected, 2, "two collected events")
	assert.Contents(`<event: "lorem" / payload: <"default": 4711>>`, collected)
	assert.Contents(`<event: "ipsum" / payload: <"default": 1234>>`, collected)
}

//--------------------
// HELPERS
//--------------------

const (
	// iterateTopic lets the test behavior iterate over its subscribers.
	iterateTopic = "iterate!!"

	// panicTopic lets the test behavior panic to check recovering.
	panicTopic = "panic!"

	// subscribersTopic returns the current subscribers.
	subscribersTopic = "subscribers?"
)

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
		b.ctx.Emit(event)
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

// EOF
