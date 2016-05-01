// Tideland Go Application Support - Loop
//
// Copyright (C) 2013-2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package loop

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"sync"
	"time"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/identifier"
	"github.com/tideland/golib/logger"
)

//--------------------
// API
//--------------------

// Go starts the loop function in the background. The loop can be
// stopped or killed. This leads to a signal out of the channel
// Loop.ShallStop. The loop then has to end working returning
// a possible error. Wait then waits until the loop ended and
// returns the error.
func Go(lf LoopFunc, dps ...interface{}) Loop {
	descr := identifier.SepIdentifier("::", dps...)
	return goLoop(lf, nil, nil, nil, descr)
}

// GoRecoverable starts the loop function in the background. The
// loop can be stopped or killed. This leads to a signal out of the
// channel Loop.ShallStop. The loop then has to end working returning
// a possible error. Wait then waits until the loop ended and returns
// the error.
//
// If the loop panics a Recovering is created and passed with all
// Recoverings before to the RecoverFunc. If it returns nil the
// loop will be started again. Otherwise the loop will be killed
// with that error.
func GoRecoverable(lf LoopFunc, rf RecoverFunc, dps ...interface{}) Loop {
	descr := identifier.SepIdentifier("::", dps...)
	return goLoop(lf, rf, nil, nil, descr)
}

// GoSentinel starts a new sentinel. It can start simple and
// recoverable loops as well as nested sentinels. This way a
// managing tree can be setup.
func GoSentinel(nhf NotificationHandlerFunc, dps ...interface{}) Sentinel {
	descr := identifier.SepIdentifier("::", dps...)
	return goSentinel(nhf, nil, descr)
}

//--------------------
// RECOVERING
//--------------------

// Recovering stores time and reason of one of the recoverings.
type Recovering struct {
	Time   time.Time
	Reason interface{}
}

// Recoverings is a list of recoverings a loop already had.
type Recoverings []*Recovering

// Frequency checks if a given number of restarts happened during
// a given duration.
func (rs Recoverings) Frequency(num int, dur time.Duration) bool {
	if len(rs) >= num {
		first := rs[len(rs)-num].Time
		last := rs[len(rs)-1].Time
		return last.Sub(first) <= dur
	}
	return false
}

// Len returns the length of the recoverings.
func (rs Recoverings) Len() int {
	return len(rs)
}

// Trim returns the last recoverings defined by l. This
// way the recover func can con control the length and take
// care that the list not grows too much.
func (rs Recoverings) Trim(l int) Recoverings {
	if l >= len(rs) {
		return rs
	}
	return rs[len(rs)-l:]
}

// Last returns the last recovering.
func (rs Recoverings) First() *Recovering {
	if len(rs) > 0 {
		return rs[0]
	}
	return nil
}

// Last returns the last recovering.
func (rs Recoverings) Last() *Recovering {
	if len(rs) > 0 {
		return rs[len(rs)-1]
	}
	return nil
}

// RecoverFunc decides if a loop shall be started again or
// end with an error. It is also responsible to trim the
// list of revocerings if needed.
type RecoverFunc func(rs Recoverings) (Recoverings, error)

//--------------------
// OBSERVABLE
//--------------------

// Observable is a common base interface for those objects
// that a sentinel can monitor.
type Observable interface {
	fmt.Stringer

	// Stop tells the observable to stop working and waits until it is done.
	Stop() error

	// Kill kills the observable with the passed error.
	Kill(err error)

	// Wait blocks the caller until the observable ended and returns
	// a possible error.
	Wait() error

	// Restart stops the observable and restarts it afterwards.
	Restart() error

	// Error returns information about the current status and error.
	Error() (status int, err error)

	// attachSentinel attaches the observable to a sentinel.
	attachSentinel(s *sentinel)
}

//--------------------
// LOOP
//--------------------

// Status of the loop.
const (
	Running = iota
	Stopping
	Stopped
)

// LoopFunc is managed loop function.
type LoopFunc func(l Loop) error

// Loop manages running loops in the background as goroutines.
type Loop interface {
	Observable

	// ShallStop returns a channel signalling the loop to
	// stop working.
	ShallStop() <-chan struct{}

	// IsStopping returns a channel that can be used to wait until
	// the loop is stopping or to avoid deadlocks when communicating
	// with the loop.
	IsStopping() <-chan struct{}
}

// Loop manages a loop function.
type loop struct {
	mux         sync.Mutex
	descr       string
	status      int
	err         error
	loopF       LoopFunc
	recoverF    RecoverFunc
	recoverings Recoverings
	startedC    chan struct{}
	stopC       chan struct{}
	doneC       chan struct{}
	owner       Observable
	sentinel    *sentinel
}

// goLoop starts a loop in the background.
func goLoop(lf LoopFunc, rf RecoverFunc, o Observable, s *sentinel, d string) *loop {
	l := &loop{
		descr:    d,
		loopF:    lf,
		recoverF: rf,
		startedC: make(chan struct{}),
		stopC:    make(chan struct{}),
		doneC:    make(chan struct{}),
		owner:    o,
		sentinel: s,
	}
	// Check description.
	if l.descr == "" {
		l.descr = identifier.NewUUID().String()
	}
	// Check owner, at least we should own ourself.
	if l.owner == nil {
		l.owner = l
	}
	// Start the loop.
	logger.Infof("loop %q starts", l)
	go l.run()
	<-l.startedC
	return l
}

// String implements the Stringer interface. It returns
// the description of the loop.
func (l *loop) String() string {
	return l.descr
}

// Stop implements the Observable interface.
func (l *loop) Stop() error {
	l.terminate(nil)
	return l.Wait()
}

// Kill implements the Observable interface.
func (l *loop) Kill(err error) {
	l.terminate(err)
}

// Wait implements the Observable interface.
func (l *loop) Wait() error {
	<-l.doneC
	l.mux.Lock()
	defer l.mux.Unlock()
	err := l.err
	return err
}

// Restart implements the Observable interface.
func (l *loop) Restart() error {
	l.mux.Lock()
	defer l.mux.Unlock()
	if l.status != Stopped {
		return errors.New(ErrRestartNonStopped, errorMessages, l)
	}
	l.err = nil
	l.recoverings = nil
	l.status = Running
	l.stopC = make(chan struct{})
	l.doneC = make(chan struct{})
	// Restart the goroutine.
	logger.Infof("loop %q restarts", l)
	go l.run()
	<-l.startedC
	return nil
}

// Error implements the Observable interface.
func (l *loop) Error() (status int, err error) {
	l.mux.Lock()
	defer l.mux.Unlock()
	status = l.status
	err = l.err
	return
}

// attachSentinel implements the Observable interface.
func (l *loop) attachSentinel(s *sentinel) {
	l.mux.Lock()
	defer l.mux.Unlock()
	if l.sentinel != nil {
		l.sentinel.Forget(l)
	}
	l.sentinel = s
}

// ShallStop implements the Loop interface.
func (l *loop) ShallStop() <-chan struct{} {
	return l.stopC
}

// IsStopping implements the Loop interface.
func (l *loop) IsStopping() <-chan struct{} {
	return l.stopC
}

// run operates the loop as goroutine.
func (l *loop) run() {
	l.status = Running
	// Finalize the loop.
	defer l.finalizeTermination()
	// Create a loop wrapper containing the recovering control.
	loopWrapper := func() {
		defer func() {
			// Check for recovering.
			if reason := recover(); reason != nil {
				l.checkTermination(reason)
			}
		}()
		l.checkTermination(l.loopF(l))
	}
	// Now start runnung the loop wrappr.
	l.startedC <- struct{}{}
	for l.status == Running {
		loopWrapper()
	}
}

// terminate tells the loop to stop working and stores
// the passed error if none has been stored already.
func (l *loop) terminate(err error) {
	l.mux.Lock()
	defer l.mux.Unlock()
	if l.err == nil {
		l.err = err
	}
	if l.status != Running {
		return
	}
	l.status = Stopping
	select {
	case <-l.stopC:
	default:
		close(l.stopC)
	}
}

// checkTermination checks if an error has been the reason and if
// it possibly can be recovered by a recover function.
func (l *loop) checkTermination(reason interface{}) {
	switch {
	case reason == nil:
		// Regular end.
		l.status = Stopping
	case l.recoverF == nil:
		// Error but no recover function.
		l.status = Stopping
		if l.err != nil {
			break
		}
		if err, ok := reason.(error); ok {
			l.err = err
		} else {
			l.err = errors.New(ErrLoopPanicked, errorMessages, reason)
		}
	default:
		// Try to recover.
		logger.Errorf("loop %q tries to recover", l)
		l.recoverings = append(l.recoverings, &Recovering{time.Now(), reason})
		l.recoverings, l.err = l.recoverF(l.recoverings)
		if l.err != nil {
			l.status = Stopping
		} else {
			logger.Infof("loop %q recovered", l)
		}
	}
}

// finalizeTermination notifies listeners that the loop stopped
// working and a potential sentinal about its status.
func (l *loop) finalizeTermination() {
	l.status = Stopped
	// Close stopC in case  the termination is due to an
	// error or internal.
	select {
	case <-l.stopC:
	default:
		close(l.stopC)
	}
	// Communicate that it' done.
	select {
	case <-l.doneC:
	default:
		close(l.doneC)
	}
	// If a sentinel monitors us then till him.
	if l.sentinel != nil {
		if l.err != nil {
			// Notify sentinel about error termination.
			l.sentinel.notifyC <- l.owner
		} else {
			// Tell sentinel to remove loop.
			l.sentinel.Forget(l)
		}
	}
	if l.err != nil {
		logger.Errorf("loop %q stopped with error: %v", l, l.err)
	} else {
		logger.Infof("loop %q stopped", l)
	}
}

//--------------------
// SENTINEL
//--------------------

// NotificationHandlerFunc allows a sentinel to react on
// an observers error notification.
type NotificationHandlerFunc func(s Sentinel, o Observable) error

// Sentinel manages a number of loops or other sentinels.
type Sentinel interface {
	Observable

	// Observe tells the sentinel to monitor the passed observables.
	Observe(o ...Observable)

	// Forget tells the sentinel to forget the passed observables.
	Forget(o ...Observable)

	// ObservablesDo executes the passed function for each observable,
	// e.g. to react after an error.
	ObservablesDo(f func(o Observable) error) error
}

type observableChange struct {
	observables []Observable
	doneC       chan struct{}
}

// sentinel implements the Sentinel interface.
type sentinel struct {
	mux         sync.Mutex
	descr       string
	loop        *loop
	handlerF    NotificationHandlerFunc
	observables map[Observable]struct{}
	addC        chan *observableChange
	removeC     chan *observableChange
	notifyC     chan Observable
}

// goSentinel starts a new sentinel.
func goSentinel(nhf NotificationHandlerFunc, ps *sentinel, d string) *sentinel {
	s := &sentinel{
		descr:       d,
		handlerF:    nhf,
		observables: make(map[Observable]struct{}),
		addC:        make(chan *observableChange),
		removeC:     make(chan *observableChange),
		notifyC:     make(chan Observable),
	}
	s.loop = goLoop(s.backendLoop, nil, s, ps, d)
	return s
}

// String implements the Stringer interface. It returns
// the description of the sentinel.
func (s *sentinel) String() string {
	return s.descr
}

// Stop implements the Observable interface.
func (s *sentinel) Stop() error {
	return s.loop.Stop()
}

// Kill implements the Observable interface.
func (s *sentinel) Kill(err error) {
	s.loop.Kill(err)
}

// Wait implements the Observable interface.
func (s *sentinel) Wait() error {
	return s.loop.Wait()
}

// Error implements the Observable interface.
func (s *sentinel) Error() (int, error) {
	return s.loop.Error()
}

// Restart implements the Observable interface.
func (s *sentinel) Restart() error {
	logger.Infof("sentinel %q restarts", s)
	// Start backendLoop again.
	s.loop.Restart()
	// Now restart children.
	return s.ObservablesDo(func(o Observable) error {
		return o.Restart()
	})
}

// attachSentinel implements the Observable interface.
func (s *sentinel) attachSentinel(ps *sentinel) {
	s.loop.attachSentinel(ps)
}

// Observe implements the Sentinel interface.
func (s *sentinel) Observe(os ...Observable) {
	change := &observableChange{
		observables: os,
		doneC:       make(chan struct{}),
	}
	s.addC <- change
	<-change.doneC
}

// Forget implements the Sentinel interface.
func (s *sentinel) Forget(os ...Observable) {
	change := &observableChange{
		observables: os,
		doneC:       make(chan struct{}),
	}
	s.removeC <- change
	<-change.doneC
}

// ObservablesDo implements the Sentinel interface.
func (s *sentinel) ObservablesDo(f func(o Observable) error) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	var errs []error
	for o, _ := range s.observables {
		if err := f(o); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Collect(errs...)
	}
	return nil
}

// backendLoop listens to ending managed loops.
func (s *sentinel) backendLoop(l Loop) error {
	for {
		select {
		case <-l.ShallStop():
			// We're done.
			return s.ObservablesDo(func(o Observable) error {
				o.Stop()
				return nil
			})
		case change := <-s.addC:
			// Add new observables.
			for _, o := range change.observables {
				s.observables[o] = struct{}{}
				o.attachSentinel(s)
				logger.Infof("started observing %q", o)
			}
			close(change.doneC)
		case change := <-s.removeC:
			// Remove observable.
			for _, o := range change.observables {
				delete(s.observables, o)
				logger.Infof("stopped observing %q", o)
			}
			close(change.doneC)
		case o := <-s.notifyC:
			_, err := o.Error()
			// First check if my own loop has troubles.
			if o == s {
				return err
			}
			// Recieve notification about observable
			// with error.
			if s.handlerF != nil {
				// Try to handle the notification.
				err = s.handlerF(s, o)
			}
			if err != nil {
				// Still an error, so kill all.
				logger.Errorf("sentinel %q stops all observables after error: %v", s, err)
				s.ObservablesDo(func(o Observable) error {
					logger.Errorf("stopping %q", o)
					return o.Stop()
				})
				return errors.Annotate(err, ErrHandlingFailed, errorMessages, o)
			}
		}
	}
}

// EOF
