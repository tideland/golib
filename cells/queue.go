// Tideland Go Library - Cells - Queue
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
	"github.com/tideland/golib/collections"
	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/loop"
)

//--------------------
// EVENT QUEUE
//--------------------

// EventQueueFactory describes a function returning individual
// implementations of the EventQueue interface. This way different
// types of event queues can be injected into environments.
type EventQueueFactory func(env Environment) (EventQueue, error)

// EventQueue describes the methods any queue implementation
// must provide.
type EventQueue interface {
	// Push appends an event to the end of the queue.
	Push(event Event) error

	// Events returns a channel delivering the event
	// fromt the beginning of the queue.
	Events() <-chan Event

	// Stop tells the queue to end working.
	Stop() error
}

//--------------------
// LOCAL EVENT QUEUE
//--------------------

// localEventQueue implements a local in-memory event queue
// using a simple buffered go channel.
type localEventQueue struct {
	buffer collections.RingBuffer
	pushc  chan Event
	eventc chan Event
	loop   loop.Loop
}

// makeLocalEventQueueFactory creates a factory for local
// event queues.
func makeLocalEventQueueFactory(size int) EventQueueFactory {
	return func(env Environment) (EventQueue, error) {
		queue := &localEventQueue{
			buffer: collections.NewRingBuffer(size),
			pushc:  make(chan Event, 1),
			eventc: make(chan Event, 1),
		}
		queue.loop = loop.Go(queue.backendLoop)
		return queue, nil
	}
}

// Push appends an event to the end of the queue.
func (q *localEventQueue) Push(event Event) error {
	select {
	case q.pushc <- event:
	case <-q.loop.IsStopping():
		return errors.New(ErrStopping, errorMessages, "event queue")
	}
	return nil
}

// Events returns a channel delivering the event
// fromt the beginning of the queue.
func (q *localEventQueue) Events() <-chan Event {
	return q.eventc
}

// Stop tells the queue to end working.
func (q *localEventQueue) Stop() error {
	return q.loop.Stop()
}

// backendLoop realizes the backend of the queue.
func (q *localEventQueue) backendLoop(l loop.Loop) error {
	var raw interface{}
	var ok bool
	// Start receiving loop.
	for {
		// Read event from buffer.
		if !ok {
			raw, ok = q.buffer.Pop()
		}
		if ok {
			select {
			case <-l.ShallStop():
				return nil
			case event := <-q.pushc:
				// Store new received event in buffer.
				q.buffer.Push(event)
			case q.eventc <- raw.(Event):
				// Pushed event.
				ok = false
			}
			continue
		}
		// Empty buffer.
		select {
		case <-l.ShallStop():
			return nil
		case event := <-q.pushc:
			// Store new event in buffer.
			q.buffer.Push(event)
		}
	}
}

// EOF
