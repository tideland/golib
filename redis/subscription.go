// Tideland Go Library - Redis Client - Subscription
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
)

//--------------------
// SUBSCRIPTION
//--------------------

// Subscription manages a subscription to Redis channels and allows
// to subscribe and unsubscribe from channels.
type Subscription struct {
	database *Database
	resp     *resp
}

// newSubscription creates a new subscription.
func newSubscription(db *Database) (*Subscription, error) {
	sub := &Subscription{
		database: db,
	}
	err := sub.ensureProtocol()
	if err != nil {
		return nil, err
	}
	// Perform authentication and database selection.
	err = sub.resp.authenticate()
	if err != nil {
		sub.database.pool.kill(sub.resp)
		return nil, err
	}
	return sub, nil
}

// Subscribe adds one or more channels to the subscription.
func (sub *Subscription) Subscribe(channels ...string) error {
	return sub.subUnsub("subscribe", channels...)
}

// Unsubscribe removes one or more channels from the subscription.
func (sub *Subscription) Unsubscribe(channels ...string) error {
	return sub.subUnsub("unsubscribe", channels...)
}

// subUnsub is the generic subscription and unsubscription method.
func (sub *Subscription) subUnsub(cmd string, channels ...string) error {
	err := sub.ensureProtocol()
	if err != nil {
		return err
	}
	pattern := false
	args := []interface{}{}
	for _, channel := range channels {
		if containsPattern(channel) {
			pattern = true
		}
		args = append(args, channel)
	}
	if pattern {
		cmd = "p" + cmd
	}
	err = sub.resp.sendCommand(cmd, args...)
	logCommand(cmd, args, err, sub.database.logging)
	return err
}

// Pop waits for a published value and returns it.
func (sub *Subscription) Pop() (*PublishedValue, error) {
	err := sub.ensureProtocol()
	if err != nil {
		return nil, err
	}
	result, err := sub.resp.receiveResultSet()
	if err != nil {
		return nil, err
	}
	// Analyse the result.
	kind, err := result.StringAt(0)
	if err != nil {
		return nil, err
	}
	switch {
	case strings.Contains(kind, "message"):
		channel, err := result.StringAt(1)
		if err != nil {
			return nil, err
		}
		value, err := result.ValueAt(2)
		if err != nil {
			return nil, err
		}
		return &PublishedValue{
			Kind:    kind,
			Channel: channel,
			Value:   value,
		}, nil
	case strings.Contains(kind, "subscribe"):
		channel, err := result.StringAt(1)
		if err != nil {
			return nil, err
		}
		count, err := result.IntAt(2)
		if err != nil {
			return nil, err
		}
		return &PublishedValue{
			Kind:    kind,
			Channel: channel,
			Count:   count,
		}, nil
	default:
		return nil, errors.New(ErrInvalidResponse, errorMessages, result)
	}
}

// Close ends the subscription.
func (sub *Subscription) Close() error {
	err := sub.ensureProtocol()
	if err != nil {
		return err
	}
	err = sub.resp.sendCommand("punsubscribe")
	if err != nil {
		return err
	}
	for {
		pv, err := sub.Pop()
		if err != nil {
			return err
		}
		if pv.Kind == "punsubscribe" {
			break
		}
	}
	sub.database.pool.push(sub.resp)
	return nil
}

// ensureProtocol retrieves a protocol from the pool if needed.
func (sub *Subscription) ensureProtocol() error {
	if sub.resp == nil {
		p, err := sub.database.pool.pull(forcedPull)
		if err != nil {
			return err
		}
		sub.resp = p
	}
	return nil
}

// EOF
