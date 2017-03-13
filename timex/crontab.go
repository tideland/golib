// Tideland Go Library - Time Extensions
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package timex

//--------------------
// IMPORTS
//--------------------

import (
	"time"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/logger"
	"github.com/tideland/golib/loop"
)

//--------------------
// CRONTAB
//--------------------

// Job is executed by the crontab.
type Job interface {
	// ShallExecute decides when called if the job
	// shal be executed.
	ShallExecute(t time.Time) bool

	// Execute executes the job. If the method returns
	// false or an error it will be removed.
	Execute() (bool, error)
}

// cronCommand operates on a crontab.
type command struct {
	add bool
	id  string
	job Job
}

// Crontab is one cron server. A system can run multiple in
// parallel.
type Crontab struct {
	frequency   time.Duration
	jobs        map[string]Job
	commandChan chan *command
	loop        loop.Loop
}

// NewCrontab creates a cron server.
func NewCrontab(freq time.Duration) *Crontab {
	c := &Crontab{
		frequency:   freq,
		jobs:        make(map[string]Job),
		commandChan: make(chan *command),
	}
	c.loop = loop.GoRecoverable(c.backendLoop, c.checkRecovering, "crontab", freq.String())
	return c
}

// Stop terminates the cron server.
func (c *Crontab) Stop() error {
	return c.loop.Stop()
}

// Add adds a new job to the server.
func (c *Crontab) Add(id string, job Job) {
	c.commandChan <- &command{true, id, job}
}

// Remove removes a job from the server.
func (c *Crontab) Remove(id string) {
	c.commandChan <- &command{false, id, nil}
}

// backendLoop runs the server backend.
func (c *Crontab) backendLoop(l loop.Loop) error {
	ticker := time.NewTicker(c.frequency)
	for {
		select {
		case <-l.ShallStop():
			return nil
		case cmd := <-c.commandChan:
			if cmd.add {
				c.jobs[cmd.id] = cmd.job
			} else {
				delete(c.jobs, cmd.id)
			}
		case now := <-ticker.C:
			for id, job := range c.jobs {
				c.do(id, job, now)
			}
		}
	}
}

// checkRecovering checks if the backend can be recovered.
func (c *Crontab) checkRecovering(rs loop.Recoverings) (loop.Recoverings, error) {
	if rs.Frequency(12, time.Minute) {
		logger.Errorf("crontab cannot be recovered: %v", rs.Last().Reason)
		return nil, errors.New(ErrCrontabCannotBeRecovered, errorMessages, rs.Last().Reason)
	}
	logger.Warningf("crontab recovered: %v", rs.Last().Reason)
	return rs.Trim(12), nil
}

// do checks and performs a job.
func (c *Crontab) do(id string, job Job, now time.Time) {
	if job.ShallExecute(now) {
		go func() {
			cont, err := job.Execute()
			if err != nil {
				logger.Errorf("job %q removed after error: %v", id, err)
				cont = false
			}
			if !cont {
				c.Remove(id)
			}
		}()
	}
}

// EOF
