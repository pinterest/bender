/*
Copyright 2014 Pinterest.com

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

package thrift

import (
	"bytes"
	"math/rand"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/pinterest/bender"
)

type ThriftClientExecutor func(interface{}, thrift.TTransport) (interface{}, error)

func NewThriftRequestExec(tFac thrift.TTransportFactory, clientExec ThriftClientExecutor, timeout time.Duration, hosts ...string) bender.RequestExecutor {
	return func(_ int64, request interface{}) (interface{}, error) {
		addr := hosts[rand.Intn(len(hosts))]
		socket, err := thrift.NewTSocketTimeout(addr, timeout)
		if err != nil {
			return nil, err
		}
		defer socket.Close()

		transport := tFac.GetTransport(socket)
		if err := transport.Open(); err != nil {
			return nil, err
		}
		defer transport.Close()

		return clientExec(request, transport)
	}
}

func DeserializeThriftMessage(buf *bytes.Buffer, ts thrift.TStruct) (string, thrift.TMessageType, int32, error) {
	transport := thrift.NewStreamTransportR(buf)
	protocol := thrift.NewTBinaryProtocol(transport, false, false)
	name, typeId, seqId, err := protocol.ReadMessageBegin()
	if err != nil {
		return "", 0, 0, err
	}

	err = ts.Read(protocol)
	if err != nil {
		return "", 0, 0, err
	}

	return name, typeId, seqId, nil
}
