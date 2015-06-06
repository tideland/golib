// Tideland Go Library - Web - Unit Tests - Handlers
//
// Copyright (C) 2009-2014 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package web_test

//--------------------
// IMPORTS
//--------------------

import (
	"bufio"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/web"
)

//--------------------
// TESTS
//--------------------

// TestWrapperHandler tests the usage of standard handler funcs
// wrapped to be used inside the package context.
func TestWrapperHandler(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Setup the test server.
	mux, ts, err := web.StartTestServer()
	assert.Nil(err)
	defer ts.Close()
	handler := func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Been there, done that!"))
	}
	err = mux.Register("test", "wrapper", web.NewWrapperHandler("wrapper", handler))
	assert.Nil(err)
	// Perform test requests.
	resp, err := web.DoTestRequest(ts, &web.TestRequest{
		Method: "GET",
		Path:   "/test/wrapper",
	})
	assert.Nil(err)
	assert.Equal(string(resp.Body), "Been there, done that!")
}

// TestFileServeHandler tests the serving of files.
func TestFileServeHandler(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Setup the test file.
	dir, err := ioutil.TempDir("", "gont-web")
	assert.Nil(err)
	defer os.RemoveAll(dir)
	filename := filepath.Join(dir, "foo.txt")
	f, err := os.Create(filename)
	assert.Nil(err)
	_, err = f.WriteString("Been there, done that!")
	assert.Nil(err)
	assert.Logf("written %s", f.Name())
	err = f.Close()
	assert.Nil(err)
	// Setup the test server.
	mux, ts, err := web.StartTestServer()
	assert.Nil(err)
	defer ts.Close()
	err = mux.Register("test", "files", web.NewFileServeHandler("files", dir))
	assert.Nil(err)
	// Perform test requests.
	resp, err := web.DoTestRequest(ts, &web.TestRequest{
		Method: "GET",
		Path:   "/test/files/foo.txt",
	})
	assert.Nil(err)
	assert.Equal(string(resp.Body), "Been there, done that!")
	resp, err = web.DoTestRequest(ts, &web.TestRequest{
		Method: "GET",
		Path:   "/test/files/does.not.exist",
	})
	assert.Nil(err)
	assert.Equal(string(resp.Body), "404 page not found\n")
}

// TestFileUploadHandler tests the uploading of files.
func TestFileUploadHandler(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	data := "Been there, done that!"
	// Setup the file upload processor.
	processor := func(ctx web.Context, header *multipart.FileHeader, file multipart.File) error {
		assert.Equal(header.Filename, "test.txt")
		scanner := bufio.NewScanner(file)
		assert.True(scanner.Scan())
		text := scanner.Text()
		assert.Equal(text, data)
		return nil
	}
	// Setup the test server.
	mux, ts, err := web.StartTestServer()
	assert.Nil(err)
	defer ts.Close()
	err = mux.Register("test", "files", web.NewFileUploadHandler("files", processor))
	assert.Nil(err)
	// Perform test requests.
	_, err = web.DoTestUpload(ts, "/test/files", "testfile", "test.txt", data)
	assert.Nil(err)
}

// EOF
