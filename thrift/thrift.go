package thrift

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"bytes"
	"github.com/Pinterest/bender"
	"math/rand"
)

type ThriftClientExecutor func(*bender.Request, thrift.TTransport) error

func NewThriftRequestExec(tFac thrift.TTransportFactory, clientExec ThriftClientExecutor, hosts... string) bender.RequestExecutor {
	return func(_ int64, request *bender.Request) error {
		addr := hosts[rand.Intn(len(hosts))]
		socket, err := thrift.NewTSocket(addr)
		if err != nil {
			return err
		}
		defer socket.Close()

		transport := tFac.GetTransport(socket)
		if err := transport.Open(); err != nil {
			return err
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
