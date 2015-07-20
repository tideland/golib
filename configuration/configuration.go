// Tideland Go Library - Configuration
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package configuration

//--------------------
// IMPORTS
//--------------------

import (
	"io"
	"strings"

	"github.com/tideland/golib/collections"
	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/sml"
)

//--------------------
// VALUE
//--------------------

// Value contains the found string value at a given path
// or an error if the path is invalid. A convenient converting
// can be done using the stringex.Defaulter. So e.g. accessing
// a configuration and interpret it as int woud be
//
//   myInt := myDefaulter.AsInt(myConfig.At("path", "to", "value"), 42)
//
// An error check can also be done by myValue.Error().
type Value interface {
	// List returns the configuration keys below the value.
	List() ([]string, error)

	// Value returns the value as string.
	Value() (string, error)

	// Error returns the error if the value retrieval fails.
	Error() error
}

// value implements Value.
type value struct {
	path    []string
	changer collections.KeyStringValueChanger
}

// List implements the Value interface.
func (v *value) List() ([]string, error) {
	kvs, err := v.changer.List()
	if err != nil {
		return nil, errors.New(ErrInvalidPath, errorMessages, pathToString(v.path))
	}
	var ks []string
	for _, kv := range kvs {
		ks = append(ks, kv.Key)
	}
	return ks, nil
}

// Value implements the Value interface.
func (v *value) Value() (string, error) {
	sv, err := v.changer.Value()
	if err != nil {
		return "", errors.New(ErrInvalidPath, errorMessages, pathToString(v.path))
	}
	return sv, nil
}

// Error implements the Value interface.
func (v *value) Error() error {
	return v.changer.Error()
}

//--------------------
// CONFIGURATION
//--------------------

// Configuration contains the read configuration and provides
// typed access to it. The root node "config" is automatically
// preceded to the path.
type Configuration interface {
	// At retrieves the value at a given path.
	At(path ...string) Value

	// Apply creates a new configuration by adding of overwriting
	// the passed values. The keys of the map have to be slash
	// separated configuration paths without the leading "config".
	Apply(kvs map[string]string) (Configuration, error)
}

// Read reads the SML source of the configuration from a
// reader, parses it, and returns the configuration instance.
func Read(source io.Reader) (Configuration, error) {
	builder := sml.NewKeyStringValueTreeBuilder()
	err := sml.ReadSML(source, builder)
	if err != nil {
		return nil, errors.Annotate(err, ErrIllegalSourceFormat, errorMessages)
	}
	tree, err := builder.Tree()
	if err != nil {
		return nil, errors.Annotate(err, ErrIllegalSourceFormat, errorMessages)
	}
	if err := tree.At("config").Error(); err != nil {
		return nil, errors.Annotate(err, ErrIllegalSourceFormat, errorMessages)
	}
	return &configuration{tree}, nil
}

// ReadString reads the SML source of the configuration from a
// string, parses it, and returns the configuration instance.
func ReadString(source string) (Configuration, error) {
	return Read(strings.NewReader(source))
}

// configuration implements the Configuration interface.
type configuration struct {
	values collections.KeyStringValueTree
}

// At implements the Configuration interface.
func (c *configuration) At(path ...string) Value {
	path = append([]string{"config"}, path...)
	changer := c.values.At(path...)
	return &value{path, changer}
}

// Apply implements the Configuration interface.
func (c *configuration) Apply(kvs map[string]string) (Configuration, error) {
	cc := &configuration{
		values: c.values.Copy(),
	}
	for key, value := range kvs {
		path := append([]string{"config"}, strings.Split(key, "/")...)
		_, err := cc.values.Create(path...).SetValue(value)
		if err != nil {
			return nil, errors.Annotate(err, ErrCannotApply, errorMessages)
		}
	}
	return cc, nil
}

//--------------------
// BUILDER
//--------------------

// configBuilder implements sml.Builder to parse the
// configuration source and create the tree containing
// the values.
type configBuilder struct {
	stack  collections.StringStack
	values collections.KeyStringValueTree
}

// BeginTagNode implements the sml.Builder interface.
func (b *configBuilder) BeginTagNode(tag string) error {
	switch {
	case b.values == nil && tag != "config":
		return errors.New(ErrIllegalConfigSource, errorMessages, `does not start with "config" node`)
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
func (b *configBuilder) EndTagNode() error {
	_, err := b.stack.Pop()
	return err
}

// TextNode implements the sml.Builder interface.
func (b *configBuilder) TextNode(text string) error {
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
func (b *configBuilder) RawNode(raw string) error {
	return b.TextNode(raw)
}

// Comment implements the sml.Builder interface.
func (b *configBuilder) CommentNode(comment string) error {
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
