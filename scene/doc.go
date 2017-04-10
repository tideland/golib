// Tideland Go Library - Scene
//
// Copyright (C) 2014-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// Package scene of the Tideland Go Library provides a shared access to common
// used data in a larger context.
//
// By definition a scene is a sequence of continuous action in a play,
// movie, opera, or book. Applications do know these kind of scenes too,
// especially in concurrent software. Here aspects of the action have to
// passed between the actors in a secure way, very often they are interwoven
// and depending.
//
// Here the scene package helps. Beside a simple atomic way to store and
// fetch information together with optional cleanup functions it handles
// inactivity and absolute timeouts.
//
// A scene without timeouts is started with
//
//    scn := scene.Start()
//
// Now props can be stored, fetched, and disposed.
//
//    err := scn.Store("foo", myFoo)
//    foo, err := scn.Fetch("foo")
//    foo, err := scn.Dispose("foo")
//
// It's also possible to cleanup if a prop is disposed or the whole
// scene is stopped or aborted.
//
//    myCleanup := func(key string, prop interface{}) error {
//        // Cleanup, e.g. return the prop into a pool
//        // or close handles.
//        ...
//        return nil
//    }
//    err := scn.StoreClean("foo", myFoo, myCleanup)
//
// The cleanup is called individually per prop when disposing it, when the
// scene ends due to a timeout, or when it is stopped with
//
//    err := scn.Stop()
//
// or
//
//    scn.Abort(myError)
//
// Another functionality of the scene is the signaling of a topic. So
// multiple goroutines can wait for a signal with a topic, all will be
// notified after the topic has been signaled. Additionally they can wait
// with a timeout.
//
//    go func() {
//        err := scn.WaitFlag("foo")
//        ...
//    }()
//    go func() {
//        err := scn.WaitFlagLimited("foo", 5 * time.Second)
//        ...
//    }()
//    err := scn.Flag("foo")
//
// In case a flag is already signaled wait immediatily returns. Store()
// and Flag() can also be combined to StoreAndFlag(). This way the key
// will be used as flag topic and a waiter knows that the information is
// available.
//
// A scene knows two different timeouts. The first is the time of inactivity,
// the second is the absolute maximum time of a scene.
//
//    inactivityTimeout := 5 * time.Minutes
//    absoluteTimeout := 60 * time.Minutes
//    scn := scene.StartLimited(inactivityTimeout, absoluteTimeout)
//
// Now the scene is stopped after 5 minutes without any access or at the
// latest 60 minutes after the start. Both value may be zero if not needed.
// So scene.StartLimited(0, 0) is the same as scene.Start().
package scene

// EOF
