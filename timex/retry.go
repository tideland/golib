// Tideland Go Library - Time Extensions
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package timex

//--------------------
// IMPORTS
//--------------------

import (
	"time"

	"github.com/tideland/golib/errors"
)

//--------------------
// RETRY
//--------------------

// RetryStrategy describes how often the function in Retry is executed, the
// initial break between those retries, how much this time is incremented
// for each retry, and the maximum timeout.
type RetryStrategy struct {
	Count          int
	Break          time.Duration
	BreakIncrement time.Duration
	Timeout        time.Duration
}

// ShortAttempt returns a predefined short retry strategy.
func ShortAttempt() RetryStrategy {
	return RetryStrategy{
		Count:          10,
		Break:          50 * time.Millisecond,
		BreakIncrement: 0,
		Timeout:        5 * time.Second,
	}
}

// MediumAttempt returns a predefined medium retry strategy.
func MediumAttempt() RetryStrategy {
	return RetryStrategy{
		Count:          50,
		Break:          10 * time.Millisecond,
		BreakIncrement: 10 * time.Millisecond,
		Timeout:        30 * time.Second,
	}
}

// LongAttempt returns a predefined long retry strategy.
func LongAttempt() RetryStrategy {
	return RetryStrategy{
		Count:          100,
		Break:          10 * time.Millisecond,
		BreakIncrement: 25 * time.Millisecond,
		Timeout:        5 * time.Minute,
	}
}

// Retry executes the passed function until it returns true or an error.
// These retries are restricted by the retry strategy.
func Retry(f func() (bool, error), rs RetryStrategy) error {
	timeout := time.Now().Add(rs.Timeout)
	sleep := rs.Break
	for i := 0; i < rs.Count; i++ {
		done, err := f()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		if time.Now().After(timeout) {
			return errors.New(ErrRetriedTooLong, errorMessages, rs.Timeout)
		}
		time.Sleep(sleep)
		sleep += rs.BreakIncrement
	}
	return errors.New(ErrRetriedTooOften, errorMessages, rs.Count)
}

// EOF
