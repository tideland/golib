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

	// start starts the child.
	start()

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
func goLoop(lf LoopFunc, rf RecoverFunc, s *sentinel, d string) *loop {
	l := &loop{
		descr:       d,
		loopFunc:    lf,
		recoverFunc: rf,
		startedc:    make(chan struct{}),
		stopc:       make(chan struct{}),
		donec:       make(chan struct{}),
		sentinel:    s,
	}
	// Check description.
	if l.descr == "" {
		l.descr = identifier.NewUUID().String()
	}
	// Start the loop.
	l.start()
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
	<-l.donec
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
	return l.stopc
}

// IsStopping implements the Loop interface.
func (l *loop) IsStopping() <-chan struct{} {
	return l.stopc
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

// start implements the manageable interface.
func (l *loop) start() {
	go l.run()
	<-l.startedc
	logger.Infof("loop '%s' started", l.description())
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
			logger.Errorf("loop '%s' tries to recover", l.description())
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
				logger.Errorf("loop '%s' panicked: %v", l.description(), reason)
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
		select {
		case <-l.donec:
		default:
			close(l.donec)
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
	werr := l.Wait()
	if l.sentinel != nil {
		l.sentinel.manageablec <- l
	}
	if werr != nil {
		logger.Errorf("loop '%s' stopped with error: %v", l.description(), werr)
		return werr
	}
	logger.Infof("loop '%s' stopped", l.description())
	return nil
}

// restart implements the manageable interface.
func (l *loop) restart() error {
	logger.Warningf("loop '%s' restarts", l.description())
	l.err = nil
	l.terminate(nil)
	if err := l.Wait(); err != nil {
		logger.Errorf("loop '%s' tried to restart with error: %v", l.description(), err)
		return err
	}
	l.start()
	return nil
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
func goSentinel(rf RecoverFunc, ps *sentinel, d string) *sentinel {
	s := &sentinel{
		recoverFunc: rf,
		manageables: make(map[manageable]struct{}),
		manageablec: make(chan manageable, 4),
		goablec:     make(chan *goable),
		startedc:    make(chan struct{}),
		sentinel:    ps,
	}
	s.loop = goLoop(s.backendLoop, s.recoverFunc, s.sentinel, d)
	return s
}

// Go implements the Sentinel interface.
func (s *sentinel) Go(lf LoopFunc, dps ...string) Loop {
	descr := identifier.JoinedIdentifier(dps...)
	g := &goable{
		descr:       descr,
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
func (s *sentinel) GoRecoverable(lf LoopFunc, rf RecoverFunc, dps ...string) Loop {
	descr := identifier.JoinedIdentifier(dps...)
	g := &goable{
		descr:       descr,
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
func (s *sentinel) GoSentinel(rf RecoverFunc, dps ...string) Sentinel {
	descr := identifier.JoinedIdentifier(dps...)
	g := &goable{
		descr:       descr,
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
		case g := <-s.goablec:
			// Spawn a new mangeable.
			var m manageable
			switch {
			case g.loopFunc != nil && g.recoverFunc != nil:
				// Recoverable loop.
				m = goLoop(g.loopFunc, g.recoverFunc, s, g.descr)
			case g.loopFunc != nil:
				// Simple loop.
				m = goLoop(g.loopFunc, nil, s, g.descr)
			case g.recoverFunc != nil:
				// Sentinel.
				m = goSentinel(g.recoverFunc, s, g.descr)
			}
			if m != nil {
				s.manageables[m] = struct{}{}
				mcc := cap(s.manageablec)
				if mcc == len(s.manageables) {
					s.manageablec = make(chan manageable, 2*mcc)
				}
			}
			g.manageablec <- m
		case m := <-s.manageablec:
			// Check loop error.
			if err := m.error(); err != nil {
				// Let the recovering function decide how to procede.
				var rerr error
				recoverings = append(recoverings, &Recovering{time.Now(), err})
				if recoverings, rerr = s.recoverFunc(recoverings); rerr != nil {
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
	errs := map[manageable]error{}
	for m := range s.manageables {
		if err := m.stop(err); err != nil {
			errs[m] = err
		}
	}
	// TODO: Return potential error!
	return nil
}

// restartAllChildren stops and starts all children.
func (s *sentinel) restartAllChildren() error {
	errs := map[manageable]error{}
	for m := range s.manageables {
		if err := m.restart(); err != nil {
			errs[m] = err
		}
	}
	// TODO: Return potential error!
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

// start implements the manageable interface.
func (s *sentinel) start() {
	s.loop.start()
}

// stop implements the manageable interface.
func (s *sentinel) stop(err error) error {
	return s.loop.stop(err)
}

// restart implements the manageable interface.
func (s *sentinel) restart() error {
	return s.loop.restart()
}

// EOF
