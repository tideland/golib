// Tideland Go Library - Cells - Constants
//
// Copyright (C) 2010-2015 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"time"
)

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

	// Default timeout for requests to cells.
	DefaultTimeout = 5 * time.Second

	// minEventBufferSize is the minimum size of the
	// event buffer per cell.
	minEventBufferSize = 16

	// minRecoveringNumber and minRecoveringDuration
	// control the default recovering frequency.
	minRecoveringNumber   = 10
	minRecoveringDuration = time.Second

	// minEmitTimeout is the minimum allowed timeout
	// for event emitting (see below).
	minEmitTimeout = time.Second

	// maxEmitTimeout is the maximum time to emit an
	// event into a cells event buffer before a timeout
	// error is returned to the emitter.
	maxEmitTimeout = 30 * time.Second
)

// EOF
