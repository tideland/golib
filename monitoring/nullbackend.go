// Tideland Go Library - Monitoring - Null Backend
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
	"time"
)

//--------------------
// MEASURING
//--------------------

// nullMeasuring implements the Measuring interface.
type nullMeasuring struct{}

// EndMEasuring implements the Measuring interface.
func (m *nullMeasuring) EndMeasuring() time.Duration { return 0 }

//--------------------
// MEASURING POINT
//--------------------

// nullMeasuringPoint implements the MeasuringPoint interface.
type nullMeasuringPoint struct{}

// ID implements the MeasuringPoint interface.
func (mp *nullMeasuringPoint) ID() string { return "null" }

// Count implements the MeasuringPoint interface.
func (mp *nullMeasuringPoint) Count() int64 { return 0 }

// MinDuration implements the MeasuringPoint interface.
func (mp *nullMeasuringPoint) MinDuration() time.Duration { return 0 }

// MaxDuration implements the MeasuringPoint interface.
func (mp *nullMeasuringPoint) MaxDuration() time.Duration { return 0 }

// AvgDuration implements the MeasuringPoint interface.
func (mp *nullMeasuringPoint) AvgDuration() time.Duration { return 0 }

// String implements the Stringer interface.
func (mp *nullMeasuringPoint) String() string { return "Null Measuring Point" }

//--------------------
// STAY-SET VARIABLE
//--------------------

// nullStaySetVariable implements the StaySetVariable interface.
type nullStaySetVariable struct{}

// ID implements the StaySetVariable interface.
func (ssv *nullStaySetVariable) ID() string { return "null" }

// Count implements the StaySetVariable interface.
func (ssv *nullStaySetVariable) Count() int64 { return 0 }

// ActValue implements the StaySetVariable interface.
func (ssv *nullStaySetVariable) ActValue() int64 { return 0 }

// MinValue implements the StaySetVariable interface.
func (ssv *nullStaySetVariable) MinValue() int64 { return 0 }

// MaxValue implements the StaySetVariable interface.
func (ssv *nullStaySetVariable) MaxValue() int64 { return 0 }

// MinValue implements the StaySetVariable interface.
func (ssv *nullStaySetVariable) AvgValue() int64 { return 0 }

// String implements the Stringer interface.
func (ssv *nullStaySetVariable) String() string { return "Null Stay-Set Variable" }

//--------------------
// DYNAMIC STATUS RETRIEVER
//--------------------

// nullDynamicStatusValue implements the DynamicStatusValue interface.
type nullDynamicStatusValue struct{}

// ID implements the DynamicStatusValue interface.
func (dsv *nullDynamicStatusValue) ID() string { return "null" }

// Value implements the DynamicStatusValue interface.
func (dsv *nullDynamicStatusValue) Value() string { return "" }

// String implements the Stringer interface.
func (dsv *nullDynamicStatusValue) String() string { return "Null Dynamic Status Value" }

//--------------------
// MONITORING BACKEND
//--------------------

// nullBackend implements the Backend interface.
type nullBackend struct{}

// NewNullBackend starts the null monitoring backend doing nothing.
func NewNullBackend() Backend { return &nullBackend{} }

// BeginMeasuring implements the MonitorBackend interface.
func (b *nullBackend) BeginMeasuring(id string) Measuring { return &nullMeasuring{} }

// ReadMeasuringPoint implements the MonitorBackend interface.
func (b *nullBackend) ReadMeasuringPoint(id string) (MeasuringPoint, error) {
	return &nullMeasuringPoint{}, nil
}

// MeasuringPointsDo implements the MonitorBackend interface.
func (b *nullBackend) MeasuringPointsDo(f func(MeasuringPoint)) error { return nil }

// SetVariable implements the MonitorBackend interface.
func (b *nullBackend) SetVariable(id string, v int64) {}

// IncrVariable implements the MonitorBackend interface.
func (b *nullBackend) IncrVariable(id string) {}

// DecrVariable implements the MonitorBackend interface.
func (b *nullBackend) DecrVariable(id string) {}

// ReadVariable implements the MonitorBackend interface.
func (b *nullBackend) ReadVariable(id string) (StaySetVariable, error) {
	return &nullStaySetVariable{}, nil
}

// StaySetVariablesDo implements the MonitorBackend interface.
func (b *nullBackend) StaySetVariablesDo(f func(StaySetVariable)) error { return nil }

// Register implements the MonitorBackend interface.
func (b *nullBackend) Register(id string, rf DynamicStatusRetriever) {}

// ReadStatus implements the MonitorBackend interface.
func (b *nullBackend) ReadStatus(id string) (string, error) { return "", nil }

// DynamicStatusValuesDo implements the MonitorBackend interface.
func (b *nullBackend) DynamicStatusValuesDo(f func(DynamicStatusValue)) error { return nil }

// SetMeasuringsFilter implements the MonitorBackend interface.
func (b *nullBackend) SetMeasuringsFilter(f IDFilter) IDFilter { return nil }

// SetVariablesFilter implements the MonitorBackend interface.
func (b *nullBackend) SetVariablesFilter(f IDFilter) IDFilter { return nil }

// SetRetrieversFilter implements the MonitorBackend interface.
func (b *nullBackend) SetRetrieversFilter(f IDFilter) IDFilter { return nil }

// Reset implements the MonitorBackend interface.
func (b *nullBackend) Reset() error { return nil }

// Stop implements the MonitorBackend interface.
func (b *nullBackend) Stop() {}

// EOF
