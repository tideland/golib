// Tideland Go Library - Monitoring - Standard Backend
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package monitoring

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"sort"
	"time"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/logger"
	"github.com/tideland/golib/loop"
)

//--------------------
// CONSTANTS
//--------------------

const (
	cmdReset = iota
	cmdMeasuringPointRead
	cmdMeasuringPointsReadAll
	cmdStaySetVariableRead
	cmdStaySetVariablesReadAll
	cmdDynamicStatusRetrieverRead
	cmdDynamicStatusRetrieversReadAll
)

//--------------------
// COMMAND
//--------------------

// command encapsulated the data for any command.
type command struct {
	opCode   int
	args     interface{}
	respChan chan interface{}
}

// respond allow to simply respond to a command.
func (c *command) respond(v interface{}) {
	c.respChan <- v
}

// ok is a simple positive response.
func (c *command) ok() {
	c.respond(true)
}

// close closes the response channel.
func (c *command) close() {
	close(c.respChan)
}

//--------------------
// MEASURING
//--------------------

// stdMeasuring implements the Measuring interface.
type stdMeasuring struct {
	backend   *stdBackend
	id        string
	startTime time.Time
	duration  time.Duration
}

// EndMEasuring implements the Measuring interface.
func (m *stdMeasuring) EndMeasuring() time.Duration {
	if m.backend == nil {
		return 0
	}
	m.duration = time.Since(m.startTime)
	m.backend.measuringC <- m
	return m.duration
}

//--------------------
// MEASURING POINT
//--------------------

// stdMeasuringPoint implements the MeasuringPoint interface.
type stdMeasuringPoint struct {
	id          string
	count       int64
	minDuration time.Duration
	maxDuration time.Duration
	avgDuration time.Duration
}

// newStdMeasuringPoint creates a new measuring point out of a measuring.
func newStdMeasuringPoint(m *stdMeasuring) *stdMeasuringPoint {
	return &stdMeasuringPoint{
		id:          m.id,
		count:       1,
		minDuration: m.duration,
		maxDuration: m.duration,
		avgDuration: m.duration,
	}
}

// ID implements the MeasuringPoint interface.
func (mp *stdMeasuringPoint) ID() string { return mp.id }

// Count implements the MeasuringPoint interface.
func (mp *stdMeasuringPoint) Count() int64 { return mp.count }

// MinDuration implements the MeasuringPoint interface.
func (mp *stdMeasuringPoint) MinDuration() time.Duration { return mp.minDuration }

// MaxDuration implements the MeasuringPoint interface.
func (mp *stdMeasuringPoint) MaxDuration() time.Duration { return mp.maxDuration }

// AvgDuration implements the MeasuringPoint interface.
func (mp *stdMeasuringPoint) AvgDuration() time.Duration { return mp.avgDuration }

// Uupdate a measuring point with a measuring.
func (mp *stdMeasuringPoint) update(m *stdMeasuring) {
	average := mp.avgDuration.Nanoseconds()
	mp.count++
	if mp.minDuration > m.duration {
		mp.minDuration = m.duration
	}
	if mp.maxDuration < m.duration {
		mp.maxDuration = m.duration
	}
	mp.avgDuration = time.Duration((average + m.duration.Nanoseconds()) / 2)
}

// String implements the Stringer interface.
func (mp *stdMeasuringPoint) String() string {
	return fmt.Sprintf("Measuring Point %q (%dx / min %s / max %s / avg %s)", mp.id, mp.count, mp.minDuration, mp.maxDuration, mp.avgDuration)
}

//--------------------
// STAY-SET VARIABLE
//--------------------

// stdSSVChange represents the change of a stay-set variable.
type stdSSVChange struct {
	id       string
	absolute bool
	variable int64
}

// stdStaySetVariable implements the StaySetVariable interface.
type stdStaySetVariable struct {
	id       string
	count    int64
	actValue int64
	minValue int64
	maxValue int64
	avgValue int64
	total    int64
}

// newStdStaySetVariable creates a new stay-set variable out of a variable.
func newStdStaySetVariable(v *stdSSVChange) *stdStaySetVariable {
	return &stdStaySetVariable{
		id:       v.id,
		count:    1,
		actValue: v.variable,
		minValue: v.variable,
		maxValue: v.variable,
		avgValue: v.variable,
	}
}

// ID implements the StaySetVariable interface.
func (ssv *stdStaySetVariable) ID() string { return ssv.id }

// Count implements the StaySetVariable interface.
func (ssv *stdStaySetVariable) Count() int64 { return ssv.count }

// ActValue implements the StaySetVariable interface.
func (ssv *stdStaySetVariable) ActValue() int64 { return ssv.actValue }

// MinValue implements the StaySetVariable interface.
func (ssv *stdStaySetVariable) MinValue() int64 { return ssv.minValue }

// MaxValue implements the StaySetVariable interface.
func (ssv *stdStaySetVariable) MaxValue() int64 { return ssv.maxValue }

// MinValue implements the StaySetVariable interface.
func (ssv *stdStaySetVariable) AvgValue() int64 { return ssv.avgValue }

// update a stay-set variable with a change.
func (ssv *stdStaySetVariable) update(chg *stdSSVChange) {
	ssv.count++
	if chg.absolute {
		ssv.actValue = chg.variable
	} else {
		ssv.actValue += chg.variable
	}
	ssv.total += chg.variable
	if ssv.minValue > ssv.actValue {
		ssv.minValue = ssv.actValue
	}
	if ssv.maxValue < ssv.actValue {
		ssv.maxValue = ssv.actValue
	}
	ssv.avgValue = ssv.total / ssv.count
}

// String implements the Stringer interface.
func (ssv *stdStaySetVariable) String() string {
	return fmt.Sprintf("Stay-Set Variable %q (%dx / act %d / min %d / max %d / avg %d)",
		ssv.id, ssv.count, ssv.actValue, ssv.minValue, ssv.maxValue, ssv.avgValue)
}

//--------------------
// DYNAMIC STATUS RETRIEVER
//--------------------

// stdRetrieverRegistration allows the registration of a retriever function.
type stdRetrieverRegistration struct {
	id  string
	dsr DynamicStatusRetriever
}

// stdDynamicStatusValue implements the DynamicStatusValue interface.
type stdDynamicStatusValue struct {
	id    string
	value string
}

// ID implements the DynamicStatusValue interface.
func (dsv *stdDynamicStatusValue) ID() string { return dsv.id }

// Value implements the DynamicStatusValue interface.
func (dsv *stdDynamicStatusValue) Value() string { return dsv.value }

// String implements the Stringer interface.
func (dsv *stdDynamicStatusValue) String() string {
	return fmt.Sprintf("Dynamic Status Value %q (value = %q)", dsv.id, dsv.value)
}

//--------------------
// BACKEND
//--------------------

// stdBackend implements the Backend interface.
type stdBackend struct {
	etmFilter              IDFilter
	ssvFilter              IDFilter
	dsrFilter              IDFilter
	etmData                map[string]*stdMeasuringPoint
	ssvData                map[string]*stdStaySetVariable
	dsrData                map[string]DynamicStatusRetriever
	measuringC             chan *stdMeasuring
	ssvChangeC             chan *stdSSVChange
	retrieverRegistrationC chan *stdRetrieverRegistration
	commandC               chan *command
	backend                loop.Loop
}

// NewStandardBackend starts the standard monitoring backend.
func NewStandardBackend() Backend {
	m := &stdBackend{
		measuringC:             make(chan *stdMeasuring, 1024),
		ssvChangeC:             make(chan *stdSSVChange, 1024),
		retrieverRegistrationC: make(chan *stdRetrieverRegistration),
		commandC:               make(chan *command),
	}
	m.backend = loop.GoRecoverable(m.backendLoop, m.checkRecovering, "monitoring backend")
	return m
}

// BeginMeasuring implements the MonitorBackend interface.
func (b *stdBackend) BeginMeasuring(id string) Measuring {
	if b.etmFilter != nil && !b.etmFilter(id) {
		return &stdMeasuring{}
	}
	return &stdMeasuring{b, id, time.Now(), 0}
}

// ReadMeasuringPoint implements the MonitorBackend interface.
func (b *stdBackend) ReadMeasuringPoint(id string) (MeasuringPoint, error) {
	resp, err := b.command(cmdMeasuringPointRead, id)
	if err != nil {
		return nil, err
	}
	return resp.(MeasuringPoint), nil
}

// MeasuringPointsDo implements the MonitorBackend interface.
func (b *stdBackend) MeasuringPointsDo(f func(MeasuringPoint)) error {
	resp, err := b.command(cmdMeasuringPointsReadAll, nil)
	if err != nil {
		return err
	}
	mps := resp.(MeasuringPoints)
	for _, mp := range mps {
		f(mp)
	}
	return nil
}

// SetVariable implements the MonitorBackend interface.
func (b *stdBackend) SetVariable(id string, v int64) {
	if b.ssvFilter != nil && !b.ssvFilter(id) {
		return
	}
	b.ssvChangeC <- &stdSSVChange{id, true, v}
}

// IncrVariable implements the MonitorBackend interface.
func (b *stdBackend) IncrVariable(id string) {
	if b.ssvFilter != nil && !b.ssvFilter(id) {
		return
	}
	b.ssvChangeC <- &stdSSVChange{id, false, 1}
}

// DecrVariable implements the MonitorBackend interface.
func (b *stdBackend) DecrVariable(id string) {
	if b.ssvFilter != nil && !b.ssvFilter(id) {
		return
	}
	b.ssvChangeC <- &stdSSVChange{id, false, -1}
}

// ReadVariable implements the MonitorBackend interface.
func (b *stdBackend) ReadVariable(id string) (StaySetVariable, error) {
	resp, err := b.command(cmdStaySetVariableRead, id)
	if err != nil {
		return nil, err
	}
	return resp.(StaySetVariable), nil
}

// StaySetVariablesDo implements the MonitorBackend interface.
func (b *stdBackend) StaySetVariablesDo(f func(StaySetVariable)) error {
	resp, err := b.command(cmdStaySetVariablesReadAll, nil)
	if err != nil {
		return err
	}
	ssvs := resp.(StaySetVariables)
	for _, ssv := range ssvs {
		f(ssv)
	}
	return nil
}

// Register implements the MonitorBackend interface.
func (b *stdBackend) Register(id string, rf DynamicStatusRetriever) {
	if b.dsrFilter != nil && !b.dsrFilter(id) {
		return
	}
	b.retrieverRegistrationC <- &stdRetrieverRegistration{id, rf}
}

// ReadStatus implements the MonitorBackend interface.
func (b *stdBackend) ReadStatus(id string) (string, error) {
	resp, err := b.command(cmdDynamicStatusRetrieverRead, id)
	if err != nil {
		return "", err
	}
	return resp.(string), nil
}

// DynamicStatusValuesDo implements the MonitorBackend interface.
func (b *stdBackend) DynamicStatusValuesDo(f func(DynamicStatusValue)) error {
	resp, err := b.command(cmdDynamicStatusRetrieversReadAll, f)
	if err != nil {
		return err
	}
	dsvs := resp.(DynamicStatusValues)
	for _, dsv := range dsvs {
		f(dsv)
	}
	return nil
}

// SetMeasuringsFilter implements the MonitorBackend interface.
func (b *stdBackend) SetMeasuringsFilter(f IDFilter) IDFilter {
	old := b.etmFilter
	b.etmFilter = f
	return old
}

// SetVariablesFilter implements the MonitorBackend interface.
func (b *stdBackend) SetVariablesFilter(f IDFilter) IDFilter {
	old := b.ssvFilter
	b.ssvFilter = f
	return old
}

// SetRetrieversFilter implements the MonitorBackend interface.
func (b *stdBackend) SetRetrieversFilter(f IDFilter) IDFilter {
	old := b.dsrFilter
	b.dsrFilter = f
	return old
}

// Reset implements the MonitorBackend interface.
func (b *stdBackend) Reset() error {
	_, err := b.command(cmdReset, nil)
	if err != nil {
		return err
	}
	return nil
}

// Stop implements the MonitorBackend interface.
func (b *stdBackend) Stop() {
	b.backend.Stop()
}

// command sends a command to the system monitor and waits for a response.
func (b *stdBackend) command(opCode int, args interface{}) (interface{}, error) {
	cmd := &command{opCode, args, make(chan interface{})}
	b.commandC <- cmd
	resp, ok := <-cmd.respChan
	if !ok {
		return nil, errors.New(ErrMonitoringPanicked, errorMessages)
	}
	if err, ok := resp.(error); ok {
		return nil, err
	}
	return resp, nil
}

// init the system monitor.
func (b *stdBackend) init() {
	b.etmData = make(map[string]*stdMeasuringPoint)
	b.ssvData = make(map[string]*stdStaySetVariable)
	b.dsrData = make(map[string]DynamicStatusRetriever)
}

// backendLoop runs the system monitor.
func (b *stdBackend) backendLoop(l loop.Loop) error {
	// Init the monitor.
	b.init()
	// Run loop.
	for {
		select {
		case <-l.ShallStop():
			return nil
		case measuring := <-b.measuringC:
			// Received a new measuring.
			if mp, ok := b.etmData[measuring.id]; ok {
				mp.update(measuring)
			} else {
				b.etmData[measuring.id] = newStdMeasuringPoint(measuring)
			}
		case ssvChange := <-b.ssvChangeC:
			// Received a new change.
			if ssv, ok := b.ssvData[ssvChange.id]; ok {
				ssv.update(ssvChange)
			} else {
				b.ssvData[ssvChange.id] = newStdStaySetVariable(ssvChange)
			}
		case registration := <-b.retrieverRegistrationC:
			// Received a new retriever for registration.
			b.dsrData[registration.id] = registration.dsr
		case cmd := <-b.commandC:
			// Received a command to process.
			b.processCommand(cmd)
		}
	}
}

// processCommand handles the received commands of the monitor.
func (b *stdBackend) processCommand(cmd *command) {
	defer cmd.close()
	switch cmd.opCode {
	case cmdReset:
		// Reset monitoring.
		b.init()
		cmd.ok()
	case cmdMeasuringPointRead:
		// Read just one measuring point.
		id := cmd.args.(string)
		if mp, ok := b.etmData[id]; ok {
			// Measuring point found.
			clone := *mp
			cmd.respond(&clone)
		} else {
			// Measuring point does not exist.
			cmd.respond(errors.New(ErrMeasuringPointNotExists, errorMessages, id))
		}
	case cmdMeasuringPointsReadAll:
		// Read all measuring points.
		resp := MeasuringPoints{}
		for _, mp := range b.etmData {
			clone := *mp
			resp = append(resp, &clone)
		}
		sort.Sort(resp)
		cmd.respond(resp)
	case cmdStaySetVariableRead:
		// Read just one stay-set variable.
		id := cmd.args.(string)
		if ssv, ok := b.ssvData[id]; ok {
			// Variable found.
			clone := *ssv
			cmd.respond(&clone)
		} else {
			// Variable does not exist.
			cmd.respond(errors.New(ErrStaySetVariableNotExists, errorMessages, id))
		}
	case cmdStaySetVariablesReadAll:
		// Read all stay-set variables.
		resp := StaySetVariables{}
		for _, mp := range b.ssvData {
			clone := *mp
			resp = append(resp, &clone)
		}
		sort.Sort(resp)
		cmd.respond(resp)
	case cmdDynamicStatusRetrieverRead:
		// Read just one dynamic status value.
		id := cmd.args.(string)
		if dsr, ok := b.dsrData[id]; ok {
			// Dynamic status found.
			v, err := dsr()
			if err != nil {
				cmd.respond(err)
			} else {
				cmd.respond(v)
			}
		} else {
			// Dynamic status does not exist.
			cmd.respond(errors.New(ErrDynamicStatusNotExists, errorMessages, id))
		}
	case cmdDynamicStatusRetrieversReadAll:
		// Read all dynamic status values.
		resp := DynamicStatusValues{}
		for id, dsr := range b.dsrData {
			v, err := dsr()
			if err != nil {
				cmd.respond(err)
			}
			dsv := &stdDynamicStatusValue{id, v}
			resp = append(resp, dsv)
		}
		sort.Sort(resp)
		cmd.respond(resp)
	}
}

// checkRecovering checks if the backend can be recovered.
func (b *stdBackend) checkRecovering(rs loop.Recoverings) (loop.Recoverings, error) {
	if rs.Frequency(12, time.Minute) {
		logger.Errorf("standard monitor cannot be recovered: %v", rs.Last().Reason)
		return nil, errors.New(ErrMonitoringCannotBeRecovered, errorMessages, rs.Last().Reason)
	}
	logger.Warningf("standard monitor recovered: %v", rs.Last().Reason)
	return rs.Trim(12), nil
}

// EOF
