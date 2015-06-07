// Tideland Go Library - Redis Client
//
// Copyright (C) 2009-2015 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library redis package provides a very powerful as well as
// convenient client for the Redis database.
//
// After opening the database with Open() a pooled connection can be
// retrieved using db.Connection(). It has be returnded to the pool with
// with conn.Return(), optimally done using a defer after retrieving. The
// connection provides a conn.Do() method to execute any command. It returns
// a result set with helpers to access the returned values and convert
// them into Go types. For typical returnings there are conn.DoXxx() methods.
//
// All conn.Do() methods work atomically and are able to run all commands
// except subscriptions. Also the execution of scripts is possible that
// way. Additionally the execution of commands can be pipelined. The
// pipeline can be retrieved db.Pipeline(). It provides a ppl.Do()
// method for the execution of individual commands. Their results can
// be collected with ppl.Collect(), which returns a sice of result sets
// containing the responses of the commands.
//
// Due to the nature of the subscription the client provides an own
// type which can be retrieved with db.Subscription(). Here channels,
// in the sense of the Redis Pub/Sub, can be subscribed or unsubscribed.
// Published values can be retrieved with sub.Pop(). If the subscription
// is not needed anymore it can be closed using sub.Close().
package redis

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
	return version.New(4, 0, 0)
}

// EOF
