// Tideland Go Library - Simple Markup Language - Builder
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package sml

//--------------------
// IMPORTS
//--------------------

import (
	"strings"

	"github.com/tideland/golib/collections"
	"github.com/tideland/golib/errors"
)

//--------------------
// NODE BUILDER
//--------------------

// NodeBuilder creates a node structure when a SML
// document is read.
type NodeBuilder struct {
	stack []*tagNode
	done  bool
}

// NewNodeBuilder return a new nnode builder.
func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{[]*tagNode{}, false}
}

// Root returns the root node of the read document.
func (nb *NodeBuilder) Root() (Node, error) {
	if !nb.done {
		return nil, errors.New(ErrBuilder, errorMessages, "building is not yet done")
	}
	return nb.stack[0], nil
}

// BeginTagNode implements the Builder interface.
func (nb *NodeBuilder) BeginTagNode(tag string) error {
	if nb.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	t, err := newTagNode(tag)
	if err != nil {
		return err
	}
	nb.stack = append(nb.stack, t)
	return nil
}

// EndTagNode implements the Builder interface.
func (nb *NodeBuilder) EndTagNode() error {
	if nb.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	switch l := len(nb.stack); l {
	case 0:
		return errors.New(ErrBuilder, errorMessages, "no opening tag")
	case 1:
		nb.done = true
	default:
		nb.stack[l-2].appendChild(nb.stack[l-1])
		nb.stack = nb.stack[:l-1]
	}
	return nil
}

// TextNode implements the Builder interface.
func (nb *NodeBuilder) TextNode(text string) error {
	if nb.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	if len(nb.stack) > 0 {
		nb.stack[len(nb.stack)-1].appendTextNode(text)
		return nil
	}
	return errors.New(ErrBuilder, errorMessages, "no opening tag for text")
}

// RawNode implements the Builder interface.
func (nb *NodeBuilder) RawNode(raw string) error {
	if nb.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	if len(nb.stack) > 0 {
		nb.stack[len(nb.stack)-1].appendRawNode(raw)
		return nil
	}
	return errors.New(ErrBuilder, errorMessages, "no opening tag for raw text")
}

// CommentNode implements the Builder interface.
func (nb *NodeBuilder) CommentNode(comment string) error {
	if nb.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	if len(nb.stack) > 0 {
		nb.stack[len(nb.stack)-1].appendCommentNode(comment)
		return nil
	}
	return errors.New(ErrBuilder, errorMessages, "no opening tag for comment")
}

//--------------------
// KEY/STRING VALUE TREE BUILDER
//--------------------

// KeyStringValueTreeBuilder implements Builder to parse a
// file and create a KeyStringValueTree.
type KeyStringValueTreeBuilder struct {
	stack  collections.StringStack
	tree collections.KeyStringValueTree
	done bool
}

// NewNodeBuilder return a new nnode builder.
func NewKeyStringValueTreeBuilder() *KeyStringValueTreeBuilder {
	return &KeyStringValueTreeBuilder{}
}

// Tree returns the created tree.
func (tb *KeyStringValueTreeBuilder) Tree() (collections.KeyStringValueTree, error) {
	if !tb.done {
		return nil, errors.New(ErrBuilder, errorMessages, "building is not yet done")
	}
	return tb.tree, nil
}

// BeginTagNode implements the Builder interface.
func (tb *KeyStringValueTreeBuilder) BeginTagNode(tag string) error {
	if tb.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	switch {
	case tb.tree == nil:
		tb.stack = collections.NewStringStack(tag)
		tb.tree = collections.NewKeyStringValueTree(tag, "", false)
	default:
		tb.stack.Push(tag)
		changer := tb.tree.Create(tb.stack.All()...)
		if err := changer.Error(); err != nil {
			return errors.Annotate(err, ErrBuilder, errorMessages)
		}
	}
	return nil
}

// EndTagNode implements the Builder interface.
func (tb *KeyStringValueTreeBuilder) EndTagNode() error {
	if tb.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	_, err := tb.stack.Pop()
	if tb.stack.Len() == 0 {
		tb.done = true
	}
	return err
}

// TextNode implements the Builder interface.
func (tb *KeyStringValueTreeBuilder) TextNode(text string) error {
	if tb.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	value, err := tb.tree.At(tb.stack.All()...).Value()
	if err != nil {
		return errors.Annotate(err, ErrBuilder, errorMessages)
	}
	if value != "" {
		return errors.New(ErrBuilder, errorMessages, "node has multiple values")
	}
	text = strings.TrimSpace(text)
	if text != "" {
		_, err = tb.tree.At(tb.stack.All()...).SetValue(text)
	}
	return err
}

// RawNode implements the Builder interface.
func (tb *KeyStringValueTreeBuilder) RawNode(raw string) error {
	return tb.TextNode(raw)
}

// Comment implements the Builder interface.
func (tb *KeyStringValueTreeBuilder) CommentNode(comment string) error {
	if tb.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	return nil
}

// EOF
