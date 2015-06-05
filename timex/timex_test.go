// Tideland Go Library - Time Extensions - Unit Tests
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package timex_test

//--------------------
// IMPORTS
//--------------------

import (
	"errors"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/timex"
)

//--------------------
// TESTS
//--------------------

// Test time containments.
func TestTimeContainments(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Create some test data.
	ts := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	years := []int{2008, 2009, 2010}
	months := []time.Month{10, 11, 12}
	days := []int{10, 11, 12, 13, 14}
	hours := []int{20, 21, 22, 23}
	minutes := []int{0, 5, 10, 15, 20, 25}
	seconds := []int{0, 15, 30, 45}
	weekdays := []time.Weekday{time.Monday, time.Tuesday, time.Wednesday}

	assert.True(timex.YearInList(ts, years), "Go time in year list.")
	assert.True(timex.YearInRange(ts, 2005, 2015), "Go time in year range.")
	assert.True(timex.MonthInList(ts, months), "Go time in month list.")
	assert.True(timex.MonthInRange(ts, 7, 12), "Go time in month range.")
	assert.True(timex.DayInList(ts, days), "Go time in day list.")
	assert.True(timex.DayInRange(ts, 5, 15), "Go time in day range .")
	assert.True(timex.HourInList(ts, hours), "Go time in hour list.")
	assert.True(timex.HourInRange(ts, 20, 31), "Go time in hour range .")
	assert.True(timex.MinuteInList(ts, minutes), "Go time in minute list.")
	assert.True(timex.MinuteInRange(ts, 0, 5), "Go time in minute range .")
	assert.True(timex.SecondInList(ts, seconds), "Go time in second list.")
	assert.True(timex.SecondInRange(ts, 0, 5), "Go time in second range .")
	assert.True(timex.WeekdayInList(ts, weekdays), "Go time in weekday list.")
	assert.True(timex.WeekdayInRange(ts, time.Monday, time.Friday), "Go time in weekday range .")
}

// Test crontab keeping the job.
func TestCrontabKeep(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Create test crontab with job.
	c := timex.NewCrontab(10 * time.Millisecond)
	j := &cronjob{0, false, false}

	c.Add("keep", j)
	time.Sleep(50 * time.Millisecond)
	c.Remove("keep")
	c.Stop()

	assert.Equal(j.counter, 3, "job counter increased twice")
}

// Test crontab removing the job.
func TestCrontabRemove(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Create test crontab with job.
	c := timex.NewCrontab(10 * time.Millisecond)
	j := &cronjob{0, false, false}

	c.Add("remove", j)
	time.Sleep(250 * time.Millisecond)
	c.Remove("remove")
	c.Stop()

	assert.Equal(j.counter, 10, "job counter increased max ten times")
}

// Test crontab removing the job after an error.
func TestCrontabError(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Create test crontab with job.
	c := timex.NewCrontab(10 * time.Millisecond)
	j := &cronjob{0, false, true}

	c.Add("remove", j)
	time.Sleep(250 * time.Millisecond)
	c.Remove("remove")
	c.Stop()

	assert.Equal(j.counter, 5, "job counter increased max five times")
}

//--------------------
// HELPERS
//--------------------

type cronjob struct {
	counter int
	flip    bool
	fail    bool
}

func (j *cronjob) ShallExecute(t time.Time) bool {
	j.flip = !j.flip
	return j.flip
}

func (j *cronjob) Execute() (bool, error) {
	j.counter++
	if j.fail && j.counter == 5 {
		return false, errors.New("failed")
	}
	if j.counter == 10 {
		return false, nil
	}
	return true, nil
}

// EOF
