// Tideland Go Application Support - Loop
//
// Copyright (C) 2013-2014 Frank Mueller / Tideland / Oldenburg / Germany
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
)

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
	go l.singleLoop()
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

// singleLoop is the goroutine for a loop which is not recoverable.
func (l *loop) singleLoop() {
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
	l.Kill(nil)
	return l.Wait()
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

// EOF
