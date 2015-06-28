// Tideland Go Library - Collections - Ring Buffer
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package collections

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"strings"
)

//--------------------
// RING BUFFER
//--------------------

// valueLink is one ring buffer element containing one
// value and linking to the next element.
type valueLink struct {
	used  bool
	value interface{}
	next  *valueLink
}

// ringBuffer implements the RingBuffer interface.
type ringBuffer struct {
	start   *valueLink
	end     *valueLink
	current *valueLink
}

// NewRingBuffer creates a new ring buffer.
func NewRingBuffer(size int) RingBuffer {
	rb := &ringBuffer{}
	rb.start = &valueLink{}
	rb.end = rb.start
	if size < 2 {
		size = 2
	}
	for i := 0; i < size-1; i++ {
		link := &valueLink{}
		rb.end.next = link
		rb.end = link
	}
	rb.end.next = rb.start
	return rb
}

// Push implements the RingBuffer interface.
func (rb *ringBuffer) Push(values ...interface{}) {
	for _, value := range values {
		if rb.end.next.used == false {
			rb.end.next.used = true
			rb.end.next.value = value
			rb.end = rb.end.next
			continue
		}
		link := &valueLink{
			used:  true,
			value: value,
			next:  rb.start,
		}
		rb.end.next = link
		rb.end = rb.end.next
	}
}

// Pop implements the RingBuffer interface.
func (rb *ringBuffer) Pop() (interface{}, bool) {
	if rb.start.used == false {
		return nil, false
	}
	value := rb.start.value
	rb.start.used = false
	rb.start.value = nil
	rb.start = rb.start.next
	return value, true
}

// Len implements the RingBuffer interface.
func (rb *ringBuffer) Len() int {
	l := 0
	current := rb.start
	for current.used {
		l++
		current = current.next
		if current == rb.start {
			break
		}
	}
	return l
}

// Cap implements the RingBuffer interface.
func (rb *ringBuffer) Cap() int {
	c := 1
	current := rb.start
	for current.next != rb.start {
		c++
		current = current.next
	}
	return c
}

// String implements the Stringer interface.
func (rb *ringBuffer) String() string {
	vs := []string{}
	current := rb.start
	for current.used {
		vs = append(vs, fmt.Sprintf("[%v]", current.value))
		current = current.next
		if current == rb.start {
			break
		}
	}
	return strings.Join(vs, "->")
}

// EOF
