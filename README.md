bender
======

Bender makes it easy to build load testing applications for services using protocols like HTTP,
Thrift, Protocol Buffers and many more. Bender provides a library of flexible, powerful primitives
that can be combined (with plain Go code) to build load testers customized to any use case and that
evolve with your service over time.

Bender provides two different approaches to load testing. The first, LoadTestThroughput, gives the
tester control over the throughput (QPS), but not over the concurrency. This one is very well
suited for services that are open to the Internet, like web services, and even backend Thrift or
Protocol Buffer services, since it will just keep sending requests, even if the service is
struggling. The second approach, LoadTestConcurrency, gives the tester control over the concurrency,
but not over the throughput. This approach is better suited to testing services that require lots of
concurrent connections, and need to be tested for resource limits.

That Bender is a library makes it flexible and easy to extend, but means it takes longer to create
an initial load tester. As a result, we've focused on creating easy-to-follow tutorials that are
written for people unfamiliar with Go, along with documentation for all the major functions in the
library.

## Getting Started

The easiest way to get started with Bender is to use one of the tutorials:

* [Thrift](https://github.com/pinterest/bender/blob/master/thrift/TUTORIAL.md)
* [HTTP](https://github.com/pinterest/bender/blob/master/http/TUTORIAL.md)

## Documentation

The package documentation is available on [godoc.org](http://godoc.org/github.com/pinterest/bender).
The function and data structure documentation is also available there.

## What Is Missing

Bender does not provide any support for sending load from more than one machine. If you need to
send more load than a single machine can handle, or you need the requests to come from multiple
physical hosts (or different networks, or whatever), you currently have to write your own tools. In
addition, the histogram implementation used by Bender is inefficient to send over the network,
unlike q-digest or t-digest, which we hope to implement in the future.

Bender does not provide any visualization tools, and has a relatively simple set of measurements,
including a customizable histogram of latencies, an error rate and some other summary statistics.
Bender does provide a complete log of everything that happens during a load test, so you can use
existing tools to graph any aspect of that data, but nothing in Bender makes that easy right now.

Bender only provides helper functions for HTTP and Thrift currently, because that is all we use
internally at Pinterest.

The load testers we have written internally with Bender have a lot of common command line arguments,
but we haven't finalized a set to share as part of the library.

The documentation doesn't provide guidance for tuning channel buffer sizes or Linux TCP tunables,
and these can make a big difference in the throughput from a single host.

## Comparison to Other Load Testers

#### JMeter

JMeter provides a GUI to configure and run load tests, and can also be configured via XML (really,
really not recommended by hand!) and run from the command line. JMeter uses the same approach as
LoadTestConcurrency in Bender, which is not a good approach to load testing services (see the Bender
docs and the Iago philosophy for more details on why that is). It isn't easy to extend JMeter to
handle new protocols, so it doesn't have support for Thrift or Protobuf. It is relatively easy to
extend other parts of JMeter by writing Java code, however, and the GUI makes it easy to plug all
the pieces together.

#### Iago

Iago is Twitter's load testing library and it is the inspiration for Bender's LoadTestThroughput
function. Iago is a Scala library written on top of Netty and the Twitter Finagle libraries. As a
result, Iago is powerful, but difficult to understand, extend and configure. It was frustration with
making Iago work that led to the creation of Bender.

#### The Grinder

The Grinder has the same load testing approach as JMeter, but allows scripting via Jython, which
makes it more flexible and extensible. The Grinder uses threads, which limits the concurrency at
which it can work, and makes it hard to implement things like Bender's LoadTestThroughput function.
The Grinder does have support for conveniently running distributed load tests.

## Attribution

Bender includes open source from the following sources:

Apache Thrift Libraries
Copyright 2014 Apache Software Foundation. Licensed under the Apache License v2.0 (http://www.apache.org/licenses/).

Go Libraries
Copyright 2012 The Go Authors. Licensed under the BSD license (http://golang.org/LICENSE). 
