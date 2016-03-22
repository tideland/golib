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
)

//--------------------
// API
//--------------------

// Go starts the loop function in the background. The loop can be
// stopped or killed. This leads to a signal out of the channel
// Loop.ShallStop(). The loop then has to end working returning
// a possible error. Wait() then waits until the loop ended and
// returns the error.
func Go(lf LoopFunc) Loop {
	return goLoop(lf, nil, nil)
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
func GoRecoverable(lf LoopFunc, rf RecoverFunc) Loop {
	return goLoop(lf, rf, nil)
}

// GoSentinel starts a new sentinel. It can start simple and
// recoverable loops as well as nested sentinels. This way a
// managing tree can be setup.
func GoSentinel(rf RecoverFunc) Sentinel {
	return goSentinel(rf, nil)
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
	// error returns the error of the manageable or nil.
	error() error

	// start starts the child.
	start()

	// stop terminates the child.
	stop(err error) error
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
	loopFunc    LoopFunc
	recoverFunc RecoverFunc
	err         error
	status      int
	startedc    chan struct{}
	stopc       chan struct{}
	donec       chan struct{}
	sentinel    *sentinel
}

// goLoop starts a loop in the background.
func goLoop(lf LoopFunc, rf RecoverFunc, s *sentinel) *loop {
	l := &loop{
		loopFunc:    lf,
		recoverFunc: rf,
		startedc:    make(chan struct{}),
		stopc:       make(chan struct{}),
		donec:       make(chan struct{}),
		sentinel:    s,
	}
	// Start the loop.
	l.start()
	return l
}

// Stop tells the loop to stop working without a passed error and
// waits until it is done.
func (l *loop) Stop() error {
	return l.stop(nil)
}

// Kill tells the loop to stop working due to the passed error.
// Here only the first error will be stored for later evaluation.
func (l *loop) Kill(err error) {
	l.stop(err)
}

// Wait blocks the caller until the loop ended and returns the error.
func (l *loop) Wait() (err error) {
	<-l.donec
	l.mux.Lock()
	defer l.mux.Unlock()
	err = l.err
	return
}

// Error returns the current status and error of the loop.
func (l *loop) Error() (status int, err error) {
	l.mux.Lock()
	defer l.mux.Unlock()
	status = l.status
	err = l.err
	return
}

// ShallStop returns a channel signalling the loop to
// stop working.
func (l *loop) ShallStop() <-chan struct{} {
	return l.stopc
}

// IsStopping returns a channel that can be used to wait until
// the loop is stopping or to avoid deadlocks when communicating
// with the loop.
func (l *loop) IsStopping() <-chan struct{} {
	return l.stopc
}

// error implements the manageable interface.
func (l *loop) error() error {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.err
}

// start implements the manageable interface.
func (l *loop) start() {
	go l.run()
	<-l.startedc
}

// run operates the loop as goroutine.
func (l *loop) run() {
	defer l.done()
	l.status = Running
	run := true
	recoverings := Recoverings{}
	// Create an error check function.
	checkError := func(reason interface{}) {
		if reason == nil {
			// No error.
			l.terminate(nil)
			run = false
			return
		}
		if err, ok := reason.(error); ok && l.recoverFunc == nil {
			// Error and no recover function.
			l.terminate(err)
			run = false
			return
		}
		if l.recoverFunc != nil {
			// Try recovering.
			var err error
			recoverings = append(recoverings, &Recovering{time.Now(), reason})
			if recoverings, err = l.recoverFunc(recoverings); err != nil {
				l.terminate(err)
				run = false
			}
			return
		}
		// Panic and no recover function.
		l.terminate(errors.New(ErrLoopPanicked, errorMessages, reason))
		run = false
	}
	// Create a loop wrapper containing the recovering control.
	loopWrapper := func() {
		defer func() {
			// Check for recovering.
			if reason := recover(); reason != nil {
				checkError(reason)
			}
		}()
		checkError(l.loopFunc(l))
	}
	// Now start runnung the loop wrappr.
	l.startedc <- struct{}{}
	for run {
		loopWrapper()
	}
}

// done finalizes the stopping of the loop.
func (l *loop) done() {
	l.mux.Lock()
	defer l.mux.Unlock()
	if l.status == Stopping {
		l.status = Stopped
		close(l.donec)
		if l.sentinel != nil {
			l.sentinel.manageablec <- l
		}
	}
}

// terminate the loop and store the passed error if none has
// been stored already.
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
	case <-l.stopc:
	default:
		close(l.stopc)
	}
}

// stop implements the manageable interface.
func (l *loop) stop(err error) error {
	l.terminate(err)
	return l.Wait()
}

//--------------------
// SENTINEL
//--------------------

// Sentinel manages a number of loops or other sentinels.
type Sentinel interface {
	// Go works analog to the standard Go function,
	// but the loop is managed by the sentinel.
	Go(lf LoopFunc) Loop

	// GoRecoverable works analog to the standard GoRecoverable
	// function, but the loop is managed by the sentinel.
	GoRecoverable(lf LoopFunc, rf RecoverFunc) Loop

	// GoSentinel starts a new sentinel managed by this one.
	GoSentinel(rf RecoverFunc) Sentinel
}

// goable contains the information about a loop or sentinel to start.
type goable struct {
	loopFunc    LoopFunc
	recoverFunc RecoverFunc
	manageablec chan manageable
}

// sentinel implements the Sentinel interface.
type sentinel struct {
	recoverFunc RecoverFunc
	manageables map[manageable]struct{}
	manageablec chan manageable
	goablec     chan *goable
	startedc    chan struct{}
	loop        *loop
	sentinel    *sentinel
}

// goSentinel starts a new sentinel.
func goSentinel(rf RecoverFunc, ps *sentinel) *sentinel {
	s := &sentinel{
		recoverFunc: rf,
		manageables: make(map[manageable]struct{}),
		manageablec: make(chan manageable, 1),
		goablec:     make(chan *goable),
		startedc:    make(chan struct{}),
		sentinel:    ps,
	}
	s.loop = goLoop(s.backendLoop, s.recoverFunc, s.sentinel)
	return s
}

// Go implements the Sentinel interface.
func (s *sentinel) Go(lf LoopFunc) Loop {
	g := &goable{
		loopFunc:    lf,
		manageablec: make(chan manageable),
	}
	s.goablec <- g
	m := <-g.manageablec
	if m == nil {
		return nil
	}
	return m.(Loop)
}

// GoRecoverable implements the Sentinel interface.
func (s *sentinel) GoRecoverable(lf LoopFunc, rf RecoverFunc) Loop {
	g := &goable{
		loopFunc:    lf,
		recoverFunc: rf,
		manageablec: make(chan manageable),
	}
	s.goablec <- g
	m := <-g.manageablec
	if m == nil {
		return nil
	}
	return m.(Loop)
}

// GoSentinel implements the Sentinel interface.
func (s *sentinel) GoSentinel(rf RecoverFunc) Sentinel {
	g := &goable{
		recoverFunc: rf,
		manageablec: make(chan manageable),
	}
	s.goablec <- g
	m := <-g.manageablec
	if m == nil {
		return nil
	}
	return m.(Sentinel)
}

// backendLoop listens to ending managed loops.
func (s *sentinel) backendLoop(l Loop) error {
	for {
		select {
		case <-l.ShallStop():
			// Stop all managed children.
			return s.stopAllChildren(nil)
		case g := <-s.goablec:
			// Spawn a new mangeable.
			var m manageable
			switch {
			case g.loopFunc != nil && g.recoverFunc != nil:
				// Recoverable loop.
				m = goLoop(g.loopFunc, g.recoverFunc, s)
			case g.loopFunc != nil:
				// Simple loop.
				m = goLoop(g.loopFunc, nil, s)
			case g.recoverFunc != nil:
				// Sentinel.
				m = goSentinel(g.recoverFunc, s)
			}
			if m != nil {
				s.manageables[m] = struct{}{}
			}
			g.manageablec <- m
		case m := <-s.manageablec:
			// Check loop error.
			if err := m.error(); err != nil {
				// Restart all children.
				if err := s.restartAllChildren(err); err != nil {
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
	errs := map[manageable]error{}
	for m := range s.manageables {
		if err := m.stop(err); err != nil {
			errs[m] = err
		}
	}
	return nil
}

// startAllChildren starts all children.
func (s *sentinel) startAllChildren() {
	for m := range s.manageables {
		m.start()
	}
}

// restartAllChildren stops and starts all children.
func (s *sentinel) restartAllChildren(err error) error {
	serr := errors.New(ErrSentinelRecovers, errorMessages, err)
	if err := s.stopAllChildren(serr); err != nil {
		return err
	}
	s.startAllChildren()
	return nil
}

// error implements the manageable interface.
func (s *sentinel) error() error {
	return s.loop.error()
}

// start implements the manageable interface.
func (s *sentinel) start() {
	s.loop.start()
}

// stop implements the manageable interface.
func (s *sentinel) stop(err error) error {
	return s.loop.stop(err)
}

// EOF
