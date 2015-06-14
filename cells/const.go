// Tideland Go Library- Cells - Constants
//
// Copyright (C) 2010-2015 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cells

//--------------------
// CONSTANTS
//--------------------

const (
	// Often used standard topics.
	CollectedTopic = "collected?"
	CountersTopic  = "counters?"
	PingTopic      = "ping?"
	ProcessedTopic = "processed?"
	ResetTopic     = "reset!"
	StatusTopic    = "status?"
	TickTopic      = "tick!"

	// Standard payload keys.
	DefaultPayload      = "default"
	ResponseChanPayload = "responseChan"
	TickerIDPayload     = "ticker:id"
	TickerTimePayload   = "ticker:time"

	// Special responses.
	PongResponse = "pong!"
)

// EOF
