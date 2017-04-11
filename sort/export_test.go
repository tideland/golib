// Tideland Go Library - Sort - Export Test
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package sort

//--------------------
// IMPORTS
//--------------------

import (
	"sort"
)

//--------------------
// EXPORTED FUNCTIONS
//--------------------

func Partition(data sort.Interface, lo, hi int) (int, int) {
	return partition(data, lo, hi)
}

func InsertionSort(data sort.Interface, lo, hi int) {
	insertionSort(data, lo, hi)
}

func SequentialQuickSort(data sort.Interface, lo, hi int) {
	sequentialQuickSort(data, lo, hi)
}

// EOF
