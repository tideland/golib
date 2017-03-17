// Tideland Go Library - Loop
//
// Copyright (C) 2013-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library loop package is intended to support
// the developer implementing the typical Go idiom for
// concurrent applications running in a loop in the background
// and doing a select on one or more channels. Stopping those
// loops or getting aware of internal errors requires extra
// efforts. The loop package helps to control this kind of
// goroutines.
//
// Beside the simple controlled loops the also can be made
// recoverable. In this case a user defined recovery function
// gets notified if a loop ends with an error or panics.
// The paseed passed list of recovering information helps
// to check the reason and frequency.
//
// A third way are sentinels. Those can monitor multiple
// loops and other sentinels. So hierarchies can be defined.
// In case of no handler function an error of one monitored
// instance will lead to a stop of all monitored instances.
// Otherwise the user can check the error reason inside
// the handler function and optionally restart the loop
// or sentinel.
//
// See the example functions for more information.
package loop

// EOF
