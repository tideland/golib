// Tideland Go Library - Cell Behaviors - Unit Tests - Collector
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

// TestCollectorBehavior tests the collector behavior.
func TestCollectorBehavior(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	env := cells.NewEnvironment("collector-behavior")
	defer env.Stop()

	env.StartCell("collector", behaviors.NewCollectorBehavior(10))

	for i := 0; i < 25; i++ {
		env.EmitNew("collector", "collect", i, nil)
	}

	time.Sleep(100 * time.Millisecond)

	collected, err := env.Request("collector", cells.CollectedTopic, nil, nil, cells.DefaultTimeout)
	assert.Nil(err)
	assert.Length(collected, 10)

	err = env.EmitNew("collector", cells.ResetTopic, nil, nil)
	assert.Nil(err)

	collected, err = env.Request("collector", cells.CollectedTopic, nil, nil, cells.DefaultTimeout)
	assert.Nil(err)
	assert.Length(collected, 0)
}

// EOF
