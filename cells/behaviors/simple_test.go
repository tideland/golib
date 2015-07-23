// Tideland Go Library - Cell Behaviors - Unit Tests - Simple
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

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/cells"
	"github.com/tideland/golib/cells/behaviors"
)

//--------------------
// TESTS
//--------------------

// TestSimpleBehavior tests the simple processor behavior.
func TestSimpleBehavior(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	env := cells.NewEnvironment("simple-procesor-behavior")
	defer env.Stop()

	topics := []string{}
	done := make(chan bool, 1)
	spf := func(ctx cells.Context, event cells.Event) error {
		topics = append(topics, event.Topic())
		if len(topics) == 3 {
			done <- true
		}
		return nil
	}
	env.StartCell("simple", behaviors.NewSimpleProcessorBehavior(spf))

	env.EmitNew("simple", "foo", "", nil)
	env.EmitNew("simple", "bar", "", nil)
	env.EmitNew("simple", "baz", "", nil)

	ok := <-done
	assert.True(ok)
}

// EOF
