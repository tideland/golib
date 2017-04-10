// Tideland Go Library - Map/Reduce
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package mapreduce

//--------------------
// IMPORTS
//--------------------

import (
	"hash/adler32"
	"runtime"
)

//--------------------
// KEY/VALUE
//--------------------

// KeyValue is a pair of a string key and any data as value.
type KeyValue interface {
	// Key returns the key for the mapping.
	Key() string

	// Value returns the payload value for processing.
	Value() interface{}
}

// KeyValueChan is a channel for the transfer of key/value pairs.
type KeyValueChan chan KeyValue

// Close the channel for key/value pairs.
func (c KeyValueChan) Close() {
	close(c)
}

//--------------------
// MAP/REDUCE
//--------------------

// MapReducer has to be implemented to control the map/reducing.
type MapReducer interface {
	// Input has to return the input channel for the
	// date to process.
	Input() KeyValueChan

	// Map maps a key/value pair to another one and emits it.
	Map(in KeyValue, emit KeyValueChan)

	// Reduce reduces the values delivered via the input
	// channel to the emit channel.
	Reduce(in, emit KeyValueChan)

	// Consume allows the MapReducer to consume the
	// processed data.
	Consume(in KeyValueChan) error
}

// MapReduce applies a map and a reduce function to keys and values in parallel.
func MapReduce(mr MapReducer) error {
	mapEmitChan := make(KeyValueChan)
	reduceEmitChan := make(KeyValueChan)

	go performReducing(mr, mapEmitChan, reduceEmitChan)
	go performMapping(mr, mapEmitChan)

	return mr.Consume(reduceEmitChan)
}

//--------------------
// PRIVATE
//--------------------

// closerChan signals the closing of channels.
type closerChan chan struct{}

// closerChan closes given channel after a number of signals.
func newCloserChan(kvc KeyValueChan, size int) closerChan {
	signals := make(closerChan)
	go func() {
		ctr := 0
		for {
			<-signals
			ctr++
			if ctr == size {
				kvc.Close()
				close(signals)
				return
			}
		}
	}()
	return signals
}

// performReducing runs the reducing goroutines.
func performReducing(mr MapReducer, mapEmitChan, reduceEmitChan KeyValueChan) {
	// Start a closer for the reduce emit chan.
	size := runtime.NumCPU()
	signals := newCloserChan(reduceEmitChan, size)

	// Start reduce goroutines.
	reduceChans := make([]KeyValueChan, size)
	for i := 0; i < size; i++ {
		reduceChans[i] = make(KeyValueChan)
		go func(in KeyValueChan) {
			mr.Reduce(in, reduceEmitChan)
			signals <- struct{}{}
		}(reduceChans[i])
	}

	// Read map emitted data.
	for kv := range mapEmitChan {
		hash := adler32.Checksum([]byte(kv.Key()))
		idx := hash % uint32(size)
		reduceChans[idx] <- kv
	}

	// Close reduce channels.
	for _, reduceChan := range reduceChans {
		reduceChan.Close()
	}
}

// Perform the mapping.
func performMapping(mr MapReducer, mapEmitChan KeyValueChan) {
	// Start a closer for the map emit chan.
	size := runtime.NumCPU() * 4
	signals := newCloserChan(mapEmitChan, size)

	// Start map goroutines.
	mapChans := make([]KeyValueChan, size)
	for i := 0; i < size; i++ {
		mapChans[i] = make(KeyValueChan)
		go func(in KeyValueChan) {
			for kv := range in {
				mr.Map(kv, mapEmitChan)
			}
			signals <- struct{}{}
		}(mapChans[i])
	}

	// Dispatch input data to map channels.
	idx := 0
	for kv := range mr.Input() {
		mapChans[idx%size] <- kv
		idx++
	}

	// Close map channels.
	for i := 0; i < size; i++ {
		mapChans[i].Close()
	}
}

// EOF
