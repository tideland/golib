// Tideland Go Library - Monitoring
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
// SYSTEM MONITOR
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

// systemMonitor contains all monitored informations.
type systemMonitor struct {
	etmData                   map[string]*MeasuringPoint
	ssvData                   map[string]*StaySetVariable
	dsrData                   map[string]DynamicStatusRetriever
	measuringChan             chan *Measuring
	ssvChangeChan             chan *ssvChange
	retrieverRegistrationChan chan *retrieverRegistration
	commandChan               chan *command
	backend                   loop.Loop
}

// newSystemMonitor starts the system monitor.
func newSystemMonitor() *systemMonitor {
	m := &systemMonitor{
		measuringChan:             make(chan *Measuring, 1000),
		ssvChangeChan:             make(chan *ssvChange, 1000),
		retrieverRegistrationChan: make(chan *retrieverRegistration, 10),
		commandChan:               make(chan *command),
	}
	m.backend = loop.GoRecoverable(m.backendLoop, m.checkRecovering)
	return m
}

// command sends a command to the system monitor and waits for a response.
func (m *systemMonitor) command(opCode int, args interface{}) (interface{}, error) {
	cmd := &command{opCode, args, make(chan interface{})}
	m.commandChan <- cmd
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
func (m *systemMonitor) init() {
	m.etmData = make(map[string]*MeasuringPoint)
	m.ssvData = make(map[string]*StaySetVariable)
	m.dsrData = make(map[string]DynamicStatusRetriever)
}

// backendLoop runs the system monitor.
func (m *systemMonitor) backendLoop(l loop.Loop) error {
	// Init the monitor.
	m.init()
	// Run loop.
	for {
		select {
		case <-l.ShallStop():
			return nil
		case measuring := <-m.measuringChan:
			// Received a new measuring.
			if mp, ok := m.etmData[measuring.id]; ok {
				mp.update(measuring)
			} else {
				m.etmData[measuring.id] = newMeasuringPoint(measuring)
			}
		case ssvChange := <-m.ssvChangeChan:
			// Received a new change.
			if ssv, ok := m.ssvData[ssvChange.id]; ok {
				ssv.update(ssvChange)
			} else {
				m.ssvData[ssvChange.id] = newStaySetVariable(ssvChange)
			}
		case registration := <-m.retrieverRegistrationChan:
			// Received a new retriever for registration.
			m.dsrData[registration.id] = registration.dsr
		case cmd := <-m.commandChan:
			// Received a command to process.
			m.processCommand(cmd)
		}
	}
}

// processCommand handles the received commands of the monitor.
func (m *systemMonitor) processCommand(cmd *command) {
	defer cmd.close()
	switch cmd.opCode {
	case cmdReset:
		// Reset monitoring.
		m.init()
		cmd.ok()
	case cmdMeasuringPointRead:
		// Read just one measuring point.
		id := cmd.args.(string)
		if mp, ok := m.etmData[id]; ok {
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
		for _, mp := range m.etmData {
			clone := *mp
			resp = append(resp, &clone)
		}
		sort.Sort(resp)
		cmd.respond(resp)
	case cmdStaySetVariableRead:
		// Read just one stay-set variable.
		id := cmd.args.(string)
		if ssv, ok := m.ssvData[id]; ok {
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
		for _, mp := range m.ssvData {
			clone := *mp
			resp = append(resp, &clone)
		}
		sort.Sort(resp)
		cmd.respond(resp)
	case cmdDynamicStatusRetrieverRead:
		// Read just one dynamic status value.
		id := cmd.args.(string)
		if dsr, ok := m.dsrData[id]; ok {
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
		for id, dsr := range m.dsrData {
			v, err := dsr()
			if err != nil {
				cmd.respond(err)
			}
			dsv := &DynamicStatusValue{id, v}
			resp = append(resp, dsv)
		}
		sort.Sort(resp)
		cmd.respond(resp)
	}
}

// checkRecovering checks if the backend can be recovered.
func (m *systemMonitor) checkRecovering(rs loop.Recoverings) (loop.Recoverings, error) {
	if rs.Frequency(12, time.Minute) {
		logger.Errorf("monitor cannot be recovered: %v", rs.Last().Reason)
		return nil, errors.New(ErrMonitorCannotBeRecovered, errorMessages, rs.Last().Reason)
	}
	logger.Warningf("monitor recovered: %v", rs.Last().Reason)
	return rs.Trim(12), nil
}

//--------------------
// GLOBAL MONITORING API
//--------------------

// monitor is the one global monitor instance.
var monitor *systemMonitor = newSystemMonitor()

// Reset clears all monitored values.
func Reset() error {
	_, err := monitor.command(cmdReset, nil)
	if err != nil {
		return err
	}
	return nil
}

// EOF
