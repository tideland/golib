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
	// "github.com/tideland/golib/errors"
)

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

// Go starts the loop function in the background. The loop can be
// stopped or killed. This leads to a signal out of the channel
// Loop.ShallStop(). The loop then has to end working returning
// a possible error. Wait() then waits until the loop ended and
// returns the error.
func Go(lf LoopFunc) Loop {
	l := &loop{
		loopFunc: lf,
		status:   Running,
		startedc: make(chan struct{}),
		stopc:    make(chan struct{}),
		donec:    make(chan struct{}),
	}
	// Start and wait until it's running.
	go l.simpleLoop()
	<-l.startedc
	return l
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
	l := &loop{
		loopFunc:    lf,
		recoverFunc: rf,
		status:      Running,
		startedc:    make(chan struct{}),
		stopc:       make(chan struct{}),
		donec:       make(chan struct{}),
	}
	// Start and wait until it's running.
	go l.recoverableLoop()
	<-l.startedc
	return l
}

// simpleLoop is the goroutine for a loop which is not recoverable.
func (l *loop) simpleLoop() {
	defer l.done()
	l.startedc <- struct{}{}
	l.Kill(l.loopFunc(l))
}

// recoverableLoop is the goroutine for loops which
func (l *loop) recoverableLoop() {
	defer l.done()
	l.startedc <- struct{}{}
	run := true
	rs := Recoverings{}
	loop := func() {
		defer func() {
			// Loop ended due to a panic, check recovering.
			if r := recover(); r != nil {
				var err error
				rs = append(rs, &Recovering{time.Now(), r})
				if rs, err = l.recoverFunc(rs); err != nil {
					l.Kill(err)
					run = false
				}
			}
		}()
		err := l.loopFunc(l)
		if err != nil {
			// Loop ends with error, check recovering.
			rs = append(rs, &Recovering{time.Now(), err})
			if rs, err = l.recoverFunc(rs); err != nil {
				l.Kill(err)
				run = false
			}
		} else {
			// Loop ends w/o any error.
			l.Kill(nil)
			run = false
		}
	}
	for run {
		loop()
	}
}

// done finalizes the stopping of the loop.
func (l *loop) done() {
	l.mux.Lock()
	defer l.mux.Unlock()
	if l.status == Stopping {
		l.status = Stopped
		close(l.donec)
	}
}

// Stop tells the loop to stop working without a passed error and
// waits until it is done.
func (l *loop) Stop() error {
	return l.stop()
}

// Kill tells the loop to stop working due to the passed error.
// Here only the first error will be stored for later evaluation.
func (l *loop) Kill(err error) {
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

// start starts the loop.
func (l *loop) start() error {
	return nil
}

// hasError returns true if the loop has an error.
func (l *loop) hasError() bool {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.err != nil
}

// stop terminates the loop.
func (l *loop) stop() error {
	l.Kill(nil)
	return l.Wait()
}

// restart stops and starts the loop.
func (l *loop) restart() error {
	if err := l.stop(); err != nil {
		return err
	}
	return l.start()
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
	GoSentinel() Sentinel
}

// manageable defines the common interface of loop and sentinel
// to be managed be a sentinel.
type manageable interface {
	// hasError returns true if the child has an error.
	hasError() bool

	// stop terminates the child.
	stop() error

	// restart stops and starts the child.
	restart() error
}

// sentinel implements the Sentinel interface.
type sentinel struct {
	manageables map[manageable]struct{}
	manageablec chan manageable
	restartc    chan struct{}
	loop        Loop
	sentinel    *sentinel
}

// Go implements the Sentinel interface.
func (s *sentinel) Go(lf LoopFunc) Loop {
	return nil
}

// GoRecoverable implements the Sentinel interface.
func (s *sentinel) GoRecoverable(lf LoopFunc, rf RecoverFunc) Loop {
	return nil
}

// GoSentinel implements the Sentinel interface.
func (s *sentinel) GoSentinel() Sentinel {
	return nil
}

// backendLoop listens to ending managed loops.
func (s *sentinel) backendLoop(l Loop) error {
	for {
		select {
		case <-l.ShallStop():
			// Stop all managed children.
			return s.stopAllChildren()
		case m := <-s.manageablec:
			// Check loop error.
			if m.hasError() {
				// Restart all children.
				if err := s.restartAllChildren(); err != nil {
					return err
				}
			} else {
				// Child terminated, remove it.
				delete(s.manageables, m)
			}
		case <-s.restartc:
			// Restart all children.
			if err := s.restartAllChildren(); err != nil {
				return err
			}
		}
	}
}

// checkRecovering checks if a sentinel error shall be recovered.
func (s *sentinel) checkRecovering(rs Recoverings) (Recoverings, error) {
}

// stopAllChildren terminates all children.
func (s *sentinel) stopAllChildren() error {
	return nil
}

// startAllChildren starts all children.
func (s *sentinel) startAllChildren() error {
	return nil
}

// restartAllChildren stops and starts all children.
func (s *sentinel) restartAllChildren() error {
	if err := s.stopAllChildren(); err != nil {
		return err
	}
	return s.startAllChildren()
}

func (s *sentinel) hasError() bool {
	return false
}

func (s *sentinel) stop() error {
	return nil
}

func (s *sentinel) restart() error {
	return nil
}

// GoSentinel starts a new sentinel.
func GoSentinel() Sentinel {
	s := &sentinel{
		manageables: make(map[manageable]struct{}),
		manageablec: make(chan manageable),
	}
	s.loop = GoRecoverable(s.backendLoop, s.checkRecovering)
	return s
}

// EOF
