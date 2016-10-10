/*
Copyright 2014-2016 Pinterest, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package bender makes it easy to build load testing applications for services using protocols like
HTTP, Thrift, Protocol Buffers and many others.

Bender provides two different approaches to load testing. The first, LoadTestThroughput, gives the
tester control over the throughput (QPS), but not over the concurrency (number of goroutines). The
second, LoadTestConcurrency, gives the tester control over the concurrency, but not over the
throughput.

LoadTestThroughput simulates the load caused by concurrent clients sending requests to a service. It
can be used to simulate a target throughput (QPS) and to measure the request latency and error rate
at that throughput. The load tester will keep spawning goroutines to send requests, even if the
service is sending errors or hanging, making this a good way to test the actual behavior of the
service under heavy load. This is the same approach used by Twitter's Iago library, and is nearly
always the right place to start when load testing services exposed (directly or indirectly) to the
Internet.

LoadTestConcurrency simulates a fixed number of clients, each of which sends a request, waits for a
response and then repeats. The downside to this approach is that increased latency from the service
results in decreased throughput from the load tester, as the simulated clients are all waiting for
responses. That makes this a poor way to test services, as real-world traffic doesn't behave this
way. The best use for this function is to test services that need to handle a lot of concurrent
connections, and for which you need to simulate many connections to test resource limits, latency
and other metrics. This approach is used by load testers like the Grinder and JMeter, and has been
critiqued well by Gil Tene in his talk "How Not To Measure Latency".

The next two sections provide more detail on the implementations of LoadTestThroughput and
LoadTestConcurrency. The following sections provide descriptions for the common arguments to the
load testing functions, and how they work, including the interval generators, request generators,
request executors and event recorders.

LoadTestThroughput

The LoadTestThroughput function takes four arguments. The first is a function that generates
nanosecond intervals which are used as request arrival times. The second is a channel of requests.
The third is a function that knows how to send a request and validate the response. The inner loop
of LoadTestThroughput looks like this:

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
complete, so there can be some lag between calling workerSem.Wait(n) and the number of running
connections actually decreasing by n. The worker semaphore does not protect you from reducing the
number of workers below zero, which will cause undefined behavior from the load tester.

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

The requests channel should almost certainly be buffered, unless you can generate requests much
faster than they are sent (and not just on average). The easiest way to miss your target throughput
with LoadTestThroughput is to be blocked waiting for requests to be generated, particularly when
testing a large throughput.

Request Executors

A request executor is a function that takes the current Unix Epoch time (in nanoseconds) and a
*Request, sends the request to the service, waits for the response, optionally validates it and
returns an error or nil. This function is timed by the load tester, so it should do as little else
as possible, and everything it does will be added to the reported service latency. Here, for
example, is a very simple request executor for HTTP requests:

 func HttpRequestExecutor(_ int64, request *Request) error {
     url := request.Request.(string)
     _, err := http.Get(url)
     return err
 }

The http package in Bender provides a function that generates executors that make use of the http
packages Transport and Client classes and provide an easy way to validate the body of the http
request.

RequestExecutors are called concurrently from multiple goroutines, and must be concurrency-safe.

Event Messages

The LoadTestThroughput and LoadTestConcurrency functions both take a channel of events (represented
as interface{}) as a parameter. This channel is used to output events as they happen during the load
test, including the following events:

StartEvent: sent once at the start of the load test.

EndEvent: sent once at the end of the load test, no more events are sent after this.

WaitEvent: sent only for LoadTestThroughput, see below for details.

StartRequestEvent: sent before a request is sent to the service, includes the request and the
event time. Note that the event time is not the same as the start time for the request for
stupid performance reasons. If you need to know the actual start time, see the EndRequestEvent.

EndRequestEvent: sent after a request has finished, includes the response, the actual start and
end times for the request and any error returned by the RequestExecutor.

The WaitEvent includes the time until the next request is sent (in nanoseconds) and an "overage"
time. When the inner loop sleeps, it subtracts the total time slept from the time it intended to
sleep, and adds that to the overage. The overage, therefore, is a good proxy for how overloaded the
load testing host is. If it grows over time, that means the load test is falling behind, and can't
start enough goroutines to run all the requests it needs to. In that case you will need a more
powerful load testing host, or need to distribute the load test across more hosts.

The event channel doesn't need to be buffered, but it may help if you find that Bender isn't sending
as much throughput as you expect. In general, this depends a lot on how quickly you are consuming
events from the channel, and how quickly the load tester is running. It is a good practice to
proactively buffer this channel.
*/
package bender
