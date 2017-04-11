// Tideland Go Library - Identifier - Unit Tests
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package identifier_test

//--------------------
// IMPORTS
//--------------------

import (
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/identifier"
)

//--------------------
// TESTS
//--------------------

// Test the standard UUID.
func TestStandardUUID(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Asserts.
	uuid := identifier.NewUUID()
	assert.Equal(uuid.Version(), identifier.UUIDv4)
	uuidShortStr := uuid.ShortString()
	uuidStr := uuid.String()
	assert.Equal(len(uuid), 16, "UUID length has to be 16")
	assert.Match(uuidShortStr, "[0-9a-f]{32}", "UUID short")
	assert.Match(uuidStr, "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", "UUID long")
	// Check for unique creation, but only weak test.
	uuids := make(map[string]bool)
	for i := 0; i < 1000000; i++ {
		uuid = identifier.NewUUID()
		uuidStr = uuid.String()
		assert.False(uuids[uuidStr], "UUID collision must not happen")
		uuids[uuidStr] = true
	}
	// Check for copy.
	uuidA := identifier.NewUUID()
	uuidB := uuidA.Copy()
	for i := 0; i < len(uuidA); i++ {
		uuidA[i] = 0
	}
	assert.Different(uuidA, uuidB)
}

// Test UUID versions.
func TestUUIDVersions(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	ns := identifier.UUIDNamespaceOID()
	// Asserts.
	uuidV1, err := identifier.NewUUIDv1()
	assert.Nil(err)
	assert.Equal(uuidV1.Version(), identifier.UUIDv1)
	assert.Equal(uuidV1.Variant(), identifier.UUIDVariantRFC4122)
	assert.Logf("UUID V1: %v", uuidV1)
	uuidV3, err := identifier.NewUUIDv3(ns, []byte{4, 7, 1, 1})
	assert.Nil(err)
	assert.Equal(uuidV3.Version(), identifier.UUIDv3)
	assert.Equal(uuidV3.Variant(), identifier.UUIDVariantRFC4122)
	assert.Logf("UUID V3: %v", uuidV3)
	uuidV4, err := identifier.NewUUIDv4()
	assert.Nil(err)
	assert.Equal(uuidV4.Version(), identifier.UUIDv4)
	assert.Equal(uuidV4.Variant(), identifier.UUIDVariantRFC4122)
	assert.Logf("UUID V4: %v", uuidV4)
	uuidV5, err := identifier.NewUUIDv5(ns, []byte{4, 7, 1, 1})
	assert.Nil(err)
	assert.Equal(uuidV5.Version(), identifier.UUIDv5)
	assert.Equal(uuidV5.Variant(), identifier.UUIDVariantRFC4122)
	assert.Logf("UUID V5: %v", uuidV5)
}

// Test creating UUIDs from hex strings.
func TestUUIDByHex(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Asserts.
	_, err := identifier.NewUUIDByHex("ffff")
	assert.ErrorMatch(err, `\[IDENTIFIER:.*\] invalid length of hex string, has to be 32`)
	_, err = identifier.NewUUIDByHex("012345678901234567890123456789zz")
	assert.ErrorMatch(err, `\[IDENTIFIER:.*\] invalid value of hex string: .*`)
	_, err = identifier.NewUUIDByHex("012345678901234567890123456789ab")
	assert.Nil(err)
}

// EOF
