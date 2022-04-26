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

package thrift

import (
	"bytes"
	"context"
	"math/rand"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/pinterest/bender"
)

// A ClientExecutor executes a Thrift request.
type ClientExecutor func(interface{}, thrift.TTransport) (interface{}, error)

// NewThriftRequestExec creates a new Thrift-based RequestExecutor.
func NewThriftRequestExec(tFac thrift.TTransportFactory, clientExec ClientExecutor, cfg *thrift.TConfiguration, hosts ...string) bender.RequestExecutor {
	return func(_ int64, request interface{}) (interface{}, error) {
		addr := hosts[rand.Intn(len(hosts))]
		socket := thrift.NewTSocketConf(addr, cfg)
		defer socket.Close()

		transport, err := tFac.GetTransport(socket)
		if err != nil {
			return nil, err
		}
		if err := transport.Open(); err != nil {
			return nil, err
		}
		defer transport.Close()

		return clientExec(request, transport)
	}
}

// DeserializeThriftMessage deserializes a Thrift-encoded byte array.
func DeserializeThriftMessage(buf *bytes.Buffer, ts thrift.TStruct) (string, thrift.TMessageType, int32, error) {
	transport := thrift.NewStreamTransportR(buf)
	protocol := thrift.NewTBinaryProtocolConf(transport, &thrift.TConfiguration{
		TBinaryStrictRead:  thrift.BoolPtr(false),
		TBinaryStrictWrite: thrift.BoolPtr(false),
	})
	name, typeID, seqID, err := protocol.ReadMessageBegin(context.Background())
	if err != nil {
		return "", 0, 0, err
	}

	err = ts.Read(context.Background(), protocol)
	if err != nil {
		return "", 0, 0, err
	}

	return name, typeID, seqID, nil
}
