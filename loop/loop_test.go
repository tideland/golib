// Tideland Go Library - Loop - Unit Test
//
// Copyright (C) 2013-2017 Frank Mueller / Tideland / Oldenburg / Germany
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

//--------------------
// CONSTANTS
//--------------------

// timeout is the waitng time for events from inside of loops.
var timeout time.Duration = 5 * time.Second

//--------------------
// TESTS
//--------------------

// TestSimpleStop tests the simple backend returning nil
// after a stop.
func TestSimpleStop(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	bundle := newChannelBundle()
	l := loop.Go(makeSimpleLF(bundle), "simple-stop")

	assert.Wait(bundle.startedc, true, timeout)
	assert.Nil(l.Stop(), "no error after simple stop")
	assert.Wait(bundle.donec, true, timeout)

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// TestSimpleRestart tests restarting when not stopped and
// when stopped.
func TestSimpleRestart(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	bundle := newChannelBundle()
	l := loop.Go(makeSimpleLF(bundle), "simple-restart")

	assert.Wait(bundle.startedc, true, timeout)
	assert.ErrorMatch(l.Restart(), ".*cannot restart unstopped.*")

	status, _ := l.Error()

	assert.Equal(loop.Running, status)
	assert.Nil(l.Stop())
	assert.Wait(bundle.donec, true, timeout)

	status, _ = l.Error()

	assert.Equal(loop.Stopped, status)
	assert.Nil(l.Restart())
	assert.Wait(bundle.startedc, true, timeout)

	status, _ = l.Error()

	assert.Equal(loop.Running, status)
	assert.Nil(l.Stop())
	assert.Wait(bundle.donec, true, timeout)
}

// TestSimpleKill tests the simple backend returning an error
// after a kill.
func TestSimpleKill(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	bundle := newChannelBundle()
	l := loop.Go(makeSimpleLF(bundle), "simple-kill")

	assert.Wait(bundle.startedc, true, timeout)

	l.Kill(errors.New("ouch"))

	assert.ErrorMatch(l.Stop(), "ouch")
	assert.Wait(bundle.donec, true, timeout)

	status, _ := l.Error()

	assert.Equal(status, loop.Stopped)
}

// TestError tests an internal error.
func TestError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	bundle := newChannelBundle()
	l := loop.Go(makeErrorLF(bundle), "error")

	assert.Wait(bundle.startedc, true, timeout)

	bundle.errorc <- true

	assert.ErrorMatch(l.Stop(), "internal loop error")
	assert.Wait(bundle.donec, true, timeout)

	status, _ := l.Error()

	assert.Equal(status, loop.Stopped)
}

// TestDeferredError tests an error in a deferred function inside the loop.
func TestDeferredError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	bundle := newChannelBundle()
	l := loop.Go(makeDeferredErrorLF(bundle), "deferred-error")

	assert.Wait(bundle.startedc, true, timeout)
	assert.ErrorMatch(l.Stop(), "deferred error")
	assert.Wait(bundle.donec, true, timeout)

	status, _ := l.Error()

	assert.Equal(status, loop.Stopped)
}

// TestStopRecoverings tests the regular stop of a recovered loop.
func TestStopRecoverings(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	lbundle := newChannelBundle()
	rbundle := newChannelBundle()
	l := loop.GoRecoverable(makeRecoverPanicLF(lbundle), makeIgnorePanicsRF(rbundle), "stop-recoverings")

	assert.Wait(lbundle.startedc, true, timeout)

	lbundle.errorc <- true

	assert.Nil(l.Stop())
	assert.Wait(rbundle.donec, "recovered", timeout)

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// TestEndRecoverings tests the regular internal stop of a recovered loop.
func TestEndRecoverings(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	lbundle := newChannelBundle()
	rbundle := newChannelBundle()
	l := loop.GoRecoverable(makeRecoverNoErrorLF(lbundle), makeIgnorePanicsRF(rbundle), "end-recoverings")

	assert.Wait(lbundle.startedc, true, timeout)

	lbundle.stopc <- true

	assert.Wait(lbundle.donec, true, timeout)

	status, _ := l.Error()
	assert.Equal(status, loop.Stopped)
}

// TestRecoveringsPanic test recoverings after panics.
func TestRecoveringsPanic(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	lbundle := newChannelBundle()
	rbundle := newChannelBundle()
	l := loop.GoRecoverable(makeRecoverPanicLF(lbundle), makeCheckCountRF(rbundle), "recoverings-panic")

	go func() {
		for i := 0; i < 10; i++ {
			<-lbundle.startedc

			lbundle.errorc <- true

			<-rbundle.startedc
		}
	}()

	assert.Wait(rbundle.donec, 5, timeout)
	assert.ErrorMatch(l.Stop(), "too many panics")

	status, _ := l.Error()

	assert.Equal(status, loop.Stopped)
}

// TestRecoveringsError tests recoverings after errors.
func TestRecoveringsError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	lbundle := newChannelBundle()
	rbundle := newChannelBundle()
	l := loop.GoRecoverable(makeRecoverErrorLF(lbundle), makeCatchErrorRF(rbundle), "recoverings-error")

	assert.Wait(lbundle.startedc, true, timeout)

	lbundle.errorc <- true

	assert.ErrorMatch(l.Stop(), "error")
	assert.Wait(rbundle.donec, "error", timeout)

	status, _ := l.Error()

	assert.Equal(loop.Stopped, status, "loop is stopped")
}

// TestDescription tests the handling of loop and
// sentinel descriptions.
func TestDescription(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	abundle := newChannelBundle()
	bbundle := newChannelBundle()

	s := loop.GoSentinel("one")
	lA := loop.Go(makeSimpleLF(abundle), "two", "three", "four")
	lB := loop.Go(makeSimpleLF(bbundle))

	assert.Wait(abundle.startedc, true, timeout)
	assert.Wait(bbundle.startedc, true, timeout)

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
	abundle := newChannelBundle()
	bbundle := newChannelBundle()
	cbundle := newChannelBundle()

	s := loop.GoSentinel("simple-sentinel")
	lA := loop.Go(makeSimpleLF(abundle), "loop", "a")
	lB := loop.Go(makeSimpleLF(bbundle), "loop", "b")
	lC := loop.Go(makeSimpleLF(cbundle), "loop", "c")

	assert.Wait(abundle.startedc, true, timeout)
	assert.Wait(bbundle.startedc, true, timeout)
	assert.Wait(cbundle.startedc, true, timeout)

	s.Observe(lA, lB, lC)

	assert.Nil(s.Stop())
	assert.Wait(abundle.donec, true, timeout)
	assert.Wait(bbundle.donec, true, timeout)
	assert.Wait(cbundle.donec, true, timeout)

	status, _ := s.Error()

	assert.Equal(status, loop.Stopped)
}

// TestSentinelStoppingLoop tests the stopping
// of a loop before sentinel stops.
func TestSentinelStoppingLoop(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	abundle := newChannelBundle()
	bbundle := newChannelBundle()
	cbundle := newChannelBundle()

	s := loop.GoSentinel("sentinel-stopping-loop")
	lA := loop.Go(makeSimpleLF(abundle), "loop", "a")
	lB := loop.Go(makeSimpleLF(bbundle), "loop", "b")
	lC := loop.Go(makeSimpleLF(cbundle), "loop", "c")

	assert.Wait(abundle.startedc, true, timeout)
	assert.Wait(bbundle.startedc, true, timeout)
	assert.Wait(cbundle.startedc, true, timeout)

	s.Observe(lA, lB, lC)

	assert.Nil(lB.Stop())
	assert.Wait(bbundle.donec, true, timeout)

	assert.Nil(s.Stop())
	assert.Wait(abundle.donec, true, timeout)
	assert.Wait(cbundle.donec, true, timeout)

	status, _ := s.Error()

	assert.Equal(status, loop.Stopped)
}

// TestSentinelForget tests the forgetting of loops
// by a sentinel.
func TestSentineForget(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	abundle := newChannelBundle()
	bbundle := newChannelBundle()
	cbundle := newChannelBundle()
	dbundle := newChannelBundle()

	s := loop.GoSentinel("sentinel-forget")
	lA := loop.Go(makeSimpleLF(abundle), "loop", "a")
	lB := loop.Go(makeSimpleLF(bbundle), "loop", "b")
	lC := loop.Go(makeSimpleLF(cbundle), "loop", "c")
	lD := loop.Go(makeSimpleLF(dbundle), "loop", "d")

	assert.Wait(abundle.startedc, true, timeout)
	assert.Wait(bbundle.startedc, true, timeout)
	assert.Wait(cbundle.startedc, true, timeout)
	assert.Wait(dbundle.startedc, true, timeout)

	s.Observe(lA, lB, lC, lD)
	s.Forget(lB, lC)

	assert.Nil(s.Stop())
	assert.Wait(abundle.donec, true, timeout)
	assert.Wait(dbundle.donec, true, timeout)

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

// TestSentinelKillingLoopNoHandler tests the killing
// of a loop before sentinel stops. The sentinel has
// no handler and so ignores the error.
func TestSentinelKillingLoopNoHandler(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	abundle := newChannelBundle()
	bbundle := newChannelBundle()
	cbundle := newChannelBundle()

	s := loop.GoSentinel("sentinel-killing-loop-no-handler")
	lA := loop.Go(makeSimpleLF(abundle), "loop", "a")
	lB := loop.Go(makeSimpleLF(bbundle), "loop", "b")
	lC := loop.Go(makeSimpleLF(cbundle), "loop", "c")

	s.Observe(lA, lB, lC)

	assert.Wait(abundle.startedc, true, timeout)
	assert.Wait(bbundle.startedc, true, timeout)
	assert.Wait(cbundle.startedc, true, timeout)

	lB.Kill(errors.New("bang!"))
	assert.Wait(abundle.donec, true, timeout)
	assert.Wait(bbundle.donec, true, timeout)
	assert.Wait(cbundle.donec, true, timeout)

	assert.ErrorMatch(s.Stop(), ".*bang!.*")

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

// TestSentinelKillingLoopHandlerRestarts tests the killing
// of a loop before sentinel stops. The sentinel has
// a handler and restarts the loop.
func TestSentinelKillingLoopHandlerRestarts(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	abundle := newChannelBundle()
	bbundle := newChannelBundle()
	cbundle := newChannelBundle()
	sbundle := newChannelBundle()
	shandler := func(s loop.Sentinel, o loop.Observable, rs loop.Recoverings) (loop.Recoverings, error) {
		o.Restart()
		sbundle.donec <- true
		return nil, nil
	}

	s := loop.GoNotifiedSentinel(shandler, "sentinel-killing-loop-handler-restarts")
	lA := loop.Go(makeSimpleLF(abundle), "loop", "a")
	lB := loop.Go(makeSimpleLF(bbundle), "loop", "b")
	lC := loop.Go(makeSimpleLF(cbundle), "loop", "c")

	s.Observe(lA, lB, lC)

	assert.Wait(abundle.startedc, true, timeout)
	assert.Wait(bbundle.startedc, true, timeout)
	assert.Wait(cbundle.startedc, true, timeout)

	lB.Kill(errors.New("bang!"))

	assert.Wait(bbundle.donec, true, timeout)
	assert.Wait(sbundle.donec, true, timeout)
	assert.Nil(s.Stop())
	assert.Wait(abundle.donec, true, timeout)
	assert.Wait(bbundle.donec, true, timeout)
	assert.Wait(cbundle.donec, true, timeout)

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

// TestSentinelKillingLoopHandlerStops tests the killing
// of a loop before sentinel stops. The sentinel has
// a handler which stops the processing.
func TestSentinelKillingLoopHandlerStops(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	abundle := newChannelBundle()
	bbundle := newChannelBundle()
	cbundle := newChannelBundle()
	sbundle := newChannelBundle()
	shandler := func(s loop.Sentinel, o loop.Observable, rs loop.Recoverings) (loop.Recoverings, error) {
		sbundle.donec <- true
		return nil, errors.New("oh no!")
	}

	s := loop.GoNotifiedSentinel(shandler, "sentinel-killing-loop-with-stops")
	lA := loop.Go(makeSimpleLF(abundle), "loop", "a")
	lB := loop.Go(makeSimpleLF(bbundle), "loop", "b")
	lC := loop.Go(makeSimpleLF(cbundle), "loop", "c")

	s.Observe(lA, lB, lC)

	assert.Wait(abundle.startedc, true, timeout)
	assert.Wait(bbundle.startedc, true, timeout)
	assert.Wait(cbundle.startedc, true, timeout)

	lB.Kill(errors.New("bang!"))

	assert.Wait(abundle.donec, true, timeout)
	assert.Wait(bbundle.donec, true, timeout)
	assert.Wait(cbundle.donec, true, timeout)
	assert.Wait(sbundle.donec, true, timeout)

	assert.ErrorMatch(s.Stop(), ".*oh no!.*")

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

// TestSentinelKillingLoopHandlerRestartAll tests the killing
// of a loop before sentinel stops. The sentinel has
// a handler which restarts all observables.
func TestSentinelKillingLoopHandlerRestartAll(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	abundle := newChannelBundle()
	bbundle := newChannelBundle()
	cbundle := newChannelBundle()
	sbundle := newChannelBundle()
	shandler := func(s loop.Sentinel, _ loop.Observable, _ loop.Recoverings) (loop.Recoverings, error) {
		s.ObservablesDo(func(o loop.Observable) error {
			o.Restart()
			return nil
		})
		sbundle.donec <- true
		return nil, nil
	}

	s := loop.GoNotifiedSentinel(shandler, "sentinel-killing-loop-restarting-all")
	lA := loop.Go(makeSimpleLF(abundle), "loop", "a")
	lB := loop.Go(makeSimpleLF(bbundle), "loop", "b")
	lC := loop.Go(makeSimpleLF(cbundle), "loop", "c")

	s.Observe(lA, lB, lC)

	assert.Wait(abundle.startedc, true, timeout)
	assert.Wait(bbundle.startedc, true, timeout)
	assert.Wait(cbundle.startedc, true, timeout)

	lB.Kill(errors.New("bang!"))

	assert.Wait(bbundle.donec, true, timeout)
	assert.Wait(sbundle.donec, true, timeout)
	assert.Nil(s.Stop())
	assert.Wait(abundle.donec, true, timeout)
	assert.Wait(bbundle.donec, true, timeout)
	assert.Wait(cbundle.donec, true, timeout)

	status, _ := s.Error()

	assert.Equal(loop.Stopped, status)
}

// TestNestedSentinelKill tests the killing and restarting of a
// nested sentinel.
func TestNestedSentinelKill(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	abundle := newChannelBundle()
	bbundle := newChannelBundle()
	sbundle := newChannelBundle()
	shandler := func(s loop.Sentinel, o loop.Observable, rs loop.Recoverings) (loop.Recoverings, error) {
		o.Restart()
		sbundle.donec <- true
		return nil, nil
	}

	sT := loop.GoNotifiedSentinel(shandler, "nested-sentinel-kill", "top")
	lA := loop.Go(makeSimpleLF(abundle), "loop", "a")
	sN := loop.GoNotifiedSentinel(shandler, "nested-sentinel-kill", "nested")
	lB := loop.Go(makeSimpleLF(bbundle), "loop", "b")

	sT.Observe(lA, sN)
	sN.Observe(lB)

	assert.Wait(abundle.startedc, true, timeout)
	assert.Wait(bbundle.startedc, true, timeout)

	sN.Kill(errors.New("bang!"))

	assert.Wait(sbundle.donec, true, timeout)

	assert.Nil(sT.Stop())
	assert.Wait(bbundle.donec, true, timeout)
	assert.Wait(abundle.donec, true, timeout)

	status, _ := sT.Error()

	assert.Equal(loop.Stopped, status)
}

// TestSentinelSwitch tests if the change of the assignment
// of a sentinel is handled correctly.
func TestSentinelSwitch(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	abundle := newChannelBundle()

	sA := loop.GoSentinel("sentinel-switch", "a")
	sB := loop.GoSentinel("sentinel-switch", "b")
	lA := loop.Go(makeSimpleLF(abundle), "loop", "a")

	sA.Observe(lA)

	assert.Wait(abundle.startedc, true, timeout)

	sB.Observe(lA)

	assert.Nil(sA.Stop())
	assert.Nil(sB.Stop())
	assert.Wait(abundle.donec, true, timeout)
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
		if rs.Frequency(5, 10*time.Millisecond) {
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
	handleF := func(s loop.Sentinel, o loop.Observable, rs loop.Recoverings) (loop.Recoverings, error) {
		if rs.Frequency(5, 10*time.Millisecond) {
			return nil, errors.New("too high error frequency")
		}
		return nil, o.Restart()
	}
	loopA := loop.Go(loopF, "loop", "a")
	loopB := loop.Go(loopF, "loop", "b")
	loopC := loop.Go(loopF, "loop", "c")
	loopD := loop.Go(loopF, "loop", "d")
	sentinel := loop.GoNotifiedSentinel(handleF, "sentinel demo")

	sentinel.Observe(loopA, loopB)

	// Hierarchies are possible.
	observedSentinel := loop.GoSentinel("nested sentinel w/o handler")

	sentinel.Observe(observedSentinel)
	observedSentinel.Observe(loopC)
	observedSentinel.Observe(loopD)
}

//--------------------
// HELPERS
//--------------------

// channelBundle enables communication from and to loop functions.
type channelBundle struct {
	startedc chan interface{}
	donec    chan interface{}
	stopc    chan interface{}
	errorc   chan interface{}
}

func newChannelBundle() *channelBundle {
	return &channelBundle{
		startedc: audit.MakeSigChan(),
		donec:    audit.MakeSigChan(),
		stopc:    audit.MakeSigChan(),
		errorc:   audit.MakeSigChan(),
	}
}

// makeSimpleLF creates a loop function doing nothing special.
func makeSimpleLF(bundle *channelBundle) loop.LoopFunc {
	return func(l loop.Loop) error {
		defer func() { bundle.donec <- true }()
		bundle.startedc <- true
		for {
			select {
			case <-l.ShallStop():
				return nil
			}
		}
	}
}

// makeErrorLF creates a loop function stopping with
// an error after receiving a signal.
func makeErrorLF(bundle *channelBundle) loop.LoopFunc {
	return func(l loop.Loop) error {
		defer func() { bundle.donec <- true }()
		bundle.startedc <- true
		for {
			select {
			case <-l.ShallStop():
				return nil
			case <-bundle.errorc:
				return errors.New("internal loop error")
			}
		}
	}
}

// makeDeferredErrorLF creates a loop function returning
// an error in its deferred function.
func makeDeferredErrorLF(bundle *channelBundle) loop.LoopFunc {
	return func(l loop.Loop) (err error) {
		defer func() { bundle.donec <- true }()
		defer func() {
			err = errors.New("deferred error")
		}()
		bundle.startedc <- true
		for {
			select {
			case <-l.ShallStop():
				return nil
			}
		}
	}
}

// makeRecoverPanicLF creates a loop function having a panic
// but getting recovered.
func makeRecoverPanicLF(bundle *channelBundle) loop.LoopFunc {
	return func(l loop.Loop) error {
		bundle.startedc <- true
		for {
			select {
			case <-l.ShallStop():
				return nil
			case <-bundle.errorc:
				panic("panic")
			}
		}
	}
}

// makeRecoverErrorLF creates a loop function having an error
// but getting recovered.
func makeRecoverErrorLF(bundle *channelBundle) loop.LoopFunc {
	return func(l loop.Loop) error {
		bundle.startedc <- true
		for {
			select {
			case <-l.ShallStop():
				return nil
			case <-bundle.errorc:
				return errors.New("error")
			}
		}
	}
}

// makeRecoverNoErrorLF creates a loop function stopping
// without an error so it won't be recovered.
func makeRecoverNoErrorLF(bundle *channelBundle) loop.LoopFunc {
	return func(l loop.Loop) error {
		bundle.startedc <- true
		<-bundle.stopc
		bundle.donec <- true
		return nil
	}
}

// makeCheckCountRF creates a recover function stopping
// recovering after 5 calls.
func makeCheckCountRF(bundle *channelBundle) loop.RecoverFunc {
	return func(rs loop.Recoverings) (loop.Recoverings, error) {
		bundle.startedc <- true
		if len(rs) >= 5 {
			bundle.donec <- len(rs)
			return nil, errors.New("too many panics")
		}
		return rs, nil
	}
}

// makeCatchErrorRF creates a recover function stopping
// recovering in case of the error "error".
func makeCatchErrorRF(bundle *channelBundle) loop.RecoverFunc {
	return func(rs loop.Recoverings) (loop.Recoverings, error) {
		if len(rs) > 0 {
			if err, ok := rs.Last().Reason.(error); ok {
				if err.Error() == "error" {
					bundle.donec <- "error"
					return nil, err
				}
			}
		}
		return nil, nil
	}
}

// makeIgnorePanicsRF creates a recover function always
// recovering the paniced loop function.
func makeIgnorePanicsRF(bundle *channelBundle) loop.RecoverFunc {
	return func(rs loop.Recoverings) (loop.Recoverings, error) {
		bundle.donec <- "recovered"
		return nil, nil
	}
}

// EOF
