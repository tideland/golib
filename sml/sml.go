// Tideland Go Library - Simple Markup Language
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package sml

//--------------------
// IMPORTS
//--------------------

//--------------------
// PROCESSOR
//--------------------

// Processor represents any type able to process
// a node structure. Four callbacks then will be
// called with the according data.
type Processor interface {
	// OpenTag is called if a tag is opened.
	OpenTag(tag []string) error

	// CloseTag is called if a tag is closed.
	CloseTag(tag []string) error

	// Text is called for each text data inside a node.
	Text(text string) error

	// Raw is called for each raw data inside a node.
	Raw(raw string) error

	// Comment is called for each comment data inside a node.
	Comment(comment string) error
}

//--------------------
// BUILDER
//--------------------

// Builder defines the callbacks for the reader
// to handle parts of the document.
type Builder interface {
	// BeginTagNode is called when new tag node begins.
	BeginTagNode(tag string) error

	// EndTagNode is called when the tag node ends.
	EndTagNode() error

	// TextNode is called for each text data.
	TextNode(text string) error

	// RawNode is called for each raw data.
	RawNode(raw string) error

	// Comment is called for each comment data inside a node.
	CommentNode(comment string) error
}

//--------------------
// NODES
//--------------------

// Node represents the common interface of all nodes (tags and text).
type Node interface {
	// Tag returns the tag in case of a tag node, otherwise nil.
	Tag() []string

	// Len returns the length of a text or the number of subnodes,
	// depending on the concrete type of the node.
	Len() int

	// ProcessWith is called for the processing of this node.
	ProcessWith(p Processor) error

	// String returns a simple string representation of the node.
	String() string
}

// EOF
