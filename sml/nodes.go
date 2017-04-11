// Tideland Go Library - Simple Markup Language - Nodes
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package sml

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

//--------------------
// TAG NODE
//--------------------

// tagNode represents a node with one multipart tag and zero to many
// children nodes.
type tagNode struct {
	tag      []string
	children []Node
}

// newTagNode creates a node with the given tag.
func newTagNode(tag string) (*tagNode, error) {
	vtag, err := ValidateTag(tag)
	if err != nil {
		return nil, err
	}
	return &tagNode{
		tag:      vtag,
		children: []Node{},
	}, nil
}

// appendTagNode creates a new tag node, appends it as last child
// and returns it.
func (tn *tagNode) appendTagNode(tag string) (*tagNode, error) {
	ntn, err := newTagNode(tag)
	if err != nil {
		return nil, err
	}
	tn.appendChild(ntn)
	return ntn, nil
}

// appendTextNode creates a text node, appends it as last child
// and returns it.
func (tn *tagNode) appendTextNode(text string) *textNode {
	trimmedText := strings.TrimSpace(text)
	if trimmedText == "" {
		return nil
	}
	ntn := newTextNode(trimmedText)
	tn.appendChild(ntn)
	return ntn
}

// appendRawNode creates a raw node, appends it as last child
// and returns it.
func (tn *tagNode) appendRawNode(raw string) *rawNode {
	nrn := newRawNode(raw)
	tn.appendChild(nrn)
	return nrn
}

// appendCommentNode creates a comment node, appends it as last child
// and returns it.
func (tn *tagNode) appendCommentNode(comment string) *commentNode {
	ncn := newCommentNode(comment)
	tn.appendChild(ncn)
	return ncn
}

// appendChild adds a node as last child.
func (tn *tagNode) appendChild(n Node) {
	tn.children = append(tn.children, n)
}

// Tag returns the tag parts.
func (tn *tagNode) Tag() []string {
	out := make([]string, len(tn.tag))
	copy(out, tn.tag)
	return out
}

// Len return the number of children of this node.
func (tn *tagNode) Len() int {
	return 1 + len(tn.children)
}

// ProcessWith processes the node and all chidlren recursively
// with the passed processor.
func (tn *tagNode) ProcessWith(p Processor) error {
	if err := p.OpenTag(tn.tag); err != nil {
		return err
	}
	for _, child := range tn.children {
		if err := child.ProcessWith(p); err != nil {
			return err
		}
	}
	return p.CloseTag(tn.tag)
}

// String returns the tag node as string.
func (tn *tagNode) String() string {
	var buf bytes.Buffer
	context := NewWriterContext(NewStandardSMLWriter(), &buf, true, "\t")
	WriteSML(tn, context)
	return buf.String()
}

//--------------------
// TEXT NODE
//--------------------

// textNode is a node containing some text.
type textNode struct {
	text string
}

// newTextNode creates a new text node.
func newTextNode(text string) *textNode {
	return &textNode{strings.TrimSpace(text)}
}

// Tag returns nil.
func (tn *textNode) Tag() []string {
	return nil
}

// Len returns the len of the text in the text node.
func (tn *textNode) Len() int {
	return len(tn.text)
}

// ProcessWith processes the text node with the given
// processor.
func (tn *textNode) ProcessWith(p Processor) error {
	return p.Text(tn.text)
}

// String returns the text node as string.
func (tn *textNode) String() string {
	return tn.text
}

//--------------------
// RAW NODE
//--------------------

// rawNode is a node containing some raw data.
type rawNode struct {
	raw string
}

// newRawNode creates a new raw node.
func newRawNode(raw string) *rawNode {
	return &rawNode{raw}
}

// Tag returns nil.
func (rn *rawNode) Tag() []string {
	return nil
}

// Len returns the len of the data in the raw node.
func (rn *rawNode) Len() int {
	return len(rn.raw)
}

// ProcessWith processes the raw node with the given
// processor.
func (rn *rawNode) ProcessWith(p Processor) error {
	return p.Raw(rn.raw)
}

// String returns the raw node as string.
func (rn *rawNode) String() string {
	return rn.raw
}

//--------------------
// COMMENT NODE
//--------------------

// commentNode is a node containing a comment.
type commentNode struct {
	comment string
}

// newCommentNode creates a new comment node.
func newCommentNode(comment string) *commentNode {
	return &commentNode{strings.TrimSpace(comment)}
}

// Tag returns nil.
func (cn *commentNode) Tag() []string {
	return nil
}

// Len returns the len of the data in the comment node.
func (cn *commentNode) Len() int {
	return len(cn.comment)
}

// ProcessWith processes the comment node with the given
// processor.
func (cn *commentNode) ProcessWith(p Processor) error {
	return p.Comment(cn.comment)
}

// String returns the comment node as string.
func (cn *commentNode) String() string {
	return cn.comment
}

//--------------------
// PRIVATE FUNCTIONS
//--------------------

// validTagRe contains the regular expression for
// the validation of tags.
var validTagRe *regexp.Regexp

// init the regexp for valid tags.
func init() {
	var err error
	validTagRe, err = regexp.Compile(`^([a-z][a-z0-9]*(\-[a-z0-9]+)*)(:([a-z0-9]+(\-[a-z0-9]+)*))*$`)
	if err != nil {
		panic(err)
	}
}

// ValidateTag checks if a tag is valid. Only
// the chars 'a' to 'z', '0' to '9', '-' and ':' are
// accepted. It also transforms it to lowercase
// and splits the parts at the colons.
func ValidateTag(tag string) ([]string, error) {
	ltag := strings.ToLower(tag)
	if !validTagRe.MatchString(ltag) {
		return nil, fmt.Errorf("invalid tag: %q", tag)
	}
	ltags := strings.Split(ltag, ":")
	return ltags, nil
}

// EOF
