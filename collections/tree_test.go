// Tideland Go Library - Collections - Tree - Unit Tests
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package collections_test

//--------------------
// IMPORTS
//--------------------

import (
	"errors"
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/collections"
)

//--------------------
// TEST TREE
//--------------------

// TestTreeCreate tests the correct creation of a tree.
func TestTreeCreate(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	// Tree with duplicates, no errors.
	tree := collections.NewTree("root", true)
	err := tree.At("root").Add("alpha")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.Nil(err)
	err = tree.At("root").Add("charlie")
	assert.Nil(err)
	err = tree.Create("root", "delta", 1).Add(true)
	assert.Nil(err)
	assert.Length(tree, 8)

	// Deflate tree.
	tree.Deflate("toor")
	assert.Length(tree, 1)

	// Navigate with illegal paths.
	err = tree.At("foo").Add(0)
	assert.ErrorMatch(err, ".* node not found")
	err = tree.At("root", "foo").Add(0)
	assert.ErrorMatch(err, ".* node not found")

	// Tree without duplicates, so also with errors.
	tree = collections.NewTree("root", false)
	err = tree.At("root").Add("alpha")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.ErrorMatch(err, ".* duplicates are not allowed")
}

// TestTreeRemove tests the correct removal of tree nodes.
func TestTreeRemove(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createTree(assert)

	err := tree.At("root", "alpha").Remove()
	assert.Nil(err)
	assert.Length(tree, 11)

	err = tree.At("root", "delta").Remove()
	assert.Nil(err)
	assert.Length(tree, 6)
}

// TestTreeSetValue tests the setting of a tree nodes value.
func TestTreeSetValue(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createTree(assert)

	// Tree with duplicates.
	old, err := tree.At("root", "alpha").SetValue("beta")
	assert.Nil(err)
	assert.Equal(old, "alpha")
	act, err := tree.At("root", "beta").Value()
	assert.Nil(err)
	assert.Equal(act, "beta")

	// Tree without duplicates.
	tree = collections.NewTree("root", false)
	err = tree.At("root").Add("alpha")
	assert.Nil(err)
	err = tree.At("root").Add("beta")
	assert.Nil(err)
	old, err = tree.At("root", "alpha").SetValue("beta")
	assert.Nil(old)
	assert.ErrorMatch(err, ".* duplicates are not allowed")
}

// TestTreeFind tests the correct finding in tree nodes.
func TestTreeFind(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createTree(assert)

	// Test finding the first matching.
	list, err := tree.FindFirst(func(v interface{}) (bool, error) {
		switch vt := v.(type) {
		case string:
			return vt == "bravo", nil
		default:
			return false, nil
		}
	}).List()
	assert.Nil(err)
	assert.Equal(list, []interface{}{"foo", "bar"})
	list, err = tree.FindFirst(func(v interface{}) (bool, error) {
		return false, nil
	}).List()
	assert.ErrorMatch(err, ".* node not found")
	list, err = tree.FindFirst(func(v interface{}) (bool, error) {
		return false, errors.New("ouch")
	}).List()
	assert.ErrorMatch(err, ".* cannot find first node: ouch")

	// Test finding all matching.
	changers := tree.FindAll(func(v interface{}) (bool, error) {
		switch v.(type) {
		case int:
			return true, nil
		default:
			return false, nil
		}
	})
	assert.Length(changers, 2)
	v, err := changers[0].Value()
	assert.Nil(err)
	assert.Equal(v, 1)
	v, err = changers[1].Value()
	assert.Nil(err)
	assert.Equal(v, 2)
	changers = tree.FindAll(func(v interface{}) (bool, error) {
		return false, nil
	})
	assert.Length(changers, 0)
	changers = tree.FindAll(func(v interface{}) (bool, error) {
		return false, errors.New("ouch")
	})
	assert.Length(changers, 1)
	assert.ErrorMatch(changers[0].Error(), ".* cannot find all nodes: ouch")
}

// TestTreeDo tests the iteration over the tree nodes.
func TestTreeDo(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	tree := collections.NewTree("root", true)
	err := tree.At("root").Add("alpha")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.Nil(err)
	err = tree.At("root", "bravo").Add("foo")
	assert.Nil(err)
	err = tree.At("root", "bravo").Add("bar")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.Nil(err)
	err = tree.At("root").Add("charlie")
	assert.Nil(err)
	err = tree.Create("root", "delta", 1).Add(true)
	assert.Nil(err)
	err = tree.Create("root", "delta", 2).Add(false)
	assert.Nil(err)

	// Test iteration.
	var values []interface{}
	err = tree.DoAll(func(v interface{}) error {
		values = append(values, v)
		return nil
	})
	assert.Nil(err)
	assert.Length(values, 12)
	err = tree.DoAll(func(v interface{}) error {
		return errors.New("ouch")
	})
	assert.ErrorMatch(err, ".* cannot perform function on all nodes: ouch")
}

// TestTreeCopy tests the copy of a tree.
func TestTreeCopy(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	tree := collections.NewTree("root", true)
	err := tree.Create("root", "alpha").Add("a")
	assert.Nil(err)
	err = tree.Create("root", "beta").Add("b")
	assert.Nil(err)
	err = tree.Create("root", "gamma", "one").Add("1")
	assert.Nil(err)
	err = tree.Create("root", "gamma", "two").Add("2")
	assert.Nil(err)

	ctree := tree.Copy()
	assert.Length(ctree, 10)
	value, err := ctree.At("root", "alpha", "a").Value()
	assert.Nil(err)
	assert.Equal(value, "a")
	value, err = ctree.At("root", "gamma", "two", "2").Value()
	assert.Nil(err)
	assert.Equal(value, "2")
}

//--------------------
// TEST STRING TREE
//--------------------

// TestStringTreeCreate tests the correct creation of a string tree.
func TestStringTreeCreate(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	// String tree with duplicates, no errors.
	tree := collections.NewStringTree("root", true)
	err := tree.At("root").Add("alpha")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.Nil(err)
	err = tree.At("root").Add("charlie")
	assert.Nil(err)
	err = tree.Create("root", "delta", "one").Add("true")
	assert.Nil(err)
	assert.Length(tree, 8)

	// Deflate tree.
	tree.Deflate("toor")
	assert.Length(tree, 1)

	// Navigate with illegal paths.
	err = tree.At("foo").Add("zero")
	assert.ErrorMatch(err, ".* node not found")
	err = tree.At("root", "foo").Add("zero")
	assert.ErrorMatch(err, ".* node not found")

	// Tree without duplicates, so also with errors.
	tree = collections.NewStringTree("root", false)
	err = tree.At("root").Add("alpha")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.ErrorMatch(err, ".* duplicates are not allowed")
}

// TestStringTreeRemove tests the correct removal of string tree nodes.
func TestStringTreeRemove(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createStringTree(assert)

	err := tree.At("root", "alpha").Remove()
	assert.Nil(err)
	assert.Length(tree, 11)

	err = tree.At("root", "delta").Remove()
	assert.Nil(err)
	assert.Length(tree, 6)
}

// TestStringTreeSetValue tests the setting of a string tree nodes value.
func TestStringTreeSetValue(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createStringTree(assert)

	// Tree with duplicates.
	old, err := tree.At("root", "alpha").SetValue("beta")
	assert.Nil(err)
	assert.Equal(old, "alpha")
	act, err := tree.At("root", "beta").Value()
	assert.Nil(err)
	assert.Equal(act, "beta")

	// Tree without duplicates.
	tree = collections.NewStringTree("root", false)
	err = tree.At("root").Add("alpha")
	assert.Nil(err)
	err = tree.At("root").Add("beta")
	assert.Nil(err)
	old, err = tree.At("root", "alpha").SetValue("beta")
	assert.Equal(old, "")
	assert.ErrorMatch(err, ".* duplicates are not allowed")
}

// TestStringTreeFind tests the correct finding in string tree nodes.
func TestStringTreeFind(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createStringTree(assert)

	// Test finding the first matching.
	list, err := tree.FindFirst(func(v string) (bool, error) {
		return v == "bravo", nil
	}).List()
	assert.Nil(err)
	assert.Equal(list, []string{"foo", "bar"})
	list, err = tree.FindFirst(func(v string) (bool, error) {
		return false, nil
	}).List()
	assert.ErrorMatch(err, ".* node not found")
	list, err = tree.FindFirst(func(v string) (bool, error) {
		return false, errors.New("ouch")
	}).List()
	assert.ErrorMatch(err, ".* cannot find first node: ouch")

	// Test finding all matching.
	changers := tree.FindAll(func(v string) (bool, error) {
		return v == "bravo", nil
	})
	assert.Length(changers, 2)
	v, err := changers[0].Value()
	assert.Nil(err)
	assert.Equal(v, "bravo")
	v, err = changers[1].Value()
	assert.Nil(err)
	assert.Equal(v, "bravo")
	changers = tree.FindAll(func(v string) (bool, error) {
		return false, nil
	})
	assert.Length(changers, 0)
	changers = tree.FindAll(func(v string) (bool, error) {
		return false, errors.New("ouch")
	})
	assert.Length(changers, 1)
	assert.ErrorMatch(changers[0].Error(), ".* cannot find all nodes: ouch")
}

// TestStringTreeDo tests the iteration over the string tree nodes.
func TestStringTreeDo(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createStringTree(assert)

	// Test iteration.
	var values []string
	err := tree.DoAll(func(v string) error {
		values = append(values, v)
		return nil
	})
	assert.Nil(err)
	assert.Length(values, 12)
	err = tree.DoAll(func(v string) error {
		return errors.New("ouch")
	})
	assert.ErrorMatch(err, ".* cannot perform function on all nodes: ouch")
}

// TestStringTreeCopy tests the copy of a string tree.
func TestStringTreeCopy(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	tree := collections.NewStringTree("root", true)
	err := tree.Create("root", "alpha").Add("a")
	assert.Nil(err)
	err = tree.Create("root", "beta").Add("b")
	assert.Nil(err)
	err = tree.Create("root", "gamma", "one").Add("1")
	assert.Nil(err)
	err = tree.Create("root", "gamma", "two").Add("2")
	assert.Nil(err)

	ctree := tree.Copy()
	assert.Length(ctree, 10)
	value, err := ctree.At("root", "alpha", "a").Value()
	assert.Nil(err)
	assert.Equal(value, "a")
	value, err = ctree.At("root", "gamma", "two", "2").Value()
	assert.Nil(err)
	assert.Equal(value, "2")
}

//--------------------
// TEST KEY/VALUE TREE
//--------------------

// TestKeyValueTreeCreate tests the correct creation of a key/value tree.
func TestKeyValueTreeCreate(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	// Key/value tree with duplicates, no errors.
	tree := collections.NewKeyValueTree("root", 1, true)
	err := tree.At("root").Add("alpha", 2)
	assert.Nil(err)
	err = tree.At("root").Add("bravo", 3)
	assert.Nil(err)
	err = tree.At("root").Add("bravo", true)
	assert.Nil(err)
	err = tree.At("root").Add("charlie", 1.0)
	assert.Nil(err)
	err = tree.Create("root", "delta", "one").Add("true", "false")
	assert.Nil(err)
	assert.Length(tree, 8)

	// Deflate tree.
	tree.Deflate("toor", 0)
	assert.Length(tree, 1)

	// Navigate with illegal paths.
	err = tree.At("foo").Add("zero", 0)
	assert.ErrorMatch(err, ".* node not found")
	err = tree.At("root", "foo").Add("zero", 0)
	assert.ErrorMatch(err, ".* node not found")

	// Tree without duplicates, so also with errors.
	tree = collections.NewKeyValueTree("root", 0, false)
	err = tree.At("root").Add("alpha", "a")
	assert.Nil(err)
	err = tree.At("root").Add("bravo", "b")
	assert.Nil(err)
	err = tree.At("root").Add("bravo", 2)
	assert.ErrorMatch(err, ".* duplicates are not allowed")
}

// TestKeyValueTreeRemove tests the correct removal of key/value tree nodes.
func TestKeyValueTreeRemove(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createKeyValueTree(assert)

	err := tree.At("root", "alpha").Remove()
	assert.Nil(err)
	assert.Length(tree, 11)

	err = tree.At("root", "delta").Remove()
	assert.Nil(err)
	assert.Length(tree, 6)
}

// TestKeyValueTreeSetValue tests the setting of a key/value tree nodes value.
func TestKeyValueTreeSetValue(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createKeyValueTree(assert)

	// Tree with duplicates.
	old, err := tree.At("root", "alpha").SetValue("beta")
	assert.Nil(err)
	assert.Equal(old, 2)
	act, err := tree.At("root", "alpha").Value()
	assert.Nil(err)
	assert.Equal(act, "beta")

	// Tree without duplicates.
	tree = collections.NewKeyValueTree("root", 1, false)
	err = tree.At("root").Add("alpha", 2)
	assert.Nil(err)
	err = tree.At("root").Add("beta", 3)
	assert.Nil(err)
	old, err = tree.At("root", "alpha").SetValue("beta")
	assert.Nil(err)
	assert.Equal(old, 2)
}

// TestKeyValueTreeFind tests the correct finding in key/value tree nodes.
func TestKeyValueTreeFind(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createKeyValueTree(assert)

	// Test finding the first matching.
	list, err := tree.FindFirst(func(k string, v interface{}) (bool, error) {
		return k == "bravo", nil
	}).List()
	assert.Nil(err)
	assert.Equal(list, []collections.KeyValue{{"foo", "bar"}, {"bar", "foo"}})
	list, err = tree.FindFirst(func(k string, v interface{}) (bool, error) {
		return false, nil
	}).List()
	assert.ErrorMatch(err, ".* node not found")
	list, err = tree.FindFirst(func(k string, v interface{}) (bool, error) {
		return false, errors.New("ouch")
	}).List()
	assert.ErrorMatch(err, ".* cannot find first node: ouch")

	// Test finding all matching.
	changers := tree.FindAll(func(k string, v interface{}) (bool, error) {
		return k == "bravo", nil
	})
	assert.Length(changers, 2)
	v, err := changers[0].Value()
	assert.Nil(err)
	assert.Equal(v, 3)
	v, err = changers[1].Value()
	assert.Nil(err)
	assert.Equal(v, 4)
	changers = tree.FindAll(func(k string, v interface{}) (bool, error) {
		return false, nil
	})
	assert.Length(changers, 0)
	changers = tree.FindAll(func(k string, v interface{}) (bool, error) {
		return false, errors.New("ouch")
	})
	assert.Length(changers, 1)
	assert.ErrorMatch(changers[0].Error(), ".* cannot find all nodes: ouch")
}

// TestKeyValueTreeDo tests the iteration over the key/value tree nodes.
func TestKeyValueTreeDo(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createKeyValueTree(assert)

	// Test iteration.
	var values []interface{}
	err := tree.DoAll(func(k string, v interface{}) error {
		values = append(values, v)
		return nil
	})
	assert.Nil(err)
	assert.Length(values, 12)
	err = tree.DoAll(func(k string, v interface{}) error {
		return errors.New("ouch")
	})
	assert.ErrorMatch(err, ".* cannot perform function on all nodes: ouch")
}

// TestKeyValueTreeCopy tests the copy of a key/value tree.
func TestKeyValueTreeCopy(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	tree := collections.NewKeyValueTree("root", "0", true)
	err := tree.Create("root", "alpha").Add("a", "1")
	assert.Nil(err)
	err = tree.Create("root", "beta").Add("b", "2")
	assert.Nil(err)
	err = tree.Create("root", "gamma", "one").Add("1", "3.1")
	assert.Nil(err)
	err = tree.Create("root", "gamma", "two").Add("2", "3.2")
	assert.Nil(err)

	ctree := tree.Copy()
	assert.Length(ctree, 10)
	value, err := ctree.At("root", "alpha", "a").Value()
	assert.Nil(err)
	assert.Equal(value, "1")
	value, err = ctree.At("root", "gamma", "two", "2").Value()
	assert.Nil(err)
	assert.Equal(value, "3.2")
}

//--------------------
// TEST KEY/STRING VALUE TREE
//--------------------

// TestKeyStringValueTreeCreate tests the correct creation of a
// key/string value tree.
func TestKeyStringValueTreeCreate(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	// Key/string value tree with duplicates, no errors.
	tree := collections.NewKeyStringValueTree("root", "one", true)
	err := tree.At("root").Add("alpha", "two")
	assert.Nil(err)
	err = tree.At("root").Add("bravo", "three")
	assert.Nil(err)
	err = tree.At("root").Add("bravo", "true")
	assert.Nil(err)
	err = tree.At("root").Add("charlie", "1.0")
	assert.Nil(err)
	err = tree.Create("root", "delta", "one").Add("true", "false")
	assert.Nil(err)
	assert.Length(tree, 8)

	// Deflate tree.
	tree.Deflate("toor", "zero")
	assert.Length(tree, 1)

	// Navigate with illegal paths.
	err = tree.At("foo").Add("zero", "0")
	assert.ErrorMatch(err, ".* node not found")
	err = tree.At("root", "foo").Add("zero", "0")
	assert.ErrorMatch(err, ".* node not found")

	// Tree without duplicates, so also with errors.
	tree = collections.NewKeyStringValueTree("root", "0", false)
	err = tree.At("root").Add("alpha", "a")
	assert.Nil(err)
	err = tree.At("root").Add("bravo", "b")
	assert.Nil(err)
	err = tree.At("root").Add("bravo", "2")
	assert.ErrorMatch(err, ".* duplicates are not allowed")
}

// TestKeyStringValueTreeRemove tests the correct removal of
// key/string value tree nodes.
func TestKeyStringValueTreeRemove(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createKeyStringValueTree(assert)

	err := tree.At("root", "alpha").Remove()
	assert.Nil(err)
	assert.Length(tree, 11)

	err = tree.At("root", "delta").Remove()
	assert.Nil(err)
	assert.Length(tree, 6)
}

// TestKeyStringValueTreeSetValue tests the setting of a
// key/string value tree nodes value.
func TestKeyStringValueTreeSetValue(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createKeyStringValueTree(assert)

	// Tree with duplicates.
	old, err := tree.At("root", "alpha").SetValue("beta")
	assert.Nil(err)
	assert.Equal(old, "two")
	act, err := tree.At("root", "alpha").Value()
	assert.Nil(err)
	assert.Equal(act, "beta")

	// Tree without duplicates.
	tree = collections.NewKeyStringValueTree("root", "one", false)
	err = tree.At("root").Add("alpha", "two")
	assert.Nil(err)
	err = tree.At("root").Add("beta", "three")
	assert.Nil(err)
	old, err = tree.At("root", "alpha").SetValue("beta")
	assert.Nil(err)
	assert.Equal(old, "two")
}

// TestKeyStringValueTreeFind tests the correct finding in
// key/string value tree nodes.
func TestKeyStringValueTreeFind(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createKeyStringValueTree(assert)

	// Test finding the first matching.
	list, err := tree.FindFirst(func(k, v string) (bool, error) {
		return k == "bravo", nil
	}).List()
	assert.Nil(err)
	assert.Equal(list, []collections.KeyStringValue{{"foo", "bar"}, {"bar", "foo"}})
	list, err = tree.FindFirst(func(k, v string) (bool, error) {
		return false, nil
	}).List()
	assert.ErrorMatch(err, ".* node not found")
	list, err = tree.FindFirst(func(k, v string) (bool, error) {
		return false, errors.New("ouch")
	}).List()
	assert.ErrorMatch(err, ".* cannot find first node: ouch")

	// Test finding all matching.
	changers := tree.FindAll(func(k, v string) (bool, error) {
		return k == "bravo", nil
	})
	assert.Length(changers, 2)
	v, err := changers[0].Value()
	assert.Nil(err)
	assert.Equal(v, "three")
	v, err = changers[1].Value()
	assert.Nil(err)
	assert.Equal(v, "four")
	changers = tree.FindAll(func(k, v string) (bool, error) {
		return false, nil
	})
	assert.Length(changers, 0)
	changers = tree.FindAll(func(k, v string) (bool, error) {
		return false, errors.New("ouch")
	})
	assert.Length(changers, 1)
	assert.ErrorMatch(changers[0].Error(), ".* cannot find all nodes: ouch")
}

// TestKeyStringValueTreeDo tests the iteration over the
// key/string value tree nodes.
func TestKeyStringValueTreeDo(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	tree := createKeyStringValueTree(assert)

	// Test iteration.
	var values []interface{}
	err := tree.DoAll(func(k, v string) error {
		values = append(values, v)
		return nil
	})
	assert.Nil(err)
	assert.Length(values, 12)
	err = tree.DoAll(func(k, v string) error {
		return errors.New("ouch")
	})
	assert.ErrorMatch(err, ".* cannot perform function on all nodes: ouch")
}

// TestKeyStringValueTreeCopy tests the copy of a key/string value tree.
func TestKeyStringValueTreeCopy(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	tree := collections.NewKeyStringValueTree("root", "0", true)
	err := tree.Create("root", "alpha").Add("a", "1")
	assert.Nil(err)
	err = tree.Create("root", "beta").Add("b", "2")
	assert.Nil(err)
	err = tree.Create("root", "gamma", "one").Add("1", "3.1")
	assert.Nil(err)
	err = tree.Create("root", "gamma", "two").Add("2", "3.2")
	assert.Nil(err)

	ctree := tree.Copy()
	assert.Length(ctree, 10)
	value, err := ctree.At("root", "alpha", "a").Value()
	assert.Nil(err)
	assert.Equal(value, "1")
	value, err = ctree.At("root", "gamma", "two", "2").Value()
	assert.Nil(err)
	assert.Equal(value, "3.2")
}

//--------------------
// HELPERS
//--------------------

func createTree(assert audit.Assertion) collections.Tree {
	tree := collections.NewTree("root", true)
	err := tree.At("root").Add("alpha")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.Nil(err)
	err = tree.At("root", "bravo").Add("foo")
	assert.Nil(err)
	err = tree.At("root", "bravo").Add("bar")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.Nil(err)
	err = tree.At("root").Add("charlie")
	assert.Nil(err)
	err = tree.Create("root", "delta", 1).Add(true)
	assert.Nil(err)
	err = tree.Create("root", "delta", 2).Add(false)
	assert.Nil(err)
	assert.Length(tree, 12)

	return tree
}

func createStringTree(assert audit.Assertion) collections.StringTree {
	tree := collections.NewStringTree("root", true)
	err := tree.At("root").Add("alpha")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.Nil(err)
	err = tree.At("root", "bravo").Add("foo")
	assert.Nil(err)
	err = tree.At("root", "bravo").Add("bar")
	assert.Nil(err)
	err = tree.At("root").Add("bravo")
	assert.Nil(err)
	err = tree.At("root").Add("charlie")
	assert.Nil(err)
	err = tree.Create("root", "delta", "one").Add("true")
	assert.Nil(err)
	err = tree.Create("root", "delta", "two").Add("false")
	assert.Nil(err)
	assert.Length(tree, 12)

	return tree
}

func createKeyValueTree(assert audit.Assertion) collections.KeyValueTree {
	tree := collections.NewKeyValueTree("root", 1, true)
	err := tree.At("root").Add("alpha", 2)
	assert.Nil(err)
	err = tree.At("root").Add("bravo", 3)
	assert.Nil(err)
	err = tree.At("root", "bravo").Add("foo", "bar")
	assert.Nil(err)
	err = tree.At("root", "bravo").Add("bar", "foo")
	assert.Nil(err)
	err = tree.At("root").Add("bravo", 4)
	assert.Nil(err)
	err = tree.At("root").Add("charlie", 5)
	assert.Nil(err)
	err = tree.Create("root", "delta", "one").Add("true", 1)
	assert.Nil(err)
	err = tree.Create("root", "delta", "two").Add("false", 0)
	assert.Nil(err)
	assert.Length(tree, 12)

	return tree
}

func createKeyStringValueTree(assert audit.Assertion) collections.KeyStringValueTree {
	tree := collections.NewKeyStringValueTree("root", "one", true)
	err := tree.At("root").Add("alpha", "two")
	assert.Nil(err)
	err = tree.At("root").Add("bravo", "three")
	assert.Nil(err)
	err = tree.At("root", "bravo").Add("foo", "bar")
	assert.Nil(err)
	err = tree.At("root", "bravo").Add("bar", "foo")
	assert.Nil(err)
	err = tree.At("root").Add("bravo", "four")
	assert.Nil(err)
	err = tree.At("root").Add("charlie", "five")
	assert.Nil(err)
	err = tree.Create("root", "delta", "one").Add("true", "one")
	assert.Nil(err)
	err = tree.Create("root", "delta", "two").Add("false", "zero")
	assert.Nil(err)
	assert.Length(tree, 12)

	return tree
}

// EOF
