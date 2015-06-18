// Tideland Go Library - Cell Behaviors - Unit Tests - Finite State Machine
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
	"fmt"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/cells"
	"github.com/tideland/golib/cells/behaviors"
)

//--------------------
// TESTS
//--------------------

// TestFSMBehavior tests the finite state machine behavior.
func TestFSMBehavior(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	env := cells.NewEnvironment("fsm-behavior")
	defer env.Stop()

	checkCents := func(id string) int {
		cents, err := env.Request(id, "cents?", nil, nil, cells.DefaultTimeout)
		assert.Nil(err)
		return cents.(int)
	}
	info := func(id string) string {
		info, err := env.Request(id, "info?", nil, nil, cells.DefaultTimeout)
		assert.Nil(err)
		return info.(string)
	}
	grabCents := func() int {
		cents, err := env.Request("restorer", "grab!", nil, nil, cells.DefaultTimeout)
		assert.Nil(err)
		return cents.(int)
	}

	lockA := lockMachine{}
	lockB := lockMachine{}

	env.StartCell("lock-a", behaviors.NewFSMBehavior(lockA.Locked))
	env.StartCell("lock-b", behaviors.NewFSMBehavior(lockB.Locked))
	env.StartCell("restorer", newRestorerBehavior())

	env.Subscribe("lock-a", "restorer")
	env.Subscribe("lock-b", "restorer")

	// 1st run: emit not enough and press button.
	env.EmitNew("lock-a", "coin!", 20, nil)
	env.EmitNew("lock-a", "coin!", 20, nil)
	env.EmitNew("lock-a", "coin!", 20, nil)
	env.EmitNew("lock-a", "button-press!", nil, nil)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(checkCents("lock-a"), 0)
	assert.Equal(grabCents(), 60)

	// 2nd run: unlock the lock and lock it again.
	env.EmitNew("lock-a", "coin!", 50, nil)
	env.EmitNew("lock-a", "coin!", 20, nil)
	env.EmitNew("lock-a", "coin!", 50, nil)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(info("lock-a"), "state 'unlocked' with 20 cents")

	env.EmitNew("lock-a", "button-press!", nil, nil)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(checkCents("lock-a"), 00)
	assert.Equal(info("lock-a"), "state 'locked' with 0 cents")
	assert.Equal(grabCents(), 20)

	// 3rd run: put a screwdriwer in the lock.
	env.EmitNew("lock-a", "screwdriver!", nil, nil)

	time.Sleep(100 * time.Millisecond)

	status := behaviors.RequestFSMStatus(env, "lock-a")
	assert.Equal(status.Done, true)
	assert.Nil(status.Error)

	// 4th run: try an illegal action.
	env.EmitNew("lock-b", "chewing-gum", nil, nil)

	time.Sleep(100 * time.Millisecond)

	status = behaviors.RequestFSMStatus(env, "lock-b")
	assert.Equal(status.Done, true)
	assert.ErrorMatch(status.Error, "illegal topic in state 'locked': chewing-gum")
}

//--------------------
// HELPERS
//--------------------

// cents retrieves the cents out of the payload of an event.
func payloadCents(event cells.Event) int {
	cents, ok := event.Payload().Get(cells.DefaultPayload)
	if !ok {
		return -1
	}
	return cents.(int)
}

// lockMachine will be unlocked if enough money is inserted. After
// that it can be locked again.
type lockMachine struct {
	cents int
}

// Locked represents the locked state receiving coins.
func (m *lockMachine) Locked(ctx cells.Context, event cells.Event) (behaviors.FSMState, error) {
	switch event.Topic() {
	case "cents?":
		return m.Locked, event.Respond(m.cents)
	case "info?":
		info := fmt.Sprintf("state 'locked' with %d cents", m.cents)
		return m.Locked, event.Respond(info)
	case "coin!":
		cents := payloadCents(event)
		if cents < 1 {
			return nil, fmt.Errorf("do not insert buttons")
		}
		m.cents += cents
		if m.cents > 100 {
			m.cents -= 100
			return m.Unlocked, nil
		}
		return m.Locked, nil
	case "button-press!":
		if m.cents > 0 {
			ctx.Environment().EmitNew("restorer", "drop!", m.cents, event.Scene())
			m.cents = 0
		}
		return m.Locked, nil
	case "screwdriver!":
		// Allow a screwdriver to bring the lock into an undefined state.
		return nil, nil
	}
	return m.Locked, fmt.Errorf("illegal topic in state 'locked': %s", event.Topic())
}

// Unlocked represents the unlocked state receiving coins.
func (m *lockMachine) Unlocked(ctx cells.Context, event cells.Event) (behaviors.FSMState, error) {
	switch event.Topic() {
	case "cents?":
		return m.Unlocked, event.Respond(m.cents)
	case "info?":
		info := fmt.Sprintf("state 'unlocked' with %d cents", m.cents)
		return m.Unlocked, event.Respond(info)
	case "coin!":
		cents := payloadCents(event)
		ctx.EmitNew("return", cents, event.Scene())
		return m.Unlocked, nil
	case "button-press!":
		ctx.Environment().EmitNew("restorer", "drop!", m.cents, event.Scene())
		m.cents = 0
		return m.Locked, nil
	}
	return m.Unlocked, fmt.Errorf("illegal topic in state 'unlocked': %s", event.Topic())
}

type restorerBehavior struct {
	ctx   cells.Context
	cents int
}

func newRestorerBehavior() cells.Behavior {
	return &restorerBehavior{nil, 0}
}

func (b *restorerBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	return nil
}

func (b *restorerBehavior) Terminate() error {
	return nil
}

func (b *restorerBehavior) ProcessEvent(event cells.Event) error {
	switch event.Topic() {
	case "grab!":
		cents := b.cents
		b.cents = 0
		return event.Respond(cents)
	case "drop!":
		b.cents += payloadCents(event)
	}
	return nil
}

func (b *restorerBehavior) Recover(err interface{}) error {
	return nil
}

// EOF
