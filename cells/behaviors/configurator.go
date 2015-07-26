// Tideland Go Library - Cell Behaviors - Configurator
//
// Copyright (C) 2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package behaviors

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/cells"
	"github.com/tideland/golib/configuration"
	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/logger"
)

//--------------------
// CONVENIENCE
//--------------------

// Configuration returns the configuration payload
// of the passed event or an empty configuration.
func Configuration(event cells.Event) configuration.Configuration {
	payload, ok := event.Payload().Get(ConfigurationPayload)
	if !ok {
		logger.Warningf("event does not contain configuration payload")
		config, _ := configuration.ReadString("{config}")
		return config
	}
	config, ok := payload.(configuration.Configuration)
	if !ok {
		logger.Warningf("configuration payload has illegal type")
		config, _ := configuration.ReadString("{config}")
		return config
	}
	return config
}

//--------------------
// CONFIGURATOR BEHAVIOR
//--------------------

// ConfigurationValidator defines a function for the validation of
// a new read configuration.
type ConfigurationValidator func(configuration.Configuration) error

// configuratorBehavior implements the configurator behavior.
type configuratorBehavior struct {
	ctx      cells.Context
	validate ConfigurationValidator
}

// NewConfiguratorBehavior creates the configurator behavior. It loads a
// configuration file and emits the it to its subscribers. If a validator
// is passed the read configuration will be validated using it. Errors
// will be logged.
func NewConfiguratorBehavior(validator ConfigurationValidator) cells.Behavior {
	return &configuratorBehavior{
		validate: validator,
	}
}

// Init implements the cells.Behavior interface.
func (b *configuratorBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	return nil
}

// Terminate implements the cells.Behavior interface.
func (b *configuratorBehavior) Terminate() error {
	return nil
}

// ProcessEvent reads, validates and emits a configuration.
func (b *configuratorBehavior) ProcessEvent(event cells.Event) error {
	switch event.Topic() {
	case ReadConfigurationTopic:
		// Read configuration
		filename, ok := event.Payload().GetString(ConfigurationFilenamePayload)
		if !ok {
			logger.Errorf("cannot read configuration without filename payload")
			return nil
		}
		logger.Infof("reading configuration from %q", filename)
		config, err := configuration.ReadFile(filename)
		if err != nil {
			return errors.Annotate(err, ErrCannotReadConfiguration, errorMessages)
		}
		// If wanted then validate it.
		if b.validate != nil {
			err = b.validate(config)
			if err != nil {
				return errors.Annotate(err, ErrCannotValidateConfiguration, errorMessages)
			}
		}
		// All done, emit it.
		pvs := cells.PayloadValues{
			ConfigurationPayload: config,
		}
		b.ctx.EmitNew(ConfigurationTopic, pvs, event.Scene())
	}
	return nil
}

// Recover implements the cells.Behavior interface.
func (b *configuratorBehavior) Recover(err interface{}) error {
	return nil
}

// EOF
