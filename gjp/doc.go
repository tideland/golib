// Tideland Go Library - Generic JSON Processor
//
// Copyright (C) 2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// Package gjp of the Tideland Go Library package provides the
// generic parsing and processing of JSON documents by paths. The
// returned values are typed, also a default has to be provided.
// The path separator for accessing can be defined when parsing
// a document.
//
//     doc, err := gjp.Parse(myDoc, "/")
//     if err != nil {
//         ...
//     }
//     name := doc.ValueAt("name").AsString("")
//     street := doc.ValueAt("address/street").AsString("unknown")
//
// The value passed to AsString() and the others are default values if
// there's none at the path. Another way is to create an empty document
// with
//
//     doc := gjp.NewDocument("::")
//
// Here and at parsed documents values can be set with
//
//     err := doc.SetValueAt("a/b/3/c", 4711)
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
