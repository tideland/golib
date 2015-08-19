// Tideland Go Library - Monitoring - Standard Backend
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

// standardMeasuring implements the Measuring interface.
type standardMeasuring struct {
	backend   *standardMonitoringBackend
	id        string
	startTime time.Time
	duration  time.Duration
}

// EndMEasuring implements the Measuring interface.
func (m *standardMeasuring) EndMeasuring() time.Duration {
	m.duration = time.Now().Sub(m.startTime)
	m.backend.measuringChan <- m
	return m.duration
}

//--------------------
// MEASURING POINT
//--------------------

// standardMeasuringPoint implements the MeasuringPoint interface.
type standardMeasuringPoint struct {
	id          string
	count       int64
	minDuration time.Duration
	maxDuration time.Duration
	avgDuration time.Duration
}

// newStandardMeasuringPoint creates a new measuring point out of a measuring.
func newStandardMeasuringPoint(m *standardMeasuring) *standardMeasuringPoint {
	return &standardMeasuringPoint{
		id:          m.id,
		count:       1,
		minDuration: m.duration,
		maxDuration: m.duration,
		avgDuration: m.duration,
	}
}

// Uupdate a measuring point with a measuring.
func (mp *standardMeasuringPoint) update(m *standardMeasuring) {
	average := mp.avgDuration.Nanoseconds()
	mp.Count++
	if mp.minDuration > m.duration {
		mp.minDuration = m.duration
	}
	if mp.maxDuration < m.duration {
		mp.maxDuration = m.duration
	}
	mp.avgDuration = time.Duration((average + m.duration.Nanoseconds()) / 2)
}

// String implements the Stringer interface.
func (mp *standardMeasuringPoint) String() string {
	return fmt.Sprintf("Measuring Point %q (%dx / min %s / max %s / avg %s)", mp.id, mp.count, mp.minDuration, mp.maxDuration, mp.avgDuration)
}

//--------------------
// STAY-SET VARIABLE
//--------------------

// standardSSVChange represents the change of a stay-set variable.
type standardSSVChange struct {
	id       string
	absolute bool
	variable int64
}

// standardStaySetVariable implements the StaySetVariable interface.
type standardStaySetVariable struct {
	id       string
	count    int64
	actValue int64
	minValue int64
	maxValue int64
	avgValue int64
	total    int64
}

// newStaySetVariable creates a new stay-set variable out of a variable.
func newStaySetVariable(v *standardSSVChange) *standardStaySetVariable {
	return &standardStaySetVariable{
		id:       v.id,
		count:    1,
		actValue: v.variable,
		minValue: v.variable,
		maxValue: v.variable,
		avgValue: v.variable,
	}
}

// update a stay-set variable with a change.
func (ssv *standardStaySetVariable) update(chg *standardSSVChange) {
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
func (ssv *standardStaySetVariable) String() string {
	return fmt.Sprintf("Stay-Set Variable %q (%dx / act %d / min %d / max %d / avg %d)",
		ssv.id, ssv.count, ssv.actValue, ssv.minValue, ssv.maxValue, ssv.avgValue)
}

//--------------------
// DYNAMIC STATUS RETRIEVER
//--------------------

// standardRetrieverRegistration allows the registration of a retriever function.
type standardRetrieverRegistration struct {
	id  string
	dsr DynamicStatusRetriever
}

// standardDynamicStatusValue implements the DynamicStatusValue interface.
type standardDynamicStatusValue struct {
	id    string
	value string
}

// String implements the Stringer interface.
func (dsv *standardDynamicStatusValue) String() string {
	return fmt.Sprintf("Dynamic Status Value %q (value = %q)", dsv.id, dsv.value)
}

//--------------------
// MONITORING BACKEND
//--------------------

// standardMonitoringBackend implements the MonitoringBackend interface.
type standardMonitoringBackend struct {
	etmData                   map[string]*MeasuringPoint
	ssvData                   map[string]*StaySetVariable
	dsrData                   map[string]DynamicStatusRetriever
	measuringChan             chan *Measuring
	ssvChangeChan             chan *ssvChange
	retrieverRegistrationChan chan *retrieverRegistration
	commandChan               chan *command
	backend                   loop.Loop
}

// NewStandardMonitoringBackend starts the standard monitoring backend.
func NewStandardMonitoringBackend() MonitoringBackend {
	m := &standardMonitoringBackend{
		measuringChan:             make(chan *Measuring, 1000),
		ssvChangeChan:             make(chan *ssvChange, 1000),
		retrieverRegistrationChan: make(chan *retrieverRegistration, 10),
		commandChan:               make(chan *command),
	}
	m.backend = loop.GoRecoverable(m.backendLoop, m.checkRecovering)
	return m
}

// BeginMeasuring implements the MonitorBackend interface.
func (b *standardMonitoringBackend) BeginMeasuring(id string) *Measuring {
	return &standardMeasuring{b, id, time.Now(), time.Now()}
}

// ReadMeasuringPoint implements the MonitorBackend interface.
func (b *standardMonitoringBackend) ReadMeasuringPoint(id string) (MeasuringPoint, error) {
	resp, err := b.command(cmdMeasuringPointRead, id)
	if err != nil {
		return nil, err
	}
	return resp.(MeasuringPoint), nil
}

// MeasuringPointsDo implements the MonitorBackend interface.
func (b *standardMonitoringBackend) MeasuringPointsDo(f func(MeasuringPoint)) error {
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
func (b *standardMonitoringBackend) SetVariable(id string, v int64) {
	b.ssvChangeChan <- &standardSSVChange{id, true, v}
}

// IncrVariable implements the MonitorBackend interface.
func (b *standardMonitoringBackend) IncrVariable(id string) {
	b.ssvChangeChan <- &standardSSVChange{id, false, 1}
}

// DecrVariable implements the MonitorBackend interface.
func (b *standardMonitoringBackend) DecrVariable(id string) {
	b.ssvChangeChan <- &standardSSVChange{id, false, -1}
}

// ReadVariable implements the MonitorBackend interface.
func (b *standardMonitoringBackend) ReadVariable(id string) (StaySetVariable, error) {
	resp, err := b.command(cmdStaySetVariableRead, id)
	if err != nil {
		return nil, err
	}
	return resp.(StaySetVariable), nil
}

// StaySetVariablesDo implements the MonitorBackend interface.
func (b *standardMonitoringBackend) StaySetVariablesDo(f func(StaySetVariable)) error {
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
func  (b *standardMonitoringBackend) Register(id string, rf DynamicStatusRetriever) {
	v.retrieverRegistrationChan <- &retrieverRegistration{id, rf}
}

// ReadStatus implements the MonitorBackend interface.
func (b *standardMonitoringBackend) ReadStatus(id string) (string, error) {
	resp, err := b.command(cmdDynamicStatusRetrieverRead, id)
	if err != nil {
		return "", err
	}
	return resp.(string), nil
}

// DynamicStatusValuesDo implements the MonitorBackend interface.
func (b *standardMonitoringBackend) DynamicStatusValuesDo(f func(DynamicStatusValue)) error {
	resp, err := monitor.command(cmdDynamicStatusRetrieversReadAll, f)
	if err != nil {
		return err
	}
	dsvs := resp.(DynamicStatusValues)
	for _, dsv := range dsvs {
		f(dsv)
	}
	return nil
}

// Reset implements the MonitorBackend interface.
func (b *standardMonitoringBackend) Reset() error {
	_, err := m.command(cmdReset, nil)
	if err != nil {
		return err
	}
	return nil
}

// command sends a command to the system monitor and waits for a response.
func (b *standardMonitoringBackend) command(opCode int, args interface{}) (interface{}, error) {
	cmd := &command{opCode, args, make(chan interface{})}
	b.commandChan <- cmd
	resp, ok := <-cmd.respChan
	if !ok {
		return nil, errors.New(ErrMonitorPanicked, errorMessages)
	}
	if err, ok := resp.(error); ok {
		return nil, err
	}
	return resp, nil
}

// init the system monitor.
func (b *standardMonitoringBackend) init() {
	b.etmData = make(map[string]*MeasuringPoint)
	b.ssvData = make(map[string]*StaySetVariable)
	b.dsrData = make(map[string]DynamicStatusRetriever)
}

// backendLoop runs the system monitor.
func (b *standardMonitoringBackend) backendLoop(l loop.Loop) error {
	// Init the monitor.
	b.init()
	// Run loop.
	for {
		select {
		case <-l.ShallStop():
			return nil
		case measuring := <-b.measuringChan:
			// Received a new measuring.
			if mp, ok := b.etmData[measuring.id]; ok {
				mp.update(measuring)
			} else {
				b.etmData[measuring.id] = newMeasuringPoint(measuring)
			}
		case ssvChange := <-m.ssvChangeChan:
			// Received a new change.
			if ssv, ok := m.ssvData[ssvChange.id]; ok {
				ssv.update(ssvChange)
			} else {
				b.ssvData[ssvChange.id] = newStaySetVariable(ssvChange)
			}
		case registration := <-m.retrieverRegistrationChan:
			// Received a new retriever for registration.
			b.dsrData[registration.id] = registration.dsr
		case cmd := <-b.commandChan:
			// Received a command to process.
			b.processCommand(cmd)
		}
	}
}

// processCommand handles the received commands of the monitor.
func (b *standardMonitoringBackend) processCommand(cmd *command) {
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
			dsv := &standardDynamicStatusValue{id, v}
			resp = append(resp, dsv)
		}
		sort.Sort(resp)
		cmd.respond(resp)
	}
}

// checkRecovering checks if the backend can be recovered.
func (b *standardMonitoringBackend) checkRecovering(rs loop.Recoverings) (loop.Recoverings, error) {
	if rs.Frequency(12, time.Minute) {
		logger.Errorf("standard monitor cannot be recovered: %v", rs.Last().Reason)
		return nil, errors.New(ErrMonitorCannotBeRecovered, errorMessages, rs.Last().Reason)
	}
	logger.Warningf("standard monitor recovered: %v", rs.Last().Reason)
	return rs.Trim(12), nil
}
