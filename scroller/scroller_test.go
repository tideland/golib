// Tideland Go Library - Scroller - Unit Tests
//
// Copyright (C) 2014-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package scroller_test

//--------------------
// IMPORTS
//--------------------

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/loop"
	"github.com/tideland/golib/scroller"
)

//--------------------
// TESTS
//--------------------

// tests contains the data descibing the tests.
var tests = []struct {
	description      string
	initialLeg       int
	options          func() []scroller.Option
	initialExpected  []int
	appendedExpected []int
	injector         func(*scroller.Scroller, *readSeeker) func([]string)
	err              string
}{{
	description: "no lines existing; initially no lines scrolled",
	options: func() []scroller.Option {
		return []scroller.Option{
			scroller.PollTime(2 * time.Millisecond),
		}
	},
	initialExpected:  []int{},
	appendedExpected: intRange(0, 99),
}, {
	description: "no lines existing; initially five lines scrolled",
	options: func() []scroller.Option {
		return []scroller.Option{
			scroller.Lines(5),
			scroller.PollTime(2 * time.Millisecond),
		}
	},
	initialExpected:  []int{},
	appendedExpected: intRange(0, 99),
}, {
	description: "ten lines existing; initially no lines scrolled",
	initialLeg:  10,
	options: func() []scroller.Option {
		return []scroller.Option{
			scroller.PollTime(2 * time.Millisecond),
		}
	},
	initialExpected:  []int{},
	appendedExpected: intRange(10, 99),
}, {
	description: "ten lines existing; initially five lines scrolled",
	initialLeg:  10,
	options: func() []scroller.Option {
		return []scroller.Option{
			scroller.Lines(5),
			scroller.PollTime(2 * time.Millisecond),
		}
	},
	initialExpected:  intRange(5, 9),
	appendedExpected: intRange(10, 99),
}, {
	description: "ten lines existing; initially twenty lines scrolled",
	initialLeg:  10,
	options: func() []scroller.Option {
		return []scroller.Option{
			scroller.Lines(20),
			scroller.PollTime(2 * time.Millisecond),
		}
	},
	initialExpected:  intRange(0, 9),
	appendedExpected: intRange(10, 99),
}, {
	description: "ten lines existing; initially twenty lines scrolled; buffer smaller than lines",
	initialLeg:  10,
	options: func() []scroller.Option {
		return []scroller.Option{
			scroller.Lines(20),
			scroller.PollTime(2 * time.Millisecond),
			scroller.BufferSize(10),
		}
	},
	initialExpected:  intRange(0, 9),
	appendedExpected: intRange(10, 99),
}, {
	description: "ten lines existing; initially three lines scrolled; filter lines with special prefix",
	initialLeg:  10,
	options: func() []scroller.Option {
		return []scroller.Option{
			scroller.Lines(3),
			scroller.Filter(func(line []byte) bool { return bytes.HasPrefix(line, specialPrefix) }),
			scroller.PollTime(2 * time.Millisecond),
		}
	},
	initialExpected:  []int{3, 5, 8},
	appendedExpected: []int{13, 21, 44, 65},
}, {
	description: "ten lines existing; initially five lines scrolled; error after further 25 lines",
	initialLeg:  10,
	options: func() []scroller.Option {
		return []scroller.Option{
			scroller.Lines(5),
			scroller.PollTime(2 * time.Millisecond),
		}
	},
	initialExpected:  intRange(5, 9),
	appendedExpected: intRange(10, 99),
	injector: func(s *scroller.Scroller, rs *readSeeker) func([]string) {
		return func(lines []string) {
			if len(lines) == 25 {
				rs.setError("ouch")
			}
		}
	},
	err: "ouch",
}, {
	description: "ten lines existing; initially five lines scrolled; simply stop after 25 lines",
	initialLeg:  10,
	options: func() []scroller.Option {
		return []scroller.Option{
			scroller.Lines(5),
			scroller.PollTime(2 * time.Millisecond),
		}
	},
	initialExpected:  intRange(5, 9),
	appendedExpected: intRange(10, 99),
	injector: func(s *scroller.Scroller, rs *readSeeker) func([]string) {
		return func(lines []string) {
			if len(lines) == 25 {
				s.Stop()
			}
		}
	},
}, {
	description: "unterminated last line is not scrolled",
	initialLeg:  103,
	options: func() []scroller.Option {
		return []scroller.Option{
			scroller.Lines(5),
			scroller.PollTime(2 * time.Millisecond),
		}
	},
	initialExpected:  intRange(95, 97),
	appendedExpected: intRange(98, 99),
}}

// TestScroller runs the different scroller test.
func TestScroller(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	input, output := generateTestData()

	for i, test := range tests {
		assert.Logf("test #%d/%d: %s", i+1, len(tests), test.description)

		rs, sigc := newReadSeeker(input, test.initialLeg)
		receiver := newReceiver(assert, output)

		// Set options.
		options := []scroller.Option{}
		if test.options != nil {
			options = test.options()
		}

		// Start scroller.
		s, err := scroller.NewScroller(rs, receiver.writer, options...)
		assert.Nil(err)
		receiver.autoClose(s)

		// Prepare injection.
		var injection func([]string)
		if test.injector != nil {
			injection = test.injector(s, rs)
		}

		// Make assertions.
		receiver.assertCollected(test.initialExpected, nil)

		sigc <- struct{}{}

		receiver.assertCollected(test.appendedExpected, injection)

		// Test success or error.
		if test.err == "" {
			assert.Nil(s.Stop())
		} else {
			st, err := s.Error()
			assert.Equal(st, loop.Stopped)
			assert.ErrorMatch(err, test.err)
		}
	}
}

//--------------------
// TEST HELPERS
//--------------------

// intRange creates a set of ints.
func intRange(lo, hi int) []int {
	is := []int{}
	for i := lo; i <= hi; i++ {
		is = append(is, i)
	}
	return is
}

var (
	regularPrefix = []byte("[REGULAR]")
	specialPrefix = []byte("[SPECIAL]")
)

// generateTestData returns slices with input and output data for tests.
func generateTestData() (input, output []string) {
	tagged := []int{1, 2, 3, 5, 8, 13, 21, 44, 65}
	rand := audit.FixedRand()
	gen := audit.NewGenerator(rand)
	line := ""
	// Generate 98 standard lines.
	for i := 0; i < 98; i++ {
		switch {
		case i%10 == 0:
			// Spread some empty lines.
			line = "\n"
		case len(tagged) > 0 && i == tagged[0]:
			// Special prefixed lines.
			line = fmt.Sprintf("%s #%d ", specialPrefix, i) + gen.Sentence() + "\n"
			tagged = tagged[1:]
		default:
			// Regular prefixed lines.
			line = fmt.Sprintf("%s #%d ", regularPrefix, i) + gen.Sentence() + "\n"
		}
		input = append(input, line)
		output = append(output, line)
	}
	// Add two longer lines, each time the first half not terminated.
	tmp := ""
	for i := 0; i < 4; i++ {
		if i%2 == 0 {
			line = fmt.Sprintf("%s #%d ", regularPrefix, i) +
				gen.Sentence() + " " +
				gen.Sentence() + " "
			tmp = line
		} else {
			line = gen.Sentence() + " " + gen.Sentence() + "\n"
			tmp += line
		}
		input = append(input, line)
		if i%2 != 0 {
			output = append(output, tmp)
			tmp = ""
		}
	}
	// Add an unterminated line.
	line = fmt.Sprintf("%s #%d ", specialPrefix, 100) + gen.Sentence()
	input = append(input, line)
	return input, output
}

// readSeeker simulates the ReadSeeker in the tests.
type readSeeker struct {
	mux    sync.Mutex
	buffer []byte
	pos    int
	err    error
}

// newReadSeeker creates the ReadSeeker with the passed input. The data
// is written with an initial number of lines and then waits for a signal
// to continue.
func newReadSeeker(input []string, initialLeg int) (*readSeeker, chan struct{}) {
	sigc := make(chan struct{})
	rs := &readSeeker{}
	i := 0
	for ; i < initialLeg; i++ {
		rs.write(input[i])
	}
	go func() {
		<-sigc

		for ; i < len(input); i++ {
			time.Sleep(5 * time.Millisecond)
			rs.write(input[i])
		}
	}()
	return rs, sigc
}

func (rs *readSeeker) write(s string) {
	rs.mux.Lock()
	defer rs.mux.Unlock()
	rs.buffer = append(rs.buffer, []byte(s)...)
}

func (rs *readSeeker) setError(msg string) {
	rs.mux.Lock()
	defer rs.mux.Unlock()
	rs.err = errors.New(msg)
}

func (rs *readSeeker) Read(p []byte) (n int, err error) {
	rs.mux.Lock()
	defer rs.mux.Unlock()
	if rs.err != nil {
		return 0, rs.err
	}
	if rs.pos >= len(rs.buffer) {
		return 0, io.EOF
	}
	n = copy(p, rs.buffer[rs.pos:])
	rs.pos += n
	return n, nil
}

func (rs *readSeeker) Seek(offset int64, whence int) (ret int64, err error) {
	rs.mux.Lock()
	defer rs.mux.Unlock()
	var newPos int64
	switch whence {
	case 0:
		newPos = offset
	case 1:
		newPos = int64(rs.pos) + offset
	case 2:
		newPos = int64(len(rs.buffer)) + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}
	if newPos < 0 {
		return 0, fmt.Errorf("negative position: %d", newPos)
	}
	if newPos >= 1<<31 {
		return 0, fmt.Errorf("position out of range: %d", newPos)
	}
	rs.pos = int(newPos)
	return newPos, nil
}

// receiver is responsible for receiving the scrolled lines and
// performing the assertions
type receiver struct {
	assert audit.Assertion
	data   []string
	reader *io.PipeReader
	writer *io.PipeWriter
	linec  chan string
}

// newReceiver creates a new receiver.
func newReceiver(assert audit.Assertion, data []string) *receiver {
	r := &receiver{
		assert: assert,
		data:   data,
		linec:  make(chan string),
	}
	r.reader, r.writer = io.Pipe()
	go r.loop()
	return r
}

func (r *receiver) autoClose(scroller *scroller.Scroller) {
	go func() {
		scroller.Wait()
		r.writer.Close()
	}()
}

func (r *receiver) assertCollected(expected []int, injection func([]string)) {
	expectedLines := []string{}
	for _, lineNo := range expected {
		expectedLines = append(expectedLines, r.data[lineNo])
	}
	timeout := time.After(2 * time.Second)
	lines := []string{}
	for {
		select {
		case line, ok := <-r.linec:
			if ok {
				lines = append(lines, line)
				if injection != nil {
					injection(lines)
				}
				if len(lines) == len(expectedLines) {
					// All data received.
					r.assert.Equal(lines, expectedLines)
					return
				}
			} else {
				// linec closed after stopping or error.
				r.assert.Equal(lines, expectedLines[:len(lines)])
				return
			}
		case <-timeout:
			if len(expected) == 0 || injection != nil {
				return
			}
			r.assert.Fail("timeout during tailer collection")
			return
		}
	}
}

func (r *receiver) loop() {
	defer close(r.linec)
	reader := bufio.NewReader(r.reader)
	for {
		line, err := reader.ReadString('\n')
		switch err {
		case nil:
			r.linec <- line
		case io.EOF:
			return
		default:
			r.assert.Fail()
		}
	}
}

// EOF
