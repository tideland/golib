// Tideland Go Library - Web - Handlers
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package web

//--------------------
// IMPORTS
//--------------------

import (
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/logger"
)

//--------------------
// WRAPPER HANDLER
//--------------------

// WrapperHandler wraps existing handler functions for a usage inside
// the web package.
type WrapperHandler struct {
	id     string
	handle http.HandlerFunc
}

// NewWrapperHandler creates a new wrapper around a handler function.
func NewWrapperHandler(id string, hf http.HandlerFunc) *WrapperHandler {
	return &WrapperHandler{id, hf}
}

// ID is specified on the ResourceHandler interface.
func (h *WrapperHandler) ID() string {
	return h.id
}

// Init is specified on the ResourceHandler interface.
func (h *WrapperHandler) Init(mux Multiplexer, domain, resource string) error {
	return nil
}

// Get is specified on the GetResourceHandler interface.
func (h *WrapperHandler) Get(ctx Context) (bool, error) {
	h.handle(ctx.ResponseWriter(), ctx.Request())
	return true, nil
}

// Head is specified on the HeadResourceHandler interface.
func (h *WrapperHandler) Head(ctx Context) (bool, error) {
	h.handle(ctx.ResponseWriter(), ctx.Request())
	return true, nil
}

// Put is specified on the PutResourceHandler interface.
func (h *WrapperHandler) Put(ctx Context) (bool, error) {
	h.handle(ctx.ResponseWriter(), ctx.Request())
	return true, nil
}

// Post is specified on the PostResourceHandler interface.
func (h *WrapperHandler) Post(ctx Context) (bool, error) {
	h.handle(ctx.ResponseWriter(), ctx.Request())
	return true, nil
}

// Delete is specified on the DeleteResourceHandler interface.
func (h *WrapperHandler) Delete(ctx Context) (bool, error) {
	h.handle(ctx.ResponseWriter(), ctx.Request())
	return true, nil
}

// Options is specified on the OptionsResourceHandler interface.
func (h *WrapperHandler) Options(ctx Context) (bool, error) {
	h.handle(ctx.ResponseWriter(), ctx.Request())
	return true, nil
}

//--------------------
// FILE SERVER HANDLER
//--------------------

// FileServeHandler serves files identified by the resource ID part out
// of the configured local directory.
type FileServeHandler struct {
	id  string
	dir string
}

// NewFileServeHandler creates a new handler with a directory.
func NewFileServeHandler(id, dir string) *FileServeHandler {
	pdir := filepath.FromSlash(dir)
	if !strings.HasSuffix(pdir, string(filepath.Separator)) {
		pdir += string(filepath.Separator)
	}
	return &FileServeHandler{id, pdir}
}

// ID is specified on the ResourceHandler interface.
func (h *FileServeHandler) ID() string {
	return h.id
}

// Init is specified on the ResourceHandler interface.
func (h *FileServeHandler) Init(mux Multiplexer, domain, resource string) error {
	return nil
}

// Get is specified on the GetResourceHandler interface.
func (h *FileServeHandler) Get(ctx Context) (bool, error) {
	filename := h.dir + ctx.ResourceID()
	logger.Infof("serving file %q", filename)
	http.ServeFile(ctx.ResponseWriter(), ctx.Request(), filename)
	return true, nil
}

//--------------------
// FILE UPLOAD HANDLER
//--------------------

const defaultMaxMemory = 32 << 20 // 32 MB

// FileUploadProcessor defines the function used for the processing
// of the uploaded file. It has to be specified by the user of the
// handler and e.g. persists the received data in the file system or
// a database.
type FileUploadProcessor func(ctx Context, header *multipart.FileHeader, file multipart.File) error

// FileUploadHandler handles uploading POST requests.
type FileUploadHandler struct {
	id        string
	processor FileUploadProcessor
}

// NewFileUploadHandler creates a new handler for the uploading of files.
func NewFileUploadHandler(id string, processor FileUploadProcessor) *FileUploadHandler {
	return &FileUploadHandler{
		id:        id,
		processor: processor,
	}
}

// Init is specified on the ResourceHandler interface.
func (h *FileUploadHandler) ID() string {
	return h.id
}

// ID is specified on the ResourceHandler interface.
func (h *FileUploadHandler) Init(mux Multiplexer, domain, resource string) error {
	return nil
}

// Post is specified on the PostResourceHandler interface.
func (h *FileUploadHandler) Post(ctx Context) (bool, error) {
	if err := ctx.Request().ParseMultipartForm(defaultMaxMemory); err != nil {
		return false, errors.Annotate(err, ErrUploadingFile, errorMessages)
	}
	for _, headers := range ctx.Request().MultipartForm.File {
		for _, header := range headers {
			logger.Infof("receiving file %q", header.Filename)
			// Open file and process it.
			if infile, err := header.Open(); err != nil {
				return false, errors.Annotate(err, ErrUploadingFile, errorMessages)
			} else if err := h.processor(ctx, header, infile); err != nil {
				return false, errors.Annotate(err, ErrUploadingFile, errorMessages)
			}
		}
	}
	return true, nil
}

// EOF
