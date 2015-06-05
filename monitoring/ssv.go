// Tideland Go Library - Monitoring - Stay-Set Variable
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
)

//--------------------
// CONSTANTS
//--------------------

const (
	ssvTLine  = "+----------------------------------------------------+-----------+---------------+---------------+---------------+---------------+\n"
	ssvHeader = "| Stay-Set Variable Name                             | Count     | Act Value     | Min Value     | Max Value     | Avg Value     |\n"
	ssvFormat = "| %-50s | %9d | %13d | %13d | %13d | %13d |\n"
	ssvString = "Stay-Set Variable %q (%dx / act %d / min %d / max %d / avg %d)"
)

//--------------------
// API
//--------------------

// SetVariable sets a value of a stay-set variable.
func SetVariable(id string, v int64) {
	monitor.ssvChangeChan <- &ssvChange{id, true, v}
}

// IncrVariable increases a variable.
func IncrVariable(id string) {
	monitor.ssvChangeChan <- &ssvChange{id, false, 1}
}

// DecrVariable decreases a variable.
func DecrVariable(id string) {
	monitor.ssvChangeChan <- &ssvChange{id, false, -1}
}

// ReadVariable returns the stay-set variable for an id.
func ReadVariable(id string) (*StaySetVariable, error) {
	resp, err := monitor.command(cmdStaySetVariableRead, id)
	if err != nil {
		return nil, err
	}
	return resp.(*StaySetVariable), nil
}

// StaySetVariablesDo performs the function f for all
// variables.
func StaySetVariablesDo(f func(*StaySetVariable)) error {
	resp, err := monitor.command(cmdStaySetVariablesReadAll, nil)
	if err != nil {
		return err
	}
	ssvs := resp.(StaySetVariables)
	for _, ssv := range ssvs {
		f(ssv)
	}
	return nil
}

// StaySetVariablesWrite prints the stay-set variables for which
// the passed function returns true to the passed writer.
func StaySetVariablesWrite(w io.Writer, ff func(*StaySetVariable) bool) error {
	fmt.Fprint(w, ssvTLine)
	fmt.Fprint(w, ssvHeader)
	fmt.Fprint(w, ssvTLine)
	if err := StaySetVariablesDo(func(ssv *StaySetVariable) {
		if ff(ssv) {
			fmt.Fprintf(w, ssvFormat, ssv.Id, ssv.Count, ssv.ActValue, ssv.MinValue, ssv.MaxValue, ssv.AvgValue)
		}
	}); err != nil {
		return err
	}
	fmt.Fprint(w, ssvTLine)
	return nil
}

// StaySetVariablesPrintAll prints all stay-set variables
// to STDOUT.
func StaySetVariablesPrintAll() error {
	return StaySetVariablesWrite(os.Stdout, func(ssv *StaySetVariable) bool { return true })
}

//--------------------
// HELPERS
//--------------------

// ssvChange represents the change of a stay-set variable.
type ssvChange struct {
	id       string
	absolute bool
	variable int64
}

// StaySetVariable contains the cumulated values
// for one stay-set variable.
type StaySetVariable struct {
	Id       string
	Count    int64
	ActValue int64
	MinValue int64
	MaxValue int64
	AvgValue int64
	total    int64
}

// newStaySetVariable creates a new stay-set variable out of a variable.
func newStaySetVariable(v *ssvChange) *StaySetVariable {
	ssv := &StaySetVariable{
		Id:       v.id,
		Count:    1,
		ActValue: v.variable,
		MinValue: v.variable,
		MaxValue: v.variable,
		AvgValue: v.variable,
	}
	return ssv
}

// update a stay-set variable with a variable.
func (ssv *StaySetVariable) update(v *ssvChange) {
	ssv.Count++
	if v.absolute {
		ssv.ActValue = v.variable
	} else {
		ssv.ActValue += v.variable
	}
	ssv.total += v.variable
	if ssv.MinValue > ssv.ActValue {
		ssv.MinValue = ssv.ActValue
	}
	if ssv.MaxValue < ssv.ActValue {
		ssv.MaxValue = ssv.ActValue
	}
	ssv.AvgValue = ssv.total / ssv.Count
}

// String implements the Stringer interface.
func (ssv StaySetVariable) String() string {
	return fmt.Sprintf(ssvString, ssv.Id, ssv.Count, ssv.ActValue, ssv.MinValue, ssv.MaxValue, ssv.AvgValue)
}

// StaySetVariables is a set of stay-set variables.
type StaySetVariables []*StaySetVariable

// Implement the sort interface.

func (s StaySetVariables) Len() int           { return len(s) }
func (s StaySetVariables) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s StaySetVariables) Less(i, j int) bool { return s[i].Id < s[j].Id }

//--------------------
// EOF
//--------------------
