// Tideland Go Library - Cell Behaviors - Unit Tests - Counter
//
// Copyright (C) 2010-2015 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package behaviors_test

//--------------------
// IMPORTS
//--------------------

import (
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/cells"
	"github.com/tideland/golib/cells/behaviors"
)

//--------------------
// TESTS
//--------------------

// TestCounterBehavior tests the counting of events.
func TestCounterBehavior(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	env := cells.NewEnvironment(cells.ID("counter-behavior"))
	defer env.Stop()

	cf := func(id string, event cells.Event) []string {
		payload, ok := event.Payload().Get(cells.DefaultPayload)
		if !ok {
			return []string{}
		}
		return payload.([]string)
	}
	env.StartCell("counter", behaviors.NewCounterBehavior(cf))

	env.EmitNew("counter", "count", []string{"a", "b"}, nil)
	env.EmitNew("counter", "count", []string{"a", "c", "d"}, nil)
	env.EmitNew("counter", "count", []string{"a", "d"}, nil)

	time.Sleep(100 * time.Millisecond)

	counters, err := env.Request("counter", cells.CountersTopic, nil, nil, cells.DefaultTimeout)
	assert.Nil(err)
	assert.Length(counters, 4, "four counted events")

	c := counters.(behaviors.Counters)

	assert.Equal(c["a"], int64(3))
	assert.Equal(c["b"], int64(1))
	assert.Equal(c["c"], int64(1))
	assert.Equal(c["d"], int64(2))

	err = env.EmitNew("counter", cells.ResetTopic, nil, nil)
	assert.Nil(err)

	counters, err = env.Request("counter", cells.CountersTopic, nil, nil, cells.DefaultTimeout)
	assert.Nil(err)
	assert.Empty(counters, "zero counted events")
}

// EOF
