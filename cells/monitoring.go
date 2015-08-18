// Tideland Go Library - Cells - Monitoring
//
// Copyright (C) 2010-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/golib/monitoring"
)

//--------------------
// MONITORING INTERFACES
//--------------------

// MonitoringMeasuring contains one measuring.
type MonitoringMeasuring interface {
	// End ends the measuring and passes its values
	// to the backend.
	End()
}

// Monitoring defines the interface to a pluggable monitoring
// system.
type Monitoring interface {
	// IncrVariable allows increase the counter identified
	// by the ID. One use case in Cells is the count of a
	// given cell.
	IncrVariable(id string)

	// DecrVariable decreases the counter identified by
	// the ID.
	DecrVariable(id string)

	// BeginMeasuring is intended to start a runtime measuring
	// of a code block. It returns the MonitoringMeasuring
	// containing all vital information of the measuring.
	// Calling End on it, e.g. via defer, stops the measuring.
	// One use case in Cells is the measuring of the processing
	// time of events.
	BeginMeasuring(id string) MonitoringMeasuring
}

//--------------------
// MONITORING IMPLEMENTATIONS
//--------------------

// nullMonitoringMeasuring does nothing.
type nullMonitoringMeasuring struct{}

// End implements MonitoringMeasuring.
func (mm *nullMonitoringMeasuring) End() {}

// nullMonitoring does nothing.
type nullMonitoring struct{}

// NewNullMonitoring returns a monitoring that does nothing and
// consumes almost no ressources.
func NewNullMonitoring() Monitoring {
	return &nullMonitoring{}
}

// IncrVariable implements Monitoring.
func (m *nullMonitoring) IncrVariable(id string) {}

// DecrVariable implements Monitoring.
func (m *nullMonitoring) DecrVariable(id string) {}

// BeginMeasuring implements Monitoring.
func (m *nullMonitoring) BeginMeasuring(id string) MonitoringMeasuring {
	return &nullMonitoringMeasuring{}
}

// standardMonitoringMeasuring uses the GoLib
// monitoring package.
type standardMonitoringMeasuring struct {
	measuring *monitoring.Measuring
}

// End implements MonitoringMeasuring.
func (mm *standardMonitoringMeasuring) End() {
	mm.measuring.EndMeasuring()
}

// standardMonitoring uses the GoLib monitoring.
type standardMonitoring struct{}

// NewStandardMonitoring returns a monitoring that uses
// the GoLib monitoring package.
func NewStandardMonitoring() Monitoring {
	return &standardMonitoring{}
}

// IncrVariable implements Monitoring.
func (m *standardMonitoring) IncrVariable(id string) {
	monitoring.IncrVariable(id)
}

// DecrVariable implements Monitoring.
func (m *standardMonitoring) DecrVariable(id string) {
	monitoring.DecrVariable(id)
}

// BeginMeasuring implements Monitoring.
func (m *standardMonitoring) BeginMeasuring(id string) MonitoringMeasuring {
	return &standardMonitoringMeasuring{
		measuring: monitoring.BeginMeasuring(id),
	}
}

// EOF
