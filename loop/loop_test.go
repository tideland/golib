// Tideland Go Library - Loop - Unit Test
//
// Copyright (C) 2013-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package loop_test

//--------------------
// IMPORTS
//--------------------

import (
	"errors"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/loop"
)

var (
	shortDelay    time.Duration = 20 * time.Millisecond
	longDelay     time.Duration = 50 * time.Millisecond
	veryLongDelay time.Duration = 200 * time.Millisecond
)

//--------------------
// TESTS
//--------------------

// Test the simple backend returning nil after stop.
func TestSimpleStop(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	done := false
	l := loop.Go(generateSimpleBackend(&done))

	assert.Nil(l.Stop(), "no error after simple stop")
	assert.True(done, "backend has done")

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// Test the simple backend returning an error after kill.
func TestSimpleKill(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	done := false
	l := loop.Go(generateSimpleBackend(&done))

	l.Kill(errors.New("ouch"))

	assert.ErrorMatch(l.Stop(), "ouch", "error has to be 'ouch'")
	assert.True(done, "backend has done")

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// Test an internal error.
func TestError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	done := false
	l := loop.Go(generateErrorBackend(&done))

	time.Sleep(longDelay)

	assert.ErrorMatch(l.Stop(), "timed out", "error has to be 'time out'")
	assert.True(done, "backend has done")

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// Test an error in a deferred function inside the loop.
func TestDeferredError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	done := false
	l := loop.Go(generateDeferredErrorBackend(&done))

	assert.ErrorMatch(l.Stop(), "deferred error", "error has to be 'deferred error'")
	assert.True(done, "backend has done")

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// Test recoverings after panics.
func TestRecoverings(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	done := false
	count := 0
	l := loop.GoRecoverable(generateSimplePanicBackend(&done, &count), checkRecovering)

	time.Sleep(veryLongDelay)

	assert.ErrorMatch(l.Stop(), "too many panics", "error has to be 'too many panics'")
	assert.True(done, "backend has done")
	assert.Equal(count, 5, "loop has to be restarted 5 times")

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// Test regular stop of a recovered loop.
func TestStopRecoverings(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	done := false
	count := 0
	l := loop.GoRecoverable(generateSimplePanicBackend(&done, &count), ignorePanics)

	time.Sleep(longDelay)

	assert.Nil(l.Stop(), "no error after simple stop")
	assert.True(done, "backend has done")

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// Test error inside a recovered loop.
func TestRecoveringsError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	done := false
	count := 0
	l := loop.GoRecoverable(generateErrorPanicBackend(&done, &count), catchTimeout)

	time.Sleep(longDelay)

	assert.ErrorMatch(l.Stop(), "timed out", "error has to be 'timed out'")
	assert.True(done, "backend has done")

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

//--------------------
// EXAMPLES
//--------------------

func ExampleLoopFunc() {
	printChan := make(chan string)
	loopFunc := func(l loop.Loop) error {
		for {
			select {
			case <-l.ShallStop():
				return nil
			case str := <-printChan:
				println(str)
			}
		}
	}
	loop.Go(loopFunc)
}

func ExampleRecoverFunc() {
	printChan := make(chan string)
	loopFunc := func(l loop.Loop) error {
		for {
			select {
			case <-l.ShallStop():
				return nil
			case str := <-printChan:
				println(str)
			}
		}
	}
	recoverFunc := func(rs loop.Recoverings) (loop.Recoverings, error) {
		if len(rs) >= 5 {
			return nil, errors.New("too many panics")
		}
		return rs, nil
	}
	loop.GoRecoverable(loopFunc, recoverFunc)
}

//--------------------
// HELPERS
//--------------------

func generateSimpleBackend(done *bool) loop.LoopFunc {
	return func(l loop.Loop) error {
		defer func() { t := true; *done = t }()
		for {
			select {
			case <-l.ShallStop():
				return nil
			}
		}
	}
}

func generateErrorBackend(done *bool) loop.LoopFunc {
	return func(l loop.Loop) error {
		defer func() { t := true; *done = t }()
		for {
			select {
			case <-l.ShallStop():
				return nil
			case <-time.After(shortDelay):
				return errors.New("timed out")
			}
		}
	}
}

func generateDeferredErrorBackend(done *bool) loop.LoopFunc {
	return func(l loop.Loop) (err error) {
		defer func() { t := true; *done = t }()
		defer func() {
			err = errors.New("deferred error")
		}()
		for {
			select {
			case <-l.ShallStop():
				return nil
			}
		}
	}
}

func generateSimplePanicBackend(done *bool, count *int) loop.LoopFunc {
	return func(l loop.Loop) error {
		defer func() { t := true; *done = t }()
		c := *count
		*count = c + 1
		for {
			select {
			case <-l.ShallStop():
				return nil
			case <-time.After(shortDelay):
				panic("ouch")
			}
		}
	}
}

func generateErrorPanicBackend(done *bool, count *int) loop.LoopFunc {
	return func(l loop.Loop) error {
		defer func() { t := true; *done = t }()
		for {
			select {
			case <-l.ShallStop():
				return nil
			case <-time.After(shortDelay):
				return errors.New("timed out")
			}
		}
	}
}

func checkRecovering(rs loop.Recoverings) (loop.Recoverings, error) {
	if len(rs) >= 5 {
		return nil, errors.New("too many panics")
	}
	return rs, nil
}

func catchTimeout(rs loop.Recoverings) (loop.Recoverings, error) {
	if len(rs) > 0 {
		if err, ok := rs.Last().Reason.(error); ok {
			if err.Error() == "timed out" {
				return nil, err
			}
		}
	}
	return nil, nil
}

func ignorePanics(rs loop.Recoverings) (loop.Recoverings, error) {
	return nil, nil
}

// EOF
