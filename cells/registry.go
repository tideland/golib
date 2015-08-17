// Tideland Go Library - Cells - Registry
//
// Copyright (C) 2010-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"sync"

	"github.com/tideland/golib/errors"
)

//--------------------
// CELLER
//--------------------

// celler manages one cell, its subscriptions and subscribers.
type celler struct {
	registry      *registry
	cell          *cell
	subscriptions map[string]struct{}
	subscribers   map[string]struct{}
}

// stop unsubscribes from the subscriptions and stops the cell.
func (cr *celler) stop() error {
	subscriberID := cr.cell.ID()
	for id := range cr.subscribers {
		ucr := cr.registry.cellers[id]
		delete(ucr.subscriptions, subscriberID)
		ucr.updateSubscribers()
	}
	for id := range cr.subscriptions {
		ucr, ok := cr.registry.cellers[id]
		if !ok {
			panic("subscriptions out of sync with cellers")
		}
		delete(ucr.subscribers, subscriberID)
		ucr.updateSubscribers()
	}
	return cr.cell.stop()
}

// updateSubscribers notifies the cell about the
// current subscribers.
func (cr *celler) updateSubscribers() {
	var cells []*cell
	for id := range cr.subscribers {
		if scr, ok := cr.registry.cellers[id]; ok {
			cells = append(cells, scr.cell)
		}
	}
	cr.cell.subscribers.update(cells)
}

//--------------------
// CELL REGISTRY
//--------------------

// registry manages cells and their subscriptions.
type registry struct {
	mutex   sync.RWMutex
	cellers map[string]*celler
}

// newRegistry creates a new cell registry.
func newRegistry() *registry {
	return &registry{
		cellers: make(map[string]*celler),
	}
}

// stop stops the registry.
func (r *registry) stop() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, cr := range r.cellers {
		if err := cr.cell.stop(); err != nil {
			return err
		}
	}
	return nil
}

// startCell starts and adds a new cell to the registry if the
// ID does not already exist.
func (r *registry) startCell(env *environment, id string, behavior Behavior) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// Check if the ID already exists.
	if _, ok := r.cellers[id]; ok {
		return errors.New(ErrDuplicateID, errorMessages, id)
	}
	// Create and add celler.
	c, err := newCell(env, id, behavior)
	if err != nil {
		return err
	}
	cr := &celler{
		registry:      r,
		cell:          c,
		subscriptions: make(map[string]struct{}),
		subscribers:   make(map[string]struct{}),
	}
	r.cellers[id] = cr
	return nil
}

// stopCell stops a cell.
func (r *registry) stopCell(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	cr, ok := r.cellers[id]
	if !ok {
		return errors.New(ErrInvalidID, errorMessages, id)
	}
	// Stopping the celler will stop the cell and unsubscribe
	// from its subscriptions.
	if err := cr.stop(); err != nil {
		return err
	}
	// Remove the cell from the registry.
	delete(r.cellers, id)
	return nil
}

// subscribe subscribes cells to an emitter.
func (r *registry) subscribe(emitterID string, subscriberIDs ...string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	cr, ok := r.cellers[emitterID]
	if !ok {
		return errors.New(ErrInvalidID, errorMessages, emitterID)
	}
	// Subscribe to subscriber IDs.
	if err := r.checkIDs(emitterID, subscriberIDs...); err != nil {
		return err
	}
	for _, subscriberID := range subscriberIDs {
		ucr := r.cellers[subscriberID]
		ucr.subscriptions[emitterID] = struct{}{}
		cr.subscribers[subscriberID] = struct{}{}
	}
	cr.updateSubscribers()
	return nil
}

// unsubscribe usubscribes cells from an emitter.
func (r *registry) unsubscribe(emitterID string, subscriberIDs ...string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	cr, ok := r.cellers[emitterID]
	if !ok {
		return errors.New(ErrInvalidID, errorMessages, emitterID)
	}
	// Unsubscribe from subscriber IDs.
	if err := r.checkIDs(emitterID, subscriberIDs...); err != nil {
		return err
	}
	for _, subscriberID := range subscriberIDs {
		ucr := r.cellers[subscriberID]
		delete(ucr.subscriptions, emitterID)
		delete(cr.subscribers, subscriberID)
	}
	cr.updateSubscribers()
	return nil
}

// subscribers returns the IDs of the subscribers of one cell.
func (r *registry) subscribers(emitterID string) ([]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	cr, ok := r.cellers[emitterID]
	if !ok {
		return nil, errors.New(ErrInvalidID, errorMessages, emitterID)
	}
	var ids []string
	for id := range cr.subscribers {
		ids = append(ids, id)
	}
	return ids, nil
}

// cells returns the cells with the given ids.
func (r *registry) cells(ids ...string) ([]*cell, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	var cs []*cell
	for _, id := range ids {
		c, ok := r.cellers[id]
		if !ok {
			return nil, errors.New(ErrInvalidID, errorMessages, id)
		}
		cs = append(cs, c.cell)
	}
	return cs, nil
}

// checkIDs checks if the passed IDs are valid. It is only
// called internally, so no locking.
func (r *registry) checkIDs(emitterID string, subscriberIDs ...string) error {
	for _, subscriberID := range subscriberIDs {
		if subscriberID == emitterID {
			return errors.New(ErrInvalidID, errorMessages, subscriberID)
		}
		if _, ok := r.cellers[subscriberID]; !ok {
			return errors.New(ErrInvalidID, errorMessages, subscriberID)
		}
	}
	return nil
}

// EOF
