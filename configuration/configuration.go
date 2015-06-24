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
	"strconv"
	"strings"

	"github.com/tideland/golib/collections"
	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/sml"
)

//--------------------
// CONFIGURATION
//--------------------

// Configuration contains the read configuration and provides
// typed access to it. The root node "config" is automatically
// preceded to the path.
type Configuration interface {
	// List returns the configuration keys below the given path.
	List(path ...string) ([]string, error)

	// Get returns the string value at the given path.
	Get(path ...string) (string, error)

	// GetBool returns the boolean value at the given path.
	GetBool(path ...string) (bool, error)

	// GetInt returns the int value at the given path.
	GetInt(path ...string) (int, error)

	// GetFloat64 returns the float value at the given path.
	GetFloat64(path ...string) (float64, error)
}

// Read reads the SML source of the configuration from a
// reader, parses it, and returns the configuration instance.
func Read(source io.Reader) (Configuration, error) {
	builder := &configBuilder{}
	err := sml.ReadSML(source, builder)
	if err != nil {
		if errors.IsError(err, ErrIllegalConfigSource) {
			return nil, err
		}
		return nil, errors.Annotate(err, ErrIllegalSourceFormat, errorMessages)
	}
	return &configuration{builder.values}, nil
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

// List implements the Configuration interface.
func (c *configuration) List(path ...string) ([]string, error) {
	path = fullPath(path)
	kvs, err := c.values.At(path...).List()
	if err != nil {
		return nil, errors.New(ErrInvalidConfigurationPath, errorMessages, pathToString(path))
	}
	var ks []string
	for _, kv := range kvs {
		ks = append(ks, kv.Key)
	}
	return ks, nil
}

// Get implements the Configuration interface.
func (c *configuration) Get(path ...string) (string, error) {
	path = fullPath(path)
	value, err := c.values.At(path...).Value()
	if err != nil {
		return "", errors.New(ErrInvalidConfigurationPath, errorMessages, pathToString(path))
	}
	return value, nil
}

// GetBool implements the Configuration interface.
func (c *configuration) GetBool(path ...string) (bool, error) {
	raw, err := c.Get(path...)
	if err != nil {
		return false, err
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, errors.Annotate(err, ErrInvalidFormat, errorMessages, raw)
	}
	return value, nil
}

// GetInt implements the Configuration interface.
func (c *configuration) GetInt(path ...string) (int, error) {
	raw, err := c.Get(path...)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseInt(raw, 10, 0)
	if err != nil {
		return 0, errors.Annotate(err, ErrInvalidFormat, errorMessages, raw)
	}
	return int(value), nil
}

// GetFloat64 implements the Configuration interface.
func (c *configuration) GetFloat64(path ...string) (float64, error) {
	raw, err := c.Get(path...)
	if err != nil {
		return 0.0, err
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0.0, errors.Annotate(err, ErrInvalidFormat, errorMessages, raw)
	}
	return value, nil
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

// fullPath returns the passed path with a preceded "config".
func fullPath(path []string) []string {
	return append([]string{"config"}, path...)
}

// pathToString returns the path in a filesystem like notation.
func pathToString(path []string) string {
	return "/" + strings.Join(path, "/")
}

// EOF
