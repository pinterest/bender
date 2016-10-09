Bender Thrift Tutorial
======================

This tutorial walks through the steps to create a simple "Hello, World" Thrift service with a load
tester. It assumes no prior knowledge of Go or Thrift. If you already know how to use Thrift with
Go, skip down to the "Load Testing" section to get started with Bender and Thrift.

## Getting Started

You will need a copy of Thrift installed on your machine, which will allow you to run the `thrift`
command. You can follow the "Getting Started" instructions on the
[Apache Thrift](https://thrift.apache.org/) page to download and install it. These instructions were
written with Thrift version 0.9.1 and have not been tested with other versions, please send a pull
request if you find issues with other Thrift versions!

Next you will need to install and configure Go by following the instructions on the
[Getting Started](https://golang.org/doc/install) page. Then follow the instructions on the
[How To Write Go Code](https://golang.org/doc/code.html) page, particularly for setting up your
workspace and the `GOPATH` environment variable, which we will use throughout this tutorial.

Next we need the Apache Thrift libraries for Go, which you can fetch using `go get` as:

```
cd $GOPATH
go get git.apache.org/thrift.git/lib/go/thrift
```

Note, this command downloads a lot of data, so it can take awhile. You should see a directory in
src named git.apache.org.

Finally you will need the latest version of Bender, which you can get by running:

```
go get github.com/pinterest/bender
```

## Writing the Thrift Server and Client

This section will walk through the creation of a Thrift client and server, which we will use to
test Bender in the following section. If you already have a Thrift definition file and server, you
can use them instead, but should still follow the instructions for creating the import paths to the
generated Thrift client for Go. Those instructions should work for any Thrift client.

In the following, all commands should be run from the `$GOROOT` directory, unless otherwise noted.

### Create the Go Package

Create a new Go package for your Thrift service and client. We'll refer to this as `$PKG` in this
document, and it can be any path you want. At Pinterest, for example, we use `github.com/pinterest`.

```
cd $GOPATH
mkdir -p src/$PKG/hellothrift
```

### Thrift Service Definition and Code Generation

Now create a file named `src/$PKG/hellothrift/hello.thrift` and add these lines to it using your
text editor:

```thrift
struct HelloRequest {
    1: optional string message;
}

struct HelloResponse {
    1: optional string message;
}

service Hello {
    HelloResponse hello(1: HelloRequest request);
}
```

This defines a Thrift service with one API endpoint named `hello` that takes a `HelloRequest` and
returns a `HelloResponse`. Now we need to generate the Go code for the service, which we can do
using these commands:

```
thrift --out src/$PKG/hellothrift --gen go:package_prefix=$PKG/hellothrift src/$PKG/hellothrift/hello.thrift
```

This will create a directory `src/$PKG/hellothrift/hello` that contains the following files:

* `hello.go` - the interfaces for the client and server, and the serialization logic for the service
arguments.
* `ttypes.go` - functions to create request and response types.
* `constants.go` - empty for this service, but normally contains any constant definitions.

The thrift command above will work for multiple Thrift files and includes, although you may need to
use the "-I PATH" argument to add include paths for other Thrift files. The `package_prefix`
argument isn't necessary for this small service, but is important when using includes and multiple
Thrift files.

### Thrift Service Implementation

Now we will create a simple service definition that just echoes the request string to the response.
First, create a new directory:

```
mkdir -p src/$PKG/hellothrift/server
```

Then create a file named `main.go` in that directory and add these lines to it:

```go
package main

import (
	"$PKG/hellothrift/hello"
	"git.apache.org/thrift.git/lib/go/thrift"
	"fmt"
	"time"
)

type HelloHandler struct {

}

func (*HelloHandler) Hello(request *hello.HelloRequest) (*hello.HelloResponse, error) {
	resp := hello.NewHelloResponse()
	resp.Message = request.Message
	fmt.Printf("%d - %s\n", time.Now().UnixNano(), request.Message)
	return resp, nil
}

func NewHelloHandler() hello.Hello {
	return new(HelloHandler)
}

func RunServer(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string) error {
	transport, err := thrift.NewTServerSocket(addr)
	if err != nil {
		return err
	}

	handler := NewHelloHandler()
	processor := hello.NewHelloProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)

	return server.Serve()
}

func main() {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	addr := "localhost:3636"
	RunServer(transportFactory, protocolFactory, addr)
}
```

Make sure to change `$PKG` to the name of your package, this won't compile as-is.

### Thrift Client Implementation

Now we will create a simple client library and a command line tool to test the server we created in
the last step. First, create a new directory:

```
mkdir -p src/$PKG/hellothrift/client
```

Then create a file named `main.go` in that directory and add these lines to it:

```go
package main

import (
	"fmt"
	"$PKG/hellothrift/hello"
	"git.apache.org/thrift.git/lib/go/thrift"
)

func RunClient(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string) error {
	socket, err := thrift.NewTSocket(addr)
	if err != nil {
		return err
	}

	transport := transportFactory.GetTransport(socket)
	defer transport.Close()
	if err := transport.Open(); err != nil {
		return err
	}

	client := hello.NewHelloClientFactory(transport, protocolFactory)
	request := hello.NewHelloRequest()
	request.Message = "hello, world!"
	response, err := client.Hello(request)
	if err != nil {
		return err
	}
	fmt.Println(response.Message)

	return nil
}

func main() {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	addr := "localhost:3636"
	err := RunClient(transportFactory, protocolFactory, addr)
	if err != nil {
	    panic(err)
	}
}
```

Make sure to change `$PKG` to the name of your package, this won't compile as-is.

### Run Server and Client

Run these commands to compile and install the server and client:

```
go install $PKG/hellothrift/hello
go install $PKG/hellothrift/server
go install $PKG/hellothrift/client
```

In one terminal window, run the server as `./bin/server` and in another terminal window run the
client as `./bin/client`. You should see "hello, world!" print in both terminals.

## Load Testing

Now that we have a Thrift server we can use Bender to build a simple load tester for it. This
section uses the same directories and packages as the previous section, but it is easy to use the
same instructions for any Thrift server, just replace `$PKG` with the package in which you have your
Thrift server code. The next few sections walk through the various parts of the load tester. If you
are in a hurry skip to the section "Final Load Tester Program" and just follow the instructions from
there.

### Intervals

The first thing we need is a function to generate intervals (in nanoseconds) between executing
requests. The Bender library comes with some predefined intervals, including a uniform distribution
(always wait the same amount of time between each request) and an exponential distribution. In this
case we will use the exponential distribution, which means our server will experience load as
generated by a [Poisson process](http://en.wikipedia.org/wiki/Poisson_process), which is fairly
typical of server workloads on the Internet (with the usual caveats that every service is a special
snowflake, etc, etc). We get the interval function with this code:

```go
intervals := bender.ExponentialIntervalGenerator(qps)
```

Where `qps` is our desired throughput measured in queries per second. It is also the reciprocal of
the mean value of the exponential distribution used to generate the request arrival times (see the
wikipedia article above). In practice this means you will see an average QPS that fluctuates around
the target QPS (with less fluctuation as you increase the time interval over which you are
averaging).

### Request Generator

The second thing we need is a channel of requests to send to the server. When an interval has been
generated and Bender is ready to send the request, it pulls the next request from this channel and
spawns a goroutine to send the request to the server. This function creates a simple synthetic
request generator:

```go
func SyntheticHelloRequests(n int) chan interface{} {
	c := make(chan interface{}, 100)
	go func() {
		for i := 0; i < n; i++ {
			request := hello.NewHelloRequest()
			request.Message = "hello"
			c <- request
		}
		close(c)
	}()
	return c
}
```

This creates a separate goroutine to generate the requests and send them on the channel so that
request generation works concurrently with request execution. That means the inner loop of Bender
isn't spending time creating requests (which could be expensive if we're reading them from disk or
from a remote database, for instance). You can modify this function to read requests from anywhere
including server query logs, remote services or databases. At Pinterest, for instance, our Thrift
servers log the binary Thrift requests to disk before deserializing them on the server, so we have
a request generator that reads those files and reconstructs the production requests from them.

### Request Executor

The next thing we need is a request executor, which takes the requests generated above and sends
them to the service. We will use a helper function from Bender's thrift library to do most of the
work (connection management, error handling, etc), so all we have to do is write code to send the
request:

```go
func HelloExecutor(request interface{}, transport thrift.TTransport) (interface{}, error) {
	pFac := thrift.NewTBinaryProtocolFactoryDefault()
	client := hello.NewHelloClientFactory(transport, pFac)
	return client.Hello(request.(*hello.HelloRequest))
}
```

### Recording Results

The last thing we need is a channel that will output events as the load tester runs. This will let
us listen to the load testers progress and record stats. We want this channel to be buffered so that
we can run somewhat independently of the load test without slowing it down:

```go
recorder := make(chan interface{}, 128)
```

The `LoadTestThroughput` function returns a channel on which it will send events for things like
the start of the load test, how long it waits between requests, how much overage it is currently
experiencing, and when requests start and end, how long they took and whether or not they had
errors. That raw event stream makes it possible to analyze the results of a load test. Bender has
a couple simple "recorders" that provide basic functionality for result analysis and all of which
use the `Record` function:

* `NewLoggingRecorder` creates a recorder that takes a `log.Logger` and outputs each event to it in
a well-defined format.
* `NewHistogramRecorder` creates a recorder that manages a histogram of latencies from requests and
error counts.

You can combine recorders using the `Record` function, so you can both log events and manage a
histogram using code like this:

```go
l := log.New(os.Stdout, "", log.LstdFlags)
h := hist.NewHistogram(60000, 1000000)
bender.Record(recorder, bender.NewLoggingRecorder(l), bender.NewHistogramRecorder(h))
```

The histogram takes two arguments: the number of buckets and a scaling factor for times. In this
case we are going to record times in milliseconds and allow 60,000 buckets for times up to one
minute. The scaling factor is 1,000,000 which converts from nanoseconds (the timer values) to
milliseconds.

It is relatively easy to build recorders, or to just process the events from the channel yourself,
see the Bender documentation for more details on what events can be sent, and what data they
contain.

### Final Load Tester Program

Create a directory for the load tester:

```
mkdir -p src/$PKG/hellobender
```

Then create a file named `main.go` in that directory and add these lines to it:

```go
package main

import (
	"github.com/pinterest/bender"
	bthrift "github.com/pinterest/bender/thrift"
	"git.apache.org/thrift.git/lib/go/thrift"
	"log"
	"os"
	"github.com/pinterest/bender/hist"
	"fmt"
	"time"
	"$PKG/hellothrift/hello"
)

func SyntheticRequests(n int) chan interface{} {
	c := make(chan interface{}, 100)
	go func() {
		for i := 0; i < n; i++ {
			request := hello.NewHelloRequest()
			request.Message = "hello " + i
			c <- request
		}
		close(c)
	}()
	return c
}

func HelloExecutor(request interface{}, transport thrift.TTransport) (interface{}, error) {
	pFac := thrift.NewTBinaryProtocolFactoryDefault()
	client := hello.NewHelloClientFactory(transport, pFac)
	return client.Hello(request.(*hello.HelloRequest))
}

func main() {
	intervals := bender.ExponentialIntervalGenerator(10.0)
	requests := SyntheticRequests(10)
	exec := bthrift.NewThriftRequestExec(thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory()), HelloExecutor, 10 * time.Second, "localhost:3636")
	recorder := make(chan interface{}, 128)
	bender.LoadTestThroughput(intervals, requests, exec, recorder)
	l := log.New(os.Stdout, "", log.LstdFlags)
	h := hist.NewHistogram(60000, 1000000)
	bender.Record(recorder, bender.NewLoggingRecorder(l), bender.NewHistogramRecorder(h))
	fmt.Println(h)
}
```

### Run Server and Load Tester

Run these commands to compile and install the server and load tester:

```
go install $PKG/hellothrift/hello
go install $PKG/hellothrift/server
go install $PKG/hellobender
```

In one terminal window, run the server as `./bin/server` and in another window run the load tester
as `./bin/hellobender`. You should see a long sequence of outputs in the server window, and a final
print out of the histogram data in the other window.
