// Tideland Go Library - Cell Behaviors - Finite State Machine
//
// Copyright (C) 2010-2015 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package behaviors

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"

	"github.com/tideland/golib/cells"
)

//--------------------
// FSM BEHAVIOR
//--------------------

// FSMState is the signature of a function or method which processes
// an event and returns the following state or an error.
type FSMState func(ctx cells.Context, event cells.Event) (FSMState, error)

// FSMStatus contains information about the current status of the FSM.
type FSMStatus struct {
	Done  bool
	Error error
}

// String is specified on the Stringer interface.
func (s FSMStatus) String() string {
	return fmt.Sprintf("<FSM done: %v / error: %v>", s.Done, s.Error)
}

// fsmBehavior runs the finite state machine.
type fsmBehavior struct {
	ctx   cells.Context
	state FSMState
	done  bool
	err   error
}

// NewFSMBehavior creates a finite state machine behavior based on the
// passed initial state function. The function is called with the event
// has to return the next state, which can be the same one. In case of
// nil the stae will be transfered into a generic end state, if an error
// is returned the state is a generic error state.
func NewFSMBehavior(state FSMState) cells.Behavior {
	return &fsmBehavior{nil, state, false, nil}
}

// Init the behavior.
func (b *fsmBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	return nil
}

// Terminate the behavior.
func (b *fsmBehavior) Terminate() error {
	return nil
}

// ProcessEvent executes the state function and stores
// the returned new state.
func (b *fsmBehavior) ProcessEvent(event cells.Event) error {
	switch event.Topic() {
	case cells.StatusTopic:
		status := FSMStatus{
			Done:  b.done,
			Error: b.err,
		}
		if err := event.Respond(status); err != nil {
			return err
		}
	default:
		if b.done {
			return nil
		}
		state, err := b.state(b.ctx, event)
		if err != nil {
			b.done = true
			b.err = err
		} else if state == nil {
			b.done = true
		}
		b.state = state
	}
	return nil
}

// Recover from an error.
func (b *fsmBehavior) Recover(err interface{}) error {
	b.done = true
	b.err = cells.NewCannotRecoverError(b.ctx.ID(), err)
	return nil
}

// RequestFSMStatus retrieves the status of a FSM cell.
func RequestFSMStatus(env cells.Environment, id string) FSMStatus {
	response, err := env.Request(id, cells.StatusTopic, nil, nil, cells.DefaultTimeout)
	if err != nil {
		return FSMStatus{
			Error: err,
		}
	}
	status, ok := response.(FSMStatus)
	if !ok {
		return FSMStatus{
			Error: cells.NewInvalidResponseError(response),
		}
	}
	return status
}

// EOF
