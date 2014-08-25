bender
======

Bender is a library for building load testing programs. It currently has support and documentation
for Thrift and Http, but can be extended to support any protocol.

*WARNING:* this library is not ready for production. We are busy "dogfooding" it right now, and will
update this README when it is ready. Until then, expect huge changes in the API!

## Getting Started

The easiest way to get started with Bender is to use one of the tutorials:

* [Thrift](http://github.com/Pinterest/bender/thrift/TUTORIAL.md)
* [HTTP](http://github.com/Pinterest/bender/http/TUTORIAL.md)

The rest of this document discusses the design of Bender, the core library functions and how they
can be extended. It also briefly covers the differences between Bender and other load testing
libraries and programs, and why/when you would want to use Bender.

## Design

Bender is a [Go](http://golang.org) library that makes it easy to build load testers for any kind
of service over just about any protocol. The core of the library supports two different approaches
to load testing. The first, `LoadTestThroughput`, gives you control over how often requests are
sent (i.e., the total QPS) and starts as many goroutines as necessary to send that load. The second,
`LoadTestConcurrency`, gives you control over how many goroutines are run and how fast each can send
requests.

It is usually correct to use `LoadTestThroughput`, since it better simulates the kind of load a
service receives from distributed clients, who aren't going to slow down their request rate just
because your service is slowing down. The other method, `LoadTestConcurrency`, is better for testing
the connection limits of services that must accept a lot of connections, but isn't good for testing
latency or throughput, since the individual goroutines are naturally throttled by the performance
of the service.

Bender is modeled after the [Iago](http://twitter.github.io/iago/philosophy.html) library, which
uses the same approach as `LoadTestThroughput`. We found Iago challenging to use as a library, and
it relies heavily on the Twitter Finagle libraries, which require a lot of dependencies. Bender, on
the other hand, just relies on Go's built-in goroutines, requires significantly less code and
configuration, and produces much smaller binaries. It is just as easy to build load testers and,
in our opinion, even easier to extend them with your own code.

This library is structured around two primitives, `LoadTestThroughput` and `LoadTestConcurrency`,
which are described below. In addition, the library provides convenience functions for building
load testers that use the Thrift or HTTP protocol (more to come in the future). Finally, the library
includes tools for logging events and times, as well as basic histograms for viewing percentiles,
averages and other statistics for each run.

## LoadTestThroughput

The `LoadTestThroughput` function takes three arguments:

1. A `chan int` that provides successive intervals in nanoseconds. The inner loop reads the next
interval from this channel, sleeps for that duration and then sends the next request. If this 
channel is closed, the inner loop will shutdown, no new requests will be executed and the currently
running requests are waited on.
2. A `chan *Request` that provides requests to be executed. The `Request` struct contains an integer
ID, which identifies the request through the pipeline (this can be anything, even a random number,
as long as it uniquely identifies the request), and a request, which is an `interface{}` and can
contain anything you can use to execute a request against your service (a Thrift `TStruct` or an
`http.Request`, for instance). If this channel is closed the inner loop will shutdown, no new
requests will be executed and the currently running requests are waited on.
3. A function that takes a `*Request` and returns an `error`. This is the execution function, which
will be used by the inner loop to execute requests in a goroutine (which will also include code to
time the request and report the results). This function takes the most recent `*Request` from the
request channel and returns `nil` if there was no error, and an `error` otherwise.

The `LoadTestThroughput` function returns a `chan interface{}` which is used to send event messages.
The following messages are sent:

1. `StartMsg`: sent when the load tester starts, and includes the Unix epoch in nanoseconds.
2. `EndMsg`: sent when either the intervals or requests channel has closed and all running requests
have completed. Contains the Unix epoch in nanoseconds for the start and end times.
3. `WaitMsg`: sent every time the inner loop sleeps and includes the sleep interval (taken from
the intervals channel) and the current "overage" (see below).
4. `StartRequestMsg`: sent right before a request is executed using the function provided to the
library. Includes the Unix epoch in nanoseconds and the request ID.
5. `EndRequestMsg`: sent when the request has finished executing. Includes the Unix epoch in
nanoseconds for the start and end of the request, the request ID and any `error` returned by the
executor (or `nil` if there was no error).

The "overage" number reported by `WaitMsg` is a measure of how overloaded the load tester is on the
local computer. The inner loop sleeps for the next interval and then checks how long it actually
slept. If it slept longer than expected, it increments the overage by that amount. You will almost
certainly see this number fluctuate over the course of a load test, particularly if you are sending
high QPS. You only need to worry when the number is increasing monotonically, as that indicates you
are sending too much load for the computer on which the load tester is running.

## LoadTestConcurrency

TBD
