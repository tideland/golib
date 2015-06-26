// Tideland Go Library - Cell Behaviors - Unit Tests - Scene
//
// Copyright (C) 2015 Frank Mueller / Oldenburg / Germany
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
	"github.com/tideland/golib/scene"
)

//--------------------
// TESTS
//--------------------

// TestSceneBehavior tests the scene behavior.
func TestSceneBehavior(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	env := cells.NewEnvironment("scene-behavior")
	defer env.Stop()

	env.StartCell("broadcast", behaviors.NewBroadcasterBehavior())
	env.StartCell("scene", behaviors.NewSceneBehavior())
	env.Subscribe("broadcast", "scene")

	scn := scene.Start()
	defer scn.Stop()

	env.EmitNew("broadcast", "foo", "bar", scn)
	value, err := scn.WaitFlagLimitedAndFetch("foo", 5*time.Second)
	assert.Nil(err)
	assert.Equal(value, cells.NewPayload("bar"))

	env.EmitNew("broadcast", "yadda", 42, nil)
	value, err = scn.WaitFlagLimitedAndFetch("yadda", 1*time.Second)
	assert.Nil(value)
	assert.ErrorMatch(err, `.* waiting for signal "yadda" timed out`)
}

// EOF
