Bender
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

* [DHCPv6](https://github.com/pinterest/bender/blob/master/dhcpv6/TUTORIAL.md)
* [DNS](https://github.com/pinterest/bender/blob/master/dns/TUTORIAL.md)
* [HTTP](https://github.com/pinterest/bender/blob/master/http/TUTORIAL.md)
* [TFTP](https://github.com/pinterest/bender/blob/master/tftp/TUTORIAL.md)
* [Thrift](https://github.com/pinterest/bender/blob/master/thrift/TUTORIAL.md)

## Documentation

The package documentation is available on [godoc.org](http://godoc.org/github.com/pinterest/bender).
The function and data structure documentation is also available there.

## Performance

We have only informal, anecdotal evidence for the maximum performance of Bender. For example, in a
very simple load test using a Thrift server that just echoes a message, Bender was able to send
7,000 QPS from a single EC2 m3.2xlarge host. At higher throughput the Bender "overage" counter
increased, indicating that either the Go runtime, the OS or the host was struggling to keep up.

We have found a few things that make a big difference when running load tests. First, the Go
runtime needs some tuning. In particular, the Go GC is very immature, so we prefer to disable it
using the `GOGC=off` environment variable. In addition, we have seen some gains from setting
`GOMAXPROCS` to twice the number of CPUs.

Secondly, the Linux TCP stack for a default server installation is usually not tuned to high
throughput servers or load testers. After some experimentation, we have settled on adding these
lines to `/etc/sysctl.conf`, after which you can run `sysctl -p` to load them (although it is
recommended to restart your host at this point to make sure these take effect).

```
# /etc/sysctl.conf
# Increase system file descriptor limit
fs.file-max = 100000

# Increase ephermeral IP ports
net.ipv4.ip_local_port_range = 10000 65000

# Increase Linux autotuning TCP buffer limits
# Set max to 16MB for 1GE and 32M (33554432) or 54M (56623104) for 10GE
# Don't set tcp_mem itself! Let the kernel scale it based on RAM.
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.core.rmem_default = 16777216
net.core.wmem_default = 16777216
net.core.optmem_max = 40960
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216

# Make room for more TIME_WAIT sockets due to more clients,
# and allow them to be reused if we run out of sockets
# Also increase the max packet backlog
net.core.netdev_max_backlog = 50000
net.ipv4.tcp_max_syn_backlog = 30000
net.ipv4.tcp_max_tw_buckets = 2000000
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 10

# Disable TCP slow start on idle connections
net.ipv4.tcp_slow_start_after_idle = 0
```

This is a slightly modified version of advice taken from this source:
http://www.nateware.com/linux-network-tuning-for-2013.html#.VBjahC5dVyE

In addition, it helps to increase the open file limit with something like:

```
ulimit -n 100000
```

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

Bender currently provides helper functions for DHCPv6, DNS, HTTP, Thrift and TFTP. We appreciate
Pull Requests for other protocols.

The load testers we have written internally with Bender have a lot of common command line arguments,
but we haven't finalized a set to share as part of the library.

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

## Copyright

Copyright 2014-2018 Pinterest, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Attribution

Bender includes open source from the following sources:

* Apache Thrift Libraries. Copyright 2014 Apache Software Foundation. Licensed under the Apache License v2.0 (http://www.apache.org/licenses/).
* Miek Gieben DNS library for Go. Licensed under the BSD license (https://github.com/miekg/dns/blob/master/COPYRIGHT)
* Insomniacslk DHCP library for Go. Licensed under the BSD license (https://github.com/insomniacslk/dhcp/blob/master/LICENSE)
* Dmitri Popov TFTP library for Go. Licensed under the MIT license (https://github.com/pin/tftp/blob/master/LICENSE)
* Go Libraries. Copyright 2012 The Go Authors. Licensed under the BSD license (http://golang.org/LICENSE).
