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
)

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
// RETRY
//--------------------

// RetryStrategy describes how often the function in Retry is executed, the
// initial break between those retries, how much this time is incremented
// for each retry, and the maximum timeout.
type RetryStrategy struct {
	Count          int
	Break          time.Duration
	BreakIncrement time.Duration
	Timeout        time.Duration
}

// ShortAttempt returns a predefined short retry strategy.
func ShortAttempt() RetryStrategy {
	return RetryStrategy{
		Count:          10,
		Break:          50 * time.Millisecond,
		BreakIncrement: 0,
		Timeout:        5 * time.Second,
	}
}

// MediumAttempt returns a predefined medium retry strategy.
func MediumAttempt() RetryStrategy {
	return RetryStrategy{
		Count:          50,
		Break:          10 * time.Millisecond,
		BreakIncrement: 10 * time.Millisecond,
		Timeout:        30 * time.Second,
	}
}

// LongAttempt returns a predefined long retry strategy.
func LongAttempt() RetryStrategy {
	return RetryStrategy{
		Count:          100,
		Break:          10 * time.Millisecond,
		BreakIncrement: 25 * time.Millisecond,
		Timeout:        5 * time.Minute,
	}
}

// Retry executes the passed function until it returns true or an error.
// These retries are restricted by the retry strategy.
func Retry(f func() (bool, error), rs RetryStrategy) error {
	timeout := time.Now().Add(rs.Timeout)
	sleep := rs.Break
	for i := 0; i < rs.Count; i++ {
		done, err := f()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		if time.Now().After(timeout) {
			return errors.New(ErrRetriedTooLong, errorMessages, rs.Timeout)
		}
		time.Sleep(sleep)
		sleep += rs.BreakIncrement
	}
	return errors.New(ErrRetriedTooOften, errorMessages, rs.Count)
}

// EOF
