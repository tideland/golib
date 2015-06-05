// Tideland Go Library - Logger - No SysLogger
//
// Copyright (C) 2012-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// +build windows,plan9

package logger

//--------------------
// IMPORTS
//--------------------

import (
	"log"
)

//--------------------
// SYSLOGGER
//--------------------

// SysLogger uses the Go syslog package as logging backend. It does
// not work on Windows or Plan9. Here it uses the standard Go logger.
type SysLogger struct {
	tag string
}

// NewGoLogger returns a logger implementation using the
// Go syslog package.
func NewSysLogger(tag string) (Logger, error) {
	if len(tag) > 0 {
		tag = "(" + tag + ")"
	}
	return &SysLogger{tag}
}

// Debug logs a message at debug level.
func (sl *SysLogger) Debug(info, msg string) {
	log.Println("[DEBUG]", sl.tag, info, msg)
}

// Info logs a message at info level.
func (sl *SysLogger) Info(info, msg string) {
	log.Println("[INFO]", sl.tag, info, msg)
}

// Warning logs a message at warning level.
func (sl *SysLogger) Warning(info, msg string) {
	log.Println("[WARNING]", sl.tag, info, msg)
}

// Error logs a message at error level.
func (sl *SysLogger) Error(info, msg string) {
	log.Println("[ERROR]", sl.tag, info, msg)
}

// Critical logs a message at critical level.
func (sl *SysLogger) Critical(info, msg string) {
	log.Println("[CRITICAL]", sl.tag, info, msg)
}

// EOF
