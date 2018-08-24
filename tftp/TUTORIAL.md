# Bender TFTP Tutorial

This tutorial walks through the steps to create a simple TFTP server with a load tester. If you
already know how to Go's TFTP libraries, skip to the "Load Testing" section to get started with
Bender and TFTP.

## Getting Started

You will need to install and configure Go by following the instructions on the
[Getting Started](https://golang.org/doc/install) page. Then follow the instructions on the
[How To Write Go Code](https://golang.org/doc/code.html) page, particularly for setting up your
workspace and the `GOPATH` environment variable, which we will use throughout this tutorial.

Next we need the TFTP library for Go, which you can fetch using `go get` as:
```sh
cd $GOPATH
go get github.com/pin/tftp
```

Finally you will need the latest version of Bender, which you can get by running:

```sh
go get github.com/pinterest/bender
```

## A TFTP Server

We won't go indepth into creating a TFTP server in this tutorial. Example implementation of TFTP
server in Go using the same library we'll be using for loadtesting can be found in the pin/tftp
[README.md](https://github.com/pin/tftp#tftp-server). For simplicity I will quickly explain how
to setup a tftp server on RedHat Linux. We will also create a file to download later during the
actual load testing. Run following commands as sudo:

```sh
# Install the tftp server
yum install tftp-server
# Create a 4kB file filled with zeroes
sudo dd if=/dev/zero of=/var/lib/tftpboot/myfile.txt bs=4096 count=1
# Start the server
systemctl start tftp
```

You can verify if everything is running correctly by downloading the file with busybox
```sh
busybox tftp -g -r /myfile.txt localhost
```

## Load Testing

Now that we have a TFTP server we can use Bender to build a simple load tester for it. The next
few sections walk through the various parts of the load tester. If you are in a hurry skip to the
section "Final Load Tester Program" and just follow the instructions from there.

In the following, all commands should be run from the `$GOROOT` directory, unless otherwise noted.

### Intervals

Downloading a single file with TFTP takes a lot of UDP packets, which means a Concurrency test is
more suitable than a standard Throughput test. We will create a workers semaphore which will limit
concurrent files being downloaded to a fixed amount. As semaphore `Signal` is blocking we need to
run it in goroutine before starting the test.

```go
ws := bender.NewWorkerSemaphore()
go func() { ws.Signal(10) }()
```

### Request Generator

The second thing we need is a channel of requests to send to the TFTP server. When an interval has
been generated and Bender is ready to send the request, it pulls the next request from this channel
and spawns a goroutine to send the request to the server. This function creates a simple synthetic
request generator:

```go
func SyntheticTFTPRequests(n int) chan interface{} {
	c := make(chan interface{}, n)
	go func() {
		for i := 0; i < n; i++ {
			c <- &tftp.Request{File: "myfile.txt", Type: tftp.ModeOctet}
		}
		close(c)
	}()
	return c
}
```

### Request Executor

The next thing we need is a request executor, which takes the requests generated above and sends
them to the service. We will use a helper function from Bender's tftp library to do most of the
work (connection management, error handling, etc), so all we have to do is write code to verify the
request:

```go
func validator(r *btftp.Request, w io.WriterTo) (interface{}, error) {
	buffer := new(bytes.Buffer)
	n, err := w.WriteTo(buffer)
	if err != nil {
		return nil, err
	}
	if n != 4096 {
		return nil, fmt.Errorf("invalid response size %d, expoected 4096", n)
	}
	for i, b := range buffer.Bytes() {
		if b != 0 {
			return nil, fmt.Errorf("invalid response byte %d='0x%02x', want '0x00'", i, b)
		}
	}
	return nil, nil
}

exec := btftp.CreateExecutor(client, validator)
```

This validates that the response has size of 4kB and is filled with 0's. When creating your own
validator make sure to always read from the writer as most of the download is not started before
calling `WriteTo` method.


### Recorder

The last thing we need is a channel that will output events as the load tester runs. This will let
us listen to the load testers progress and record stats. We want this channel to be buffered so that
we can run somewhat independently of the load test without slowing it down:

```go
recorder := make(chan interface{}, 100)
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
h := hist.NewHistogram(int(2 * time.Second / time.Millisecond), int(time.Millisecond))
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

Create a new Go package for your load tester. We'll refer to this as `$PKG` in this document, and
it can be any path you want. At Facebook, for example, we use `github.com/facebook`.

```sh
mkdir -p src/$PKG/hellobender
```

Then create a file named `main.go` in that directory and add these lines to it:

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	btftp "github.com/pinterest/bender/tftp"

	"github.com/pin/tftp"
	"github.com/pinterest/bender"
	"github.com/pinterest/bender/hist"
)

// SyntheticTFTPRequests generates n dummy requests to the tftp server
func SyntheticTFTPRequests(n int) chan interface{} {
	c := make(chan interface{}, n)
	go func() {
		for i := 0; i < n; i++ {
			c <- &btftp.Request{Filename: "myfile.txt", Mode: btftp.ModeOctet}
		}
		close(c)
	}()
	return c
}

const expected = "Lorem ipsum dolor sit amet"

func validator(r *btftp.Request, w io.WriterTo) (interface{}, error) {
	buffer := new(bytes.Buffer)
	n, err := w.WriteTo(buffer)
	if err != nil {
		return nil, err
	}
	if n != 4096 {
		return nil, fmt.Errorf("invalid response size %d, expoected 4096", n)
	}
	for i, b := range buffer.Bytes() {
		if b != 0 {
			return nil, fmt.Errorf("invalid response byte %d='0x%02x', want '0x00'", i, b)
		}
	}
	return nil, nil
}

func main() {
	client, err := tftp.NewClient("localhost:69")
	if err != nil {
		panic(err)
	}
	client.SetTimeout(500 * time.Millisecond)
	exec := btftp.CreateExecutor(client, validator)
	requests := SyntheticTFTPRequests(100)
	ws := bender.NewWorkerSemaphore()
	go func() { ws.Signal(10) }()
	recorder := make(chan interface{}, 100)

	bender.LoadTestConcurrency(ws, requests, exec, recorder)

	l := log.New(os.Stdout, "", log.LstdFlags)
	h := hist.NewHistogram(int(2 * time.Second / time.Millisecond), int(time.Millisecond))
	bender.Record(recorder, bender.NewLoggingRecorder(l), bender.NewHistogramRecorder(h))
	fmt.Println(h)
}
```

### Run the Load Tester
Run these commands to compile and install the load tester:

```sh
go install $PKG/hellobender
```

Run the load tester as `./bin/hellobender`. You should see logs as files are being
downloaded and final print out of the histogram when all downloads finish.
