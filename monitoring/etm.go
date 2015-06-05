// Tideland Go Library - Monitoring - Execution Time Measuring
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package monitoring

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"io"
	"os"
	"time"
)

//--------------------
// CONSTANTS
//--------------------

const (
	etmTLine  = "+----------------------------------------------------+------------+--------------------+--------------------+--------------------+\n"
	etmHeader = "| Measuring Point Name                               | Count      | Min Dur            | Max Dur            | Avg Dur            |\n"
	etmFormat = "| %-50s | %10d | %18s | %18s | %18s |\n"
	etmString = "Measuring Point %q (%dx / min %s / max %s / avg %s)"
)

//--------------------
// API
//--------------------

// BeginMeasuring starts a new measuring with a given id.
// All measurings with the same id will be aggregated.
func BeginMeasuring(id string) *Measuring {
	return &Measuring{id, time.Now(), time.Now()}
}

// Measure the execution of a function.
func Measure(id string, f func()) time.Duration {
	m := BeginMeasuring(id)
	f()
	return m.EndMeasuring()
}

// ReadMeasuringPoint returns the measuring point for an id.
func ReadMeasuringPoint(id string) (*MeasuringPoint, error) {
	resp, err := monitor.command(cmdMeasuringPointRead, id)
	if err != nil {
		return nil, err
	}
	return resp.(*MeasuringPoint), nil
}

// MeasuringPointsDo performs the function f for
// all measuring points.
func MeasuringPointsDo(f func(*MeasuringPoint)) error {
	resp, err := monitor.command(cmdMeasuringPointsReadAll, nil)
	if err != nil {
		return err
	}
	mps := resp.(MeasuringPoints)
	for _, mp := range mps {
		f(mp)
	}
	return nil
}

// MeasuringPointsWrite prints the measuring points for which
// the passed function returns true to the passed writer.
func MeasuringPointsWrite(w io.Writer, ff func(*MeasuringPoint) bool) error {
	fmt.Fprint(w, etmTLine)
	fmt.Fprint(w, etmHeader)
	fmt.Fprint(w, etmTLine)
	if err := MeasuringPointsDo(func(mp *MeasuringPoint) {
		if ff(mp) {
			fmt.Fprintf(w, etmFormat, mp.Id, mp.Count, mp.MinDuration, mp.MaxDuration, mp.AvgDuration)
		}
	}); err != nil {
		return err
	}
	fmt.Fprint(w, etmTLine)
	return nil
}

// MeasuringPointsPrintAll prints all measuring points
// to STDOUT.
func MeasuringPointsPrintAll() error {
	return MeasuringPointsWrite(os.Stdout, func(mp *MeasuringPoint) bool { return true })
}

//--------------------
// HELPERS
//--------------------

// Measuring contains one measuring.
type Measuring struct {
	id        string
	startTime time.Time
	endTime   time.Time
}

// EndMEasuring ends a measuring and passes it to the
// measuring server in the background.
func (m *Measuring) EndMeasuring() time.Duration {
	m.endTime = time.Now()
	monitor.measuringChan <- m
	return m.endTime.Sub(m.startTime)
}

// MeasuringPoint contains the cumulated measuring
// data of one measuring point.
type MeasuringPoint struct {
	Id          string
	Count       int64
	MinDuration time.Duration
	MaxDuration time.Duration
	AvgDuration time.Duration
}

// newMeasuringPoint creates a new measuring point out of a measuring.
func newMeasuringPoint(m *Measuring) *MeasuringPoint {
	duration := m.endTime.Sub(m.startTime)
	mp := &MeasuringPoint{
		Id:          m.id,
		Count:       1,
		MinDuration: duration,
		MaxDuration: duration,
		AvgDuration: duration,
	}
	return mp
}

// Uupdate a measuring point with a measuring.
func (mp *MeasuringPoint) update(m *Measuring) {
	duration := m.endTime.Sub(m.startTime)
	average := mp.AvgDuration.Nanoseconds()
	mp.Count++
	if mp.MinDuration > duration {
		mp.MinDuration = duration
	}
	if mp.MaxDuration < duration {
		mp.MaxDuration = duration
	}
	mp.AvgDuration = time.Duration((average + duration.Nanoseconds()) / 2)
}

// String implements the Stringer interface.
func (mp MeasuringPoint) String() string {
	return fmt.Sprintf(etmString, mp.Id, mp.Count, mp.MinDuration, mp.MaxDuration, mp.AvgDuration)
}

// MeasuringPoints is a set of measuring points.
type MeasuringPoints []*MeasuringPoint

// Implement the sort interface.

func (m MeasuringPoints) Len() int           { return len(m) }
func (m MeasuringPoints) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m MeasuringPoints) Less(i, j int) bool { return m[i].Id < m[j].Id }

//--------------------
// EOF
//--------------------
