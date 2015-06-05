// Tideland Go Library - Monitoring - Dynamic Status Retriever
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
	dsrTLine  = "+----------------------------------------------------+---------------------------------------------------------------------------+\n"
	dsrHeader = "| Dynamic Status                                     | Value                                                                     |\n"
	dsrFormat = "| %-50s | %-73s |\n"
)

//--------------------
// API
//--------------------

// Register registers a new dynamic status retriever function.
func Register(id string, rf DynamicStatusRetriever) {
	monitor.retrieverRegistrationChan <- &retrieverRegistration{id, rf}
}

// ReadStatus returns the dynamic status for an id.
func ReadStatus(id string) (string, error) {
	resp, err := monitor.command(cmdDynamicStatusRetrieverRead, id)
	if err != nil {
		return "", err
	}
	return resp.(string), nil
}

// DynamicStatusValuesDo performs the function f for all
// status values.
func DynamicStatusValuesDo(f func(*DynamicStatusValue)) error {
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

// DynamicStatusValuesWrite prints the status values for which
// the passed function returns true to the passed writer.
func DynamicStatusValuesWrite(w io.Writer, ff func(*DynamicStatusValue) bool) error {
	fmt.Fprint(w, dsrTLine)
	fmt.Fprint(w, dsrHeader)
	fmt.Fprint(w, dsrTLine)
	if err := DynamicStatusValuesDo(func(dsv *DynamicStatusValue) {
		if ff(dsv) {
			fmt.Fprintf(w, dsrFormat, dsv.Id, dsv.Value)
		}
	}); err != nil {
		return err
	}
	fmt.Fprint(w, dsrTLine)
	return nil
}

// DynamicStatusValuesPrintAll prints all status values to STDOUT.
func DynamicStatusValuesPrintAll() error {
	return DynamicStatusValuesWrite(os.Stdout, func(v *DynamicStatusValue) bool { return true })
}

//--------------------
// HELPER
//--------------------

// DynamicStatusRetriever is called by the server and
// returns a current status as string.
type DynamicStatusRetriever func() (string, error)

// retrieverRegistration allows the registration of a retriever function.
type retrieverRegistration struct {
	id  string
	dsr DynamicStatusRetriever
}

// DynamicStatusValue contains one retrieved value.
type DynamicStatusValue struct {
	Id    string
	Value string
}

// DynamicStatusValues is a set of dynamic status values.
type DynamicStatusValues []*DynamicStatusValue

// Implement the sort interface.

func (d DynamicStatusValues) Len() int           { return len(d) }
func (d DynamicStatusValues) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d DynamicStatusValues) Less(i, j int) bool { return d[i].Id < d[j].Id }

//--------------------
// EOF
//--------------------
