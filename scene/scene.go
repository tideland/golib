// Tideland Go Library - Scene
//
// Copyright (C) 2014-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package scene

//--------------------
// IMPORTS
//--------------------

import (
	"time"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/identifier"
	"github.com/tideland/golib/loop"
)

//--------------------
// SCENE
//--------------------

const (
	Active    = loop.Running
	Finishing = loop.Stopping
	Over      = loop.Stopped
)

// CleanupFunc is a function for the cleanup of props after
// a scene ended.
type CleanupFunc func(key string, prop interface{}) error

// box contains a prop and a possible cleanup function.
type box struct {
	key     string
	prop    interface{}
	cleanup CleanupFunc
}

// signaling contains a topic and a signal channel.
type signaling struct {
	topic      string
	signalChan chan struct{}
}

const (
	storeProp = iota
	fetchProp
	disposeProp
	flag
	unflag
	wait
)

// envelope contains information transferred between client and scene.
type envelope struct {
	kind      int
	box       *box
	signaling *signaling
	err       error
	respChan  chan *envelope
}

// Scene is the access point to one scene. It has to be created once
// for a continuous flow of operations and then passed between all
// functions and goroutine which are actors of the scene.
type Scene interface {
	// ID returns the unique ID of the scene.
	ID() identifier.UUID

	// Stop tells the scene to end and waits until it is done.
	Stop() error

	// Abort tells the scene to end due to the passed error.
	// Here only the first error will be stored for later evaluation.
	Abort(err error)

	// Wait blocks the caller until the scene ended and returns a
	// possible error or nil.
	Wait() error

	// Status returns information about the current status of the scene.
	Status() (int, error)

	// Store stores a prop with a given key. The key must not exist.
	Store(key string, prop interface{}) error

	// StoreAndFlag stores a prop with a given key. The key must not exist.
	// The storing is signaled with the key as topic.
	StoreAndFlag(key string, prop interface{}) error

	// StoreClean stores a prop with a given key and a cleanup
	// function called when a scene ends. The key must not exist.
	StoreClean(key string, prop interface{}, cleanup CleanupFunc) error

	// StoreClean stores a prop with a given key and a cleanup
	// function called when a scene ends. The key must not exist.
	// The storing is signaled with the key as topic.
	StoreCleanAndFlag(key string, prop interface{}, cleanup CleanupFunc) error

	// Fetch retrieves a prop.
	Fetch(key string) (interface{}, error)

	// Dispose retrieves a prop and deletes it from the store.
	Dispose(key string) (interface{}, error)

	// Flag allows to signal a topic to interested actors.
	Flag(topic string) error

	// Unflag drops the signal for a given topic.
	Unflag(topic string) error

	// WaitFlag waits until the passed topic has been signaled.
	WaitFlag(topic string) error

	// WaitFlagAndFetch waits until the passed topic has been signaled.
	// A prop stored at the topic as key is fetched.
	WaitFlagAndFetch(topic string) (interface{}, error)

	// WaitFlagLimited waits until the passed topic has been signaled
	// or the timeout happened.
	WaitFlagLimited(topic string, timeout time.Duration) error

	// WaitFlagLimitedAndFetch waits until the passed topic has been signaled
	// or the timeout happened. A prop stored at the topic as key is fetched.
	WaitFlagLimitedAndFetch(topic string, timeout time.Duration) (interface{}, error)
}

// scene implements Scene.
type scene struct {
	id          identifier.UUID
	props       map[string]*box
	flags       map[string]bool
	signalings  map[string][]chan struct{}
	inactivity  time.Duration
	absolute    time.Duration
	commandChan chan *envelope
	backend     loop.Loop
}

// Start creates and runs a new scene.
func Start() Scene {
	return StartLimited(0, 0)
}

// StartLimited creates and runs a new scene with an inactivity
// and an absolute timeout. They may be zero.
func StartLimited(inactivity, absolute time.Duration) Scene {
	s := &scene{
		id:          identifier.NewUUID(),
		props:       make(map[string]*box),
		flags:       make(map[string]bool),
		signalings:  make(map[string][]chan struct{}),
		inactivity:  inactivity,
		absolute:    absolute,
		commandChan: make(chan *envelope, 1),
	}
	s.backend = loop.Go(s.backendLoop, "scene", s.id.String())
	return s
}

// ID is specified on the Scene interface.
func (s *scene) ID() identifier.UUID {
	return s.id.Copy()
}

// Stop is specified on the Scene interface.
func (s *scene) Stop() error {
	return s.backend.Stop()
}

// Abort is specified on the Scene interface.
func (s *scene) Abort(err error) {
	s.backend.Kill(err)
}

// Wait is specified on the Scene interface.
func (s *scene) Wait() error {
	return s.backend.Wait()
}

// Status is specified on the Scene interface.
func (s *scene) Status() (int, error) {
	return s.backend.Error()
}

// Store is specified on the Scene interface.
func (s *scene) Store(key string, prop interface{}) error {
	return s.StoreClean(key, prop, nil)
}

// StoreAndFlag is specified on the Scene interface.
func (s *scene) StoreAndFlag(key string, prop interface{}) error {
	err := s.StoreClean(key, prop, nil)
	if err != nil {
		return err
	}
	return s.Flag(key)
}

// StoreClean is specified on the Scene interface.
func (s *scene) StoreClean(key string, prop interface{}, cleanup CleanupFunc) error {
	command := &envelope{
		kind: storeProp,
		box: &box{
			key:     key,
			prop:    prop,
			cleanup: cleanup,
		},
		respChan: make(chan *envelope, 1),
	}
	_, err := s.command(command)
	return err
}

// StoreCleanAndFlag is specified on the Scene interface.
func (s *scene) StoreCleanAndFlag(key string, prop interface{}, cleanup CleanupFunc) error {
	err := s.StoreClean(key, prop, cleanup)
	if err != nil {
		return err
	}
	return s.Flag(key)
}

// Fetch is specified on the Scene interface.
func (s *scene) Fetch(key string) (interface{}, error) {
	command := &envelope{
		kind: fetchProp,
		box: &box{
			key: key,
		},
		respChan: make(chan *envelope, 1),
	}
	resp, err := s.command(command)
	if err != nil {
		return nil, err
	}
	return resp.box.prop, nil
}

// Dispose is specified on the Scene interface.
func (s *scene) Dispose(key string) (interface{}, error) {
	command := &envelope{
		kind: disposeProp,
		box: &box{
			key: key,
		},
		respChan: make(chan *envelope, 1),
	}
	resp, err := s.command(command)
	if err != nil {
		return nil, err
	}
	return resp.box.prop, nil
}

// Flag is specified on the Scene interface.
func (s *scene) Flag(topic string) error {
	command := &envelope{
		kind: flag,
		signaling: &signaling{
			topic: topic,
		},
		respChan: make(chan *envelope, 1),
	}
	_, err := s.command(command)
	return err
}

// Unflag is specified on the Scene interface.
func (s *scene) Unflag(topic string) error {
	command := &envelope{
		kind: unflag,
		signaling: &signaling{
			topic: topic,
		},
		respChan: make(chan *envelope, 1),
	}
	_, err := s.command(command)
	return err
}

// WaitFlag is specified on the Scene interface.
func (s *scene) WaitFlag(topic string) error {
	return s.WaitFlagLimited(topic, 0)
}

// WaitFlagAndFetch is specified on the Scene interface.
func (s *scene) WaitFlagAndFetch(topic string) (interface{}, error) {
	err := s.WaitFlag(topic)
	if err != nil {
		return nil, err
	}
	return s.Fetch(topic)
}

// WaitFlagLimited is specified on the Scene interface.
func (s *scene) WaitFlagLimited(topic string, timeout time.Duration) error {
	// Add signal channel.
	command := &envelope{
		kind: wait,
		signaling: &signaling{
			topic:      topic,
			signalChan: make(chan struct{}, 1),
		},
		respChan: make(chan *envelope, 1),
	}
	_, err := s.command(command)
	if err != nil {
		return err
	}
	// Wait for signal.
	var timeoutChan <-chan time.Time
	if timeout > 0 {
		timeoutChan = time.After(timeout)
	}
	select {
	case <-s.backend.IsStopping():
		err = s.Wait()
		if err == nil {
			err = errors.New(ErrSceneEnded, errorMessages)
		}
		return err
	case <-command.signaling.signalChan:
		return nil
	case <-timeoutChan:
		return errors.New(ErrWaitedTooLong, errorMessages, topic)
	}
}

// WaitFlagLimitedAndFetch is specified on the Scene interface.
func (s *scene) WaitFlagLimitedAndFetch(topic string, timeout time.Duration) (interface{}, error) {
	err := s.WaitFlagLimited(topic, timeout)
	if err != nil {
		return nil, err
	}
	return s.Fetch(topic)
}

// command sends a command envelope to the backend and
// waits for the response.
func (s *scene) command(command *envelope) (*envelope, error) {
	select {
	case s.commandChan <- command:
	case <-s.backend.IsStopping():
		err := s.Wait()
		if err == nil {
			err = errors.New(ErrSceneEnded, errorMessages)
		}
		return nil, err
	}
	select {
	case <-s.backend.IsStopping():
		err := s.Wait()
		if err == nil {
			err = errors.New(ErrSceneEnded, errorMessages)
		}
		return nil, err
	case resp := <-command.respChan:
		if resp.err != nil {
			return nil, resp.err
		}
		return resp, nil
	}
}

// backendLoop runs the backend loop of the scene.
func (s *scene) backendLoop(l loop.Loop) (err error) {
	// Defer cleanup.
	defer func() {
		cerr := s.cleanupAllProps()
		if err == nil {
			err = cerr
		}
	}()
	// Init timers.
	var watchdog <-chan time.Time
	var clapperboard <-chan time.Time
	if s.absolute > 0 {
		clapperboard = time.After(s.absolute)
	}
	// Run loop.
	for {
		if s.inactivity > 0 {
			watchdog = time.After(s.inactivity)
		}
		select {
		case <-l.ShallStop():
			return nil
		case timeout := <-watchdog:
			return errors.New(ErrTimeout, errorMessages, "inactivity", timeout)
		case timeout := <-clapperboard:
			return errors.New(ErrTimeout, errorMessages, "absolute", timeout)
		case command := <-s.commandChan:
			s.processCommand(command)
		}
	}
}

// processCommand processes the sent commands.
func (s *scene) processCommand(command *envelope) {
	switch command.kind {
	case storeProp:
		// Add a new prop.
		_, ok := s.props[command.box.key]
		if ok {
			command.err = errors.New(ErrPropAlreadyExist, errorMessages, command.box.key)
		} else {
			s.props[command.box.key] = command.box
		}
	case fetchProp:
		// Retrieve a prop.
		box, ok := s.props[command.box.key]
		if !ok {
			command.err = errors.New(ErrPropNotFound, errorMessages, command.box.key)
		} else {
			command.box = box
		}
	case disposeProp:
		// Remove a prop.
		box, ok := s.props[command.box.key]
		if !ok {
			command.err = errors.New(ErrPropNotFound, errorMessages, command.box.key)
		} else {
			delete(s.props, command.box.key)
			command.box = box
			if box.cleanup != nil {
				cerr := box.cleanup(box.key, box.prop)
				if cerr != nil {
					command.err = errors.Annotate(cerr, ErrCleanupFailed, errorMessages, box.key)
				}
			}
		}
	case flag:
		// Signal a topic.
		s.flags[command.signaling.topic] = true
		// Notify subscribers.
		subscribers, ok := s.signalings[command.signaling.topic]
		if ok {
			delete(s.signalings, command.signaling.topic)
			for _, subscriber := range subscribers {
				subscriber <- struct{}{}
			}
		}
	case unflag:
		// Drop a topic.
		delete(s.flags, command.signaling.topic)
	case wait:
		// Add a waiter for a topic.
		active := s.flags[command.signaling.topic]
		if active {
			command.signaling.signalChan <- struct{}{}
		} else {
			waiters := s.signalings[command.signaling.topic]
			s.signalings[command.signaling.topic] = append(waiters, command.signaling.signalChan)
		}
	default:
		panic("illegal command")
	}
	// Return the changed command as response.
	command.respChan <- command
}

// cleanupAllProps cleans all props.
func (s *scene) cleanupAllProps() error {
	for _, box := range s.props {
		if box.cleanup != nil {
			err := box.cleanup(box.key, box.prop)
			if err != nil {
				return errors.Annotate(err, ErrCleanupFailed, errorMessages, box.key)
			}
		}
	}
	return nil
}

// EOF
