/*
Package bender load tests services that use request/reply protocols like Thrift, Protobuf or HTTP.

Bender supports two different approaches to load testing. The first, LoadTestThroughput, gives the
caller control over when requests are sent and starts as many goroutines as necessary to send those
requests. The second, LoadTestConcurrency, gives the caller control over how many goroutines are
used, and sends as many requests as it can using those goroutines.

LoadTestThroughput can simulate the load caused by concurrent users interacting with a service, as
is the case for web services, or internal and external remote APIs. It is typically used to simulate
a target QPS, and to measure the latency and error rate of a service at that QPS. The load tester
will not stop sending requests because the service slows down or has errors, making it a good way
to test actual behavior under load. This is usually the right approach for load testing user facing
services.

LoadTestConcurrency is most useful for testing a services ability to handle many concurrent
connections. It can be used to simulate a fixed or varying number of connections and to measure the
throughput, latency and error rate of the service. The individual connections will loop, sending
a single request, waiting for the response and then repeating. As a result, if the service slows
down, the load tester will also slow down because the connections will all be waiting for responses.
That makes this approach a poor way to test latency and throughput, but a good way to test resource
limitations caused by many concurrent connections.

LoadTestThroughput

The LoadTestThroughput function takes four arguments. The first is a function that generates
nanosecond intervals which are used as request arrival times. The second is a channel of requests,
each of which has an ID and the actual request. The third is a function that knows how to send a
request and validate the response. The inner loop of LoadTestThroughput looks like this:

  for {
      interval := intervals(time.Now().UnixNanos())
      time.Sleep(time.Duration(interval)
      request := <-requests
      go func() {
        err := requestExec(time.Now().UnixNano(), request)
      }()
  }

The fourth argument to LoadTestThroughput is a channel which is used to output events. There are
events for the start and end of the load test, the sending of each request and the receiving of
each response and the wait time between sending requests. The wait message includes an "overage"
time which is useful for monitoring the health of the load test program and underlying OS and host.
The overage time measures the difference between the expected wait time (the interval time) and the
actual wait time. On a heavily loaded host, or when there are long GC pauses, that difference can
be large. Bender attempts to compensate for the overage by reducing the subsequent wait times, but
under heavy load, the overage will continue to increase until it cannot be compensated for. At that
point the wait events will report a monotonically increasing overage which means the load test
isn't keeping up with the desired throughput.

A load test ends when the request channel is closed and all remaining requests in the channel have
been executed.

LoadTestConcurrency

The LoadTestConcurrency function takes four arguments. The first is a semaphore that controls the
maximum number of concurrently executing requests, and makes it possible to dynamically control that
number over the lifetime of the load test. The second, third and fourth arguments are identical to
those for LoadTestThroughput. The inner loop of LoadTestConcurrency does something like this:

  for {
      workerSem.Wait(1)
      request := <-requests
      go func() {
          err := requestExec(time.Now().UnixNano(), request)
          workerSem.Signal(1)
      }
  }

Reducing the semaphore count will reduce the number of running connections as existing connections
complete, so there can be some lag.

As with LoadTestThroughput, the load test ends when the request channel is closed and all remaining
requests have been executed.

Interval Generators

An IntervalGenerator is a function that takes the current Unix epoch time (in nanoseconds) and
returns a non-negative time (also in nanoseconds) until the next request should be sent. Bender
provides functions to create interval generators for uniform and exponential distributions, each
of which takes the target throughput (requests per second) and returns an IntervalGenerator. Neither
of the included generators makes use of the function argument, but it is there for cases in which
the simulated intervals are time dependent (you want to simulate the daily traffice variation of a
web site, for example).

Request Channels

The request channel decouples creation of requests from execution of requests and allows them to
run concurrently. A typical approach to creating a request channel is code like this:

 c := make(chan *Request)
 go func() {
     for {
		 // create service request r with request ID rid
		 c <- &Request{rid, r}
     }
     close(c)
 }()

Requests can be generated randomly, read from files (like access logs) or generated any other way
you like. The important part is that the request generation be done in a separate goroutine that
communicates with the load tester via a channel. In addition, the channel must be closed to indicate
that the load test is done.

TODO(charles): how to set the buffer size on this channel? Any other performance notes?

Request Executors

A request executor is a function that takes the current Unix Epoch time (in nanoseconds) and a
*Request, sends the request to the service, waits for the response, validates it and returns an
error or nil. This function is timed by the load tester, so it should do as little else as possible,
and everything it does will be added to the reported service latency. Here, for example, is a very
simple request executor for HTTP requests:

 func HttpRequestExecutor(_ int64, request *Request) error {
     url := request.Request.(string)
     _, err := http.Get(url)
     return err
 }

The http package in Bender provides a function that generates executors that make use of the http
packages Transport and Client classes and provide an easy way to validate the body of the http
request.

Event Messages

The output of both LoadTestThroughput and LoadTestConcurrency is a channel of event messages.

TODO(charles): finish this
*/
package bender

