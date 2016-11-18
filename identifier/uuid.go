// Tideland Go Library - Identifier - UUID
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package identifier

//--------------------
// IMPORTS
//--------------------

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/tideland/golib/errors"
)

//--------------------
// UUID
//--------------------

const (
	UUIDv1 byte = 1
	UUIDv3 byte = 3
	UUIDv4 byte = 4
	UUIDv5 byte = 5

	UUIDVariantNCS       byte = 0
	UUIDVariantRFC4122   byte = 4
	UUIDVariantMicrosoft byte = 6
	UUIDVariantFuture    byte = 7
)

// UUID represents a universal identifier with 16 bytes.
// See http://en.wikipedia.org/wiki/Universally_unique_identifier.
type UUID [16]byte

// NewUUID returns a new UUID with based on the default version 4.
func NewUUID() UUID {
	uuid, err := NewUUIDv4()
	if err != nil {
		// Panic due to compatibility reasons.
		panic(err)
	}
	return uuid
}

// NewUUIDv1 generates a new UUID based on version 1 (MAC address and
// date-time).
func NewUUIDv1() (UUID, error) {
	uuid := UUID{}
	epoch := int64(0x01b21dd213814000)
	now := uint64(time.Now().UnixNano()/100 + epoch)

	clockSeqRand := [2]byte{}
	rand.Read(clockSeqRand[:])
	clockSeq := binary.LittleEndian.Uint16(clockSeqRand[:])

	timeLow := uint32(now & (0x100000000 - 1))
	timeMid := uint16((now >> 32) & 0xffff)
	timeHighVer := uint16((now >> 48) & 0x0fff)
	clockSeq &= 0x3fff

	binary.LittleEndian.PutUint32(uuid[0:4], timeLow)
	binary.LittleEndian.PutUint16(uuid[4:6], timeMid)
	binary.LittleEndian.PutUint16(uuid[6:8], timeHighVer)
	binary.LittleEndian.PutUint16(uuid[8:10], clockSeq)
	copy(uuid[10:16], cachedMACAddress)

	uuid.setVersion(UUIDv1)
	uuid.setVariant(UUIDVariantRFC4122)
	return uuid, nil
}

// NewUUIDv3 generates a new UUID based on version 3 (MD5 hash of a namespace
// and a name).
func NewUUIDv3(ns UUID, name []byte) (UUID, error) {
	uuid := UUID{}
	hash := md5.New()
	hash.Write(ns.dump())
	hash.Write(name)
	copy(uuid[:], hash.Sum([]byte{})[:16])

	uuid.setVersion(UUIDv3)
	uuid.setVariant(UUIDVariantRFC4122)
	return uuid, nil
}

// NewUUIDv4 generates a new UUID based on version 4 (strong random number).
func NewUUIDv4() (UUID, error) {
	uuid := UUID{}
	_, err := rand.Read([]byte(uuid[:]))
	if err != nil {
		return uuid, err
	}

	uuid.setVersion(UUIDv4)
	uuid.setVariant(UUIDVariantRFC4122)
	return uuid, nil
}

// NewUUIDv5 generates a new UUID based on version 5 (SHA1 hash of a namespace
// and a name).
func NewUUIDv5(ns UUID, name []byte) (UUID, error) {
	uuid := UUID{}
	hash := sha1.New()
	hash.Write(ns.dump())
	hash.Write(name)
	copy(uuid[:], hash.Sum([]byte{})[:16])

	uuid.setVersion(UUIDv5)
	uuid.setVariant(UUIDVariantRFC4122)
	return uuid, nil
}

// NewUUIDByHex creates a UUID based on the passed hex string which has to
// have the length of 32 bytes.
func NewUUIDByHex(source string) (UUID, error) {
	uuid := UUID{}
	if len([]byte(source)) != 32 {
		return uuid, errors.New(ErrInvalidHexLength, errorMessages)
	}
	raw, err := hex.DecodeString(source)
	if err != nil {
		return uuid, errors.Annotate(err, ErrInvalidHexValue, errorMessages)
	}
	copy(uuid[:], raw)
	return uuid, nil
}

// Version returns the version number of the UUID algorithm.
func (uuid UUID) Version() byte {
	return uuid[6] & 0xf0 >> 4
}

// Variant returns the variant of the UUID.
func (uuid UUID) Variant() byte {
	return uuid[8] & 0xe0 >> 5
}

// Copy returns a copy of the UUID.
func (uuid UUID) Copy() UUID {
	uuidCopy := uuid
	return uuidCopy
}

// Raw returns a copy of the UUID bytes.
func (uuid UUID) Raw() [16]byte {
	uuidCopy := uuid.Copy()
	return [16]byte(uuidCopy)
}

// dump creates a copy as a byte slice.
func (uuid UUID) dump() []byte {
	dump := make([]byte, len(uuid))

	copy(dump, uuid[:])

	return dump
}

// ShortString returns a hexadecimal string representation
// without separators.
func (uuid UUID) ShortString() string {
	return fmt.Sprintf("%x", uuid[0:16])
}

// String returns a hexadecimal string representation with
// standardized separators.
func (uuid UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// setVersion sets the version part of the UUID.
func (uuid *UUID) setVersion(v byte) {
	uuid[6] = (uuid[6] & 0x0f) | (v << 4)
}

// setVariant sets the variant part of the UUID.
func (uuid *UUID) setVariant(v byte) {
	uuid[8] = (uuid[8] & 0x1f) | (v << 5)
}

// UUIDNamespaceDNS returns the DNS namespace UUID.
func UUIDNamespaceDNS() UUID {
	uuid, _ := NewUUIDByHex("6ba7b8109dad11d180b400c04fd430c8")
	return uuid
}

// UUIDNamespaceURL returns the URL namespace UUID.
func UUIDNamespaceURL() UUID {
	uuid, _ := NewUUIDByHex("6ba7b8119dad11d180b400c04fd430c8")
	return uuid
}

// UUIDNamespaceOID returns the OID namespace UUID.
func UUIDNamespaceOID() UUID {
	uuid, _ := NewUUIDByHex("6ba7b8129dad11d180b400c04fd430c8")
	return uuid
}

// UUIDNamespaceX500 returns the X.500 namespace UUID.
func UUIDNamespaceX500() UUID {
	uuid, _ := NewUUIDByHex("6ba7b8149dad11d180b400c04fd430c8")
	return uuid
}

//--------------------
// PRIVATE HELPERS
//--------------------

// macAddress retrieves the MAC address of the computer.
func macAddress() []byte {
	address := [6]byte{}
	ifaces, err := net.Interfaces()
	// Try to get address from interfaces.
	if err == nil {
		set := false
		for _, iface := range ifaces {
			if len(iface.HardwareAddr.String()) != 0 {
				copy(address[:], []byte(iface.HardwareAddr))
				set = true
				break
			}
		}
		if set {
			// Had success.
			return address[:]
		}
	}
	// Need a random address.
	rand.Read(address[:])
	address[0] |= 0x01
	return address[:]
}

var cachedMACAddress []byte

func init() {
	cachedMACAddress = macAddress()
}

// EOF
