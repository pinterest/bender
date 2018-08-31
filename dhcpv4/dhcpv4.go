package dhcpv4

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/async"
	"github.com/pinterest/bender"
)

// ResponseValidator validates a DHCPv4 response.
type ResponseValidator func(request, response *dhcpv4.DHCPv4) error

// CreateExecutor creates a new DHCPv4 RequestExecutor.
func CreateExecutor(client *async.Client, validator ResponseValidator) (bender.RequestExecutor, error) {
	send, err := newSendFunc(client)
	if err != nil {
		return nil, err
	}
	return func(_ int64, request interface{}) (interface{}, error) {
		dis, ok := request.(*dhcpv4.DHCPv4)
		if !ok {
			return nil, fmt.Errorf("invalid request type %T, want: *dhcpv4.DHCPv4", request)
		}
		off, err := send(dis, dhcpv4.MessageTypeOffer)
		if err != nil {
			return nil, err
		}
		req, err := dhcpv4.NewRequestFromOffer(off)
		if err != nil {
			return nil, err
		}
		ack, err := send(req, dhcpv4.MessageTypeAck)
		if err != nil {
			return nil, err
		}
		err = validator(dis, ack)
		return ack, err
	}, nil
}

// sendFunc represents a function used to send dhcpv4 datagrams
type sendFunc = func(*dhcpv4.DHCPv4, dhcpv4.MessageType) (*dhcpv4.DHCPv4, error)

// newSendFunc creates a function which will send messages using the given client
func newSendFunc(client *async.Client) (sendFunc, error) {
	addr, ok := client.LocalAddr.(*net.UDPAddr)
	if !ok {
		return nil, fmt.Errorf("invalid local address %T, want *net.UDPAddr", client.LocalAddr)
	}
	m := dhcpv4.WithRelay(addr.IP)
	t := uint(client.ReadTimeout / time.Millisecond)

	// send a message and check that the response is of a given type
	return func(d *dhcpv4.DHCPv4, mt dhcpv4.MessageType) (*dhcpv4.DHCPv4, error) {
		response, err, timeout := client.Send(d, m).GetOrTimeout(t)
		if timeout {
			return nil, errors.New("timeout")
		} else if err != nil {
			return nil, err
		}
		res, ok := response.(*dhcpv4.DHCPv4)
		if !ok {
			return nil, fmt.Errorf("invalid response type %T, want: *dhcpv4.DHCPv4", response)
		}
		if err := assertMessageType(res, mt); err != nil {
			return nil, err
		}
		return res, nil
	}, nil
}

// assertMessageType extracts message type and checks if it matches expected
func assertMessageType(d *dhcpv4.DHCPv4, mt dhcpv4.MessageType) error {
	t := d.MessageType()
	if t == nil {
		return fmt.Errorf("unable to extract message type")
	}
	if *t != mt {
		return fmt.Errorf("invalid message type %s, want: %s", t.String(), mt.String())
	}
	return nil
}
