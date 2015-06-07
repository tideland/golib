// Tideland Go Library - Redis Client - Connection
//
// Copyright (C) 2009-2015 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package redis

//--------------------
// IMPORTS
//--------------------

import (
	"strings"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/identifier"
	"github.com/tideland/golib/monitoring"
)

//--------------------
// CONNECTION
//--------------------

// Connection manages one connection to a Redis database.
type Connection struct {
	database *Database
	resp     *resp
}

// newConnection creates a new connection instance.
func newConnection(db *Database) (*Connection, error) {
	conn := &Connection{
		database: db,
	}
	err := conn.ensureProtocol()
	if err != nil {
		return nil, err
	}
	// Perform authentication and database selection.
	err = conn.resp.authenticate()
	if err != nil {
		conn.database.pool.kill(conn.resp)
		return nil, err
	}
	err = conn.resp.selectDatabase()
	if err != nil {
		conn.database.pool.kill(conn.resp)
		return nil, err
	}
	return conn, nil
}

// Do executes one Redis command and returns
// the result as result set.
func (conn *Connection) Do(cmd string, args ...interface{}) (*ResultSet, error) {
	cmd = strings.ToLower(cmd)
	if strings.Contains(cmd, "subscribe") {
		return nil, errors.New(ErrUseSubscription, errorMessages)
	}
	err := conn.ensureProtocol()
	if err != nil {
		return nil, err
	}
	if conn.database.monitoring {
		m := monitoring.BeginMeasuring(identifier.Identifier("redis", "command", cmd))
		defer m.EndMeasuring()
	}
	err = conn.resp.sendCommand(cmd, args...)
	logCommand(cmd, args, err, conn.database.logging)
	if err != nil {
		return nil, err
	}
	result, err := conn.resp.receiveResultSet()
	return result, err
}

// DoValue executes one Redis command and returns a single value.
func (conn *Connection) DoValue(cmd string, args ...interface{}) (Value, error) {
	result, err := conn.Do(cmd, args...)
	if err != nil {
		return nil, err
	}
	return result.ValueAt(0)
}

// DoOK executes one Redis command and checks if
// it returns the OK string.
func (conn *Connection) DoOK(cmd string, args ...interface{}) (bool, error) {
	value, err := conn.DoValue(cmd, args...)
	if err != nil {
		return false, err
	}
	return value.IsOK(), nil
}

// DoBool executes one Redis command and interpretes
// the result as bool value.
func (conn *Connection) DoBool(cmd string, args ...interface{}) (bool, error) {
	result, err := conn.Do(cmd, args...)
	if err != nil {
		return false, err
	}
	return result.BoolAt(0)
}

// DoInt executes one Redis command and interpretes
// the result as int value.
func (conn *Connection) DoInt(cmd string, args ...interface{}) (int, error) {
	result, err := conn.Do(cmd, args...)
	if err != nil {
		return 0, err
	}
	return result.IntAt(0)
}

// DoString executes one Redis command and interpretes
// the result as string value.
func (conn *Connection) DoString(cmd string, args ...interface{}) (string, error) {
	result, err := conn.Do(cmd, args...)
	if err != nil {
		return "", err
	}
	return result.StringAt(0)
}

// DoStrings executes one Redis command and interpretes
// the result as a slice of strings.
func (conn *Connection) DoStrings(cmd string, args ...interface{}) ([]string, error) {
	result, err := conn.Do(cmd, args...)
	if err != nil {
		return nil, err
	}
	return result.Strings(), nil
}

// DoKeyValues executes on Redis command and interpretes
// the result as a list of keys and values.
func (conn *Connection) DoKeyValues(cmd string, args ...interface{}) (KeyValues, error) {
	result, err := conn.Do(cmd, args...)
	if err != nil {
		return nil, err
	}
	return result.KeyValues()
}

// DoHash executes on Redis command and interpretes
// the result as a hash.
func (conn *Connection) DoHash(cmd string, args ...interface{}) (Hash, error) {
	result, err := conn.Do(cmd, args...)
	if err != nil {
		return nil, err
	}
	return result.Hash()
}

// DoScoredValues executes on Redis command and interpretes
// the result as scored values.
func (conn *Connection) DoScoredValues(cmd string, args ...interface{}) (ScoredValues, error) {
	var withScores bool
	for _, arg := range args {
		if s, ok := arg.(string); ok {
			if strings.ToLower(s) == "withscores" {
				withScores = true
				break
			}
		}
	}
	result, err := conn.Do(cmd, args...)
	if err != nil {
		return nil, err
	}
	return result.ScoredValues(withScores)
}

// DoScan executes one Redis command which should be one of the
// scan commands. It returns the cursor and the result set containing
// the key, values or scored values depending on the scan command.
func (conn *Connection) DoScan(cmd string, args ...interface{}) (int, *ResultSet, error) {
	result, err := conn.Do(cmd, args...)
	if err != nil {
		return 0, nil, err
	}
	return result.Scanned()
}

// Return passes the connection back into the database pool.
func (conn *Connection) Return() error {
	err := conn.database.pool.push(conn.resp)
	conn.resp = nil
	return err
}

// ensureProtocol retrieves a protocol from the pool if needed.
func (conn *Connection) ensureProtocol() error {
	if conn.resp == nil {
		p, err := conn.database.pool.pull(unforcedPull)
		if err != nil {
			return err
		}
		conn.resp = p
	}
	return nil
}

// EOF
