# Tideland Go Library

## Description

The *Tideland Go Library* contains a larger set of useful Google Go packages
for different purposes.

**ATTENTION:** The `cells` package has been migrated into an own repository
at [https://github.com/tideland/gocells](https://github.com/tideland/gocells).

**ATTENTION:** The `web` package is now deprecated. It has been migrated
and extended into the repository
[https://github.com/tideland/gorest](https://github.com/tideland/gorest).

I hope you like them. ;)

[![GoDoc](https://godoc.org/github.com/tideland/golib?status.svg)](https://godoc.org/github.com/tideland/golib)
[![Sourcegraph](https://sourcegraph.com/github.com/tideland/golib/-/badge.svg)](https://sourcegraph.com/github.com/tideland/golib?badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/tideland/golib)](https://goreportcard.com/report/github.com/tideland/golib)

## Version

Version 4.23.0

## Packages

### Audit

Support for unit tests with mutliple different assertion types and functions
to generate test data.

### Cache

Individual caches for types implementing the Cacheable interface.

### Collections

Different additional collection types like ring buffer, stack, tree, and more.

### Errors

Detailed error values.

### Etc

Reading and parsing of SML-formatted configurations including substituion
of templates.

### Feed

Atom and RSS feed client.

### Generic JSON Parser

Instead of unmarshalling a JSON into a struct parse it and provide access
to the content by path and value converters to native types.

### Identifier

Identifier generation, like UUIDs or composed values.

### Logger

Flexible logging.

### Loop

Control of goroutines and their possible errors. Additional option of recovering
in case of an error or a panic. Sentinels can monitor multiple loops and restart
them all in case of an abnormal end of one of them.

### Map/Reduce

Map/Reduce for data analysis.

### Monitoring

Monitoring of execution times, stay-set indicators, and configurable system variables.

### Numerics

Different functions for statistical analysis.

### Redis Client

Client for the Redis database.

### Scene

Context-based shared data access, e.g. for web sessions or in cells.

### Scroller

Continuous filtered reading/writing of data.

### SML

Simple Markup Language, looking lispy, only with curly braces.

### Sort

Parallel Quicksort.

### Stringex

Helpful functions around strings extending the original `strings` package and
help processing strings.

### Timex

Helpful functions around dates and times.

### Version

Documentation of semantic versions.

## Contributors

- Frank Mueller (https://github.com/themue / https://github.com/tideland)
- Alex Browne (https://github.com/albrow)
- Tim Heckman (https://github.com/theckman)
- Benedikt Lang (https://github.com/blang)
- Pellaeon Lin (https://github.com/pellaeon)

## License

*Tideland Go Library* is distributed under the terms of the BSD 3-Clause license.
