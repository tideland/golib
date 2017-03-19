// Tideland Go Library - Cache - Loader
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cache

//--------------------
// IMPORTS
//--------------------

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"bytes"

	"github.com/tideland/golib/errors"
)

//--------------------
// LOADER
//--------------------

// FileCacheable contains a file.
type FileCacheable interface {
	Cacheable

	// ReadCloser returns the io.ReadCloser for the
	// cached file or the file itself if it's too large.
	ReadCloser() (io.ReadCloser, error)
}

// fileBuffer encapsulates a bytes.Buffer as io.ReadCloser.
type fileBuffer struct {
	b *bytes.Buffer
}

// Read implements the io.ReadCloser interface.
func (fb *fileBuffer) Read(p []byte) (int, error) {
	return fb.b.Read(p)
}

// Close implements the io.ReadCloser interface.
func (fb *fileBuffer) Close() error {
	return nil
}

// fileCacheable implements the FileCacheable interface.
type fileCacheable struct {
	name     string
	fullname string
	tooLarge bool
	modTime  time.Time
	data     []byte
}

// ID implements the Cacheable interface.
func (c *fileCacheable) ID() string {
	return c.name
}

// IsOutdated implements the Cacheable interface.
func (c *fileCacheable) IsOutdated() (bool, error) {
	fi, err := os.Stat(c.fullname)
	if err != nil {
		return false, errors.Annotate(err, ErrFileChecking, errorMessages)
	}
	if fi.ModTime().After(c.modTime) {
		return true, nil
	}
	return false, nil
}

// Discard implements the Cacheable interface.
func (c *fileCacheable) Discard() error {
	return nil
}

// ReadCloser implements the FileCacheable interface.
func (c *fileCacheable) ReadCloser() (io.ReadCloser, error) {
	// Check if the file has to be returned directly because
	// it is too large.
	if c.tooLarge {
		f, err := os.Open(c.fullname)
		if err != nil {
			return nil, errors.Annotate(err, ErrFileLoading, errorMessages, c.name)
		}
		return f, nil
	}
	// It's a cached buffer, so return this.
	return &fileBuffer{bytes.NewBuffer(c.data)}, nil
}

// NewFileLoader returns a CacheableLoader for files. It
// starts at the given root directory.
func NewFileLoader(root string, maxSize int64) CacheableLoader {
	return func(name string) (Cacheable, error) {
		fn := filepath.Join(root, name)
		fi, err := os.Stat(fn)
		if err != nil {
			return nil, errors.Annotate(err, ErrFileLoading, errorMessages, name)
		}
		if fi.Size() > maxSize {
			return &fileCacheable{
				name:     name,
				fullname: fn,
				tooLarge: true,
				modTime:  fi.ModTime(),
			}, nil
		}
		data, err := ioutil.ReadFile(fn)
		if err != nil {
			return nil, errors.Annotate(err, ErrFileLoading, errorMessages, name)
		}
		return &fileCacheable{
			name:     name,
			fullname: fn,
			modTime:  fi.ModTime(),
			data:     data,
		}, nil
	}
}

// EOF
