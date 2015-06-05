// Tideland Go Library - Monitoring - Unit Tests
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package monitoring_test

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/monitoring"
)

//--------------------
// TESTS
//--------------------

// Test of the ETM monitor.
func TestEtmMonitor(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Generate measurings.
	for i := 0; i < 500; i++ {
		n := rand.Intn(10)
		id := fmt.Sprintf("mp:task:%d", n)
		m := monitoring.BeginMeasuring(id)
		work(n * 5000)
		m.EndMeasuring()
	}
	// Need some time to let that backend catch up queued mesurings.
	time.Sleep(time.Millisecond)
	// Asserts.
	mp, err := monitoring.ReadMeasuringPoint("foo")
	assert.ErrorMatch(err, `\[MONITORING:.*\] measuring point "foo" does not exist`, "reading non-existent measuring point")
	mp, err = monitoring.ReadMeasuringPoint("mp:task:5")
	assert.Nil(err, "No error expected.")
	assert.Equal(mp.Id, "mp:task:5", "should get the right one")
	assert.True(mp.Count > 0, "should be measured several times")
	assert.Match(mp.String(), `Measuring Point "mp:task:5" \(.*\)`, "string representation should look fine")
	monitoring.MeasuringPointsDo(func(mp *monitoring.MeasuringPoint) {
		assert.Match(mp.Id, "mp:task:[0-9]", "id has to match the pattern")
		assert.True(mp.MinDuration <= mp.AvgDuration && mp.AvgDuration <= mp.MaxDuration,
			"avg should be somewhere between min and max")
	})
}

// Test of the SSI monitor.
func TestSsiMonitor(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Generate values.
	for i := 0; i < 500; i++ {
		n := rand.Intn(10)
		id := fmt.Sprintf("ssv:value:%d", n)
		monitoring.SetVariable(id, rand.Int63n(2001)-1000)
	}
	// Need some time to let that backend catch up queued mesurings.
	time.Sleep(time.Millisecond)
	// Asserts.
	ssv, err := monitoring.ReadVariable("foo")
	assert.ErrorMatch(err, `\[MONITORING:.*\] stay-set variable "foo" does not exist`, "reading non-existent variable")
	ssv, err = monitoring.ReadVariable("ssv:value:5")
	assert.Nil(err, "no error expected")
	assert.Equal(ssv.Id, "ssv:value:5", "should get the right one")
	assert.True(ssv.Count > 0, "should be set several times")
	assert.Match(ssv.String(), `Stay-Set Variable "ssv:value:5" (.*)`, "string representation should look fine")
	monitoring.StaySetVariablesDo(func(ssv *monitoring.StaySetVariable) {
		assert.Match(ssv.Id, "ssv:value:[0-9]", "id has to match the pattern")
		assert.True(ssv.MinValue <= ssv.AvgValue && ssv.AvgValue <= ssv.MaxValue,
			"avg should be somewhere between min and max")
	})
}

// Test of the DSR monitor.
func TestDsrMonitor(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Register monitoring funcs.
	monitoring.Register("dsr:a", func() (string, error) { return "A", nil })
	monitoring.Register("dsr:b", func() (string, error) { return "4711", nil })
	monitoring.Register("dsr:c", func() (string, error) { return "2012-02-15", nil })
	monitoring.Register("dsr:d", func() (string, error) { a := 1; a = a / (a - a); return fmt.Sprintf("%d", a), nil })
	// Need some time to let that backend catch up queued registerings.
	time.Sleep(time.Millisecond)
	// Asserts.
	dsv, err := monitoring.ReadStatus("foo")
	assert.ErrorMatch(err, `\[MONITORING:.*\] dynamic status "foo" does not exist`, "reading non-existent status")
	dsv, err = monitoring.ReadStatus("dsr:b")
	assert.Nil(err, "no error expected")
	assert.Equal(dsv, "4711", "status value should be correct")
	dsv, err = monitoring.ReadStatus("dsr:d")
	assert.NotNil(err, "error should be returned")
	assert.ErrorMatch(err, `\[MONITORING:.*\] monitor backend panicked`, "error inside retrieval has to be catched")
}

// Test the behavior after an internal panic.
func TestInternalPanic(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Register monitoring func with panic.
	monitoring.Register("panic", func() (string, error) { panic("ouch"); return "panic", nil })
	// Need some time to let that backend catch up queued registering.
	time.Sleep(time.Millisecond)
	// Asserts.
	dsv, err := monitoring.ReadStatus("panic")
	assert.Empty(dsv, "no dynamic status value")
	assert.ErrorMatch(err, `\[MONITORING:.*\] monitor backend panicked`, "monitor restarted due to panic")
}

//--------------------
// HELPERS
//--------------------

// Do some work.
func work(n int) int {
	if n < 0 {
		return 0
	}
	return n * work(n-1)
}

// EOF
