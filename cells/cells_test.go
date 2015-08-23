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
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/cells"
	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/monitoring"
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
	assert.Equal(event.String(), "<topic: \"foo\" / payload: <\"default\": bar>>")

	bar, ok := event.Payload().Get(cells.DefaultPayload)
	assert.True(ok)
	assert.Equal(bar, "bar")

	_, err = cells.NewEvent("", nil, nil)
	assert.True(cells.IsNoTopicError(err))

	_, err = cells.NewEvent("yadda", nil, nil)
	assert.Nil(err)
}

// TestPayload tests the payload creation and access.
func TestPayload(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	now := time.Now()
	dur := 30 * time.Second
	pvs := cells.PayloadValues{
		"bool":     true,
		"int":      42,
		"float64":  47.11,
		"string":   "hello, world",
		"time":     now,
		"duration": dur,
	}
	pl := cells.NewPayload(pvs)

	b, ok := pl.GetBool("bool")
	assert.True(ok)
	assert.True(b)
	i, ok := pl.GetInt("int")
	assert.True(ok)
	assert.Equal(i, 42)
	f, ok := pl.GetFloat64("float64")
	assert.True(ok)
	assert.Equal(f, 47.11)
	s, ok := pl.GetString("string")
	assert.True(ok)
	assert.Equal(s, "hello, world")
	tt, ok := pl.GetTime("time")
	assert.True(ok)
	assert.Equal(tt, now)
	td, ok := pl.GetDuration("duration")
	assert.True(ok)
	assert.Equal(td, dur)

	pln := pl.Apply("foo")
	s, ok = pln.GetString(cells.DefaultPayload)
	assert.True(ok)
	assert.Equal(s, "foo")
	assert.Length(pln, 7)

	keys := pln.Keys()
	assert.Contents("string", keys)
	assert.Contents("int", keys)

	plab := cells.NewPayload(map[string]interface{}{"a": 1, "b": 2})
	plnab := pln.Apply(plab)
	assert.Length(plnab, 9)
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

	err := env.StartCell("foo", newCollectBehavior())
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

	err := env.StartCell("negative", newEventBufferBehavior(-8))
	assert.Nil(err)
	ci := cells.InspectCell(env, "negative")
	assert.Equal(ci.EventBufferSize(), cells.MinEventBufferSize)

	err = env.StartCell("low", newEventBufferBehavior(1))
	assert.Nil(err)
	ci = cells.InspectCell(env, "low")
	assert.Equal(ci.EventBufferSize(), cells.MinEventBufferSize)

	err = env.StartCell("high", newEventBufferBehavior(2*cells.MinEventBufferSize))
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

	err := env.StartCell("negative", newRecoveringFrequencyBehavior(-1, time.Second))
	assert.Nil(err)
	ci := cells.InspectCell(env, "negative")
	assert.Equal(ci.RecoveringNumber(), cells.MinRecoveringNumber)
	assert.Equal(ci.RecoveringDuration(), cells.MinRecoveringDuration)

	err = env.StartCell("low", newRecoveringFrequencyBehavior(10, time.Millisecond))
	assert.Nil(err)
	ci = cells.InspectCell(env, "low")
	assert.Equal(ci.RecoveringNumber(), cells.MinRecoveringNumber)
	assert.Equal(ci.RecoveringDuration(), cells.MinRecoveringDuration)

	err = env.StartCell("high", newRecoveringFrequencyBehavior(12, time.Minute))
	assert.Nil(err)
	ci = cells.InspectCell(env, "high")
	assert.Equal(ci.RecoveringNumber(), 12)
	assert.Equal(ci.RecoveringDuration(), time.Minute)
}

// TestBehaviorEmitTimeoutSetting tests the setting of
// the emit timeout.
func TestBehaviorEmitTimeoutSetting(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	minSeconds := int(cells.MinEmitTimeout.Seconds() / 5)
	maxSeconds := int(cells.MaxEmitTimeout.Seconds() / 5)

	env := cells.NewEnvironment("emit-timeout-setting")
	defer env.Stop()

	err := env.StartCell("low", newEmitTimeoutBehavior(time.Millisecond))
	assert.Nil(err)
	ci := cells.InspectCell(env, "low")
	assert.Equal(ci.EmitTimeout(), minSeconds)

	err = env.StartCell("correct", newEmitTimeoutBehavior(10*time.Second))
	assert.Nil(err)
	ci = cells.InspectCell(env, "correct")
	assert.Equal(ci.EmitTimeout(), 2)

	err = env.StartCell("high", newEmitTimeoutBehavior(2*cells.MaxEmitTimeout))
	assert.Nil(err)
	ci = cells.InspectCell(env, "high")
	assert.Equal(ci.EmitTimeout(), maxSeconds)
}

// TestBehaviorEmitTimeoutError tests the timeout error handling
// when one or more emit need too much time.
func TestBehaviorEmitTimeoutError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	env := cells.NewEnvironment("emit-timeout-error")
	defer env.Stop()

	err := env.StartCell("emitter", newEmitBehavior())
	assert.Nil(err)
	err = env.StartCell("sleeper", newSleepBehavior())
	assert.Nil(err)
	err = env.Subscribe("emitter", "sleeper")
	assert.Nil(err)

	// Emit more events than queue can take while the subscriber works.
	for i := 0; i < 25; i++ {
		env.EmitNew("emitter", emitTopic, i, nil)
	}

	time.Sleep(2 * time.Second)
}

// TestEnvironmentSubscribeStop subscribing and stopping
func TestEnvironmentSubscribeStop(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	env := cells.NewEnvironment("subscribe-unsubscribe-stop")
	defer env.Stop()

	assert.Nil(env.StartCell("foo", newCollectBehavior()))
	assert.Nil(env.StartCell("bar", newCollectBehavior()))
	assert.Nil(env.StartCell("baz", newCollectBehavior()))

	assert.Nil(env.Subscribe("foo", "bar", "baz"))
	assert.Nil(env.Subscribe("bar", "foo", "baz"))

	assert.Nil(env.StopCell("bar"))
	assert.Nil(env.StopCell("foo"))
	assert.Nil(env.StopCell("baz"))
}

// TestEnvironmentSubscribeUnsubscribe tests subscribing,
// checking and unsubscribing of cells.
func TestEnvironmentSubscribeUnsubscribe(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	env := cells.NewEnvironment("subscribe-unsubscribe")
	defer env.Stop()

	err := env.StartCell("foo", newCollectBehavior())
	assert.Nil(err)
	err = env.StartCell("bar", newCollectBehavior())
	assert.Nil(err)
	err = env.StartCell("baz", newCollectBehavior())
	assert.Nil(err)
	err = env.StartCell("yadda", newCollectBehavior())
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

	err := env.StartCell("foo", newCollectBehavior())
	assert.Nil(err)
	err = env.StartCell("bar", newCollectBehavior())
	assert.Nil(err)
	err = env.StartCell("baz", newCollectBehavior())
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

	err := env.StartCell("foo", newCollectBehavior())
	assert.Nil(err)
	err = env.StartCell("bar", newCollectBehavior())
	assert.Nil(err)
	err = env.StartCell("baz", newCollectBehavior())
	assert.Nil(err)

	err = env.Subscribe("foo", "bar", "baz")
	assert.Nil(err)
	err = env.EmitNew("foo", iterateTopic, nil, nil)
	assert.Nil(err)

	time.Sleep(200 * time.Millisecond)

	collected, err := env.Request("bar", cells.ProcessedTopic, nil, nil, cells.DefaultTimeout)
	assert.Nil(err)
	assert.Length(collected, 1)
	assert.Contents(`<topic: "love" / payload: <"default": foo loves bar>>`, collected)
	collected, err = env.Request("baz", cells.ProcessedTopic, nil, nil, cells.DefaultTimeout)
	assert.Nil(err)
	assert.Length(collected, 1)
	assert.Contents(`<topic: "love" / payload: <"default": foo loves baz>>`, collected)
}

// TestEnvironmentScenario tests creating and using the
// environment in a simple way.
func TestEnvironmentScenario(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	env := cells.NewEnvironment("scenario")
	defer env.Stop()

	err := env.StartCell("foo", newCollectBehavior())
	assert.Nil(err)
	err = env.StartCell("bar", newCollectBehavior())
	assert.Nil(err)
	err = env.StartCell("collector", newCollectBehavior())
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
	assert.Contents(`<topic: "lorem" / payload: <"default": 4711>>`, collected)
	assert.Contents(`<topic: "ipsum" / payload: <"default": 1234>>`, collected)
}

//--------------------
// BENCHMARKS
//--------------------

// BenchmarkSmpleEmitNullMonitoring is a simple emitting to one cell
// with the null monitor.
func BenchmarkSmpleEmitNullMonitoring(b *testing.B) {
	monitoring.SetBackend(monitoring.NewNullBackend())
	env := cells.NewEnvironment("simple-emit-null")
	defer env.Stop()

	env.StartCell("null", &nullBehavior{})

	event, _ := cells.NewEvent("foo", "bar", nil)

	for i := 0; i < b.N; i++ {
		env.Emit("null", event)
	}
}

// BenchmarkSmpleEmitStandardMonitoring is a simple emitting to one cell
// with the standard monitor.
func BenchmarkSmpleEmitStandardMonitoring(b *testing.B) {
	monitoring.SetBackend(monitoring.NewStandardBackend())
	env := cells.NewEnvironment("simple-emit-standard")
	defer env.Stop()

	env.StartCell("null", &nullBehavior{})

	event, _ := cells.NewEvent("foo", "bar", nil)

	for i := 0; i < b.N; i++ {
		env.Emit("null", event)
	}
}

// EOF
