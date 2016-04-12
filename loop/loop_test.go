// Tideland Go Library - Loop - Unit Test
//
// Copyright (C) 2013-2016 Frank Mueller / Tideland / Oldenburg / Germany
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
	shortTimeout    time.Duration = 25 * time.Millisecond
	longTimeout     time.Duration = 100 * time.Millisecond
	longerTimeout   time.Duration = 150 * time.Millisecond
	veryLongTimeout time.Duration = 5 * time.Second
)

//--------------------
// TESTS
//--------------------

// TestSimpleStop tests the simple backend returning nil
// after a stop.
func TestSimpleStop(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	donec := audit.MakeSigChan()
	l := loop.Go(makeSimpleLF(donec), "simple-stop")

	assert.Nil(l.Stop(), "no error after simple stop")
	assert.Wait(donec, true, shortTimeout)

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// TestSimpleRestart tests restarting when not stopped and
// when stopped.
func TestSimpleRestart(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	donec := audit.MakeSigChan()
	l := loop.Go(makeSimpleLF(donec), "simple-restart")

	assert.ErrorMatch(l.Restart(), ".*cannot restart unstopped.*")

	status, _ := l.Error()

	assert.Equal(loop.Running, status)

	assert.Nil(l.Stop())
	assert.Wait(donec, true, shortTimeout)

	status, _ = l.Error()

	assert.Equal(loop.Stopped, status)

	assert.Nil(l.Restart())

	status, _ = l.Error()

	assert.Equal(loop.Running, status)
	assert.Nil(l.Stop())
	assert.Wait(donec, true, shortTimeout)
}

// TestSimpleKill tests the simple backend returning an error
// after a kill.
func TestSimpleKill(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	donec := audit.MakeSigChan()
	l := loop.Go(makeSimpleLF(donec), "simple-kill")

	l.Kill(errors.New("ouch"))

	assert.ErrorMatch(l.Stop(), "ouch", "error has to be 'ouch'")
	assert.Wait(donec, true, shortTimeout)

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// TestError tests an internal error.
func TestError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	donec := audit.MakeSigChan()
	l := loop.Go(makeErrorLF(donec), "error")

	time.Sleep(longTimeout)

	assert.ErrorMatch(l.Stop(), "timed out", "error has to be 'time out'")
	assert.Wait(donec, true, shortTimeout)

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// TestDeferredError tests an error in a deferred function inside the loop.
func TestDeferredError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	donec := audit.MakeSigChan()
	l := loop.Go(makeDeferredErrorLF(donec), "deferred-error")

	assert.ErrorMatch(l.Stop(), "deferred error", "error has to be 'deferred error'")
	assert.Wait(donec, true, shortTimeout)

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// TestStopRecoverings tests the regular stop of a recovered loop.
func TestStopRecoverings(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	donec := audit.MakeSigChan()
	l := loop.GoRecoverable(makeRecoverPanicLF(), makeIgnorePanicsRF(donec), "stop-recoverings")

	time.Sleep(longTimeout)

	assert.Nil(l.Stop(), "no error after simple stop")
	assert.Wait(donec, "recovered", longTimeout)

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// TestEndRecoverings tests the regular internal stop of a recovered loop.
func TestEndRecoverings(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	donec := audit.MakeSigChan()
	l := loop.GoRecoverable(makeRecoverNoErrorLF(donec), makeIgnorePanicsRF(nil), "end-recoverings")

	time.Sleep(longTimeout)

	assert.Wait(donec, true, longTimeout)

	status, _ := l.Error()
	assert.Equal(loop.Stopped, status)
}

// TestRecoveringsPanic test recoverings after panics.
func TestRecoveringsPanic(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	donec := audit.MakeSigChan()
	l := loop.GoRecoverable(makeRecoverPanicLF(), makeCheckCountRF(donec), "recoverings-panic")

	time.Sleep(longerTimeout)

	assert.ErrorMatch(l.Stop(), "too many panics")
	assert.Wait(donec, true, longTimeout)

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status)
}

// TestRecoveringsError tests recoverings after errors
func TestRecoveringsError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	donec := audit.MakeSigChan()
	l := loop.GoRecoverable(makeRecoverErrorLF(), makeCatchTimeoutRF(donec), "recoverings-error")

	time.Sleep(longerTimeout)

	assert.ErrorMatch(l.Stop(), "timed out", "error has to be 'timed out'")
	assert.Wait(donec, "timed out", longTimeout)

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// TestDescription tests the handling of loop and
// sentinel descriptions.
func TestDescription(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	doneAC := audit.MakeSigChan()
	doneBC := audit.MakeSigChan()

	s := loop.GoSentinel(nil, "one")
	lA := loop.Go(makeSimpleLF(doneAC), "two", "three", "four")
	lB := loop.Go(makeSimpleLF(doneBC))

	s.Observe(lA, lB)

	assert.Equal(s.Description(), "one")
	assert.Equal(lA.Description(), "two:three:four")
	assert.Match(lB.Description(), "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")

	assert.Nil(s.Stop())
}

// TestSimpleSentinel tests the simple starting and
// stopping of a sentinel.
func TestSimpleSentinel(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	doneAC := audit.MakeSigChan()
	doneBC := audit.MakeSigChan()
	doneCC := audit.MakeSigChan()

	s := loop.GoSentinel(nil, "simple-sentinel")
	lA := loop.Go(makeSimpleLF(doneAC), "loop", "a")
	lB := loop.Go(makeSimpleLF(doneBC), "loop", "b")
	lC := loop.Go(makeSimpleLF(doneCC), "loop", "c")

	s.Observe(lA, lB, lC)

	assert.Nil(s.Stop())
	assert.Wait(doneAC, true, shortTimeout)
	assert.Wait(doneBC, true, shortTimeout)
	assert.Wait(doneCC, true, shortTimeout)

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

// TestSentinelStoppingLoop tests the stopping
// of a loop before sentinel stops.
func TestSentinelStoppingLoop(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	doneAC := audit.MakeSigChan()
	doneBC := audit.MakeSigChan()
	doneCC := audit.MakeSigChan()

	s := loop.GoSentinel(nil, "sentinel-stopping-loop")
	lA := loop.Go(makeSimpleLF(doneAC), "loop", "a")
	lB := loop.Go(makeSimpleLF(doneBC), "loop", "b")
	lC := loop.Go(makeSimpleLF(doneCC), "loop", "c")

	s.Observe(lA, lB, lC)

	assert.Nil(lB.Stop())
	assert.Wait(doneBC, true, shortTimeout)

	assert.Nil(s.Stop())
	assert.Wait(doneAC, true, shortTimeout)
	assert.Wait(doneCC, true, shortTimeout)

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

// TestSentinelKillingLoopNoHandler tests the killing
// of a loop before sentinel stops. The sentinel has
// no handler and so ignores the error.
func TestSentinelKillingLoopNoHandler(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	doneAC := audit.MakeSigChan()
	doneBC := audit.MakeSigChan()
	doneCC := audit.MakeSigChan()

	s := loop.GoSentinel(nil, "sentinel-killing-loop-no-handler")
	lA := loop.Go(makeSimpleLF(doneAC), "loop", "a")
	lB := loop.Go(makeSimpleLF(doneBC), "loop", "b")
	lC := loop.Go(makeSimpleLF(doneCC), "loop", "c")

	s.Observe(lA, lB, lC)

	lB.Kill(errors.New("bang!"))
	assert.Wait(doneBC, true, shortTimeout)

	assert.ErrorMatch(s.Stop(), ".*bang!.*")
	assert.Wait(doneAC, true, shortTimeout)
	assert.Wait(doneCC, true, shortTimeout)

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

// TestSentinelKillingLoopHandlerRestarts tests the killing
// of a loop before sentinel stops. The sentinel has
// a handler and restarts the loop.
func TestSentinelKillingLoopHandlerRestarts(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	doneAC := audit.MakeSigChan()
	doneBC := audit.MakeSigChan()
	doneCC := audit.MakeSigChan()
	handlerC := audit.MakeSigChan()
	handlerF := func(s loop.Sentinel, o loop.Observable) error {
		o.Restart()
		handlerC <- true
		return nil
	}

	s := loop.GoSentinel(handlerF, "sentinel-killing-loop-handler-restarts")
	lA := loop.Go(makeSimpleLF(doneAC), "loop", "a")
	lB := loop.Go(makeSimpleLF(doneBC), "loop", "b")
	lC := loop.Go(makeSimpleLF(doneCC), "loop", "c")

	s.Observe(lA, lB, lC)

	lB.Kill(errors.New("bang!"))
	assert.Wait(handlerC, true, shortTimeout)

	assert.Nil(s.Stop())
	assert.Wait(doneBC, true, shortTimeout)
	assert.Wait(doneAC, true, shortTimeout)
	assert.Wait(doneCC, true, shortTimeout)

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

// TestSentinelKillingLoopHandlerStops tests the killing
// of a loop before sentinel stops. The sentinel has
// a handler which stops the processing.
func TestSentinelKillingLoopHandlerStops(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	infoAC := audit.MakeSigChan()
	infoBC := audit.MakeSigChan()
	infoCC := audit.MakeSigChan()
	handlerC := audit.MakeSigChan()
	handlerF := func(s loop.Sentinel, o loop.Observable) error {
		handlerC <- true
		return errors.New("oh no!")
	}

	s := loop.GoSentinel(handlerF, "sentinel-killing-loop-with-stops")
	lA := loop.Go(makeStartStopLF(infoAC), "loop", "a")
	lB := loop.Go(makeStartStopLF(infoBC), "loop", "b")
	lC := loop.Go(makeStartStopLF(infoCC), "loop", "c")

	s.Observe(lA, lB, lC)

	assert.Wait(infoAC, "started", shortTimeout)
	assert.Wait(infoBC, "started", shortTimeout)
	assert.Wait(infoCC, "started", shortTimeout)

	time.Sleep(longTimeout)

	lB.Kill(errors.New("bang!"))

	assert.Wait(handlerC, true, shortTimeout)
	assert.Wait(infoBC, "stopped", shortTimeout)
	assert.Wait(infoAC, "stopped", shortTimeout)
	assert.Wait(infoCC, "stopped", shortTimeout)

	assert.ErrorMatch(s.Stop(), ".*oh no!.*")

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

//--------------------
// EXAMPLES
//--------------------

func ExampleLoopFunc() {
	printc := make(chan string)
	loopf := func(l loop.Loop) error {
		for {
			select {
			case <-l.ShallStop():
				return nil
			case str := <-printc:
				if str == "panic" {
					return errors.New("panic")
				}
				println(str)
			}
		}
	}
	l := loop.Go(loopf)

	printc <- "Hello"
	printc <- "World"

	if err := l.Stop(); err != nil {
		panic(err)
	}
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

func makeSimpleLF(donec chan interface{}) loop.LoopFunc {
	return func(l loop.Loop) error {
		defer func() { donec <- true }()
		for {
			select {
			case <-l.ShallStop():
				return nil
			}
		}
	}
}

func makeStartStopLF(infoc chan interface{}) loop.LoopFunc {
	return func(l loop.Loop) error {
		defer func() { infoc <- "stopped" }()
		infoc <- "started"
		for {
			select {
			case <-l.ShallStop():
				return nil
			}
		}
	}
}

func makeErrorLF(donec chan interface{}) loop.LoopFunc {
	return func(l loop.Loop) error {
		defer func() { donec <- true }()
		for {
			select {
			case <-l.ShallStop():
				return nil
			case <-time.After(shortTimeout):
				return errors.New("timed out")
			}
		}
	}
}

func makeDeferredErrorLF(donec chan interface{}) loop.LoopFunc {
	return func(l loop.Loop) (err error) {
		defer func() { donec <- true }()
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

func makeRecoverPanicLF() loop.LoopFunc {
	return func(l loop.Loop) error {
		for {
			select {
			case <-l.ShallStop():
				return nil
			case <-time.After(shortTimeout):
				panic("ouch")
			}
		}
	}
}

func makeRecoverErrorLF() loop.LoopFunc {
	return func(l loop.Loop) error {
		for {
			select {
			case <-l.ShallStop():
				return nil
			case <-time.After(shortTimeout):
				return errors.New("timed out")
			}
		}
	}
}

func makeRecoverNoErrorLF(donec chan interface{}) loop.LoopFunc {
	return func(l loop.Loop) error {
		time.Sleep(shortTimeout)
		donec <- true
		return nil
	}
}

func makeCheckCountRF(donec chan interface{}) loop.RecoverFunc {
	return func(rs loop.Recoverings) (loop.Recoverings, error) {
		if len(rs) >= 5 {
			donec <- len(rs)
			return nil, errors.New("too many panics")
		}
		donec <- true
		return rs, nil
	}
}

func makeCatchTimeoutRF(donec chan interface{}) loop.RecoverFunc {
	return func(rs loop.Recoverings) (loop.Recoverings, error) {
		if len(rs) > 0 {
			if err, ok := rs.Last().Reason.(error); ok {
				if err.Error() == "timed out" {
					donec <- "timed out"
					return nil, err
				}
			}
		}
		return nil, nil
	}
}

func makeIgnorePanicsRF(donec chan interface{}) loop.RecoverFunc {
	return func(rs loop.Recoverings) (loop.Recoverings, error) {
		if donec != nil {
			donec <- "recovered"
		}
		return nil, nil
	}
}

// EOF
