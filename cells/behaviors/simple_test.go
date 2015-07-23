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
	"sync"
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
	var wg sync.WaitGroup
	spf := func(ctx cells.Context, event cells.Event) error {
		topics = append(topics, event.Topic())
		wg.Done()
		return nil
	}
	env.StartCell("simple", behaviors.NewSimpleProcessorBehavior(spf))

	wg.Add(3)
	env.EmitNew("simple", "foo", "", nil)
	env.EmitNew("simple", "bar", "", nil)
	env.EmitNew("simple", "baz", "", nil)

	wg.Wait()
	assert.Length(topics, 3)
}

// EOF
