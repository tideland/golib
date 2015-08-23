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
	"fmt"
	"sync"
	"time"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/identifier"
	"github.com/tideland/golib/logger"
	"github.com/tideland/golib/loop"
	"github.com/tideland/golib/monitoring"
	"github.com/tideland/golib/scene"
)

//--------------------
// SUBSCRIBERS
//--------------------

// subscribers manages the subscriber cells of a cell in a
// synchronized way.
type subscribers struct {
	mutex sync.RWMutex
	cells []*cell
}

// update sets the subscriber cells to the passed ones.
func (s *subscribers) update(cells []*cell) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cells = cells
}

// do executes the passed function for all subscriber cells
// and collects potential errors.
func (s *subscribers) do(f func(s Subscriber) error) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	var errs []error
	for _, sc := range s.cells {
		if err := f(sc); err != nil {
			errs = append(errs, err)
		}
	}
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return errors.Collect(errs...)
	}
}

//--------------------
// CELL
//--------------------

// cell for event processing.
type cell struct {
	env                *environment
	id                 string
	measuringID        string
	behavior           Behavior
	subscribers        *subscribers
	eventc             chan Event
	recoveringNumber   int
	recoveringDuration time.Duration
	emitTimeoutTicker  *time.Ticker
	emitTimeout        int
	loop               loop.Loop
}

// newCell create a new cell around a behavior.
func newCell(env *environment, id string, behavior Behavior) (*cell, error) {
	logger.Infof("starting cell %q", id)
	// Init cell runtime.
	c := &cell{
		env:               env,
		id:                id,
		measuringID:       identifier.Identifier("cells", env.id, "cell", id),
		behavior:          behavior,
		subscribers:       &subscribers{},
		emitTimeoutTicker: time.NewTicker(5 * time.Second),
	}
	// Set configuration.
	if bebs, ok := behavior.(BehaviorEventBufferSize); ok {
		size := bebs.EventBufferSize()
		if size < minEventBufferSize {
			size = minEventBufferSize
		}
		c.eventc = make(chan Event, size)
	} else {
		c.eventc = make(chan Event, minEventBufferSize)
	}
	if brf, ok := behavior.(BehaviorRecoveringFrequency); ok {
		number, duration := brf.RecoveringFrequency()
		if duration.Seconds()/float64(number) < 0.1 {
			number = minRecoveringNumber
			duration = minRecoveringDuration
		}
		c.recoveringNumber = number
		c.recoveringDuration = duration
	} else {
		c.recoveringNumber = minRecoveringNumber
		c.recoveringDuration = minRecoveringDuration
	}
	if bet, ok := behavior.(BehaviorEmitTimeout); ok {
		timeout := bet.EmitTimeout()
		switch {
		case timeout < minEmitTimeout:
			timeout = minEmitTimeout
		case timeout > maxEmitTimeout:
			timeout = maxEmitTimeout
		}
		c.emitTimeout = int(timeout.Seconds() / 5)
	} else {
		c.emitTimeout = int(maxEmitTimeout.Seconds() / 5)
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
	return c.SubscribersDo(func(cs Subscriber) error {
		return cs.ProcessEvent(event)
	})
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
	emitTimeoutTicks := 0
	for {
		select {
		case c.eventc <- event:
			return nil
		case <-c.loop.IsStopping():
			return errors.New(ErrInactive, errorMessages, c.id)
		case <-c.emitTimeoutTicker.C:
			emitTimeoutTicks++
			if emitTimeoutTicks > c.emitTimeout {
				op := fmt.Sprintf("emitting %q to %q", event.Topic(), c.id)
				return errors.New(ErrTimeout, errorMessages, op)
			}
		}
	}
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
	return c.subscribers.do(f)
}

// stop terminates the cell.
func (c *cell) stop() error {
	c.emitTimeoutTicker.Stop()
	err := c.loop.Stop()
	if err != nil {
		logger.Errorf("cell %q terminated with error: %v", c.id, err)
	} else {
		logger.Infof("cell %q terminated", c.id)
	}
	return err
}

// backendLoop is the backend for the processing of messages.
func (c *cell) backendLoop(l loop.Loop) error {
	totalCellsID := identifier.Identifier("cells", c.env.ID(), "total-cells")
	monitoring.IncrVariable(totalCellsID)
	defer monitoring.DecrVariable(totalCellsID)

	for {
		select {
		case <-l.ShallStop():
			return c.behavior.Terminate()
		case event := <-c.eventc:
			if event == nil {
				panic("received illegal nil event!")
			}
			measuring := monitoring.BeginMeasuring(c.measuringID)
			err := c.behavior.ProcessEvent(event)
			measuring.EndMeasuring()
			if err != nil {
				logger.Errorf("cell %q processed event %q with error: %v", c.id, event.Topic(), err)
				return err
			}
		}
	}
}

// checkRecovering checks if the cell may recover after a panic. It will
// signal an error and let the cell stop working if there have been 12 recoverings
// during the last minute or the behaviors Recover() signals, that it cannot
// handle the error.
func (c *cell) checkRecovering(rs loop.Recoverings) (loop.Recoverings, error) {
	logger.Warningf("recovering cell %q after error: %v", c.id, rs.Last().Reason)
	// Check frequency.
	if rs.Frequency(c.recoveringNumber, c.recoveringDuration) {
		err := errors.New(ErrRecoveredTooOften, errorMessages, rs.Last().Reason)
		logger.Errorf("recovering frequency of cell %q too high", c.id)
		return nil, err
	}
	// Try to recover.
	if err := c.behavior.Recover(rs.Last().Reason); err != nil {
		err := errors.Annotate(err, ErrEventRecovering, errorMessages, rs.Last().Reason)
		logger.Errorf("recovering of cell %q failed", c.id, err)
		return nil, err
	}
	logger.Infof("successfully recovered cell %q", c.id)
	return rs.Trim(c.recoveringNumber), nil
}

// EOF
