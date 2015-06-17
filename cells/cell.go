// Tideland Go Library - Cells - Cell
//
// Copyright (C) 2010-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"time"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/identifier"
	"github.com/tideland/golib/logger"
	"github.com/tideland/golib/loop"
	"github.com/tideland/golib/monitoring"
	"github.com/tideland/golib/scene"
)

//--------------------
// CELL
//--------------------

// cell for event processing.
type cell struct {
	env                *environment
	id                 string
	behavior           Behavior
	subscribers        []*cell
	eventc             chan Event
	subscriberc        chan []*cell
	loop               loop.Loop
	recoveringNumber   int
	recoveringDuration time.Duration
	measuringID        string
}

// newCell create a new cell around a behavior.
func newCell(env *environment, id string, behavior Behavior, options ...Option) (*cell, error) {
	logger.Infof("starting cell %q", id)
	// Init cell runtime.
	c := &cell{
		env:         env,
		id:          id,
		behavior:    behavior,
		subscriberc: make(chan []*cell),
		measuringID: identifier.Identifier("cells", env.id, "cell", id),
	}
	// Set options.
	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}
	// Validate settings.
	if c.eventc == nil {
		c.eventc = make(chan Event, defaultEventBufferSize)
	}
	if c.recoveringNumber < defaultRecoveringNumber {
		c.recoveringNumber = defaultRecoveringNumber
	}
	if c.recoveringDuration < defaultRecoveringDuration {
		c.recoveringDuration = defaultRecoveringDuration
	}
	// Init behavior.
	if err := behavior.Init(c); err != nil {
		return nil, errors.Annotate(err, ErrCellInit, errorMessages, id)
	}
	// Start backend.
	c.loop = loop.GoRecoverable(c.backendLoop, c.checkRecovering)
	return c, nil
}

// Environment implements the Context interface.
func (c *cell) Environment() Environment {
	return c.env
}

// ID implements the Context interface.
func (c *cell) ID() string {
	return c.id
}

// Emit implements the Context interface.
func (c *cell) Emit(event Event) error {
	for _, sc := range c.subscribers {
		if err := sc.ProcessEvent(event); err != nil {
			return err
		}
	}
	return nil
}

// EmitNew implements the Context interface.
func (c *cell) EmitNew(topic string, payload interface{}, scene scene.Scene) error {
	event, err := NewEvent(topic, payload, scene)
	if err != nil {
		return err
	}
	return c.Emit(event)
}

// ProcessEvent implements the Subscriber interface.
func (c *cell) ProcessEvent(event Event) error {
	select {
	case c.eventc <- event:
	case <-c.loop.IsStopping():
		return errors.New(ErrInactive, errorMessages, c.id)
	}
	return nil
}

// ProcessNewEvent implements the Subscriber interface.
func (c *cell) ProcessNewEvent(topic string, payload interface{}, scene scene.Scene) error {
	event, err := NewEvent(topic, payload, scene)
	if err != nil {
		return err
	}
	return c.ProcessEvent(event)
}

// SubscribersDo implements the Subscriber interface.
func (c *cell) SubscribersDo(f func(s Subscriber) error) error {
	for _, sc := range c.subscribers {
		if err := f(sc); err != nil {
			return err
		}
	}
	return nil
}

// updateSubscribers sets the subscribers of the cell.
func (c *cell) updateSubscribers(cells []*cell) error {
	select {
	case c.subscriberc <- cells:
	case <-c.loop.IsStopping():
		return errors.New(ErrInactive, errorMessages, c.id)
	}
	return nil
}

// stop terminates the cell.
func (c *cell) stop() error {
	defer logger.Infof("cell %q terminated", c.id)
	return c.loop.Stop()
}

// backendLoop is the backend for the processing of messages.
func (c *cell) backendLoop(l loop.Loop) error {
	monitoring.IncrVariable(identifier.Identifier("cells", c.env.ID(), "total-cells"))
	defer monitoring.DecrVariable(identifier.Identifier("cells", c.env.ID(), "total-cells"))

	for {
		select {
		case <-l.ShallStop():
			return c.behavior.Terminate()
		case subscribers := <-c.subscriberc:
			c.subscribers = subscribers
		case event := <-c.eventc:
			if event == nil {
				panic("received illegal nil event!")
			}
			measuring := monitoring.BeginMeasuring(c.measuringID)
			err := c.behavior.ProcessEvent(event)
			if err != nil {
				c.loop.Kill(err)
				continue
			}
			measuring.EndMeasuring()
		}
	}
}

// checkRecovering checks if the cell may recover after a panic. It will
// signal an error and let the cell stop working if there have been 12 recoverings
// during the last minute or the behaviors Recover() signals, that it cannot
// handle the error.
func (c *cell) checkRecovering(rs loop.Recoverings) (loop.Recoverings, error) {
	logger.Errorf("recovering cell %q after error: %v", c.id, rs.Last().Reason)
	// Check frequency.
	if rs.Frequency(c.recoveringNumber, c.recoveringDuration) {
		return nil, errors.New(ErrRecoveredTooOften, errorMessages, rs.Last().Reason)
	}
	// Try to recover.
	if err := c.behavior.Recover(rs.Last().Reason); err != nil {
		return nil, errors.Annotate(err, ErrEventRecovering, errorMessages, rs.Last().Reason)
	}
	return rs.Trim(c.recoveringNumber), nil
}

// EOF
