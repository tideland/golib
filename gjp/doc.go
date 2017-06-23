// Tideland Go Library - Generic JSON Parser
//
// Copyright (C) 2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// Package gjp of the Tideland Go Library package provides the parsing
// and accessing of a JSON document content by paths. The values are
// returned typed, also a default has to be provided. The path separator
// can be defined when parsing the document.
//
//     doc, err := gjp.Parse(myDoc, "/")
//     if err != nil {
//         ...
//     }
//     name := doc.ValueAsString("name", "")
//     street := doc.ValueAsString("address/street", "unknown")
//
// Additionally values of the document can be processed using
//
//     err := doc.Process(func(path string, value gjp.Value) error {
//         ...
//     })
//
// Sometimes one is more interested in the differences between two
// documents. Here
//
//     diff, err := gjp.Compare(firstDoc, secondDoc, "/")
//
// privides a gjp.Diff instance which helps to compare individual
// paths of the two document.
package gjp

// EOF
