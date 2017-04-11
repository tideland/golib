// Tideland Go Library - Map/Reduce
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// Package mapreduce of the Tideland Go Library implements the Map/Reduce
// algorithm for the processing and aggregation mass data.
//
// A type implementing the MapReducer interface has to be implemented
// and passed to the MapReduce() function. The type is responsible
// for the input, the mapping, the reducing and the consuming while
// the package provides the runtime environment for it.
package mapreduce

// EOF
