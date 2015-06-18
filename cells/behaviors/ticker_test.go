// Tideland Go Libray - Cell Behaviors - Unit Tests - Ticker
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

// TestTickerBehavior tests the ticker behavior.
func TestTickerBehavior(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	env := cells.NewEnvironment("ticker-behavior")
	defer env.Stop()

	env.StartCell("ticker", behaviors.NewTickerBehavior(50*time.Millisecond))
	env.StartCell("test", behaviors.NewCollectorBehavior(10))
	env.Subscribe("ticker", "test")

	time.Sleep(125 * time.Millisecond)

	collected, err := env.Request("test", cells.CollectedTopic, nil, nil, cells.DefaultTimeout)
	assert.Nil(err)
	assert.Length(collected, 2)
}

// EOF
