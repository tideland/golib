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

// TestLocalEventQueue tests the local event queue.
func TestLocalEventQueue(t *testing.T) {
	count := 10000
	assert := audit.NewTestingAssertion(t, true)
	factory := cells.MakeLocalEventQueueFactory(16)
	queue, err := factory(nil)
	assert.Nil(err)

	for i := 0; i < count; i++ {
		event, err := cells.NewEvent("queue-test", i, nil)
		assert.Nil(err)

		assert.Nil(queue.Push(event))
	}

	for i := 0; i < count; i++ {
		_, ok := <-queue.Events()
		assert.True(ok)
	}

	select {
	case <-queue.Events():
		assert.Fail("didn't expected any queued event")
	case <-time.After(100 * time.Millisecond):
		assert.True(true)
	}

	assert.Nil(queue.Stop())
}

// BenchmarkLocalEventQueue tests the performance of the local event queue.
func BenchmarkLocalEventQueue(b *testing.B) {
	factory := cells.MakeLocalEventQueueFactory(10)
	queue, err := factory(nil)
	if err != nil {
		b.Fatalf("cannot create queue: %v", err)
	}

	for i := 0; i < b.N; i++ {
		event, _ := cells.NewEvent("queue-test", i, nil)
		queue.Push(event)
	}

	for i := 0; i < b.N; i++ {
		<-queue.Events()
	}

	queue.Stop()
}

// TestEnvironmentAddRemove tests adding, checking and
// removing of cells.
func TestEnvironmentStartStop(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	env := cells.NewEnvironment()
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

// TestEnvironmentSubscribeUnsubscribe tests subscribing,
// checking and unsubscribing of cells.
func TestEnvironmentSubscribeUnsubscribe(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	env := cells.NewEnvironment()
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

	env := cells.NewEnvironment()
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

// TestEnvironmentScenario tests creating and using the
// environment in a simple way.
func TestEnvironmentScenario(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	env := cells.NewEnvironment(cells.ID("scenario"))
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

	time.Sleep(250 * time.Millisecond)

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
	// panicTopic lets the test behavior panic to check recovering.
	panicTopic = "panic!"

	// subscribersTopic returns the current subscribers.
	subscribersTopic = "subscribers?"
)

// testBehavior implements a simple behavior used in the tests.
type testBehavior struct {
	ctx         cells.Context
	processed   []string
	recoverings int
}

// newTestBehavior creates a behavior for testing. It collects and
// re-emits all events, returns them with the topic "processed" and
// delets all collected with the topic "reset".
func newTestBehavior() cells.Behavior {
	return &testBehavior{nil, []string{}, 0}
}

func (t *testBehavior) Init(ctx cells.Context) error {
	t.ctx = ctx
	return nil
}

func (t *testBehavior) Terminate() error {
	return nil
}

func (t *testBehavior) ProcessEvent(event cells.Event) error {
	switch event.Topic() {
	case cells.ProcessedTopic:
		processed := make([]string, len(t.processed))
		copy(processed, t.processed)
		err := event.Respond(processed)
		if err != nil {
			return err
		}
	case cells.ResetTopic:
		t.processed = []string{}
	case cells.PingTopic:
		err := event.Respond(cells.PongResponse)
		if err != nil {
			return err
		}
	case panicTopic:
		panic("Ouch!")
	case subscribersTopic:
		var ids []string
		t.ctx.SubscribersDo(func(s cells.Subscriber) error {
			ids = append(ids, s.ID())
			return nil
		})
		err := event.Respond(ids)
		if err != nil {
			return err
		}
	default:
		t.processed = append(t.processed, fmt.Sprintf("%v", event))
		t.ctx.Emit(event)
	}
	return nil
}

func (t *testBehavior) Recover(r interface{}) error {
	t.recoverings++
	if t.recoverings > 5 {
		return cells.NewCannotRecoverError(t.ctx.ID(), r)
	}
	return nil
}

// EOF
