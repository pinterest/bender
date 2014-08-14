package bender

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"bytes"
)

type ThriftClientExecutor func(*Request, thrift.TTransport) error

func NewThriftRequestExec(tFac thrift.TTransportFactory, clientExec ThriftClientExecutor) RequestExecutor {
	addr := "asterix-staging001:3636"

	return func(_ int64, request *Request) error {
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
