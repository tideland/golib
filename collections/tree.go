// Tideland Go Library - Collections - Tree
//
// Copyright (C) 2015-2017 Frank Mueller / Tideland / Oldenburg / Germany
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
// NODE VALUE
//--------------------

// nodeContent is the base interface for the different value
// containing types.
type nodeContent interface {
	// key returns the key for finding ands
	// check for duplicates.
	key() interface{}

	// value returns the value itself.
	value() interface{}

	// deepCopy creates a copy of the node content.
	deepCopy() nodeContent
}

// justValue has same key and value.
type justValue struct {
	v interface{}
}

// key implements the nodeContent interface.
func (v justValue) key() interface{} {
	return v.v
}

// value implements the nodeContent interface.
func (v justValue) value() interface{} {
	return v.v
}

// deepCopy implements the nodeContent interface.
func (v justValue) deepCopy() nodeContent {
	return justValue{v.v}
}

// String implements the Stringer interface.
func (v justValue) String() string {
	return fmt.Sprintf("%v", v.v)
}

// keyValue has different key and value.
type keyValue struct {
	k interface{}
	v interface{}
}

// key implements the nodeContent interface.
func (v keyValue) key() interface{} {
	return v.k
}

// value implements the nodeContent interface.
func (v keyValue) value() interface{} {
	return v.v
}

// deepCopy implements the nodeContent interface.
func (v keyValue) deepCopy() nodeContent {
	return keyValue{v.k, v.v}
}

// String implements the Stringer interface.
func (v keyValue) String() string {
	return fmt.Sprintf("%v = '%v'", v.k, v.v)
}

//--------------------
// NODE CONTAINER
//--------------------

// nodeContainer is the top element of all nodes and provides
// configuration.
type nodeContainer struct {
	root       *node
	duplicates bool
}

// newNodeContainer creates a new node container.
func newNodeContainer(c nodeContent, duplicates bool) *nodeContainer {
	nc := &nodeContainer{
		root: &node{
			content: c,
		},
		duplicates: duplicates,
	}
	nc.root.container = nc
	return nc
}

// deepCopy creates a copy of the container.
func (nc *nodeContainer) deepCopy() *nodeContainer {
	cnc := &nodeContainer{
		duplicates: nc.duplicates,
	}
	cnc.root = nc.root.deepCopy(cnc, nil)
	return cnc
}

//--------------------
// NODE
//--------------------

// node contains the value and structural information of a node.
type node struct {
	container *nodeContainer
	parent    *node
	content   nodeContent
	children  []*node
}

// isAllowed returns true, if adding the content or setting
// it is allowed depending on allowed duplicates.
func (n *node) isAllowed(c nodeContent, here bool) bool {
	if n.container.duplicates {
		return true
	}
	checkNode := n
	if here {
		checkNode = n.parent
	}
	for _, child := range checkNode.children {
		if child.content.key() == c.key() {
			return false
		}
	}
	return true
}

// hasDuplicateSibling checks if the node has a sibling with the same key.
func (n *node) hasDuplicateSibling(key interface{}) bool {
	if n.parent == nil {
		return false
	}
	for _, sibling := range n.parent.children {
		if sibling == n {
			continue
		}
		if sibling.content.key() == key {
			return true
		}
	}
	return false
}

// addChild adds a child node depending on allowed duplicates.
func (n *node) addChild(c nodeContent) (*node, error) {
	if !n.isAllowed(c, false) {
		return nil, errors.New(ErrDuplicate, errorMessages)
	}
	child := &node{
		container: n.container,
		parent:    n,
		content:   c,
	}
	n.children = append(n.children, child)
	return child, nil
}

// remove deletes this node from its parent.
func (n *node) remove() error {
	if n.parent == nil {
		return errors.New(ErrCannotRemoveRoot, errorMessages)
	}
	for i, child := range n.parent.children {
		if child == n {
			n.parent.children = append(n.parent.children[:i], n.parent.children[i+1:]...)
			return nil
		}
	}
	panic("cannot find node to remove at parent")
}

// at finds a node by its path.
func (n *node) at(path ...nodeContent) (*node, error) {
	if len(path) == 0 || path[0].key() != n.content.key() {
		return nil, errors.New(ErrNodeNotFound, errorMessages)
	}
	if len(path) == 1 {
		return n, nil
	}
	// Check children for rest of the path.
	for _, child := range n.children {
		found, err := child.at(path[1:]...)
		if err != nil && !IsNodeNotFoundError(err) {
			return nil, errors.Annotate(err, ErrIllegalPath, errorMessages)
		}
		if found != nil {
			return found, nil
		}
	}
	return nil, errors.New(ErrNodeNotFound, errorMessages)
}

// create acts like at but if nodes don't exist they will be created.
func (n *node) create(path ...nodeContent) (*node, error) {
	if len(path) == 0 || path[0].key() != n.content.key() {
		return nil, errors.New(ErrNodeNotFound, errorMessages)
	}
	if len(path) == 1 {
		return n, nil
	}
	// Check children for the next path element.
	var found *node
	for _, child := range n.children {
		if path[1].key() == child.content.key() {
			found = child
			break
		}
	}
	if found == nil {
		child, err := n.addChild(path[1])
		if err != nil {
			return nil, errors.Annotate(err, ErrNodeAddChild, errorMessages)
		}
		return child.create(path[1:]...)
	}
	return found.create(path[1:]...)
}

// findFirst returns the first node for which the passed function
// returns true.
func (n *node) findFirst(f func(fn *node) (bool, error)) (*node, error) {
	hasFound, err := f(n)
	if err != nil {
		return nil, errors.Annotate(err, ErrNodeFindFirst, errorMessages)
	}
	if hasFound {
		return n, nil
	}
	for _, child := range n.children {
		found, err := child.findFirst(f)
		if err != nil && !IsNodeNotFoundError(err) {
			return nil, errors.Annotate(err, ErrNodeFindFirst, errorMessages)
		}
		if found != nil {
			return found, nil
		}
	}
	return nil, errors.New(ErrNodeNotFound, errorMessages)
}

// findAll returns all nodes for which the passed function
// returns true.
func (n *node) findAll(f func(fn *node) (bool, error)) ([]*node, error) {
	var allFound []*node
	hasFound, err := f(n)
	if err != nil {
		return nil, errors.Annotate(err, ErrNodeFindAll, errorMessages)
	}
	if hasFound {
		allFound = append(allFound, n)
	}
	for _, child := range n.children {
		found, err := child.findAll(f)
		if err != nil {
			return nil, errors.Annotate(err, ErrNodeFindAll, errorMessages)
		}
		if found != nil {
			allFound = append(allFound, found...)
		}
	}
	return allFound, nil
}

// doAll performs the passed function for the node
// and all its children deep to the leafs.
func (n *node) doAll(f func(dn *node) error) error {
	if err := f(n); err != nil {
		return errors.Annotate(err, ErrNodeDoAll, errorMessages)
	}
	for _, child := range n.children {
		if err := child.doAll(f); err != nil {
			return errors.Annotate(err, ErrNodeDoAll, errorMessages)
		}
	}
	return nil
}

// doChildren performs the passed function for all children.
func (n *node) doChildren(f func(cn *node) error) error {
	for _, child := range n.children {
		if err := f(child); err != nil {
			return errors.Annotate(err, ErrNodeDoChildren, errorMessages)
		}
	}
	return nil
}

// size recursively calculates the size of the nodeContainer.
func (n *node) size() int {
	l := 1
	for _, child := range n.children {
		l += child.size()
	}
	return l
}

// deepCopy creates a copy of the node.
func (n *node) deepCopy(c *nodeContainer, p *node) *node {
	cn := &node{
		container: c,
		parent:    p,
		content:   n.content.deepCopy(),
		children:  make([]*node, len(n.children)),
	}
	for i, child := range n.children {
		cn.children[i] = child.deepCopy(c, cn)
	}
	return cn
}

// String implements the Stringer interface.
func (n *node) String() string {
	out := fmt.Sprintf("[%v", n.content)
	if len(n.children) > 0 {
		out += " "
		for _, child := range n.children {
			out += child.String()
		}
	}
	out += "]"
	return out
}

//--------------------
// TREE
//--------------------

// tree implements the Tree interface.
type tree struct {
	container *nodeContainer
}

// NewTree creates a new tree with or without duplicate
// values for children.
func NewTree(v interface{}, duplicates bool) Tree {
	return &tree{
		container: newNodeContainer(justValue{v}, duplicates),
	}
}

// At implements the Tree interface.
func (t *tree) At(values ...interface{}) Changer {
	var path []nodeContent
	for _, value := range values {
		path = append(path, justValue{value})
	}
	n, err := t.container.root.at(path...)
	return &changer{n, err}
}

// Root implements the Tree interface.
func (t *tree) Root() Changer {
	return &changer{t.container.root, nil}
}

// Create implements the Tree interface.
func (t *tree) Create(values ...interface{}) Changer {
	var path []nodeContent
	for _, value := range values {
		path = append(path, justValue{value})
	}
	n, err := t.container.root.create(path...)
	return &changer{n, err}
}

// FindFirst implements the Tree interface.
func (t *tree) FindFirst(f func(v interface{}) (bool, error)) Changer {
	n, err := t.container.root.findFirst(func(fn *node) (bool, error) {
		return f(fn.content.value())
	})
	return &changer{n, err}
}

// FindFirst implements the Tree interface.
func (t *tree) FindAll(f func(v interface{}) (bool, error)) []Changer {
	ns, err := t.container.root.findAll(func(fn *node) (bool, error) {
		return f(fn.content.value())
	})
	if err != nil {
		return []Changer{&changer{nil, err}}
	}
	var cs []Changer
	for _, n := range ns {
		cs = append(cs, &changer{n, nil})
	}
	return cs
}

// DoAll implements the Tree interface.
func (t *tree) DoAll(f func(v interface{}) error) error {
	return t.container.root.doAll(func(dn *node) error {
		return f(dn.content.value())
	})
}

// DoAllDeep implements the Tree interface.
func (t *tree) DoAllDeep(f func(vs []interface{}) error) error {
	return t.container.root.doAll(func(dn *node) error {
		values := []interface{}{}
		cn := dn
		for cn != nil {
			values = append([]interface{}{cn.content.value()}, values...)
			cn = cn.parent
		}
		return f(values)
	})
}

// Len implements the Tree interface.
func (t *tree) Len() int {
	return t.container.root.size()
}

// Copy implements the Tree interface.
func (t *tree) Copy() Tree {
	return &tree{
		container: t.container.deepCopy(),
	}
}

// Deflate implements the Tree interface.
func (t *tree) Deflate(v interface{}) {
	t.container.root = &node{
		content: justValue{v},
	}
}

// String implements the Stringer interface.
func (t *tree) String() string {
	return t.container.root.String()
}

//--------------------
// STRING TREE
//--------------------

// stringTree implements the StringTree interface.
type stringTree struct {
	container *nodeContainer
}

// NewStringTree creates a new string tree with or without
// duplicate values for children.
func NewStringTree(v string, duplicates bool) StringTree {
	return &stringTree{
		container: newNodeContainer(justValue{v}, duplicates),
	}
}

// At implements the StringTree interface.
func (t *stringTree) At(values ...string) StringChanger {
	var path []nodeContent
	for _, value := range values {
		path = append(path, justValue{value})
	}
	n, err := t.container.root.at(path...)
	return &stringChanger{n, err}
}

// Root implements the StringTree interface.
func (t *stringTree) Root() StringChanger {
	return &stringChanger{t.container.root, nil}
}

// Create implements the StringTree interface.
func (t *stringTree) Create(values ...string) StringChanger {
	var path []nodeContent
	for _, value := range values {
		path = append(path, justValue{value})
	}
	n, err := t.container.root.create(path...)
	return &stringChanger{n, err}
}

// FindFirst implements the StringTree interface.
func (t *stringTree) FindFirst(f func(v string) (bool, error)) StringChanger {
	n, err := t.container.root.findFirst(func(fn *node) (bool, error) {
		return f(fn.content.value().(string))
	})
	return &stringChanger{n, err}
}

// FindFirst implements the StringTree interface.
func (t *stringTree) FindAll(f func(v string) (bool, error)) []StringChanger {
	ns, err := t.container.root.findAll(func(fn *node) (bool, error) {
		return f(fn.content.value().(string))
	})
	if err != nil {
		return []StringChanger{&stringChanger{nil, err}}
	}
	var cs []StringChanger
	for _, n := range ns {
		cs = append(cs, &stringChanger{n, nil})
	}
	return cs
}

// DoAll implements the StringTree interface.
func (t *stringTree) DoAll(f func(v string) error) error {
	return t.container.root.doAll(func(dn *node) error {
		return f(dn.content.value().(string))
	})
}

// DoAllDeep implements the StringTree interface.
func (t *stringTree) DoAllDeep(f func(vs []string) error) error {
	return t.container.root.doAll(func(dn *node) error {
		values := []string{}
		cn := dn
		for cn != nil {
			values = append([]string{cn.content.value().(string)}, values...)
			cn = cn.parent
		}
		return f(values)
	})
}

// Len implements the StringTree interface.
func (t *stringTree) Len() int {
	return t.container.root.size()
}

// Copy implements the StringTree interface.
func (t *stringTree) Copy() StringTree {
	return &stringTree{
		container: t.container.deepCopy(),
	}
}

// Deflate implements the StringTree interface.
func (t *stringTree) Deflate(v string) {
	t.container.root = &node{
		content: justValue{v},
	}
}

// String implements the Stringer interface.
func (t *stringTree) String() string {
	return t.container.root.String()
}

//--------------------
// KEY/VALUE TREE
//--------------------

// keyValueTree implements the KeyValueTree interface.
type keyValueTree struct {
	container *nodeContainer
}

// NewKeyValueTree creates a new key/value tree with or without
// duplicate values for children.
func NewKeyValueTree(k string, v interface{}, duplicates bool) KeyValueTree {
	return &keyValueTree{
		container: newNodeContainer(keyValue{k, v}, duplicates),
	}
}

// At implements the KeyValueTree interface.
func (t *keyValueTree) At(keys ...string) KeyValueChanger {
	var path []nodeContent
	for _, key := range keys {
		path = append(path, keyValue{key, nil})
	}
	n, err := t.container.root.at(path...)
	return &keyValueChanger{n, err}
}

// Root implements the KeyValueTree interface.
func (t *keyValueTree) Root() KeyValueChanger {
	return &keyValueChanger{t.container.root, nil}
}

// Create implements the KeyValueTree interface.
func (t *keyValueTree) Create(keys ...string) KeyValueChanger {
	var path []nodeContent
	for _, key := range keys {
		path = append(path, keyValue{key, nil})
	}
	n, err := t.container.root.create(path...)
	return &keyValueChanger{n, err}
}

// FindFirst implements the KeyValueTree interface.
func (t *keyValueTree) FindFirst(f func(k string, v interface{}) (bool, error)) KeyValueChanger {
	n, err := t.container.root.findFirst(func(fn *node) (bool, error) {
		return f(fn.content.key().(string), fn.content.value())
	})
	return &keyValueChanger{n, err}
}

// FindFirst implements the KeyValueTree interface.
func (t *keyValueTree) FindAll(f func(k string, v interface{}) (bool, error)) []KeyValueChanger {
	ns, err := t.container.root.findAll(func(fn *node) (bool, error) {
		return f(fn.content.key().(string), fn.content.value())
	})
	if err != nil {
		return []KeyValueChanger{&keyValueChanger{nil, err}}
	}
	var cs []KeyValueChanger
	for _, n := range ns {
		cs = append(cs, &keyValueChanger{n, nil})
	}
	return cs
}

// DoAll implements the KeyValueTree interface.
func (t *keyValueTree) DoAll(f func(k string, v interface{}) error) error {
	return t.container.root.doAll(func(dn *node) error {
		return f(dn.content.key().(string), dn.content.value())
	})
}

// DoAllDeep implements the KeyValueTree interface.
func (t *keyValueTree) DoAllDeep(f func(ks []string, v interface{}) error) error {
	return t.container.root.doAll(func(dn *node) error {
		keys := []string{}
		cn := dn
		for cn != nil {
			keys = append([]string{cn.content.key().(string)}, keys...)
			cn = cn.parent
		}
		return f(keys, dn.content.value())
	})
}

// Len implements the KeyValueTree interface.
func (t *keyValueTree) Len() int {
	return t.container.root.size()
}

// Copy implements the KeyValueTree interface.
func (t *keyValueTree) Copy() KeyValueTree {
	return &keyValueTree{
		container: t.container.deepCopy(),
	}
}

// CopyAt implements the KeyValueTree interface.
func (t *keyValueTree) CopyAt(keys ...string) (KeyValueTree, error) {
	var path []nodeContent
	for _, key := range keys {
		path = append(path, keyValue{key, ""})
	}
	n, err := t.container.root.at(path...)
	if err != nil {
		return nil, err
	}
	nc := &nodeContainer{
		duplicates: t.container.duplicates,
	}
	nc.root = n.deepCopy(nc, nil)
	return &keyValueTree{nc}, nil
}

// Deflate implements the KeyValueTree interface.
func (t *keyValueTree) Deflate(k string, v interface{}) {
	t.container.root = &node{
		content: keyValue{k, v},
	}
}

// String implements the Stringer interface.
func (t *keyValueTree) String() string {
	return t.container.root.String()
}

//--------------------
// KEY/STRING VALUE TREE
//--------------------

// keyStringValueTree implements the KeyStringValueTree interface.
type keyStringValueTree struct {
	container *nodeContainer
}

// NewKeyStringValueTree creates a new key/value tree with or without
// duplicate values for children and strings as values.
func NewKeyStringValueTree(k, v string, duplicates bool) KeyStringValueTree {
	return &keyStringValueTree{
		container: newNodeContainer(keyValue{k, v}, duplicates),
	}
}

// At implements the KeyStringValueTree interface.
func (t *keyStringValueTree) At(keys ...string) KeyStringValueChanger {
	var path []nodeContent
	for _, key := range keys {
		path = append(path, keyValue{key, ""})
	}
	n, err := t.container.root.at(path...)
	return &keyStringValueChanger{n, err}
}

// Root implements the KeyStringValueTree interface.
func (t *keyStringValueTree) Root() KeyStringValueChanger {
	return &keyStringValueChanger{t.container.root, nil}
}

// Create implements the KeyStringValueTree interface.
func (t *keyStringValueTree) Create(keys ...string) KeyStringValueChanger {
	var path []nodeContent
	for _, key := range keys {
		path = append(path, keyValue{key, ""})
	}
	n, err := t.container.root.create(path...)
	return &keyStringValueChanger{n, err}
}

// FindFirst implements the KeyStringValueTree interface.
func (t *keyStringValueTree) FindFirst(f func(k, v string) (bool, error)) KeyStringValueChanger {
	n, err := t.container.root.findFirst(func(fn *node) (bool, error) {
		return f(fn.content.key().(string), fn.content.value().(string))
	})
	return &keyStringValueChanger{n, err}
}

// FindFirst implements the KeyStringValueTree interface.
func (t *keyStringValueTree) FindAll(f func(k, v string) (bool, error)) []KeyStringValueChanger {
	ns, err := t.container.root.findAll(func(fn *node) (bool, error) {
		return f(fn.content.key().(string), fn.content.value().(string))
	})
	if err != nil {
		return []KeyStringValueChanger{&keyStringValueChanger{nil, err}}
	}
	var cs []KeyStringValueChanger
	for _, n := range ns {
		cs = append(cs, &keyStringValueChanger{n, nil})
	}
	return cs
}

// DoAll implements the KeyStringValueTree interface.
func (t *keyStringValueTree) DoAll(f func(k, v string) error) error {
	return t.container.root.doAll(func(dn *node) error {
		return f(dn.content.key().(string), dn.content.value().(string))
	})
}

// DoAllDeep implements the KeyStringValueTree interface.
func (t *keyStringValueTree) DoAllDeep(f func(ks []string, v string) error) error {
	return t.container.root.doAll(func(dn *node) error {
		keys := []string{}
		cn := dn
		for cn != nil {
			keys = append([]string{cn.content.key().(string)}, keys...)
			cn = cn.parent
		}
		return f(keys, dn.content.value().(string))
	})
}

// Len implements the KeyStringValueTree interface.
func (t *keyStringValueTree) Len() int {
	return t.container.root.size()
}

// Copy implements the KeyStringValueTree interface.
func (t *keyStringValueTree) Copy() KeyStringValueTree {
	return &keyStringValueTree{
		container: t.container.deepCopy(),
	}
}

// CopyAt implements the KeyStringValueTree interface.
func (t *keyStringValueTree) CopyAt(keys ...string) (KeyStringValueTree, error) {
	var path []nodeContent
	for _, key := range keys {
		path = append(path, keyValue{key, ""})
	}
	n, err := t.container.root.at(path...)
	if err != nil {
		return nil, err
	}
	nc := &nodeContainer{
		duplicates: t.container.duplicates,
	}
	nc.root = n.deepCopy(nc, nil)
	return &keyStringValueTree{nc}, nil
}

// Deflate implements the KeyStringValueTree interface.
func (t *keyStringValueTree) Deflate(k, v string) {
	t.container.root = &node{
		content: keyValue{k, v},
	}
}

// String implements the Stringer interface.
func (t *keyStringValueTree) String() string {
	return t.container.root.String()
}

// EOF
