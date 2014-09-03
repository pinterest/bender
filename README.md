bender
======

Bender is a library that makes it easy to build custom load testing applications for request/reply
protocols like HTTP, Thrift and Protobuf (among many, many others). You write a Go client for your
service and generate requests (any way you like); Bender handles sending the requests and measuring
the results.

That Bender is a library makes it flexible and easy to extend, but gives it a slightly longer
learning curve. Bender provides two different approaches to load testing (by throughput or by
concurrency) along with utilities for generating Thrift and HTTP clients and requests. It also
provides a flexible way to capture events as they happen, and record them in a histogram and via a
log file. The underlying philosophy of the library is to provide powerful, orthogonal primitives and
to let programmers use Go to combine them. To make that easier, Bender has extensive documentation
and tutorials which you can find below.

*WARNING:* this library is not ready for production. We are busy "dogfooding" it right now, and will
update this README when it is ready. Until then, expect huge changes in the API!

## Getting Started

The easiest way to get started with Bender is to use one of the tutorials:

* [Thrift](https://github.com/pinterest/bender/blob/master/thrift/TUTORIAL.md)
* [HTTP](https://github.com/pinterest/bender/blob/master/http/TUTORIAL.md)

## Documentation

The package documentation is available on [godoc.org](http://godoc.org/github.com/pinterest/bender).
The function and data structure documentation is also available there.

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
