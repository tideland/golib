// Tideland Go Library - Collections
//
// Copyright (C) 2015-2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package collections

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
)

//--------------------
// EXCHANGE TYPES
//--------------------

// KeyValue wraps a key and a value for the key/value iterator.
type KeyValue struct {
	Keys  string
	Value interface{}
}

// KeyStringValue carries a combination of key and string value.
type KeyStringValue struct {
	Key   string
	Value string
}

//--------------------
// COLLECTIONS - RING BUFFER
//--------------------

// RingBuffer defines a buffer which is connected end-to-end. It
// grows if needed.
type RingBuffer interface {
	fmt.Stringer

	// Push adds values to the end of the buffer.
	Push(values ...interface{})

	// Pop removes and returns the first value of the buffer. If
	// the buffer is empty the second return value is false.
	Pop() (interface{}, bool)

	// Len returns the number of values in the buffer.
	Len() int

	// Cap returns the capacity of the buffer.
	Cap() int
}

//--------------------
// COLLECTIONS - STACKS
//--------------------

// Stack defines a stack containing any kind of values.
type Stack interface {
	fmt.Stringer

	// Push adds values to the top of the stack.
	Push(vs ...interface{})

	// Pop removes and returns the top value of the stack.
	Pop() (interface{}, error)

	// Peek returns the top value of the stack.
	Peek() (interface{}, error)

	// All returns all values bottom-up.
	All() []interface{}

	// AllReverse returns all values top-down.
	AllReverse() []interface{}

	// Len returns the number of entries in the stack.
	Len() int

	// Deflate cleans the stack.
	Deflate()
}

// StringStack defines a stack containing string values.
type StringStack interface {
	fmt.Stringer

	// Push adds strings to the top of the stack.
	Push(vs ...string)

	// Pop removes and returns the top value of the stack.
	Pop() (string, error)

	// Peek returns the top value of the stack.
	Peek() (string, error)

	// All returns all values bottom-up.
	All() []string

	// AllReverse returns all values top-down.
	AllReverse() []string

	// Len returns the number of entries in the stack.
	Len() int

	// Deflate cleans the stack.
	Deflate()
}

//--------------------
// COLLECTIONS - SETS
//--------------------

// Set defines a set containing any kind of values.
type Set interface {
	fmt.Stringer

	// Add adds values to the set.
	Add(vs ...interface{})

	// Remove removes a value out if the set. It doesn't
	// matter if the set does not contain the value.
	Remove(vs ...interface{})

	// Contains checks if the set contains a given value.
	Contains(v interface{}) bool

	// All returns all values.
	All() []interface{}

	// FindAll returns all values found by the
	// passed function.
	FindAll(f func(v interface{}) (bool, error)) ([]interface{}, error)

	// DoAll executes the passed function on all values.
	DoAll(f func(v interface{}) error) error

	// Len returns the number of entries in the set.
	Len() int

	// Deflate cleans the stack.
	Deflate()
}

// StringSet defines a set containing string values.
type StringSet interface {
	fmt.Stringer

	// Add adds values to the set.
	Add(vs ...string)

	// Remove removes a value out if the set. It doesn't
	// matter if the set does not contain the value.
	Remove(vs ...string)

	// Contains checks if a value is
	Contains(v string) bool

	// All returns all values.
	All() []string

	// FindAll returns all values found by the
	// passed function.
	FindAll(f func(v string) (bool, error)) ([]string, error)

	// DoAll executes the passed function on all values.
	DoAll(f func(v string) error) error

	// Len returns the number of entries in the set.
	Len() int

	// Deflate cleans the stack.
	Deflate()
}

//--------------------
// COLLECTIONS - TREE CHANGERS
//--------------------

// Changer defines the interface to perform changes on a tree
// node. It is returned by the addressing operations like At() and
// Create() of the Tree.
type Changer interface {
	// Value returns the changer node value.
	Value() (interface{}, error)

	// SetValue sets the changer node value. It also returns
	// the previous value.
	SetValue(value interface{}) (interface{}, error)

	// Add sets a child value.
	Add(value interface{}) error

	// Remove deletes this changer node.
	Remove() error

	// List returns the values of the children of the changer node.
	List() ([]interface{}, error)

	// Error returns a potential error of the changer.
	Error() error
}

// StringChanger defines the interface to perform changes on a string
// tree node. It is returned by the addressing operations like
// At() and Create() of the StringTree.
type StringChanger interface {
	// Value returns the changer node value.
	Value() (string, error)

	// SetValue sets the changer node value. It also returns
	// the previous value.
	SetValue(value string) (string, error)

	// Add sets a child value. If the key already exists the
	// value will be overwritten.
	Add(value string) error

	// Remove deletes this changer node.
	Remove() error

	// List returns the values of the children of the changer node.
	List() ([]string, error)

	// Error returns a potential error of the changer.
	Error() error
}

// KeyValueChanger defines the interface to perform changes on a
// key/value tree node. It is returned by the addressing operations
// like At() and Create() of the KeyValueTree.
type KeyValueChanger interface {
	// Key returns the changer node key.
	Key() (string, error)

	// SetKey sets the changer node key. Its checks if duplicate
	// keys are allowed and returns the previous key.
	SetKey(key string) (string, error)

	// Value returns the changer node value.
	Value() (interface{}, error)

	// SetValue sets the changer node value. It also returns
	// the previous value.
	SetValue(value interface{}) (interface{}, error)

	// Add sets a child key/value. If the key already exists the
	// value will be overwritten.
	Add(key string, value interface{}) error

	// Remove deletes this changer node.
	Remove() error

	// List returns the keys and values of the children of the changer node.
	List() ([]KeyValue, error)

	// Error returns a potential error of the changer.
	Error() error
}

// KeyStringValueChanger defines the interface to perform changes
// on a key/string value tree node. It is returned by the addressing
// operations like At() and Create() of the KeyStringValueTree.
type KeyStringValueChanger interface {
	// Key returns the changer node key.
	Key() (string, error)

	// SetKey sets the changer node key. Its checks if duplicate
	// keys are allowed and returns the previous key.
	SetKey(key string) (string, error)

	// Value returns the changer node value.
	Value() (string, error)

	// SetValue sets the changer node value. It also returns
	// the previous value.
	SetValue(value string) (string, error)

	// Add sets a child key/value. If the key already exists the
	// value will be overwritten.
	Add(key, value string) error

	// Remove deletes this changer node.
	Remove() error

	// List returns the keys and values of the children of the changer node.
	List() ([]KeyStringValue, error)

	// Error returns a potential error of the changer.
	Error() error
}

//--------------------
// COLLECTIONS - TREES
//--------------------

// Tree defines the interface for a tree able to store any type
// of values.
type Tree interface {
	fmt.Stringer

	// At returns the changer of the path defined by the given
	// values. If it does not exist it will not be created. Use
	// Create() here. So to set a child at a given node path do
	//
	//     err := tree.At("path", 1, "to", "use").Set(12345)
	At(values ...interface{}) Changer

	// Root returns the top level changer.
	Root() Changer

	// Create returns the changer of the path defined by the
	// given keys. If it does not exist it will be created,
	// but at least the root key has to be correct.
	Create(values ...interface{}) Changer

	// FindFirst returns the changer for the first node found
	// by the passed function.
	FindFirst(f func(value interface{}) (bool, error)) Changer

	// FindAll returns all changers for the nodes found
	// by the passed function.
	FindAll(f func(value interface{}) (bool, error)) []Changer

	// DoAll executes the passed function on all nodes.
	DoAll(f func(value interface{}) error) error

	// DoAllDeep executes the passed function on all nodes
	// passing a deep list of values ordered top-down.
	DoAllDeep(f func(values []interface{}) error) error

	// Len returns the number of nodes of the tree.
	Len() int

	// Copy creates a copy of the tree.
	Copy() Tree

	// Deflate cleans the tree with a new root value.
	Deflate(value interface{})
}

// StringTree defines the interface for a tree able to store strings.
type StringTree interface {
	fmt.Stringer

	// At returns the changer of the path defined by the given
	// values. If it does not exist it will not be created. Use
	// Create() here. So to set a child at a given node path do
	//
	//     err := tree.At("path", "one", "to", "use").Set("12345")
	At(values ...string) StringChanger

	// Root returns the top level changer.
	Root() StringChanger

	// Create returns the changer of the path defined by the
	// given keys. If it does not exist it will be created,
	// but at least the root key has to be correct.
	Create(values ...string) StringChanger

	// FindFirst returns the changer for the first node found
	// by the passed function.
	FindFirst(f func(value string) (bool, error)) StringChanger

	// FindAll returns all changers for the nodes found
	// by the passed function.
	FindAll(f func(value string) (bool, error)) []StringChanger

	// DoAll executes the passed function on all nodes.
	DoAll(f func(value string) error) error

	// DoAllDeep executes the passed function on all nodes
	// passing a deep list of values ordered top-down.
	DoAllDeep(f func(values []string) error) error

	// Len returns the number of nodes of the tree.
	Len() int

	// Copy creates a copy of the tree.
	Copy() StringTree

	// Deflate cleans the tree with a new root value.
	Deflate(value string)
}

// KeyValueTree defines the interface for a tree able to store key/value pairs.
type KeyValueTree interface {
	fmt.Stringer

	// At returns the changer of the path defined by the given
	// values. If it does not exist it will not be created. Use
	// Create() here. So to set a child at a given node path do
	//
	//     err := tree.At("path", "one", "to", "use").Set(12345)
	At(keys ...string) KeyValueChanger

	// Root returns the top level changer.
	Root() KeyValueChanger

	// Create returns the changer of the path defined by the
	// given keys. If it does not exist it will be created,
	// but at least the root key has to be correct.
	Create(keys ...string) KeyValueChanger

	// FindFirst returns the changer for the first node found
	// by the passed function.
	FindFirst(f func(key string, value interface{}) (bool, error)) KeyValueChanger

	// FindAll returns all changers for the nodes found
	// by the passed function.
	FindAll(f func(key string, value interface{}) (bool, error)) []KeyValueChanger

	// DoAll executes the passed function on all nodes.
	DoAll(f func(key string, value interface{}) error) error

	// DoAllDeep executes the passed function on all nodes
	// passing a deep list of keys ordered top-down.
	DoAllDeep(f func(keys []string, value interface{}) error) error

	// Len returns the number of nodes of the tree.
	Len() int

	// Copy creates a copy of the tree.
	Copy() KeyValueTree

	// CopyAt creates a copy of a subtree.
	CopyAt(keys ...string) (KeyValueTree, error)

	// Deflate cleans the tree with a new root value.
	Deflate(key string, value interface{})
}

// KeyStringValueTree defines the interface for a tree able to store
// key/string value pairs.
type KeyStringValueTree interface {
	fmt.Stringer

	// At returns the changer of the path defined by the given
	// values. If it does not exist it will not be created. Use
	// Create() here. So to set a child at a given node path do
	//
	//     err := tree.At("path", "one", "to", "use").Set(12345)
	At(keys ...string) KeyStringValueChanger

	// Root returns the top level changer.
	Root() KeyStringValueChanger

	// Create returns the changer of the path defined by the
	// given keys. If it does not exist it will be created,
	// but at least the root key has to be correct.
	Create(keys ...string) KeyStringValueChanger

	// FindFirst returns the changer for the first node found
	// by the passed function.
	FindFirst(f func(key, value string) (bool, error)) KeyStringValueChanger

	// FindAll returns all changers for the nodes found
	// by the passed function.
	FindAll(f func(key, value string) (bool, error)) []KeyStringValueChanger

	// DoAll executes the passed function on all nodes.
	DoAll(f func(key, value string) error) error

	// DoAllDeep executes the passed function on all nodes
	// passing a deep list of keys ordered top-down.
	DoAllDeep(f func(keys []string, value string) error) error

	// Len returns the number of nodes of the tree.
	Len() int

	// Copy creates a copy of the tree.
	Copy() KeyStringValueTree

	// CopyAt creates a copy of a subtree.
	CopyAt(keys ...string) (KeyStringValueTree, error)

	// Deflate cleans the tree with a new root value.
	Deflate(key, value string)
}

// EOF
