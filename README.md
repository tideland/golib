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

[![Sourcegraph](https://sourcegraph.com/github.com/tideland/golib/-/badge.svg)](https://sourcegraph.com/github.com/tideland/golib?badge)

## Version

Version 4.21.4

## Packages

### Audit

Support for unit tests with mutliple different assertion types and functions
to generate test data.

[![GoDoc](https://godoc.org/github.com/tideland/golib/audit?status.svg)](https://godoc.org/github.com/tideland/golib/audit)

### Cache

Lazy Loading and Caching of Values.

[![GoDoc](https://godoc.org/github.com/tideland/golib/cache?status.svg)](https://godoc.org/github.com/tideland/golib/cache)

### Collections

Different additional collection types like ring buffer, stack, tree, and more.

[![GoDoc](https://godoc.org/github.com/tideland/golib/collections?status.svg)](https://godoc.org/github.com/tideland/golib/collections)

### Errors

Detailed error values.

[![GoDoc](https://godoc.org/github.com/tideland/golib/errors?status.svg)](https://godoc.org/github.com/tideland/golib/errors)

### Etc

Reading and parsing of SML-formatted configurations including substituion
of templates.

[![GoDoc](https://godoc.org/github.com/tideland/golib/etc?status.svg)](https://godoc.org/github.com/tideland/golib/etc)

### Feed

Atom feed client.

[![GoDoc](https://godoc.org/github.com/tideland/golib/feed/atom?status.svg)](https://godoc.org/github.com/tideland/golib/feed/atom)

RSS feed client.

[![GoDoc](https://godoc.org/github.com/tideland/golib/feed/rss?status.svg)](https://godoc.org/github.com/tideland/golib/feed/rss)

### Identifier

Identifier generation, like UUIDs or composed values.

[![GoDoc](https://godoc.org/github.com/tideland/golib/identifier?status.svg)](https://godoc.org/github.com/tideland/golib/identifier)

### Logger

Flexible logging.

[![GoDoc](https://godoc.org/github.com/tideland/golib/logger?status.svg)](https://godoc.org/github.com/tideland/golib/logger)

### Loop

Control of goroutines and their possible errors. Additional option of recovering
in case of an error or a panic. Sentinels can monitor multiple loops and restart
them all in case of an abnormal end of one of them.

[![GoDoc](https://godoc.org/github.com/tideland/golib/loop?status.svg)](https://godoc.org/github.com/tideland/golib/loop)

### Map/Reduce

Map/Reduce for data analysis.

[![GoDoc](https://godoc.org/github.com/tideland/golib/mapreduce?status.svg)](https://godoc.org/github.com/tideland/golib/mapreduce)

### Monitoring

Monitoring of execution times, stay-set indicators, and configurable system variables.

[![GoDoc](https://godoc.org/github.com/tideland/golib/monitoring?status.svg)](https://godoc.org/github.com/tideland/golib/monitoring)

### Numerics

Different functions for statistical analysis.

[![GoDoc](https://godoc.org/github.com/tideland/golib/numerics?status.svg)](https://godoc.org/github.com/tideland/golib/numerics)

### Redis Client

Client for the Redis database.

[![GoDoc](https://godoc.org/github.com/tideland/golib/redis?status.svg)](https://godoc.org/github.com/tideland/golib/redis)

### Scene

Context-based shared data access, e.g. for web sessions or in cells.

[![GoDoc](https://godoc.org/github.com/tideland/golib/scene?status.svg)](https://godoc.org/github.com/tideland/golib/scene)

### Scroller

Continuous filtered reading/writing of data.

[![GoDoc](https://godoc.org/github.com/tideland/golib/scroller?status.svg)](https://godoc.org/github.com/tideland/golib/scroller)

### SML

Simple Markup Language, looking lispy, only with curly braces.

[![GoDoc](https://godoc.org/github.com/tideland/golib/sml?status.svg)](https://godoc.org/github.com/tideland/golib/sml)

### Sort

Parallel Quicksort.

[![GoDoc](https://godoc.org/github.com/tideland/golib/sort?status.svg)](https://godoc.org/github.com/tideland/golib/sort)

### Stringex

Helpful functions around strings extending the original `strings` package.

[![GoDoc](https://godoc.org/github.com/tideland/golib/stringex?status.svg)](https://godoc.org/github.com/tideland/golib/stringex)

### Timex

Helpful functions around dates and times.

[![GoDoc](https://godoc.org/github.com/tideland/golib/timex?status.svg)](https://godoc.org/github.com/tideland/golib/timex)

### Version

Documentation of semantic versions.

[![GoDoc](https://godoc.org/github.com/tideland/golib/version?status.svg)](https://godoc.org/github.com/tideland/golib/version)

## Contributors

- Frank Mueller (https://github.com/themue / https://github.com/tideland)
- Alex Browne (https://github.com/albrow)
- Tim Heckman (https://github.com/theckman)
- Benedikt Lang (https://github.com/blang)
- Pellaeon Lin (https://github.com/pellaeon)

## License

*Tideland Go Library* is distributed under the terms of the BSD 3-Clause license.
