// Tideland Go Library - Time Extensions
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package timex

//--------------------
// IMPORTS
//--------------------

import (
	"time"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/logger"
	"github.com/tideland/golib/loop"
)

//--------------------
// CONSTANTS
//--------------------

const (
	ErrCrontabCannotBeRecovered = iota + 1
)

var errorMessages = errors.Messages{
	ErrCrontabCannotBeRecovered: "crontab cannot be recovered: %v",
}

//--------------------
// RANGES
//--------------------

// YearInList test if the year of a time is in a given list.
func YearInList(time time.Time, years []int) bool {
	for _, year := range years {
		if time.Year() == year {
			return true
		}
	}

	return false
}

// YearInRange tests if a year of a time is in a given range.
func YearInRange(time time.Time, minYear, maxYear int) bool {
	return (minYear <= time.Year()) && (time.Year() <= maxYear)
}

// MonthInList tests if the month of a time is in a given list.
func MonthInList(time time.Time, months []time.Month) bool {
	for _, month := range months {
		if time.Month() == month {
			return true
		}
	}
	return false
}

// MonthInRange tests if a month of a time is in a given range.
func MonthInRange(time time.Time, minMonth, maxMonth time.Month) bool {
	return (minMonth <= time.Month()) && (time.Month() <= maxMonth)
}

// DayInList tests if the day of a time is in a given list.
func DayInList(time time.Time, days []int) bool {
	for _, day := range days {
		if time.Day() == day {
			return true
		}
	}
	return false
}

// DayInRange tests if a day of a time is in a given range.
func DayInRange(time time.Time, minDay, maxDay int) bool {
	return (minDay <= time.Day()) && (time.Day() <= maxDay)
}

// HourInList tests if the hour of a time is in a given list.
func HourInList(time time.Time, hours []int) bool {
	for _, hour := range hours {
		if time.Hour() == hour {
			return true
		}
	}
	return false
}

// HourInRange tests if a hour of a time is in a given range.
func HourInRange(time time.Time, minHour, maxHour int) bool {
	return (minHour <= time.Hour()) && (time.Hour() <= maxHour)
}

// MinuteInList tests if the minute of a time is in a given list.
func MinuteInList(time time.Time, minutes []int) bool {
	for _, minute := range minutes {
		if time.Minute() == minute {
			return true
		}
	}
	return false
}

// MinuteInRange tests if a minute of a time is in a given range.
func MinuteInRange(time time.Time, minMinute, maxMinute int) bool {
	return (minMinute <= time.Minute()) && (time.Minute() <= maxMinute)
}

// SecondInList tests if the second of a time is in a given list.
func SecondInList(time time.Time, seconds []int) bool {
	for _, second := range seconds {
		if time.Second() == second {
			return true
		}
	}
	return false
}

// SecondInRange tests if a second of a time is in a given range.
func SecondInRange(time time.Time, minSecond, maxSecond int) bool {
	return (minSecond <= time.Second()) && (time.Second() <= maxSecond)
}

// WeekdayInList tests if the weekday of a time is in a given list.
func WeekdayInList(time time.Time, weekdays []time.Weekday) bool {
	for _, weekday := range weekdays {
		if time.Weekday() == weekday {
			return true
		}
	}
	return false
}

// WeekdayInRange tests if a weekday of a time is in a given range.
func WeekdayInRange(time time.Time, minWeekday, maxWeekday time.Weekday) bool {
	return (minWeekday <= time.Weekday()) && (time.Weekday() <= maxWeekday)
}

//--------------------
// CRONTAB
//--------------------

// Job is executed by the crontab.
type Job interface {
	// ShallExecute decides when called if the job
	// shal be executed.
	ShallExecute(t time.Time) bool

	// Execute executes the job. If the method returns
	// false or an error it will be removed.
	Execute() (bool, error)
}

// cronCommand operates on a crontab.
type command struct {
	add bool
	id  string
	job Job
}

// Crontab is one cron server. A system can run multiple in
// parallel.
type Crontab struct {
	jobs        map[string]Job
	commandChan chan *command
	ticker      *time.Ticker
	loop        loop.Loop
}

// NewCrontab creates a cron server.
func NewCrontab(freq time.Duration) *Crontab {
	c := &Crontab{
		jobs:        make(map[string]Job),
		commandChan: make(chan *command),
		ticker:      time.NewTicker(freq),
	}
	c.loop = loop.GoRecoverable(c.backendLoop, c.checkRecovering)
	return c
}

// Stop terminates the cron server.
func (c *Crontab) Stop() error {
	return c.loop.Stop()
}

// Add adds a new job to the server.
func (c *Crontab) Add(id string, job Job) {
	c.commandChan <- &command{true, id, job}
}

// Remove removes a job from the server.
func (c *Crontab) Remove(id string) {
	c.commandChan <- &command{false, id, nil}
}

// backendLoop runs the server backend.
func (c *Crontab) backendLoop(l loop.Loop) error {
	for {
		select {
		case <-l.ShallStop():
			return nil
		case cmd := <-c.commandChan:
			if cmd.add {
				c.jobs[cmd.id] = cmd.job
			} else {
				delete(c.jobs, cmd.id)
			}
		case now := <-c.ticker.C:
			for id, job := range c.jobs {
				c.do(id, job, now)
			}
		}
	}
}

// checkRecovering checks if the backend can be recovered.
func (c *Crontab) checkRecovering(rs loop.Recoverings) (loop.Recoverings, error) {
	if rs.Frequency(12, time.Minute) {
		logger.Errorf("crontab cannot be recovered: %v", rs.Last().Reason)
		return nil, errors.New(ErrCrontabCannotBeRecovered, errorMessages, rs.Last().Reason)
	}
	logger.Warningf("crontab recovered: %v", rs.Last().Reason)
	return rs.Trim(12), nil
}

// do checks and performs a job.
func (c *Crontab) do(id string, job Job, now time.Time) {
	if job.ShallExecute(now) {
		go func() {
			cont, err := job.Execute()
			if err != nil {
				logger.Errorf("job %q removed after error: %v", id, err)
				cont = false
			}
			if !cont {
				c.Remove(id)
			}
		}()
	}
}

// EOF
