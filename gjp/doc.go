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
// After reading this from a file, reader, or string the number of users
// can be retrieved with a default value of 10 by calling
package gjp

// EOF
