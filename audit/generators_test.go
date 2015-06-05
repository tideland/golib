// Tideland Go Library - Audit - Unit Tests
//
// Copyright (C) 2013-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package audit_test

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
)

//--------------------
// TESTS
//--------------------

// TestInts tests the generation of ints.
func TestInts(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	gen := audit.New(audit.SimpleRand())

	// Test individual ints.
	for i := 0; i < 10000; i++ {
		lo := gen.Int(-100, 100)
		hi := gen.Int(-100, 100)
		n := gen.Int(lo, hi)
		if hi < lo {
			lo, hi = hi, lo
		}
		assert.True(lo <= n && n <= hi)
	}

	// Test int slices.
	ns := gen.Ints(0, 500, 10000)
	assert.Length(ns, 10000)
	for _, n := range ns {
		assert.True(n >= 0 && n <= 500)
	}

	// Test the generation of percent.
	for i := 0; i < 10000; i++ {
		p := gen.Percent()
		assert.True(p >= 0 && p <= 100)
	}

	// Test the flipping of coins.
	ct := 0
	cf := 0
	for i := 0; i < 10000; i++ {
		c := gen.FlipCoin(50)
		if c {
			ct++
		} else {
			cf++
		}
	}
	assert.About(float64(ct), float64(cf), 500)
}

// TestOneOf tests the generation of selections.
func TestOneOf(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	gen := audit.New(audit.SimpleRand())

	for i := 0; i < 10000; i++ {
		b := gen.OneByteOf(1, 2, 3, 4, 5)
		assert.True(b >= 1 && b <= 5)

		r := gen.OneRuneOf("abcdef")
		assert.True(r >= 'a' && r <= 'f')

		n := gen.OneIntOf(1, 2, 3, 4, 5)
		assert.True(n >= 1 && n <= 5)

		s := gen.OneStringOf("one", "two", "three", "four", "five")
		assert.Substring(s, "one/two/three/four/five")
	}
}

// TestWords tests the generation of words.
func TestWords(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	gen := audit.New(audit.SimpleRand())

	// Test single words.
	for i := 0; i < 10000; i++ {
		w := gen.Word()
		for _, r := range w {
			assert.True(r >= 'a' && r <= 'z')
		}
	}

	// Test limited words.
	for i := 0; i < 10000; i++ {
		lo := gen.Int(audit.MinWordLen, audit.MaxWordLen)
		hi := gen.Int(audit.MinWordLen, audit.MaxWordLen)
		w := gen.LimitedWord(lo, hi)
		wl := len(w)
		if hi < lo {
			lo, hi = hi, lo
		}
		assert.True(lo <= wl && wl <= hi, info("WL %d LO %d HI %d", wl, lo, hi))
	}
}

// TestPattern tests the generation based on patterns.
func TestPattern(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	gen := audit.New(audit.SimpleRand())
	assertPattern := func(pattern, runes string) {
		set := make(map[rune]bool)
		for _, r := range runes {
			set[r] = true
		}
		for i := 0; i < 10; i++ {
			result := gen.Pattern(pattern)
			for _, r := range result {
				assert.True(set[r], pattern, result, runes)
			}
		}
	}

	assertPattern("^^", "^")
	assertPattern("^0^0^0^0^0", "0123456789")
	assertPattern("^1^1^1^1^1", "123456789")
	assertPattern("^o^o^o^o^o", "01234567")
	assertPattern("^h^h^h^h^h", "0123456789abcdef")
	assertPattern("^H^H^H^H^H", "0123456789ABCDEF")
	assertPattern("^a^a^a^a^a", "abcdefghijklmnopqrstuvwxyz")
	assertPattern("^A^A^A^A^A", "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	assertPattern("^c^c^c^c^c", "bcdfghjklmnpqrstvwxyz")
	assertPattern("^C^C^C^C^C", "BCDFGHJKLMNPQRSTVWXYZ")
	assertPattern("^v^v^v^v^v", "aeiou")
	assertPattern("^V^V^V^V^V", "AEIOU")
	assertPattern("^z^z^z^z^z", "abcdefghijklmnopqrstuvwxyz0123456789")
	assertPattern("^Z^Z^Z^Z^Z", "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	assertPattern("^1^0.^0^0^0,^0^0 €", "0123456789 .,€")
}

// TestText tests the generation of text.
func TestText(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	gen := audit.New(audit.SimpleRand())

	for i := 0; i < 10000; i++ {
		s := gen.Sentence()
		ws := strings.Split(s, " ")
		lws := len(ws)
		assert.True(2 <= lws && lws <= 15, info("SL: %d", lws))
		assert.True('A' <= s[0] && s[0] <= 'Z', info("SUC: %v", s[0]))
	}

	for i := 0; i < 10000; i++ {
		p := gen.Paragraph()
		ss := strings.Split(p, ". ")
		lss := len(ss)
		assert.True(2 <= lss && lss <= 10, info("PL: %d", lss))
		for _, s := range ss {
			ws := strings.Split(s, " ")
			lws := len(ws)
			assert.True(2 <= lws && lws <= 15, info("PSL: %d", lws))
			assert.True('A' <= s[0] && s[0] <= 'Z', info("PSUC: %v", s[0]))
		}
	}
}

// TestName tests the generation of names.
func TestName(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	gen := audit.New(audit.SimpleRand())

	assert.Equal(audit.ToUpperFirst("yadda"), "Yadda")

	for i := 0; i < 10000; i++ {
		first, middle, last := gen.MaleName()

		assert.Match(first, `[A-Z][a-z]+(-[A-Z][a-z]+)?`)
		assert.Match(middle, `[A-Z][a-z]+(-[A-Z][a-z]+)?`)
		assert.Match(last, `[A-Z]['a-zA-Z]+`)

		first, middle, last = gen.FemaleName()

		assert.Match(first, `[A-Z][a-z]+(-[A-Z][a-z]+)?`)
		assert.Match(middle, `[A-Z][a-z]+(-[A-Z][a-z]+)?`)
		assert.Match(last, `[A-Z]['a-zA-Z]+`)
	}
}

// TestDomain tests the generation of domains.
func TestDomain(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	gen := audit.New(audit.SimpleRand())

	for i := 0; i < 00100; i++ {
		domain := gen.Domain()

		assert.Match(domain, `^[a-z0-9.-]+\.[a-z]{2,4}$`)
	}
}

// TestURL tests the generation of URLs.
func TestURL(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	gen := audit.New(audit.SimpleRand())

	for i := 0; i < 10000; i++ {
		url := gen.URL()

		assert.Match(url, `(http|ftp|https):\/\/[\w\-_]+(\.[\w\-_]+)+([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
	}
}

// TestEMail tests the generation of e-mail addresses.
func TestEMail(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	gen := audit.New(audit.SimpleRand())

	for i := 0; i < 10000; i++ {
		addr := gen.EMail()

		assert.Match(addr, `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,4}$`)
	}
}

// TestDimes tests the generation of durations and times.
func TestTimes(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	gen := audit.New(audit.SimpleRand())

	for i := 0; i < 10000; i++ {
		// Test durations.
		lo := gen.Duration(time.Second, time.Minute)
		hi := gen.Duration(time.Second, time.Minute)
		d := gen.Duration(lo, hi)
		if hi < lo {
			lo, hi = hi, lo
		}
		assert.True(lo <= d && d <= hi)

		// Test times.
		loc := time.Local
		now := time.Now()
		dur := gen.Duration(24*time.Hour, 30*24*time.Hour)
		t := gen.Time(loc, now, dur)
		assert.True(t.Equal(now) || t.After(now))
		assert.True(t.Before(now.Add(dur)) || t.Equal(now.Add(dur)))
	}
}

//--------------------
// HELPER
//--------------------

var info = fmt.Sprintf

// EOF
