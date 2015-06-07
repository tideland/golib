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
func (n *NodeBuilder) Root() (Node, error) {
	if !n.done {
		return nil, errors.New(ErrBuilder, errorMessages, "building is not yet done")
	}
	return n.stack[0], nil
}

// BeginTagNode is specified on the Builder interface.
func (n *NodeBuilder) BeginTagNode(tag string) error {
	if n.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	t, err := newTagNode(tag)
	if err != nil {
		return err
	}
	n.stack = append(n.stack, t)
	return nil
}

// EndTagNode is specified on the Builder interface.
func (n *NodeBuilder) EndTagNode() error {
	if n.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	switch l := len(n.stack); l {
	case 0:
		return errors.New(ErrBuilder, errorMessages, "no opening tag")
	case 1:
		n.done = true
	default:
		n.stack[l-2].appendChild(n.stack[l-1])
		n.stack = n.stack[:l-1]
	}
	return nil
}

// TextNode is specified on the Builder interface.
func (n *NodeBuilder) TextNode(text string) error {
	if n.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	if len(n.stack) > 0 {
		n.stack[len(n.stack)-1].appendTextNode(text)
		return nil
	}
	return errors.New(ErrBuilder, errorMessages, "no opening tag for text")
}

// RawNode is specified on the Builder interface.
func (n *NodeBuilder) RawNode(raw string) error {
	if n.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	if len(n.stack) > 0 {
		n.stack[len(n.stack)-1].appendRawNode(raw)
		return nil
	}
	return errors.New(ErrBuilder, errorMessages, "no opening tag for raw text")
}

// CommentNode is specified on the Builder interface.
func (n *NodeBuilder) CommentNode(comment string) error {
	if n.done {
		return errors.New(ErrBuilder, errorMessages, "building is already done")
	}
	if len(n.stack) > 0 {
		n.stack[len(n.stack)-1].appendCommentNode(comment)
		return nil
	}
	return errors.New(ErrBuilder, errorMessages, "no opening tag for comment")
}

// EOF
