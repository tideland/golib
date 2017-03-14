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

	"github.com/tideland/golib/errors"
)

//--------------------
// LOADER
//--------------------

// FileCacheable contains a file.
type FileCacheable interface {
	Cacheable

	io.Reader
}

// fileCacheable implements the FileCacheable interface.
type fileCacheable struct {
	name    string
	modTime time.Time
	data    []byte
}

// ID implements the Cacheable interface.
func (c *fileCacheable) ID() string {
	return c.name
}

// IsOutdated implements the Cacheable interface.
func (c *fileCacheable) IsOutdated() (bool, error) {
	fi, err := os.Stat(c.name)
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

// Read implements the Reader interface.
func (c *fileCacheable) Read(p []byte) (int, error) {
	n := copy(p, c.data)
	return n, nil
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
			return nil, errors.New(ErrFileSize, errorMessages, name)
		}
		data, err := ioutil.ReadFile(fn)
		if err != nil {
			return nil, errors.Annotate(err, ErrFileLoading, errorMessages, name)
		}
		return &fileCacheable{
			name:    name,
			modTime: fi.ModTime(),
			data:    data,
		}, nil
	}
}

// EOF
