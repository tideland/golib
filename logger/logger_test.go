// Tideland Go Library - Logger - Unit Tests
//
// Copyright (C) 2012-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package logger_test

//--------------------
// IMPORTS
//--------------------

import (
	"log"
	"os"
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/logger"
)

//--------------------
// TESTS
//--------------------

// Test log level.
func TestGetSetLevel(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	logger.SetLevel(logger.LevelDebug)
	assert.Equal(logger.Level(), logger.LevelDebug)
	logger.SetLevel(logger.LevelCritical)
	assert.Equal(logger.Level(), logger.LevelCritical)
	logger.SetLevel(logger.LevelDebug)
	assert.Equal(logger.Level(), logger.LevelDebug)
}

// Test log level filtering.
func TestLogLevelFiltering(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	ownLogger := &testLogger{}
	logger.SetLogger(ownLogger)
	logger.SetLevel(logger.LevelDebug)
	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")
	assert.Length(ownLogger.logs, 5)

	ownLogger = &testLogger{}
	logger.SetLogger(ownLogger)
	logger.SetLevel(logger.LevelError)
	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")
	assert.Length(ownLogger.logs, 2)
}

// Test logging with the go logger.
func TestGoLogger(t *testing.T) {
	log.SetOutput(os.Stdout)

	logger.SetLevel(logger.LevelDebug)
	logger.SetLogger(logger.NewGoLogger())

	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")
}

// Test logging with the syslogger.
func TestSysLogger(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)

	logger.SetLevel(logger.LevelDebug)

	sl, err := logger.NewSysLogger("GOAS")
	assert.Nil(err)
	logger.SetLogger(sl)

	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")
}

// Test logging with an own logger.
func TestOwnLogger(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	ownLogger := &testLogger{}

	logger.SetLevel(logger.LevelDebug)
	logger.SetLogger(ownLogger)

	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")

	assert.Length(ownLogger.logs, 5)
}

// TestFatalExit tests the call of the fatal exiter after a
// fatal error log.
func TestFatalExit(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	ownLogger := &testLogger{}
	exited := false
	fatalExiter := func() {
		exited = true
	}

	logger.SetLogger(ownLogger)
	logger.SetFatalExiter(fatalExiter)

	logger.Fatalf("fatal")
	assert.Length(ownLogger.logs, 1)
	assert.True(exited)
}

//--------------------
// LOGGER
//--------------------

type testLogger struct {
	logs []string
}

func (tl *testLogger) Debug(info, msg string) {
	tl.logs = append(tl.logs, "[DEBUG] "+info+" "+msg)
}

func (tl *testLogger) Info(info, msg string) {
	tl.logs = append(tl.logs, "[INFO] "+info+" "+msg)
}
func (tl *testLogger) Warning(info, msg string) {
	tl.logs = append(tl.logs, "[WARNING] "+info+" "+msg)
}
func (tl *testLogger) Error(info, msg string) {
	tl.logs = append(tl.logs, "[ERROR] "+info+" "+msg)
}
func (tl *testLogger) Critical(info, msg string) {
	tl.logs = append(tl.logs, "[CRITICAL] "+info+" "+msg)
}

func (tl *testLogger) Fatal(info, msg string) {
	tl.logs = append(tl.logs, "[FATAL] "+info+" "+msg)
}

// EOF
