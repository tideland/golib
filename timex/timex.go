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
)

//--------------------
// RANGES
//--------------------

// YearInList test if the year of a time is in a given list.
func YearInList(t time.Time, years []int) bool {
	for _, year := range years {
		if t.Year() == year {
			return true
		}
	}

	return false
}

// YearInRange tests if a year of a time is in a given range.
func YearInRange(t time.Time, minYear, maxYear int) bool {
	return (minYear <= t.Year()) && (t.Year() <= maxYear)
}

// MonthInList tests if the month of a time is in a given list.
func MonthInList(t time.Time, months []time.Month) bool {
	for _, month := range months {
		if t.Month() == month {
			return true
		}
	}
	return false
}

// MonthInRange tests if a month of a time is in a given range.
func MonthInRange(t time.Time, minMonth, maxMonth time.Month) bool {
	return (minMonth <= t.Month()) && (t.Month() <= maxMonth)
}

// DayInList tests if the day of a time is in a given list.
func DayInList(t time.Time, days []int) bool {
	for _, day := range days {
		if t.Day() == day {
			return true
		}
	}
	return false
}

// DayInRange tests if a day of a time is in a given range.
func DayInRange(t time.Time, minDay, maxDay int) bool {
	return (minDay <= t.Day()) && (t.Day() <= maxDay)
}

// HourInList tests if the hour of a time is in a given list.
func HourInList(t time.Time, hours []int) bool {
	for _, hour := range hours {
		if t.Hour() == hour {
			return true
		}
	}
	return false
}

// HourInRange tests if a hour of a time is in a given range.
func HourInRange(t time.Time, minHour, maxHour int) bool {
	return (minHour <= t.Hour()) && (t.Hour() <= maxHour)
}

// MinuteInList tests if the minute of a time is in a given list.
func MinuteInList(t time.Time, minutes []int) bool {
	for _, minute := range minutes {
		if t.Minute() == minute {
			return true
		}
	}
	return false
}

// MinuteInRange tests if a minute of a time is in a given range.
func MinuteInRange(t time.Time, minMinute, maxMinute int) bool {
	return (minMinute <= t.Minute()) && (t.Minute() <= maxMinute)
}

// SecondInList tests if the second of a time is in a given list.
func SecondInList(t time.Time, seconds []int) bool {
	for _, second := range seconds {
		if t.Second() == second {
			return true
		}
	}
	return false
}

// SecondInRange tests if a second of a time is in a given range.
func SecondInRange(t time.Time, minSecond, maxSecond int) bool {
	return (minSecond <= t.Second()) && (t.Second() <= maxSecond)
}

// WeekdayInList tests if the weekday of a time is in a given list.
func WeekdayInList(t time.Time, weekdays []time.Weekday) bool {
	for _, weekday := range weekdays {
		if t.Weekday() == weekday {
			return true
		}
	}
	return false
}

// WeekdayInRange tests if a weekday of a time is in a given range.
func WeekdayInRange(t time.Time, minWeekday, maxWeekday time.Weekday) bool {
	return (minWeekday <= t.Weekday()) && (t.Weekday() <= maxWeekday)
}

//--------------------
// BEGIN / END
//--------------------

// UnitOfTime describes whose begin/end is wanted.
type UnitOfTime int

const (
	Second UnitOfTime = iota + 1
	Minute
	Hour
	Day
	Month
	Year
)

// BeginOf returns the begin of the passed unit for the given time.
func BeginOf(t time.Time, unit UnitOfTime) time.Time {
	// Retrieve the individual parts of the given time.
	year := t.Year()
	month := t.Month()
	day := t.Day()
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()
	loc := t.Location()
	// Build new time.
	switch unit {
	case Second:
		return time.Date(year, month, day, hour, minute, second, 0, loc)
	case Minute:
		return time.Date(year, month, day, hour, minute, 0, 0, loc)
	case Hour:
		return time.Date(year, month, day, hour, 0, 0, 0, loc)
	case Day:
		return time.Date(year, month, day, 0, 0, 0, 0, loc)
	case Month:
		return time.Date(year, month, 1, 0, 0, 0, 0, loc)
	case Year:
		return time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	default:
		return t
	}
}

// EndOf returns the end of the passed unit for the given time.
func EndOf(t time.Time, unit UnitOfTime) time.Time {
	// Retrieve the individual parts of the given time.
	year := t.Year()
	month := t.Month()
	day := t.Day()
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()
	loc := t.Location()
	// Build new time.
	switch unit {
	case Second:
		return time.Date(year, month, day, hour, minute, second, 999999999, loc)
	case Minute:
		return time.Date(year, month, day, hour, minute, 59, 999999999, loc)
	case Hour:
		return time.Date(year, month, day, hour, 59, 59, 999999999, loc)
	case Day:
		return time.Date(year, month, day, 23, 59, 59, 999999999, loc)
	case Month:
		// Catching leap years makes the month a bit more complex.
		_, nextMonth, _ := t.AddDate(0, 1, 0).Date()
		return time.Date(year, nextMonth, 1, 23, 59, 59, 999999999, loc).AddDate(0, 0, -1)
	case Year:
		return time.Date(year, time.December, 31, 23, 59, 59, 999999999, loc)
	default:
		return t
	}
}

// EOF
