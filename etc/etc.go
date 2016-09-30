// Tideland Go Library - Etc
//
// Copyright (C) 2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package etc

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/tideland/golib/collections"
	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/sml"
	"github.com/tideland/golib/stringex"
)

//--------------------
// VALUE
//--------------------

// value helps to use the stringex.Defaulter.
type value struct {
	path    []string
	changer collections.KeyStringValueChanger
}

// Value retrieves the value or an error. It implements
// the Valuer interface.
func (v *value) Value() (string, error) {
	sv, err := v.changer.Value()
	if err != nil {
		return "", errors.New(ErrInvalidPath, errorMessages, pathToString(v.path))
	}
	return sv, nil
}

//--------------------
// ETC
//--------------------

// Etc contains the read etc configuration and provides access to
// it. The root node "etc" is automatically preceded to the path.
// The node name have to consist out of 'a' to 'z', '0' to '9', and
// '-'. The nodes of a path are separated by '/'.
type Etc interface {
	fmt.Stringer

	// ValueAsString retrieves the string value at a given path. If it
	// doesn't exist the default value dv is returned.
	ValueAsString(path, dv string) string

	// ValueAsBool retrieves the bool value at a given path. If it
	// doesn't exist the default value dv is returned.
	ValueAsBool(path string, dv bool) bool

	// ValueAsInt retrieves the int value at a given path. If it
	// doesn't exist the default value dv is returned.
	ValueAsInt(path string, dv int) int

	// ValueAsFloat64 retrieves the float64 value at a given path. If it
	// doesn't exist the default value dv is returned.
	ValueAsFloat64(path string, dv float64) float64

	// ValueAsTime retrieves the string value at a given path and
	// interprets it as time with the passed format. If it
	// doesn't exist the default value dv is returned.
	ValueAsTime(path, layout string, dv time.Time) time.Time

	// ValueAsDuration retrieves the duration value at a given path.
	// If it doesn't exist the default value dv is returned.
	ValueAsDuration(path string, dv time.Duration) time.Duration

	// Apply creates a new configuration by adding of overwriting
	// the passed values. The keys of the map have to be slash
	// separated configuration paths without the leading "etc".
	Apply(kvs map[string]string) (Etc, error)
}

// etc implements the Etc interface.
type etc struct {
	root      []string
	values    collections.KeyStringValueTree
	defaulter stringex.Defaulter
}

// Read reads the SML source of the configuration from a
// reader, parses it, and returns the etc instance.
func Read(source io.Reader) (Etc, error) {
	builder := sml.NewKeyStringValueTreeBuilder()
	err := sml.ReadSML(source, builder)
	if err != nil {
		return nil, errors.Annotate(err, ErrIllegalSourceFormat, errorMessages)
	}
	tree, err := builder.Tree()
	if err != nil {
		return nil, errors.Annotate(err, ErrIllegalSourceFormat, errorMessages)
	}
	if err := tree.At("etc").Error(); err != nil {
		return nil, errors.Annotate(err, ErrIllegalSourceFormat, errorMessages)
	}
	return &etc{
		root:      []string{"etc"},
		values:    tree,
		defaulter: stringex.NewDefaulter("etc", true),
	}, nil
}

// ReadString reads the SML source of the configuration from a
// string, parses it, and returns the etc instance.
func ReadString(source string) (Etc, error) {
	return Read(strings.NewReader(source))
}

// ReadFile reads the SML source of a configuration file,
// parses it, and returns the etc instance.
func ReadFile(filename string) (Etc, error) {
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Annotate(err, ErrCannotReadFile, errorMessages, filename)
	}
	return ReadString(string(source))
}

// ValueAsString implements the Etc interface.
func (e *etc) ValueAsString(path, dv string) string {
	value := e.valueAt(path)
	return e.defaulter.AsString(value, dv)
}

// ValueAsBool implements the Etc interface.
func (e *etc) ValueAsBool(path string, dv bool) bool {
	value := e.valueAt(path)
	return e.defaulter.AsBool(value, dv)
}

// ValueAsInt implements the Etc interface.
func (e *etc) ValueAsInt(path string, dv int) int {
	value := e.valueAt(path)
	return e.defaulter.AsInt(value, dv)
}

// ValueAsFloat64 implements the Etc interface.
func (e *etc) ValueAsFloat64(path string, dv float64) float64 {
	value := e.valueAt(path)
	return e.defaulter.AsFloat64(value, dv)
}

// ValueAsTime implements the Etc interface.
func (e *etc) ValueAsTime(path, format string, dv time.Time) time.Time {
	value := e.valueAt(path)
	return e.defaulter.AsTime(value, format, dv)
}

// ValueAsDuration implements the Etc interface.
func (e *etc) ValueAsDuration(path string, dv time.Duration) time.Duration {
	value := e.valueAt(path)
	return e.defaulter.AsDuration(value, dv)
}

// Apply implements the Etc interface.
func (e *etc) Apply(kvs map[string]string) (Etc, error) {
	ec := &etc{
		root:      e.root,
		values:    e.values.Copy(),
		defaulter: e.defaulter,
	}
	for key, value := range kvs {
		path := append(e.root, strings.Split(key, "/")...)
		_, err := ec.values.Create(path...).SetValue(value)
		if err != nil {
			return nil, errors.Annotate(err, ErrCannotApply, errorMessages)
		}
	}
	return ec, nil
}

// Apply implements the Stringer interface.
func (e *etc) String() string {
	return fmt.Sprintf("%v", e.values)
}

// valueAt retrieves and encapsulates the value
// at a given path.
func (e *etc) valueAt(path string) *value {
	fullPath := append(e.root, strings.Split(path, "/")...)
	changer := e.values.At(fullPath...)
	return &value{fullPath, changer}
}

//--------------------
// ETC BUILDER
//--------------------

// etcBuilder implements sml.Builder to parse the
// etc configuration source and create the tree
// containing the values.
type etcBuilder struct {
	stack  collections.StringStack
	values collections.KeyStringValueTree
}

// BeginTagNode implements the sml.Builder interface.
func (b *etcBuilder) BeginTagNode(tag string) error {
	switch {
	case b.values == nil && tag != "etc":
		return errors.New(ErrIllegalConfigSource, errorMessages, `does not start with "etc" node`)
	case b.values == nil:
		b.stack = collections.NewStringStack(tag)
		b.values = collections.NewKeyStringValueTree(tag, "", false)
	default:
		b.stack.Push(tag)
		changer := b.values.Create(b.stack.All()...)
		if changer.Error() != nil {
			return errors.New(ErrIllegalConfigSource, errorMessages, changer.Error())
		}
	}
	return nil
}

// EndTagNode implements the sml.Builder interface.
func (b *etcBuilder) EndTagNode() error {
	_, err := b.stack.Pop()
	return err
}

// TextNode implements the sml.Builder interface.
func (b *etcBuilder) TextNode(text string) error {
	value, err := b.values.At(b.stack.All()...).Value()
	if err != nil {
		return err
	}
	if value != "" {
		return errors.New(ErrIllegalConfigSource, errorMessages, `node has multiple values`)
	}
	text = strings.TrimSpace(text)
	if text != "" {
		_, err = b.values.At(b.stack.All()...).SetValue(text)
	}
	return err
}

// RawNode implements the sml.Builder interface.
func (b *etcBuilder) RawNode(raw string) error {
	return b.TextNode(raw)
}

// Comment implements the sml.Builder interface.
func (b *etcBuilder) CommentNode(comment string) error {
	return nil
}

//--------------------
// HELPERS
//--------------------

// pathToString returns the path in a filesystem like notation.
func pathToString(path []string) string {
	return "/" + strings.Join(path, "/")
}

// EOF
