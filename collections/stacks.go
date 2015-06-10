// Tideland Go Library - Collections - Stacks
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

	"github.com/tideland/golib/errors"
)

//--------------------
// STACK
//--------------------

// stack implements the Stack interface.
type stack struct {
	values []interface{}
}

// NewStack creates a stack with the passed values
// as initial content.
func NewStack(vs ...interface{}) Stack {
	return &stack{
		values: vs,
	}
}

// Push implements the Stack interface.
func (s *stack) Push(vs ...interface{}) {
	s.values = append(s.values, vs...)
}

// Pop implements the Stack interface.
func (s *stack) Pop() (interface{}, error) {
	lv := len(s.values)
	if lv == 0 {
		return nil, errors.New(ErrEmpty, errorMessages)
	}
	v := s.values[lv-1]
	s.values = s.values[:lv-1]
	return v, nil
}

// Peek implements the Stack interface.
func (s stack) Peek() (interface{}, error) {
	lv := len(s.values)
	if lv == 0 {
		return nil, errors.New(ErrEmpty, errorMessages)
	}
	v := s.values[lv-1]
	return v, nil
}

// All implements the Stack interface.
func (s *stack) All() []interface{} {
	sl := len(s.values)
	all := make([]interface{}, sl)
	copy(all, s.values)
	return all
}

// AllReverse implements the Stack interface.
func (s *stack) AllReverse() []interface{} {
	sl := len(s.values)
	all := make([]interface{}, sl)
	for i, value := range s.values {
		all[sl-1-i] = value
	}
	return all
}

// Len implements the Stack interface.
func (s *stack) Len() int {
	return len(s.values)
}

// Deflate implements the Stack interface.
func (s *stack) Deflate() {
	s.values = []interface{}{}
}

// Deflate implements the Stringer interface.
func (s *stack) String() string {
	return fmt.Sprintf("%v", s.values)
}

//--------------------
// STRING STACK
//--------------------

// stringStack implements the StringStack interface.
type stringStack struct {
	values []string
}

// NewStringStack creates a string stack with the passed values
// as initial content.
func NewStringStack(vs ...string) StringStack {
	return &stringStack{
		values: vs,
	}
}

// Push implements the StringStack interface.
func (s *stringStack) Push(vs ...string) {
	s.values = append(s.values, vs...)
}

// Pop implements the StringStack interface.
func (s *stringStack) Pop() (string, error) {
	lv := len(s.values)
	if lv == 0 {
		return "", errors.New(ErrEmpty, errorMessages)
	}
	v := s.values[lv-1]
	s.values = s.values[:lv-1]
	return v, nil
}

// Peek implements the StringStack interface.
func (s *stringStack) Peek() (string, error) {
	lv := len(s.values)
	if lv == 0 {
		return "", errors.New(ErrEmpty, errorMessages)
	}
	v := s.values[lv-1]
	return v, nil
}

// All implements the StringStack interface.
func (s *stringStack) All() []string {
	sl := len(s.values)
	all := make([]string, sl)
	copy(all, s.values)
	return all
}

// AllReverse implements the StringStack interface.
func (s *stringStack) AllReverse() []string {
	sl := len(s.values)
	all := make([]string, sl)
	for i, value := range s.values {
		all[sl-1-i] = value
	}
	return all
}

// Len implements the Base interface.
func (s *stringStack) Len() int {
	return len(s.values)
}

// Deflate implements the Base interface.
func (s *stringStack) Deflate() {
	s.values = []string{}
}

// Deflate implements the Stringer interface.
func (s *stringStack) String() string {
	return fmt.Sprintf("%v", s.values)
}

// EOF
