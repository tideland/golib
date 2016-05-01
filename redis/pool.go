// Tideland Go Library - Redis Client - resp Pool
//
// Copyright (C) 2009-2016 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package redis

//--------------------
// IMPORTS
//--------------------

import (
	"time"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/loop"
)

//--------------------
// CONNECTION POOL
//--------------------

const (
	forcedPull   = true
	unforcedPull = false

	forcedPullRequest = iota
	unforcedPullRequest
	pushRequest
	killRequest
	closeRequest
)

// poolResponse is returned as result of a pool request.
type poolResponse struct {
	resp *resp
	err  error
}

type poolRequest struct {
	kind         int
	resp         *resp
	responseChan chan *poolResponse
}

// pool manages a number of Redis resp instances.
type pool struct {
	database    *Database
	available   map[*resp]*resp
	inUse       map[*resp]*resp
	backend     loop.Loop
	requestChan chan *poolRequest
}

// newPool creates a connection pool with uninitialized
// protocol instances.
func newPool(db *Database) *pool {
	p := &pool{
		database:    db,
		available:   make(map[*resp]*resp),
		inUse:       make(map[*resp]*resp),
		requestChan: make(chan *poolRequest),
	}
	p.backend = loop.Go(p.backendLoop, "redis", db.address, db.index)
	return p
}

// pull returns a protocol out of the pool. If none is available
// but the configured pool sized isn't reached a new one will be
// established.
func (p *pool) pull(forced bool) (*resp, error) {
	if forced {
		return p.do(forcedPullRequest, nil)
	} else {
		wait := 5 * time.Millisecond
		for i := 0; i < 5; i++ {
			resp, err := p.do(unforcedPullRequest, nil)
			if err != nil {
				return nil, err
			}
			if resp != nil {
				return resp, nil
			}
			time.Sleep(wait)
			wait = wait * 2
		}
		return nil, errors.New(ErrPoolLimitReached, errorMessages, p.database.poolsize)
	}
}

// push returns a protocol back into the pool.
func (p *pool) push(resp *resp) error {
	_, err := p.do(pushRequest, resp)
	return err
}

// kill closes the connection and removes it from the pool.
func (p *pool) kill(resp *resp) error {
	_, err := p.do(killRequest, resp)
	return err
}

// close closes all pooled protocol instances, first the available ones,
// then the ones in use.
func (p *pool) close() error {
	_, err := p.do(closeRequest, nil)
	return err
}

// do executes one request.
func (p *pool) do(kind int, resp *resp) (*resp, error) {
	request := &poolRequest{
		kind:         kind,
		resp:         resp,
		responseChan: make(chan *poolResponse, 1),
	}
	p.requestChan <- request
	response := <-request.responseChan
	return response.resp, response.err
}

// respond answers to a request.
func (p *pool) respond(request *poolRequest, resp *resp, err error) {
	response := &poolResponse{
		resp: resp,
		err:  err,
	}
	request.responseChan <- response
}

// backendLoop manages the pool in a serialized way.
func (p *pool) backendLoop(l loop.Loop) error {
	for {
		select {
		case <-l.ShallStop():
			return nil
		case request := <-p.requestChan:
			// Handle the request.
			switch request.kind {
			case forcedPullRequest:
				// Always return a new protocol.
				resp, err := newResp(p.database)
				if err != nil {
					p.respond(request, nil, err)
				} else {
					p.respond(request, resp, nil)
				}
			case unforcedPullRequest:
				// Fetch a protocol out of the pool.
				switch {
				case len(p.available) > 0:
				fetch:
					for resp := range p.available {
						delete(p.available, resp)
						p.inUse[resp] = resp
						p.respond(request, resp, nil)
						break fetch
					}
				case len(p.inUse) < p.database.poolsize:
					resp, err := newResp(p.database)
					if err != nil {
						p.respond(request, nil, err)
					} else {
						p.respond(request, resp, nil)
					}
				default:
					p.respond(request, nil, nil)
				}
			case pushRequest:
				// Return a protocol.
				delete(p.inUse, request.resp)
				if len(p.available) < p.database.poolsize {
					p.available[request.resp] = request.resp
					p.respond(request, nil, nil)
				} else {
					p.respond(request, nil, request.resp.close())
				}
			case killRequest:
				// Close w/o reusing.
				delete(p.inUse, request.resp)
				p.respond(request, nil, request.resp.close())
			case closeRequest:
				// Close all protocols.
				for resp := range p.available {
					resp.close()
				}
				for resp := range p.inUse {
					resp.close()
				}
				p.respond(request, nil, nil)
			}
		}
	}
}

// EOF
