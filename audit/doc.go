// Tideland Go Library - Audit
//
// Copyright (C) 2012-2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// The Tideland Go Library audit package helps writing convenient and
// powerful unit tests. One part of it are assertions to compare expected
// and obtained values. Additional text output for failing tests can be
// added.
//
// In the beginning of a test function a new assertion instance is created with:
//
// assert := audit.NewTestingAssertion(t, shallFail)
//
// Inside the test an assert looks like:
//
// assert.Equal(obtained, expected, "obtained value has to be like expected")
//
// If shallFail is set to true a failing assert also lets fail the Go test.
// Otherwise the failing is printed but the tests continue. Other functions
// help with temporary directories, environment variables, and the generating
// of test data.
//
// Additional helpers support in generating test data or work with the
// environment, like temporary directories or environment variables, in
// a safe and convenient way.
package audit

// EOF
