// Tideland Go Library - Sort - Unit Tests
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package sort_test

//--------------------
// IMPORTS
//--------------------

import (
	"math/rand"
	stdsort "sort"
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/sort"
)

//--------------------
// TESTS
//--------------------

// Test pivot.
func TestPivot(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Make some test data.
	td := ByteSlice{17, 20, 13, 15, 51, 6, 21, 11, 23, 47, 59, 88, 78, 67, 94}
	plh, puh := sort.Partition(td, 0, len(td)-1)
	// Asserts.
	assert.Equal(plh, 3, "Pivot lower half.")
	assert.Equal(puh, 5, "Pivot upper half.")
	assert.Equal(td[puh-1], byte(17), "Data at median.")
	assert.Equal(td, ByteSlice{11, 13, 15, 6, 17, 20, 21, 94, 23, 47, 59, 88, 78, 67, 51}, "Prepared data.")
}

// Benchmark the standart integer sort.
func BenchmarkStandardSort(b *testing.B) {
	is := generateIntSlice(b.N)
	stdsort.Sort(is)
}

// Benchmark the insertion sort used insed of sort.
func BenchmarkInsertionSort(b *testing.B) {
	is := generateIntSlice(b.N)
	sort.InsertionSort(is, 0, len(is)-1)
}

// Benchmark the sequential quicksort used insed of sort.
func BenchmarkSequentialQuickSort(b *testing.B) {
	is := generateIntSlice(b.N)
	sort.SequentialQuickSort(is, 0, len(is)-1)
}

// Benchmark the parallel quicksort provided by the package.
func BenchmarkParallelQuickSort(b *testing.B) {
	is := generateIntSlice(b.N)
	sort.Sort(is)
}

//--------------------
// HELPERS
//--------------------

// ByteSlice is a number of bytes for sorting implementing the sort.Interface.
type ByteSlice []byte

func (bs ByteSlice) Len() int           { return len(bs) }
func (bs ByteSlice) Less(i, j int) bool { return bs[i] < bs[j] }
func (bs ByteSlice) Swap(i, j int)      { bs[i], bs[j] = bs[j], bs[i] }

// generateIntSlice generates a slice of ints.
func generateIntSlice(count int) stdsort.IntSlice {
	is := make([]int, count)
	for i := 0; i < count; i++ {
		is[i] = rand.Int()
	}
	return is
}

// duration measures the duration of a function execution.
func duration(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Now().Sub(start)
}

// EOF
