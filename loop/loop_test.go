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
	shortTimeout  time.Duration = 25 * time.Millisecond
	longTimeout   time.Duration = 100 * time.Millisecond
	longerTimeout time.Duration = 150 * time.Millisecond
	stayCalm      time.Duration = 5 * time.Second
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

	assert.ErrorMatch(l.Stop(), "timed out", "error has to be 'timed out'")
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

	assert.Equal(s.String(), "one")
	assert.Equal(lA.String(), "two::three::four")
	assert.Match(lB.String(), "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")

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
	doneC := audit.MakeSigChan()

	s := loop.GoSentinel(nil, "sentinel-stopping-loop")
	lA := loop.Go(makeSimpleLF(doneC), "loop", "a")
	lB := loop.Go(makeSimpleLF(doneC), "loop", "b")
	lC := loop.Go(makeSimpleLF(doneC), "loop", "c")

	s.Observe(lA, lB, lC)

	assert.Nil(lB.Stop())
	assert.Wait(doneC, true, shortTimeout)

	assert.Nil(s.Stop())
	assert.Wait(doneC, true, shortTimeout)
	assert.Wait(doneC, true, shortTimeout)

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

// TestSentinelForget tests the forgetting of loops
// by a sentinel.
func TestSentineForget(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	doneC := audit.MakeSigChan()

	s := loop.GoSentinel(nil, "sentinel-forget")
	lA := loop.Go(makeSimpleLF(doneC), "loop", "a")
	lB := loop.Go(makeSimpleLF(doneC), "loop", "b")
	lC := loop.Go(makeSimpleLF(doneC), "loop", "c")
	lD := loop.Go(makeSimpleLF(doneC), "loop", "d")

	s.Observe(lA, lB, lC, lD)
	time.Sleep(longTimeout)
	s.Forget(lB, lC)

	assert.Nil(s.Stop())
	assert.Wait(doneC, true, shortTimeout)
	assert.Wait(doneC, true, shortTimeout)

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

// TestSentinelKillingLoopNoHandler tests the killing
// of a loop before sentinel stops. The sentinel has
// no handler and so ignores the error.
func TestSentinelKillingLoopNoHandler(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	doneC := audit.MakeSigChan()

	s := loop.GoSentinel(nil, "sentinel-killing-loop-no-handler")
	lA := loop.Go(makeSimpleLF(doneC), "loop", "a")
	lB := loop.Go(makeSimpleLF(doneC), "loop", "b")
	lC := loop.Go(makeSimpleLF(doneC), "loop", "c")

	s.Observe(lA, lB, lC)

	lB.Kill(errors.New("bang!"))
	assert.Wait(doneC, true, stayCalm)
	assert.Wait(doneC, true, stayCalm)
	assert.Wait(doneC, true, stayCalm)

	assert.ErrorMatch(s.Stop(), ".*bang!.*")

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

	time.Sleep(shortTimeout)
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

	assert.Wait(infoAC, "started", stayCalm)
	assert.Wait(infoBC, "started", stayCalm)
	assert.Wait(infoCC, "started", stayCalm)

	time.Sleep(shortTimeout)
	lB.Kill(errors.New("bang!"))

	assert.Wait(infoBC, "stopped", stayCalm)
	assert.Wait(infoAC, "stopped", stayCalm)
	assert.Wait(infoCC, "stopped", stayCalm)
	assert.Wait(handlerC, true, stayCalm)

	assert.ErrorMatch(s.Stop(), ".*oh no!.*")

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

// TestSentinelKillingLoopHandlerRestartAll tests the killing
// of a loop before sentinel stops. The sentinel has
// a handler which restarts all observables.
func TestSentinelKillingLoopHandlerRestartAll(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	infoAC := audit.MakeSigChan()
	infoBC := audit.MakeSigChan()
	infoCC := audit.MakeSigChan()
	handlerC := audit.MakeSigChan()
	handlerF := func(s loop.Sentinel, _ loop.Observable) error {
		s.ObservablesDo(func(o loop.Observable) error {
			o.Restart()
			return nil
		})
		handlerC <- true
		return nil
	}

	s := loop.GoSentinel(handlerF, "sentinel-killing-loop-restarting-all")
	lA := loop.Go(makeStartStopLF(infoAC), "loop", "a")
	lB := loop.Go(makeStartStopLF(infoBC), "loop", "b")
	lC := loop.Go(makeStartStopLF(infoCC), "loop", "c")

	s.Observe(lA, lB, lC)

	assert.Wait(infoAC, "started", stayCalm)
	assert.Wait(infoBC, "started", stayCalm)
	assert.Wait(infoCC, "started", stayCalm)

	time.Sleep(shortTimeout)
	lB.Kill(errors.New("bang!"))

	assert.Wait(handlerC, true, stayCalm)

	assert.Nil(s.Stop())
	assert.Wait(infoBC, "stopped", stayCalm)
	assert.Wait(infoAC, "stopped", stayCalm)
	assert.Wait(infoCC, "stopped", stayCalm)

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

// TestNestedSentinelKill tests the killing and restarting of a
// nested sentinel.
func TestNestedSentinelKill(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	infoAC := audit.MakeSigChan()
	infoBC := audit.MakeSigChan()
	handlerC := audit.MakeSigChan()
	handlerF := func(s loop.Sentinel, o loop.Observable) error {
		o.Restart()
		handlerC <- true
		return nil
	}

	sT := loop.GoSentinel(handlerF, "nested-sentinel-kill", "top")
	lA := loop.Go(makeStartStopLF(infoAC), "loop", "a")
	sN := loop.GoSentinel(handlerF, "nested-sentinel-kill", "nested")
	lB := loop.Go(makeStartStopLF(infoBC), "loop", "b")

	sT.Observe(lA, sN)
	sN.Observe(lB)

	assert.Wait(infoAC, "started", stayCalm)
	assert.Wait(infoBC, "started", stayCalm)

	time.Sleep(shortTimeout)
	sN.Kill(errors.New("bang!"))
	time.Sleep(longerTimeout)

	assert.Nil(sT.Stop())
	assert.Wait(infoBC, "stopped", stayCalm)
	assert.Wait(infoAC, "stopped", stayCalm)

	status, _ := sT.Error()

	assert.Equal(loop.Stopped, status)
}

// TestSentinelSwitch tests if the change of the assignment
// of a sentinel is handled correctly.
func TestSentinelSwitch(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	infoC := audit.MakeSigChan()

	sA := loop.GoSentinel(nil, "sentinel-switch", "a")
	sB := loop.GoSentinel(nil, "sentinel-switch", "b")

	lA := loop.Go(makeStartStopLF(infoC), "loop", "a")

	sA.Observe(lA)

	assert.Wait(infoC, "started", stayCalm)

	sB.Observe(lA)

	assert.Nil(sA.Stop())
	time.Sleep(longTimeout)
	assert.Length(infoC, 0)

	assert.Nil(sB.Stop())
	assert.Wait(infoC, "stopped", stayCalm)
}

//--------------------
// EXAMPLES
//--------------------

// ExampleLoopFunc shows the usage of loop.Go with one
// loop function and no recovery. The inner loop contains
// a select listening to the channel returned by ShallStop.
// Other channels are for the standard communication
// with the loop.
func ExampleLoopFunc() {
	printC := make(chan string)
	// Sample loop function.
	loopF := func(l loop.Loop) error {
		for {
			select {
			case <-l.ShallStop():
				// We shall stop.
				return nil
			case str := <-printC:
				if str == "panic" {
					return errors.New("panic")
				}
				println(str)
			}
		}
	}
	l := loop.Go(loopF, "simple loop demo")

	printC <- "Hello"
	printC <- "World"

	if err := l.Stop(); err != nil {
		panic(err)
	}
}

// ExampleRecoverFunc demonstrates the usage of a recovery
// function when using loop.GoRecoverable. Here the frequency
// of the recoverings (more than five in 10 milliseconds)
// or the total number is checked. If the total number is
// not interesting the recoverings could be trimmed by
// e.g. rs.Trim(5). The fields Time and Reason per
// recovering allow even more diagnosis.
func ExampleRecoverFunc() {
	printC := make(chan string)
	loopF := func(l loop.Loop) error {
		for {
			select {
			case <-l.ShallStop():
				return nil
			case str := <-printC:
				println(str)
			}
		}
	}
	// Recovery function checking frequency and total number.
	recoverF := func(rs loop.Recoverings) (loop.Recoverings, error) {
		if rs.Frequency(5, 10 * time.Millisecond) {
			return nil, errors.New("too high error frequency")
		}
		if rs.Len() >= 10 {
			return nil, errors.New("too many errors")
		}
		return rs, nil
	}
	loop.GoRecoverable(loopF, recoverF, "recoverable loop demo")
}

// ExampleSentinel demonstrates the monitoring of loops and sentinel
// with a handler function trying to restart the faulty observable.
// The nested sentinel has no handler function. An error of a monitored
// observable would lead to the stop of all observables.
func ExampleSentinel() {
	loopF := func(l loop.Loop) error {
		for {
			select {
			case <-l.ShallStop():
				return nil
			}
		}
	}
	handleF := func(s loop.Sentinel, o loop.Observable) error {
		return o.Restart()
	}
	loopA := loop.Go(loopF, "loop", "a")
	loopB := loop.Go(loopF, "loop", "b")
	loopC := loop.Go(loopF, "loop", "c")
	loopD := loop.Go(loopF, "loop", "d")
	sentinel := loop.GoSentinel(handleF, "sentinel demo")

	sentinel.Observe(loopA, loopB)

	// Hierarchies are possible.
	observedSentinel := loop.GoSentinel(nil, "nested sentinel w/o handler")

	sentinel.Observe(observedSentinel)
	observedSentinel.Observe(loopC)
	observedSentinel.Observe(loopD)
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
