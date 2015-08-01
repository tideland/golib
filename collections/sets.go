// Tideland Go Library - Collections - Sets
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
// SET
//--------------------

// set implements the Set interface.
type set struct {
	values map[interface{}]struct{}
}

// NewSet creates a set with the passed values
// as initial content.
func NewSet(vs ...interface{}) Set {
	s := &set{make(map[interface{}]struct{})}
	s.Add(vs...)
	return s
}

// Add implements the Set interface.
func (s *set) Add(vs ...interface{}) {
	for _, v := range vs {
		s.values[v] = struct{}{}
	}
}

// Remove implements the Set interface.
func (s *set) Remove(vs ...interface{}) {
	for _, v := range vs {
		delete(s.values, v)
	}
}

// Contains implements the Set interface.
func (s *set) Contains(v interface{}) bool {
	_, ok := s.values[v]
	return ok
}

// All implements the Set interface.
func (s *set) All() []interface{} {
	all := []interface{}{}
	for v := range s.values {
		all = append(all, v)
	}
	return all
}

// FindAll implements the Set interface.
func (s *set) FindAll(f func(v interface{}) (bool, error)) ([]interface{}, error) {
	found := []interface{}{}
	for v := range s.values {
		ok, err := f(v)
		if err != nil {
			return nil, errors.Annotate(err, ErrFindAll, errorMessages)
		}
		if ok {
			found = append(found, v)
		}
	}
	return found, nil
}

// DoAll implements the Set interface.
func (s *set) DoAll(f func(v interface{}) error) error {
	for v := range s.values {
		if err := f(v); err != nil {
			return errors.Annotate(err, ErrDoAll, errorMessages)
		}
	}
	return nil
}

// Len implements the Set interface.
func (s *set) Len() int {
	return len(s.values)
}

// Deflate implements the Set interface.
func (s *set) Deflate() {
	s.values = make(map[interface{}]struct{})
}

// Deflate implements the Stringer interface.
func (s *set) String() string {
	all := s.All()
	return fmt.Sprintf("%v", all)
}

//--------------------
// STRING SET
//--------------------

// stringSet implements the StringSet interface.
type stringSet struct {
	values map[string]struct{}
}

// NewStringSet creates a string set with the passed values
// as initial content.
func NewStringSet(vs ...string) StringSet {
	s := &stringSet{make(map[string]struct{})}
	s.Add(vs...)
	return s
}

// Add implements the StringSet interface.
func (s *stringSet) Add(vs ...string) {
	for _, v := range vs {
		s.values[v] = struct{}{}
	}
}

// Remove implements the StringSet interface.
func (s *stringSet) Remove(vs ...string) {
	for _, v := range vs {
		delete(s.values, v)
	}
}

// Contains implements the StringSet interface.
func (s *stringSet) Contains(v string) bool {
	_, ok := s.values[v]
	return ok
}

// All implements the StringSet interface.
func (s *stringSet) All() []string {
	all := []string{}
	for v := range s.values {
		all = append(all, v)
	}
	return all
}

// FindAll implements the StringSet interface.
func (s *stringSet) FindAll(f func(v string) (bool, error)) ([]string, error) {
	found := []string{}
	for v := range s.values {
		ok, err := f(v)
		if err != nil {
			return nil, errors.Annotate(err, ErrFindAll, errorMessages)
		}
		if ok {
			found = append(found, v)
		}
	}
	return found, nil
}

// DoAll implements the StringSet interface.
func (s *stringSet) DoAll(f func(v string) error) error {
	for v := range s.values {
		if err := f(v); err != nil {
			return errors.Annotate(err, ErrDoAll, errorMessages)
		}
	}
	return nil
}

// Len implements the StringSet interface.
func (s *stringSet) Len() int {
	return len(s.values)
}

// Deflate implements the StringSet interface.
func (s *stringSet) Deflate() {
	s.values = make(map[string]struct{})
}

// Deflate implements the Stringer interface.
func (s *stringSet) String() string {
	all := s.All()
	return fmt.Sprintf("%v", all)
}

// EOF
