package tftp

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/pin/tftp"
	"github.com/pinterest/bender"
)

// RequestMode represents a mode of the tftp request
type RequestMode string

// Available request modes
const (
	ModeOctet    RequestMode = "octet"
	ModeNetascii             = "netascii"
)

// Request represents a tftp receive request
type Request struct {
	Filename string
	Mode     RequestMode
}

// ResponseValidator validates a udp response.
type ResponseValidator func(request *Request, writer io.WriterTo) (interface{}, error)

// DiscardingValidator reads the whole request and discards its body.
func DiscardingValidator(request *Request, writer io.WriterTo) (interface{}, error) {
	_, err := writer.WriteTo(ioutil.Discard)
	return request, err
}

// CreateExecutor creates a new TFTP RequestExecutor.
func CreateExecutor(client *tftp.Client, validator ResponseValidator) bender.RequestExecutor {
	return func(_ int64, request interface{}) (interface{}, error) {
		r, ok := request.(*Request)
		if !ok {
			return nil, fmt.Errorf("invalid request type %T, want: *tftp.Request", request)
		}
		w, err := client.Receive(r.Filename, string(r.Mode))
		if err != nil {
			return nil, err
		}
		return validator(r, w)
	}
}
