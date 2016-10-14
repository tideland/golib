// Tideland Go Library - Monitoring - Unit Tests
//
// Copyright (C) 2009-2016 Frank Mueller / Tideland / Oldenburg / Germany
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
func TestETMMonitor(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	monitoring.SetBackend(monitoring.NewStandardBackend())
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
	assert.ErrorMatch(err, `.* measuring point "foo" does not exist`)
	mp, err = monitoring.ReadMeasuringPoint("mp:task:5")
	assert.Nil(err, "No error expected.")
	assert.Equal(mp.ID(), "mp:task:5", "should get the right one")
	assert.True(mp.Count() > 0, "should be measured several times")
	assert.Match(mp.String(), `Measuring Point "mp:task:5" \(.*\)`, "string representation should look fine")
	monitoring.MeasuringPointsDo(func(mp monitoring.MeasuringPoint) {
		assert.Match(mp.ID(), "mp:task:[0-9]", "id has to match the pattern")
		assert.True(mp.MinDuration() <= mp.AvgDuration() && mp.AvgDuration() <= mp.MaxDuration(),
			"avg should be somewhere between min and max")
	})
}

// Test of the SSI monitor.
func TestSSIMonitor(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	monitoring.SetBackend(monitoring.NewStandardBackend())
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
	assert.ErrorMatch(err, `.* stay-set variable "foo" does not exist`)
	ssv, err = monitoring.ReadVariable("ssv:value:5")
	assert.Nil(err, "no error expected")
	assert.Equal(ssv.ID(), "ssv:value:5", "should get the right one")
	assert.True(ssv.Count() > 0, "should be set several times")
	assert.Match(ssv.String(), `Stay-Set Variable "ssv:value:5" (.*)`)
	monitoring.StaySetVariablesDo(func(ssv monitoring.StaySetVariable) {
		assert.Match(ssv.ID(), "ssv:value:[0-9]", "id has to match the pattern")
		assert.True(ssv.MinValue() <= ssv.AvgValue() && ssv.AvgValue() <= ssv.MaxValue(),
			"avg should be somewhere between min and max")
	})
}

// Test of the DSR monitor.
func TestDSRMonitor(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	monitoring.SetBackend(monitoring.NewStandardBackend())
	// Register monitoring funcs.
	monitoring.Register("dsr:a", func() (string, error) { return "A", nil })
	monitoring.Register("dsr:b", func() (string, error) { return "4711", nil })
	monitoring.Register("dsr:c", func() (string, error) { return "2012-02-15", nil })
	monitoring.Register("dsr:d", func() (string, error) { a := 1; a = a / (a - a); return fmt.Sprintf("%d", a), nil })
	// Need some time to let that backend catch up queued registerings.
	time.Sleep(time.Millisecond)
	// Asserts.
	dsv, err := monitoring.ReadStatus("foo")
	assert.ErrorMatch(err, `.* dynamic status "foo" does not exist`)
	dsv, err = monitoring.ReadStatus("dsr:b")
	assert.Nil(err, "no error expected")
	assert.Equal(dsv, "4711", "status value should be correct")
	dsv, err = monitoring.ReadStatus("dsr:d")
	assert.NotNil(err, "error should be returned")
	assert.ErrorMatch(err, `.* monitoring backend panicked`)
}

// TestStandardInternalPanic tests the clean handling of panics
// when retrieving a status with the standard backend.
func TestInternalPanic(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	monitoring.SetBackend(monitoring.NewStandardBackend())
	// Register monitoring func with panic.
	monitoring.Register("panic", func() (string, error) { panic("ouch"); return "panic", nil })
	// Need some time to let that backend catch up queued registering.
	time.Sleep(time.Millisecond)
	// Asserts.
	status, err := monitoring.ReadStatus("panic")
	assert.Empty(status, "no dynamic status value")
	assert.ErrorMatch(err, `.* monitoring backend panicked`)
}

// TestBackendSwitch tests the correct switching between backends.
func TestBackendSwitch(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	sleep := 10 * time.Millisecond
	// First standard.
	monitoring.SetBackend(monitoring.NewStandardBackend())
	monitoring.Measure("test-a", func() { time.Sleep(sleep) })
	time.Sleep(sleep)
	mp, err := monitoring.ReadMeasuringPoint("test-a")
	assert.Nil(err)
	assert.Equal(mp.Count(), int64(1))
	assert.True(sleep <= mp.AvgDuration() && mp.AvgDuration() <= 2*sleep)
	// Then null.
	monitoring.SetBackend(monitoring.NewNullBackend())
	monitoring.Measure("test", func() { time.Sleep(sleep) })
	mp, err = monitoring.ReadMeasuringPoint("test")
	assert.Nil(err)
	assert.Equal(mp.ID(), "null")
	assert.Equal(mp.Count(), int64(0))
	// Finally standard again.
	monitoring.SetBackend(monitoring.NewStandardBackend())
	monitoring.Measure("test-b", func() { time.Sleep(sleep) })
	time.Sleep(sleep)
	mp, err = monitoring.ReadMeasuringPoint("test-a")
	assert.ErrorMatch(err, `.* measuring point "test-a" does not exist`)
	mp, err = monitoring.ReadMeasuringPoint("test-b")
	assert.Nil(err)
	assert.Equal(mp.Count(), int64(1))
	assert.True(sleep <= mp.AvgDuration() && mp.AvgDuration() <= 2*sleep)
}

//--------------------
// BENCHMARKS
//--------------------

// BenchmarkStandardBackendETM checks the performance of the ETM of the
// standard backend.
func BenchmarkStandardBackendETM(b *testing.B) {
	monitoring.SetBackend(monitoring.NewStandardBackend())

	for i := 0; i < b.N; i++ {
		monitoring.Measure("test", func() {})
	}
}

// BenchmarkFilteredStandardBackendETM checks the performance of the ETM
// of the standard backend when filtered.
func BenchmarkFilteredStandardBackendETM(b *testing.B) {
	monitoring.SetBackend(monitoring.NewStandardBackend())
	monitoring.SetMeasuringsFilter(func(id string) bool {
		return id != "test"
	})
	defer monitoring.SetMeasuringsFilter(nil)

	for i := 0; i < b.N; i++ {
		monitoring.Measure("test", func() {})
	}
}

// BenchmarkNullBackendETM checks the performance of the ETM of the
// null backend.
func BenchmarkNullBackendETM(b *testing.B) {
	monitoring.SetBackend(monitoring.NewNullBackend())

	for i := 0; i < b.N; i++ {
		monitoring.Measure("test", func() {})
	}
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
