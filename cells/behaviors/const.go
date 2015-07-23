// Tideland Go Library - Cell Behaviors - Constants
//
// Copyright (C) 2010-2015 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package behaviors

//--------------------
// CONSTANTS
//--------------------

const (
	// Topics.
	ReadConfigurationTopic = "read-configuration!"
	ConfigurationTopic     = "configuration"
	TickerTopic            = "tick!"

	// Payload keys.
	ConfigurationFilenamePayload = "configuration:filename"
	ConfigurationPayload         = "configuration"
	TickerIDPayload              = "ticker:id"
	TickerTimePayload            = "ticker:time"
)

// EOF
