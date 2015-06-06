// Tideland Go Library - Web - Test Exporting
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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
)

//--------------------
// TEST SERVER
//--------------------

// StartTestServer starts a test server based on the web package
// server and its multiplexer. The returned test server has to be
// closed at the end of the tests.
func StartTestServer(options ...Option) (Multiplexer, *httptest.Server, error) {
	mux, err := NewMultiplexer(options...)
	if err != nil {
		return nil, nil, err
	}
	return mux, httptest.NewServer(mux), nil
}

//--------------------
// TEST TOOLS
//--------------------

// TestSettings handles keys and values for request headers and cookies.
type TestSettings map[string]string

// TestRequest wraps all infos for a test request.
type TestRequest struct {
	Method  string
	Path    string
	Header  TestSettings
	Cookies TestSettings
	Body    []byte
}

// TestResponse wraps all infos of a test response.
type TestResponse struct {
	Cookies TestSettings
	Body    []byte
}

// DoTestRequest performs a request against the test server.
func DoTestRequest(ts *httptest.Server, tr *TestRequest) (*TestResponse, error) {
	// First prepare it.
	transport := &http.Transport{}
	c := &http.Client{Transport: transport}
	url := ts.URL + tr.Path

	var bodyReader io.Reader
	if tr.Body != nil {
		bodyReader = ioutil.NopCloser(bytes.NewBuffer(tr.Body))
	}
	req, err := http.NewRequest(tr.Method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	for key, value := range tr.Header {
		req.Header.Set(key, value)
	}
	for key, value := range tr.Cookies {
		cookie := &http.Cookie{
			Name:  key,
			Value: value,
		}
		req.AddCookie(cookie)
	}

	// Now do it.
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot perform test request: %v", err)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	cookies := TestSettings{}
	for _, cookie := range resp.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	return &TestResponse{
		Cookies: cookies,
		Body:    respBody,
	}, err
}

// DoTestUpload is a special request for uploading a file.
func DoTestUpload(ts *httptest.Server, path, fieldname, filename, data string) (*TestResponse, error) {
	// Prepare request.
	transport := &http.Transport{}
	c := &http.Client{Transport: transport}
	url := ts.URL + path

	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)
	part, err := writer.CreateFormFile(fieldname, filename)
	if err != nil {
		return nil, fmt.Errorf("cannot create form file: %v", err)
	}
	_, err = io.WriteString(part, data)
	if err != nil {
		return nil, fmt.Errorf("cannot write data: %v", err)
	}
	contentType := writer.FormDataContentType()
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot close multipart wariter: %v", err)
	}

	// And now do it.
	resp, err := c.Post(url, contentType, buffer)
	if err != nil {
		return nil, fmt.Errorf("cannot perform test upload: %v", err)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	cookies := TestSettings{}
	for _, cookie := range resp.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	return &TestResponse{
		Cookies: cookies,
		Body:    respBody,
	}, err
}

// EOF
