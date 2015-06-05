// Tideland Go Library - Logger
//
// Copyright (C) 2012-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package logger

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

//--------------------
// LOG LEVEL
//--------------------

// LogLevel describes the chosen log level between
// debug and critical.
type LogLevel int

// Log levels to control the logging output.
const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelCritical
)

// logLevel is the global log level.
var logLevel LogLevel = LevelInfo

// Level returns the current log level.
func Level() LogLevel {
	return logLevel
}

// SetLevel switches to a new log level.
func SetLevel(level LogLevel) {
	switch {
	case level <= LevelDebug:
		logLevel = LevelDebug
	case level >= LevelCritical:
		logLevel = LevelCritical
	default:
		logLevel = level
	}
}

//--------------------
// LOGGING
//--------------------

// Debugf logs a message at debug level.
func Debugf(format string, args ...interface{}) {
	if logLevel <= LevelDebug {
		ci := retrieveCallInfo()
		fi := fmt.Sprintf(format, args...)

		logBackend.Debug(ci.verboseFormat(), fi)
	}
}

// Infof logs a message at info level.
func Infof(format string, args ...interface{}) {
	if logLevel <= LevelInfo {
		ci := retrieveCallInfo()
		fi := fmt.Sprintf(format, args...)

		logBackend.Info(ci.shortFormat(), fi)
	}
}

// Warningf logs a message at warning level.
func Warningf(format string, args ...interface{}) {
	if logLevel <= LevelWarning {
		ci := retrieveCallInfo()
		fi := fmt.Sprintf(format, args...)

		logBackend.Warning(ci.shortFormat(), fi)
	}
}

// Errorf logs a message at error level.
func Errorf(format string, args ...interface{}) {
	if logLevel <= LevelError {
		ci := retrieveCallInfo()
		fi := fmt.Sprintf(format, args...)

		logBackend.Error(ci.shortFormat(), fi)
	}
}

// Criticalf logs a message at critical level.
func Criticalf(format string, args ...interface{}) {
	ci := retrieveCallInfo()
	fi := fmt.Sprintf(format, args...)

	logBackend.Critical(ci.verboseFormat(), fi)
}

//--------------------
// LOGGER
//--------------------

// Logger is the interface for different logger backends.
type Logger interface {
	// Debug logs a message at debug level.
	Debug(info, msg string)
	// Info logs a message at info level.
	Info(info, msg string)
	// Warning logs a message at warning level.
	Warning(info, msg string)
	// Error logs a message at error level.
	Error(info, msg string)
	// Critical logs a message at critical level.
	Critical(info, msg string)
}

// logger references the used application logger.
var logBackend Logger = NewStandardLogger(os.Stdout)

// SetLogger sets a new logger.
func SetLogger(l Logger) {
	logBackend = l
}

// timeFormat controls how the timestamp of the standard logger is printed.
const timeFormat = "2006-01-02 15:04:05 Z07:00"

// StandardLogger is a simple logger writing to the given writer. It
// doesn't handle the levels differently.
type StandardLogger struct {
	mutex sync.Mutex
	out   io.Writer
}

// NewStandardLogger creates the standard logger.
func NewStandardLogger(out io.Writer) Logger {
	return &StandardLogger{out: out}
}

// Debug logs a message at debug level.
func (sl *StandardLogger) Debug(info, msg string) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	io.WriteString(sl.out, time.Now().Format(timeFormat))
	io.WriteString(sl.out, " [DEBUG] ")
	io.WriteString(sl.out, info)
	io.WriteString(sl.out, " ")
	io.WriteString(sl.out, msg)
	io.WriteString(sl.out, "\n")
}

// Info logs a message at info level.
func (sl *StandardLogger) Info(info, msg string) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	io.WriteString(sl.out, time.Now().Format(timeFormat))
	io.WriteString(sl.out, " [INFO] ")
	io.WriteString(sl.out, info)
	io.WriteString(sl.out, " ")
	io.WriteString(sl.out, msg)
	io.WriteString(sl.out, "\n")
}

// Warning logs a message at warning level.
func (sl *StandardLogger) Warning(info, msg string) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	io.WriteString(sl.out, time.Now().Format(timeFormat))
	io.WriteString(sl.out, " [WARNING] ")
	io.WriteString(sl.out, info)
	io.WriteString(sl.out, " ")
	io.WriteString(sl.out, msg)
	io.WriteString(sl.out, "\n")
}

// Error logs a message at error level.
func (sl *StandardLogger) Error(info, msg string) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	io.WriteString(sl.out, time.Now().Format(timeFormat))
	io.WriteString(sl.out, " [ERROR] ")
	io.WriteString(sl.out, info)
	io.WriteString(sl.out, " ")
	io.WriteString(sl.out, msg)
	io.WriteString(sl.out, "\n")
}

// Critical logs a message at critical level.
func (sl *StandardLogger) Critical(info, msg string) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	io.WriteString(sl.out, time.Now().Format(timeFormat))
	io.WriteString(sl.out, " [CRITICAL] ")
	io.WriteString(sl.out, info)
	io.WriteString(sl.out, " ")
	io.WriteString(sl.out, msg)
	io.WriteString(sl.out, "\n")
}

// GoLogger just uses the standard go log package.
type GoLogger struct{}

// NewGoLogger returns a logger implementation using the
// Go log package.
func NewGoLogger() Logger {
	return &GoLogger{}
}

// Debug logs a message at debug level.
func (gl *GoLogger) Debug(info, msg string) {
	log.Println("[DEBUG]", info, msg)
}

// Info logs a message at info level.
func (gl *GoLogger) Info(info, msg string) {
	log.Println("[INFO]", info, msg)
}

// Warning logs a message at warning level.
func (gl *GoLogger) Warning(info, msg string) {
	log.Println("[WARNING]", info, msg)
}

// Error logs a message at error level.
func (gl *GoLogger) Error(info, msg string) {
	log.Println("[ERROR]", info, msg)
}

// Critical logs a message at critical level.
func (gl *GoLogger) Critical(info, msg string) {
	log.Println("[CRITICAL]", info, msg)
}

//--------------------
// HELPER
//--------------------

// callInfo bundles the info about the call environment
// when a logging statement occured.
type callInfo struct {
	packageName string
	fileName    string
	funcName    string
	line        int
}

// shortFormat returns a string representation in a short variant.
func (ci *callInfo) shortFormat() string {
	return fmt.Sprintf("[%s]", ci.packageName)
}

// verboseFormat returns a string representation in a more verbose variant.
func (ci *callInfo) verboseFormat() string {
	return fmt.Sprintf("[%s] (%s:%s:%d)", ci.packageName, ci.fileName, ci.funcName, ci.line)
}

// retrieveCallInfo
func retrieveCallInfo() *callInfo {
	pc, file, line, _ := runtime.Caller(2)
	_, fileName := path.Split(file)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	packageName := ""
	funcName := parts[pl-1]

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
		packageName = strings.Join(parts[0:pl-2], ".")
	} else {
		packageName = strings.Join(parts[0:pl-1], ".")
	}

	return &callInfo{
		packageName: packageName,
		fileName:    fileName,
		funcName:    funcName,
		line:        line,
	}
}

// EOF
