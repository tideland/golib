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
// Loop.ShallStop(). The loop then has to end working returning
// a possible error. Wait() then waits until the loop ended and
// returns the error.
func Go(lf LoopFunc, dps ...string) Loop {
	descr := identifier.JoinedIdentifier(dps...)
	return goLoop(lf, nil, nil, descr)
}

// GoRecoverable starts the loop function in the background. The
// loop can be stopped or killed. This leads to a signal out of the
// channel Loop.ShallStop(). The loop then has to end working returning
// a possible error. Wait() then waits until the loop ended and returns
// the error.
//
// If the loop panics a Recovering is created and passed with all
// Recoverings before to the RecoverFunc. If it returns nil the
// loop will be started again. Otherwise the loop will be killed
// with that error.
func GoRecoverable(lf LoopFunc, rf RecoverFunc, dps ...string) Loop {
	descr := identifier.JoinedIdentifier(dps...)
	return goLoop(lf, rf, nil, descr)
}

// GoSentinel starts a new sentinel. It can start simple and
// recoverable loops as well as nested sentinels. This way a
// managing tree can be setup.
func GoSentinel(rf RecoverFunc, dps ...string) Sentinel {
	descr := identifier.JoinedIdentifier(dps...)
	return goSentinel(rf, nil, descr)
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
// MANAGEABLE
//--------------------

// manageable is the common interface of loop and sentinel
// to be managed be a sentinel.
type manageable interface {
	// description returns the description of the manageable.
	description() string

	// error returns the error of the manageable or nil.
	error() error

	// stop terminates the child and tells
	// a potential sentinel that it's leaving.
	stop(err error) error

	// restart stops the child w/o an error and
	// restarts it afterwards.
	restart() error
}

//--------------------
// LOOP
//--------------------

// Status of the loop.
const (
	Running = iota
	Restarting
	Stopping
	Stopped
)

// LoopFunc is managed loop function.
type LoopFunc func(l Loop) error

// Loop manages running loops in the background as goroutines.
type Loop interface {
	// Description returns a descriptive information about the loop. If
	// it is started without a description a UUID is generated.
	Description() string

	// Stop tells the loop to stop working without a passed error and
	// waits until it is done.
	Stop() error

	// Kill tells the loop to stop working due to the passed error.
	// Here only the first error will be stored for later evaluation.
	Kill(err error)

	// Wait blocks the caller until the loop ended and returns the error.
	Wait() (err error)

	// Error returns the current status and error of the loop.
	Error() (status int, err error)

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
	loopF       LoopFunc
	recoverF    RecoverFunc
	recoverings Recoverings
	err         error
	status      int
	startedC    chan struct{}
	stopC       chan struct{}
	doneC       chan struct{}
	sentinel    *sentinel
}

// goLoop starts a loop in the background.
func goLoop(lf LoopFunc, rf RecoverFunc, s *sentinel, d string) *loop {
	l := &loop{
		descr:    d,
		loopF:    lf,
		recoverF: rf,
		startedC: make(chan struct{}),
		stopC:    make(chan struct{}),
		doneC:    make(chan struct{}),
		sentinel: s,
	}
	// Check description.
	if l.descr == "" {
		l.descr = identifier.NewUUID().String()
	}
	// Start the loop.
	logger.Infof("loop %q starts", l.description())
	go l.run()
	<-l.startedC
	return l
}

// Description implements the Loop interface.
func (l *loop) Description() string {
	return l.descr
}

// Stop implements the Loop interface.
func (l *loop) Stop() error {
	return l.stop(nil)
}

// Kill implements the Loop interface.
func (l *loop) Kill(err error) {
	l.stop(err)
}

// Wait implements the Loop interface.
func (l *loop) Wait() error {
	<-l.doneC
	l.mux.Lock()
	defer l.mux.Unlock()
	err := l.err
	return err
}

// Error implements the Loop interface.
func (l *loop) Error() (status int, err error) {
	l.mux.Lock()
	defer l.mux.Unlock()
	status = l.status
	err = l.err
	return
}

// ShallStop implements the Loop interface.
func (l *loop) ShallStop() <-chan struct{} {
	return l.stopC
}

// IsStopping implements the Loop interface.
func (l *loop) IsStopping() <-chan struct{} {
	return l.stopC
}

// description implements the manageable interface.
func (l *loop) description() string {
	return l.descr
}

// error implements the manageable interface.
func (l *loop) error() error {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.err
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

// stop implements the manageable interface.
func (l *loop) stop(err error) error {
	l.terminate(err)
	return l.Wait()
}

// restart implements the manageable interface.
func (l *loop) restart() error {
	logger.Warningf("loop %q restarts", l.description())
	l.mux.Lock()
	l.err = nil
	l.status = Restarting
	l.mux.Unlock()
	select {
	case <-l.stopC:
	default:
		close(l.stopC)
	}
	if err := l.Wait(); err != nil {
		logger.Errorf("loop %q failed restarting after error: %v", l.description(), err)
		return err
	}
	l.stopC = make(chan struct{})
	go l.run()
	<-l.startedC
	logger.Infof("loop %q restarted", l.description())
	return nil
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

// checkTermination checks if an error has been the reason and if
// it possibly can be recovered by a recover function.
func (l *loop) checkTermination(reason interface{}) {
	switch {
	case l.status == Restarting:
		// Quick exit, we are restarting.
		return
	case reason == nil:
		// Regular end.
		l.status = Stopping
	case l.recoverF == nil:
		// Error but no recover function.
		l.status = Stopping
		if err, ok := reason.(error); ok {
			l.err = err
		} else {
			l.err = errors.New(ErrLoopPanicked, errorMessages, reason)
		}
	default:
		// Try to recover.
		logger.Errorf("loop %q tries to recover", l.description())
		l.recoverings = append(l.recoverings, &Recovering{time.Now(), reason})
		l.recoverings, l.err = l.recoverF(l.recoverings)
		if l.err != nil {
			l.status = Stopping
		} else {
			logger.Infof("loop %q recovered", l.description())
		}
	}
}

// finalizeTermination notifies listeners that the loop stopped
// working and a potential sentinal about its status.
func (l *loop) finalizeTermination() {
	if l.status == Restarting {
		// Quickly handling a restart.
		select {
		case <-l.doneC:
		default:
			close(l.doneC)
		}
		return
	}
	// Handle a regular stopping.
	l.status = Stopped
	select {
	case <-l.doneC:
	default:
		close(l.doneC)
	}
	if l.sentinel != nil {
		l.sentinel.manageableC <- l
	}
	if l.err != nil {
		logger.Errorf("loop %q stopped with error: %v", l.description(), l.err)
	} else {
		logger.Infof("loop %q stopped", l.description())
	}
}

//--------------------
// SENTINEL
//--------------------

// Sentinel manages a number of loops or other sentinels.
type Sentinel interface {
	// Description returns a descriptive information about the loop. If
	// it is started without a description a UUID is generated.
	Description() string

	// Go works analog to the standard Go function,
	// but the loop is managed by the sentinel.
	Go(lf LoopFunc, dps ...string) Loop

	// GoRecoverable works analog to the standard GoRecoverable
	// function, but the loop is managed by the sentinel.
	GoRecoverable(lf LoopFunc, rf RecoverFunc, dps ...string) Loop

	// GoSentinel starts a new sentinel managed by this one.
	GoSentinel(rf RecoverFunc, dps ...string) Sentinel

	// Stop tells the loop to stop working without a passed error and
	// waits until it is done.
	Stop() error

	// Kill tells the loop to stop working due to the passed error.
	// Here only the first error will be stored for later evaluation.
	Kill(err error)

	// Wait blocks the caller until the loop ended and returns the error.
	Wait() (err error)

	// Error returns the current status and error of the loop.
	Error() (status int, err error)
}

// goable contains the information about a loop or sentinel to start.
type goable struct {
	descr       string
	loopF       LoopFunc
	recoverF    RecoverFunc
	manageableC chan manageable
}

// sentinel implements the Sentinel interface.
type sentinel struct {
	recoverF    RecoverFunc
	manageables map[manageable]struct{}
	manageableC chan manageable
	goableC     chan *goable
	startedC    chan struct{}
	loop        *loop
	sentinel    *sentinel
}

// goSentinel starts a new sentinel.
func goSentinel(rf RecoverFunc, ps *sentinel, d string) *sentinel {
	s := &sentinel{
		recoverF:    rf,
		manageables: make(map[manageable]struct{}),
		manageableC: make(chan manageable, 4),
		goableC:     make(chan *goable),
		startedC:    make(chan struct{}),
		sentinel:    ps,
	}
	s.loop = goLoop(s.backendLoop, s.recoverF, s.sentinel, d)
	return s
}

// Go implements the Sentinel interface.
func (s *sentinel) Go(lf LoopFunc, dps ...string) Loop {
	descr := identifier.JoinedIdentifier(dps...)
	g := &goable{
		descr:       descr,
		loopF:       lf,
		manageableC: make(chan manageable),
	}
	s.goableC <- g
	m := <-g.manageableC
	if m == nil {
		return nil
	}
	return m.(Loop)
}

// GoRecoverable implements the Sentinel interface.
func (s *sentinel) GoRecoverable(lf LoopFunc, rf RecoverFunc, dps ...string) Loop {
	descr := identifier.JoinedIdentifier(dps...)
	g := &goable{
		descr:       descr,
		loopF:       lf,
		recoverF:    rf,
		manageableC: make(chan manageable),
	}
	s.goableC <- g
	m := <-g.manageableC
	if m == nil {
		return nil
	}
	return m.(Loop)
}

// GoSentinel implements the Sentinel interface.
func (s *sentinel) GoSentinel(rf RecoverFunc, dps ...string) Sentinel {
	descr := identifier.JoinedIdentifier(dps...)
	g := &goable{
		descr:       descr,
		recoverF:    rf,
		manageableC: make(chan manageable),
	}
	s.goableC <- g
	m := <-g.manageableC
	if m == nil {
		return nil
	}
	return m.(Sentinel)
}

// Description implements the Sentinel interface.
func (s *sentinel) Description() string {
	return s.loop.Description()
}

// Stop implements the Sentinel interface.
func (s *sentinel) Stop() error {
	return s.loop.Stop()
}

// Kill implements the Sentinel interface.
func (s *sentinel) Kill(err error) {
	s.loop.Kill(err)
}

// Wait implements the Sentinel interface.
func (s *sentinel) Wait() error {
	return s.loop.Wait()
}

// Error implements the Sentinel interface.
func (s *sentinel) Error() (int, error) {
	return s.loop.Error()
}

// backendLoop listens to ending managed loops.
func (s *sentinel) backendLoop(l Loop) error {
	recoverings := Recoverings{}
	for {
		select {
		case <-l.ShallStop():
			// Stop all managed children.
			return s.stopAllChildren(nil)
		case g := <-s.goableC:
			// Spawn a new mangeable.
			var m manageable
			switch {
			case g.loopF != nil && g.recoverF != nil:
				// Recoverable loop.
				m = goLoop(g.loopF, g.recoverF, s, g.descr)
			case g.loopF != nil:
				// Simple loop.
				m = goLoop(g.loopF, nil, s, g.descr)
			case g.recoverF != nil:
				// Sentinel.
				m = goSentinel(g.recoverF, s, g.descr)
			}
			if m != nil {
				// Store manageable and ensure channel size.
				s.manageables[m] = struct{}{}
				mcc := cap(s.manageableC)
				if cap(s.manageableC) < len(s.manageables) {
					s.manageableC = make(chan manageable, mcc+4)
				}
			}
			g.manageableC <- m
		case m := <-s.manageableC:
			// Check loop error.
			if err := m.error(); err != nil {
				// Let the recovering function decide how to procede.
				var rerr error
				recoverings = append(recoverings, &Recovering{time.Now(), err})
				if recoverings, rerr = s.recoverF(recoverings); rerr != nil {
					// Ouch, we'll stop with an error. Let's hope we've
					// got a sentinel.
					return rerr
				}
				// Restart all children.
				if err := s.restartAllChildren(); err != nil {
					return err
				}
			} else {
				// Child terminated, remove it.
				delete(s.manageables, m)
			}
		}
	}
}

// stopAllChildren terminates all children.
func (s *sentinel) stopAllChildren(err error) error {
	errs := []error{}
	for m := range s.manageables {
		if err := m.stop(err); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Collect(errs...)
	}
	return nil
}

// restartAllChildren stops and starts all children.
func (s *sentinel) restartAllChildren() error {
	errs := []error{}
	for m := range s.manageables {
		if err := m.restart(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Collect(errs...)
	}
	return nil
}

// description implementes the manageable interface.
func (s *sentinel) description() string {
	return s.loop.description()
}

// error implements the manageable interface.
func (s *sentinel) error() error {
	return s.loop.error()
}

// stop implements the manageable interface.
func (s *sentinel) stop(err error) error {
	return s.loop.stop(err)
}

// restart implements the manageable interface.
func (s *sentinel) restart() error {
	logger.Warningf("sentinel %q restarts", s.description())
	err := s.restartAllChildren()
	logger.Infof("sentinel %q restarted", s.description())
	return err
}

// EOF
