// Tideland Go Library - Scene - Unit Tests
//
// Copyright (C) 2014-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package scene_test

//--------------------
// IMPORTS
//--------------------

import (
	"errors"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/scene"
)

//--------------------
// TESTS
//--------------------

// TestSimpleNoTimeout tests a simple scene usage without
// any timeout.
func TestSimpleNoTimeout(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	scn := scene.Start()

	id := scn.ID()
	assert.Length(id, 16)
	err := scn.Store("foo", 4711)
	assert.Nil(err)
	foo, err := scn.Fetch("foo")
	assert.Nil(err)
	assert.Equal(foo, 4711)
	_, err = scn.Fetch("bar")
	assert.True(scene.IsPropNotFoundError(err))
	err = scn.Store("foo", "bar")
	assert.True(scene.IsPropAlreadyExistError(err))
	_, err = scn.Dispose("bar")
	assert.True(scene.IsPropNotFoundError(err))
	foo, err = scn.Dispose("foo")
	assert.Nil(err)
	assert.Equal(foo, 4711)
	_, err = scn.Fetch("foo")
	assert.True(scene.IsPropNotFoundError(err))

	status, err := scn.Status()
	assert.Nil(err)
	assert.Equal(status, scene.Active)

	err = scn.Stop()
	assert.Nil(err)

	status, err = scn.Status()
	assert.Nil(err)
	assert.Equal(status, scene.Over)
}

// TestAccessAfterStopping tests an access after the
// scene already has been stopped.
func TestAccessAfterStopping(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	scn := scene.Start()

	err := scn.Store("foo", 4711)
	assert.Nil(err)
	foo, err := scn.Fetch("foo")
	assert.Nil(err)
	assert.Equal(foo, 4711)

	err = scn.Stop()
	assert.Nil(err)

	foo, err = scn.Fetch("foo")
	assert.True(scene.IsSceneEndedError(err))
	assert.Nil(foo)
}

// TestCleanupNoError tests the cleanup of props with
// no errors.
func TestCleanupNoError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	cleanups := make(map[string]interface{})
	cleanup := func(key string, prop interface{}) error {
		cleanups[key] = prop
		return nil
	}
	scn := scene.Start()

	err := scn.StoreClean("foo", 4711, cleanup)
	assert.Nil(err)
	err = scn.StoreClean("bar", "yadda", cleanup)
	assert.Nil(err)

	foo, err := scn.Dispose("foo")
	assert.Nil(err)
	assert.Equal(foo, 4711)

	err = scn.Stop()
	assert.Nil(err)

	assert.Length(cleanups, 2)
	assert.Equal(cleanups["foo"], 4711)
	assert.Equal(cleanups["bar"], "yadda")
}

// TestCleanupWithErrors tests the cleanup of props with errors.
func TestCleanupWithErrors(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	cleanup := func(key string, prop interface{}) error {
		return errors.New("ouch")
	}
	scn := scene.Start()

	err := scn.StoreClean("foo", 4711, cleanup)
	assert.Nil(err)
	err = scn.StoreClean("bar", true, cleanup)
	assert.Nil(err)
	err = scn.StoreClean("yadda", "OK", cleanup)
	assert.Nil(err)

	foo, err := scn.Dispose("foo")
	assert.True(scene.IsCleanupFailedError(err))
	assert.Nil(foo)
	bar, err := scn.Fetch("bar")
	assert.Nil(err)
	assert.Equal(bar, true)

	err = scn.Stop()
	assert.True(scene.IsCleanupFailedError(err))
}

// TestSimpleInactivityTimeout tests a simple scene usage
// with inactivity timeout.
func TestSimpleInactivityTimeout(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	scn := scene.StartLimited(100*time.Millisecond, 0)

	err := scn.Store("foo", 4711)
	assert.Nil(err)

	for i := 0; i < 5; i++ {
		foo, err := scn.Fetch("foo")
		assert.Nil(err)
		assert.Equal(foo, 4711)
		time.Sleep(50)
	}

	time.Sleep(100 * time.Millisecond)

	foo, err := scn.Fetch("foo")
	assert.True(scene.IsTimeoutError(err))
	assert.Nil(foo)

	status, err := scn.Status()
	assert.True(scene.IsTimeoutError(err))
	assert.Equal(status, scene.Over)

	err = scn.Stop()
	assert.True(scene.IsTimeoutError(err))
}

// TestSimpleAbsoluteTimeout tests a simple scene usage
// with absolute timeout.
func TestSimpleAbsoluteTimeout(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	scn := scene.StartLimited(0, 250*time.Millisecond)

	err := scn.Store("foo", 4711)
	assert.Nil(err)

	for {
		_, err = scn.Fetch("foo")
		if err != nil {
			assert.True(scene.IsTimeoutError(err))
			break
		}
		time.Sleep(50)
	}

	err = scn.Stop()
	assert.True(scene.IsTimeoutError(err))
}

// TestCleanupAfterTimeout tests the cleanup of props after
// a timeout.
func TestCleanupAfterTimeout(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	cleanups := make(map[string]interface{})
	cleanup := func(key string, prop interface{}) error {
		cleanups[key] = prop
		return nil
	}
	scn := scene.StartLimited(0, 100*time.Millisecond)

	err := scn.StoreClean("foo", 4711, cleanup)
	assert.Nil(err)
	err = scn.StoreClean("bar", "yadda", cleanup)
	assert.Nil(err)

	time.Sleep(250 * time.Millisecond)

	err = scn.Stop()
	assert.True(scene.IsTimeoutError(err))

	assert.Length(cleanups, 2)
	assert.Equal(cleanups["foo"], 4711)
	assert.Equal(cleanups["bar"], "yadda")
}

// TestAbort tests the aborting of a scene. A cleanup error
// will not be reported.
func TestAbort(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	cleanup := func(key string, prop interface{}) error {
		return errors.New("ouch")
	}
	scn := scene.Start()

	err := scn.StoreClean("foo", 4711, cleanup)
	assert.Nil(err)

	scn.Abort(errors.New("aborted"))

	foo, err := scn.Fetch("foo")
	assert.ErrorMatch(err, "aborted")
	assert.Nil(foo)

	err = scn.Stop()
	assert.ErrorMatch(err, "aborted")
}

// TestFlagNoTimeout tests the waiting for a signal without
// a timeout.
func TestFlagNoTimeout(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	scn := scene.Start()

	go func() {
		err := scn.WaitFlag("foo")
		assert.Nil(err)
		err = scn.Store("foo-a", true)
		assert.Nil(err)
	}()
	go func() {
		err := scn.WaitFlag("foo")
		assert.Nil(err)
		err = scn.Store("foo-b", true)
		assert.Nil(err)
	}()

	time.Sleep(100 * time.Millisecond)

	err := scn.Flag("foo")
	assert.Nil(err)

	time.Sleep(250 * time.Millisecond)

	fooA, err := scn.Fetch("foo-a")
	assert.Nil(err)
	assert.Equal(fooA, true)
	fooB, err := scn.Fetch("foo-b")
	assert.Nil(err)
	assert.Equal(fooB, true)

	err = scn.Stop()
	assert.Nil(err)
}

// TestNoFlagDueToStop tests the waiting for a signal while
// a scene is stopped.
func TestNoFlagDueToStop(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	scn := scene.Start()

	go func() {
		err := scn.WaitFlag("foo")
		assert.True(scene.IsSceneEndedError(err))
	}()
	go func() {
		err := scn.WaitFlag("foo")
		assert.True(scene.IsSceneEndedError(err))
	}()

	time.Sleep(100 * time.Millisecond)

	err := scn.Stop()
	assert.Nil(err)
}

// TestStoreAndFlag tests the signaling after storing
// after value.
func TestStoreAndFlag(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	scn := scene.Start()

	go func() {
		time.Sleep(100 * time.Millisecond)
		err := scn.StoreAndFlag("foo", 4711)
		assert.Nil(err)
	}()

	err := scn.WaitFlag("foo")
	assert.Nil(err)
	foo, err := scn.Fetch("foo")
	assert.Nil(err)
	assert.Equal(foo, 4711)

	err = scn.Stop()
	assert.Nil(err)
}

// TestEarlyFlag tests the signaling before a waiting.
func TestEarlyFlag(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	scn := scene.Start()
	err := scn.Flag("foo")
	assert.Nil(err)

	go func() {
		err := scn.WaitFlag("foo")
		assert.Nil(err)
		err = scn.Store("foo-a", true)
		assert.Nil(err)
	}()
	go func() {
		err := scn.WaitFlag("foo")
		assert.Nil(err)
		err = scn.Store("foo-b", true)
		assert.Nil(err)
	}()

	time.Sleep(100 * time.Millisecond)

	fooA, err := scn.Fetch("foo-a")
	assert.Nil(err)
	assert.Equal(fooA, true)
	fooB, err := scn.Fetch("foo-b")
	assert.Nil(err)
	assert.Equal(fooB, true)

	err = scn.Stop()
	assert.Nil(err)
}

// TestFlagTimeout tests the waiting for a signal with
// a timeout.
func TestFlagTimeout(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	doneC := audit.MakeSigChan()
	scn := scene.Start()

	go func() {
		err := scn.WaitFlag("foo")
		assert.Nil(err)
		doneC <- true
	}()
	go func() {
		err := scn.WaitFlagLimited("foo", 50*time.Millisecond)
		assert.True(scene.IsWaitedTooLongError(err))
		doneC <- true
	}()

	time.Sleep(100 * time.Millisecond)

	err := scn.Flag("foo")
	assert.Nil(err)

	assert.Wait(doneC, true, time.Second)
	assert.Wait(doneC, true, time.Second)

	err = scn.Stop()
	assert.Nil(err)
}

// TestFlagUnflag tests the removal of a flag.
func TestFlagUnflag(t *testing.T) {
	assert := audit.NewTestingAssertion(t, false)
	scn := scene.Start()

	err := scn.Flag("foo")
	assert.Nil(err)
	err = scn.Unflag("foo")
	assert.Nil(err)
	err = scn.WaitFlagLimited("foo", 50*time.Millisecond)
	assert.True(scene.IsWaitedTooLongError(err))

	err = scn.Stop()
	assert.Nil(err)
}

// EOF
