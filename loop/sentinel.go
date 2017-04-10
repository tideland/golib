// Tideland Go Application Support - Loop
//
// Copyright (C) 2013-2017 Frank Mueller / Tideland / Oldenburg / Germany
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
	"github.com/tideland/golib/logger"
)

//--------------------
// SENTINEL
//--------------------

// NotificationHandlerFunc allows a sentinel to react on
// an observers error notification.
type NotificationHandlerFunc func(s Sentinel, o Observable, rs Recoverings) (Recoverings, error)

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
	for observable := range s.observables {
		if err := f(observable); err != nil {
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
	var recoverings Recoverings
	for {
		select {
		case <-l.ShallStop():
			// We're done.
			return s.ObservablesDo(func(o Observable) error {
				logger.Infof("sentinel %q stops observable %q", s, o)
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
			// Receive notification about observable
			// with error.
			if s.handlerF != nil {
				// Try to handle the notification.
				recoverings = append(recoverings, &Recovering{time.Now(), err})
				recoverings, err = s.handlerF(s, o, recoverings)
			}
			if err != nil {
				// Still an error, so kill all.
				logger.Errorf("sentinel %q kills all observables after error: %v", s, err)
				s.ObservablesDo(func(o Observable) error {
					logger.Errorf("killing %q", o)
					o.Kill(errors.Annotate(err, ErrKilledBySentinel, errorMessages, o))
					return o.Wait()
				})
				return errors.Annotate(err, ErrHandlingFailed, errorMessages, o)
			}
		}
	}
}

// EOF
