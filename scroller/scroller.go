// Tideland Go Library - Scroller
//
// Copyright (C) 2014-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package scroller

//--------------------
// IMPORTS
//--------------------

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"time"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/loop"
)

//--------------------
// CONSTANTS
//--------------------

const (
	defaultBufferSize = 4096
	defaultPollTime   = time.Second
	delimiter         = '\n'
)

var (
	delimiters = []byte{delimiter}
)

//--------------------
// FILTER
//--------------------

// FilterFunc decides if a line shall be scrolled (func is nil or
// returns true) or not (func returns false).
type FilterFunc func(line []byte) bool

//--------------------
// OPTIONS
//--------------------

// Option defines a function setting an option.
type Option func(s *Scroller) error

// Lines sets the number of lines ro scroll initially.
func Lines(l int) Option {
	return func(s *Scroller) error {
		if l < 0 {
			return errors.New(ErrNegativeLines, errorMessages, l)
		}
		s.lines = l
		return nil
	}
}

// Filter sets the filter function of the scroller.
func Filter(ff FilterFunc) Option {
	return func(s *Scroller) error {
		s.filter = ff
		return nil
	}
}

// BufferSize allows to set the initial size of the buffer
// used for reading.
func BufferSize(bs int) Option {
	return func(s *Scroller) error {
		s.bufferSize = bs
		return nil
	}
}

// PollTime defines the frequency the source is polled.
func PollTime(pt time.Duration) Option {
	return func(s *Scroller) error {
		if pt == 0 {
			pt = defaultPollTime
		}
		s.pollTime = pt
		return nil
	}
}

//--------------------
// SCROLLER
//--------------------

// Scroller scrolls and filters a ReadSeeker line by line and
// writes the data into a Writer.
type Scroller struct {
	source io.ReadSeeker
	target io.Writer

	lines      int
	filter     FilterFunc
	bufferSize int
	pollTime   time.Duration

	reader *bufio.Reader
	writer *bufio.Writer
	loop   loop.Loop
}

// NewScroller starts a Scroller for the given source and target.
// The options can control the number of lines, a filter, the buffer
// size and the poll time.
func NewScroller(source io.ReadSeeker, target io.Writer, options ...Option) (*Scroller, error) {
	if source == nil {
		return nil, errors.New(ErrNoSource, errorMessages)
	}
	if target == nil {
		return nil, errors.New(ErrNoTarget, errorMessages)
	}
	s := &Scroller{
		source:     source,
		target:     target,
		bufferSize: defaultBufferSize,
		pollTime:   defaultPollTime,
	}
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}
	s.reader = bufio.NewReaderSize(s.source, s.bufferSize)
	s.writer = bufio.NewWriter(s.target)
	s.loop = loop.Go(s.backendLoop, "scroller")
	return s, nil
}

// Stop tells the scroller to end working.
func (s *Scroller) Stop() error {
	return s.loop.Stop()
}

// Wait blocks until the scroller has stopped.
func (s *Scroller) Wait() error {
	return s.loop.Wait()
}

// Error returns the status and a possible error of the scroller.
func (s *Scroller) Error() (int, error) {
	return s.loop.Error()
}

// backendLoop is the goroutine for reading, filtering and writing.
func (s *Scroller) backendLoop(l loop.Loop) error {
	// Initial positioning.
	if err := s.seekInitial(); err != nil {
		return err
	}
	// Polling loop.
	timer := time.NewTimer(0)
	for {
		select {
		case <-l.ShallStop():
			return nil
		case <-timer.C:
			for {
				line, readErr := s.readLine()
				_, writeErr := s.writer.Write(line)
				if writeErr != nil {
					return writeErr
				}
				if readErr != nil {
					if readErr != io.EOF {
						return readErr
					}
					break
				}
			}
			if writeErr := s.writer.Flush(); writeErr != nil {
				return writeErr
			}
			timer.Reset(s.pollTime)
		}
	}
}

// seekInitial sets the initial position to start reading. This
// position depends on the number lines and the filter st.
func (s *Scroller) seekInitial() error {
	offset, err := s.source.Seek(0, os.SEEK_END)
	if err != nil {
		return err
	}
	if s.lines < 1 {
		// Simple case, no initial lines wanted.
		return nil
	}
	seekPos := int64(0)
	found := 0
	buffer := make([]byte, s.bufferSize)
SeekLoop:
	for offset > 0 {
		// bufferf partly filled, check if large enough.
		space := cap(buffer) - len(buffer)
		if space < s.bufferSize {
			// Grow buffer.
			newBuffer := make([]byte, len(buffer), cap(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
			space = cap(buffer) - len(buffer)
		}
		if int64(space) > offset {
			// Use exactly the right amount of space if there's
			// only a small amount remaining.
			space = int(offset)
		}
		// Copy remaining data to the end of the buffer.
		copy(buffer[space:cap(buffer)], buffer)
		buffer = buffer[0 : len(buffer)+space]
		offset -= int64(space)
		_, err := s.source.Seek(offset, os.SEEK_SET)
		if err != nil {
			return err
		}
		// Read into the beginning of the buffer.
		_, err = io.ReadFull(s.source, buffer[0:space])
		if err != nil {
			return err
		}
		// Find the end of the last line in the buffer.
		// This will discard any unterminated line at the end
		// of the file.
		end := bytes.LastIndex(buffer, delimiters)
		if end == -1 {
			// No end of line found - discard incomplete
			// line and continue looking. If this happens
			// at the beginning of the file, we don't care
			// because we're going to stop anyway.
			buffer = buffer[:0]
			continue
		}
		end++
		for {
			start := bytes.LastIndex(buffer[0:end-1], delimiters)
			if start == -1 && offset >= 0 {
				break
			}
			start++
			if s.isValid(buffer[start:end]) {
				found++
				if found >= s.lines {
					seekPos = offset + int64(start)
					break SeekLoop
				}
			}
			end = start
		}
		// Leave the last line in the buffer. It's not
		// clear if it is complete or not.
		buffer = buffer[0:end]
	}
	// Final positioning.
	s.source.Seek(seekPos, os.SEEK_SET)
	return nil
}

// readLine reads the next valid line from the reader, even if it is
// larger than the reader buffer.
func (s *Scroller) readLine() ([]byte, error) {
	for {
		slice, err := s.reader.ReadSlice(delimiter)
		if err == nil {
			if s.isValid(slice) {
				return slice, nil
			}
			continue
		}
		line := append([]byte(nil), slice...)
		for err == bufio.ErrBufferFull {
			slice, err = s.reader.ReadSlice(delimiter)
			line = append(line, slice...)
		}
		switch err {
		case nil:
			if s.isValid(line) {
				return line, nil
			}
		case io.EOF:
			// Reached EOF without a delimiter,
			// so step back for next time.
			s.source.Seek(-int64(len(line)), os.SEEK_CUR)
			return nil, err
		default:
			return nil, err
		}
	}
}

// isValid checks if the passed line is valid by using a
// possibly set filter.
func (s *Scroller) isValid(line []byte) bool {
	if s.filter == nil {
		return true
	}
	return s.filter(line)
}

// EOF
