// Tideland Go Library - Cells - Event
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
	"strings"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/scene"
)

//--------------------
// PAYLOAD
//--------------------

// PayloadValues is intended to set and get the information
// of a payload as bulk.
type PayloadValues map[string]interface{}

// Payload is a write-once/read-multiple container for the
// transport of additional information with events. In case
// one item is a reference type it's in the responsibility
// of the users to avoid concurrent changes of their values.
type Payload interface {
	fmt.Stringer

	// Len returns the number of values.
	Len() int

	// Get returns one of the payload values.
	Get(key string) (interface{}, bool)

	// Keys return all keys of the payload.
	Keys() []string

	// Do iterates a function over all keys and values.
	Do(f func(key string, value interface{}) error) error

	// Apply creates a new payload containing the values
	// of this one and the passed values. Values of this
	// payload are overwritten by those which are passed
	// if they share the key.
	Apply(values interface{}) Payload
}

// payload implements the Payload interface.
type payload struct {
	values PayloadValues
}

// NewPayload creates a new payload containing the passed
// values. In case of a Payload this is used directly, in
// case of a PayloadValues or a map[string]interface{} their
// content is used, and when passing any other type the
// value is stored with the key cells.DefaultPayload.
func NewPayload(values interface{}) Payload {
	if p, ok := values.(Payload); ok {
		return p
	}
	p := &payload{
		values: PayloadValues{},
	}
	switch vs := values.(type) {
	case PayloadValues:
		for key, value := range vs {
			p.values[key] = value
		}
	case map[string]interface{}:
		for key, value := range vs {
			p.values[key] = value
		}
	default:
		p.values[DefaultPayload] = values
	}
	return p
}

// Len is specified on the Payload interface.
func (p *payload) Len() int {
	return len(p.values)
}

// Get is specified on the Payload interface.
func (p *payload) Get(key string) (interface{}, bool) {
	value, ok := p.values[key]
	return value, ok
}

// Keys is specified on the Payload interface.
func (p *payload) Keys() []string {
	keys := []string{}
	for key := range p.values {
		keys = append(keys, key)
	}
	return keys
}

// Do is specified on the Payload interface.
func (p *payload) Do(f func(key string, value interface{}) error) error {
	for key, value := range p.values {
		if err := f(key, value); err != nil {
			return err
		}
	}
	return nil
}

// Apply is specified on the Payload interface.
func (p *payload) Apply(values interface{}) Payload {
	applied := &payload{
		values: PayloadValues{},
	}
	for key, value := range p.values {
		applied.values[key] = value
	}
	switch vs := values.(type) {
	case PayloadValues:
		for key, value := range vs {
			applied.values[key] = value
		}
	case map[string]interface{}:
		for key, value := range vs {
			applied.values[key] = value
		}
	default:
		applied.values[DefaultPayload] = values
	}
	return applied
}

// String returns the payload represented as string.
func (p *payload) String() string {
	ps := []string{}
	for key, value := range p.values {
		ps = append(ps, fmt.Sprintf("<%q: %v>", key, value))
	}
	return strings.Join(ps, ", ")
}

//--------------------
// EVENT
//--------------------

// Event transports what to process.
type Event interface {
	fmt.Stringer

	// Topic returns the topic of the event.
	Topic() string

	// Payload returns the payload of the event.
	Payload() Payload

	// Scene returns a scene that is possibly emitted
	// with the event.
	Scene() scene.Scene

	// Respond responds to a request event emitted
	// with Environment.Request().
	Respond(response interface{}) error
}

// event implements the Event interface.
type event struct {
	topic   string
	payload Payload
	scene   scene.Scene
}

// NewEvent creates a new event with the given topic and payload.
func NewEvent(topic string, payload interface{}, scene scene.Scene) (Event, error) {
	if topic == "" {
		return nil, errors.New(ErrNoTopic, errorMessages)
	}
	p := NewPayload(payload)
	return &event{topic, p, scene}, nil
}

// Topic is specified on the Event interface.
func (e *event) Topic() string {
	return e.topic
}

// Payload is specified on the Event interface.
func (e *event) Payload() Payload {
	return e.payload
}

// Scene is specified on the Event interface.
func (e *event) Scene() scene.Scene {
	return e.scene
}

// Respond is specified on the Event interface.
func (e *event) Respond(response interface{}) error {
	responseChanPayload, ok := e.Payload().Get(ResponseChanPayload)
	if !ok {
		return errors.New(ErrInvalidResponseEvent, errorMessages, "no response channel")
	}
	responseChan, ok := responseChanPayload.(chan interface{})
	if !ok {
		return errors.New(ErrInvalidResponseEvent, errorMessages, "invalid response channel")
	}
	responseChan <- response
	return nil
}

// String is specified on the Stringer interface.
func (e *event) String() string {
	if e.payload == nil {
		return fmt.Sprintf("<event: %q>", e.topic)
	}
	return fmt.Sprintf("<event: %q / payload: %v>", e.topic, e.payload)
}

// EOF
