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
	"time"

	"github.com/tideland/golib/collections"
	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/sml"
)

//--------------------
// VALUE
//--------------------

// Value contains the found value at a given path of
// an error if the path is invalid.
type Value interface {
	// List returns the configuration keys below the value.
	List() ([]string, error)

	// Get returns the value as string.
	Get() (string, error)

	// GetDefault returns the value as string or the default.
	GetDefault(d string) string

	// GetBool returns the value as bool.
	GetBool() (bool, error)

	// GetBoolDefault returns the value as bool or the default.
	GetBoolDefault(d bool) bool

	// GetInt returns the value as int.
	GetInt() (int, error)

	// GetIntDefault returns the value as int or the default.
	GetIntDefault(d int) int

	// GetFloat64 returns the value as float64.
	GetFloat64() (float64, error)

	// GetFloat64Default returns the value as float64 or the default.
	GetFloat64Default(d float64) float64

	// GetTime returns the value as time. It has
	// to be written in RfC 3339 format.
	GetTime() (time.Time, error)

	// GetTimeDefault returns the value as time or the default.
	GetTimeDefault(d time.Time) time.Time

	// GetDuration returns the value as duration.
	GetDuration() (time.Duration, error)

	// GetDurationDefault returns the value as duration or the default.
	GetDurationDefault(d time.Duration) time.Duration

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

// Get implements the Value interface.
func (v *value) Get() (string, error) {
	sv, err := v.changer.Value()
	if err != nil {
		return "", errors.New(ErrInvalidPath, errorMessages, pathToString(v.path))
	}
	return sv, nil
}

// GetDefault implements the Value interface.
func (v *value) GetDefault(d string) string {
	sv, err := v.Get()
	if err != nil {
		return d
	}
	return sv
}

// GetBool implements the Value interface.
func (v *value) GetBool() (bool, error) {
	raw, err := v.Get()
	if err != nil {
		return false, err
	}
	bv, err := strconv.ParseBool(raw)
	if err != nil {
		return false, errors.Annotate(err, ErrInvalidFormat, errorMessages, raw)
	}
	return bv, nil
}

// GetBoolDefault implements the Value interface.
func (v *value) GetBoolDefault(d bool) bool {
	bv, err := v.GetBool()
	if err != nil {
		return d
	}
	return bv
}

// GetInt implements the Value interface.
func (v *value) GetInt() (int, error) {
	raw, err := v.Get()
	if err != nil {
		return 0, err
	}
	iv, err := strconv.ParseInt(raw, 10, 0)
	if err != nil {
		return 0, errors.Annotate(err, ErrInvalidFormat, errorMessages, raw)
	}
	return int(iv), nil
}

// GetIntDefault implements the Value interface.
func (v *value) GetIntDefault(d int) int {
	iv, err := v.GetInt()
	if err != nil {
		return d
	}
	return iv
}

// GetFloat64 implements the Value interface.
func (v *value) GetFloat64() (float64, error) {
	raw, err := v.Get()
	if err != nil {
		return 0.0, err
	}
	fv, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0.0, errors.Annotate(err, ErrInvalidFormat, errorMessages, raw)
	}
	return fv, nil
}

// GetFloat64Default implements the Value interface.
func (v *value) GetFloat64Default(d float64) float64 {
	fv, err := v.GetFloat64()
	if err != nil {
		return d
	}
	return fv
}

// GetTime implements the Value interface.
func (v *value) GetTime() (time.Time, error) {
	raw, err := v.Get()
	if err != nil {
		return time.Time{}, err
	}
	tv, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return tv, errors.Annotate(err, ErrInvalidFormat, errorMessages, raw)
	}
	return tv, nil
}

// GetTimeDefault implements the Value interface.
func (v *value) GetTimeDefault(d time.Time) time.Time {
	tv, err := v.GetTime()
	if err != nil {
		return d
	}
	return tv
}

// GetDuration implements the Value interface.
func (v *value) GetDuration() (time.Duration, error) {
	raw, err := v.Get()
	if err != nil {
		return time.Duration(0), err
	}
	dv, err := time.ParseDuration(raw)
	if err != nil {
		return dv, errors.Annotate(err, ErrInvalidFormat, errorMessages, raw)
	}
	return dv, nil
}

// GetDurationDefault implements the Value interface.
func (v *value) GetDurationDefault(d time.Duration) time.Duration {
	dv, err := v.GetDuration()
	if err != nil {
		return d
	}
	return dv
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
