// Tideland Go Library - Cell Behaviors
//
// Copyright (C) 2010-2015 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The behaviors package provides several generic and always
// useful standard behaviors for the Tideland Go Library Cells.
// They are simply created with NewXyzBehavior(). The configuration
// is done by constructor arguments. Additionally some of them take
// functions or implementations of interfaces to control their
// processing.
//
// The behaviors are:
//
// - the broadcaster behavior simply emits all received events to all
// subscribers;
//
// - the collector behavior collects all received events and also emits
// them, they can be retrieved and resetted;
//
// - the counter behavior increments and emits counters identified by
// the return value of a configurable function and the individual events,
// the counters can be retrieved and resetted;
//
// - the filter behavior is created with a filtering function which is
// called for each event, when it returns true the event is emitted;
//
// - the FSM behavior implements a finite state machine, state functions
// process the events and return the following state;
//
// - the logger behavior logs every event at info level;
//
// - the mapper behavior is created with a mapping function processing
// each event and returning a new mapped one;
//
// - the round robin behavior distributes each received event round robin
// to its subscribers;
//
// - the scene behavior stores a received payload at the event topic as
// key in the event scene, useful in testing scenarios;
//
// - the ticker behavior emits a tick event in a defined interval to its
// subscribers.
package behaviors

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/version"
)

//--------------------
// VERSION
//--------------------

// PackageVersion returns the version of the version package.
func PackageVersion() version.Version {
	return version.New(4, 1, 0)
}

// EOF
