package dhcpv6

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/insomniacslk/dhcp/dhcpv6/async"
	"github.com/mdlayher/eui64"
	"github.com/pinterest/bender"
)

// ResponseValidator validates a DHCPv6 response.
type ResponseValidator func(request, response dhcpv6.DHCPv6) error

// CreateExecutor creates a new DHCPv6 RequestExecutor.
func CreateExecutor(client *async.Client, validator ResponseValidator) bender.RequestExecutor {
	return func(_ int64, request interface{}) (interface{}, error) {
		solicit, ok := request.(dhcpv6.DHCPv6)
		if !ok {
			return nil, fmt.Errorf("invalid request type %T, want: dhcpv6.DHCPv6", request)
		}
		sol, err := relaySolicit(solicit)
		if err != nil {
			return nil, err
		}
		adv, err := send(client, sol)
		if err != nil {
			return nil, err
		}
		req, err := relayRequestFromAdvertise(adv)
		if err != nil {
			return nil, err
		}
		res, err := send(client, req)
		if err != nil {
			return nil, err
		}
		_, rep, err := unpack(res, dhcpv6.MessageTypeReply)
		if err != nil {
			return nil, err
		}
		err = validator(solicit, rep)
		return rep, err
	}
}

// send sends a message, asserts the response type and returns error if the
// request timed out
func send(client *async.Client, message dhcpv6.DHCPv6) (dhcpv6.DHCPv6, error) {
	res, err, timeout := client.Send(message).GetOrTimeout(uint(client.ReadTimeout / time.Millisecond))
	if timeout {
		return nil, errors.New("timeout")
	} else if err != nil {
		return nil, err
	}
	if res, ok := res.(dhcpv6.DHCPv6); ok {
		return res, nil
	}
	return nil, fmt.Errorf("invalid response type %T, want: dhcpv6.DHCPv6", res)
}

// unpack extracts the relay inner message and asserts for the expected type,
// returns a relay and an inner message
func unpack(response dhcpv6.DHCPv6, messageType dhcpv6.MessageType) (*dhcpv6.DHCPv6Relay, dhcpv6.DHCPv6, error) {
	relay, ok := response.(*dhcpv6.DHCPv6Relay)
	if !ok {
		return nil, nil, fmt.Errorf("invalid response type %T, want: *dhcpv6.DHCPv6Relay", response)
	}
	mes, err := relay.GetInnerMessage()
	if err != nil {
		return nil, nil, err
	}
	if mes.Type() != messageType {
		return nil, nil, fmt.Errorf("invalid message type %s, want: %s", mes.Type().String(), messageType.String())
	}
	return relay, mes, nil
}

// relaySolicit encapsulates a solicit message in relay
func relaySolicit(solicit dhcpv6.DHCPv6) (dhcpv6.DHCPv6, error) {
	cid := solicit.GetOneOption(dhcpv6.OptionClientID)
	if cid == nil {
		return nil, errors.New("client id cannot be nil")
	}
	mac := cid.(*dhcpv6.OptClientId).Cid.LinkLayerAddr
	peer, err := eui64.ParseMAC(net.ParseIP("fe80::"), mac)
	if err != nil {
		return nil, err
	}
	return dhcpv6.EncapsulateRelay(solicit, dhcpv6.MessageTypeRelayForward, net.ParseIP("::"), peer)
}

// relayRequestFromAdvertise unpacks relayed advertise and creates a relayed
// dhcpv6 request based on it
func relayRequestFromAdvertise(advertise dhcpv6.DHCPv6) (dhcpv6.DHCPv6, error) {
	relay, adv, err := unpack(advertise, dhcpv6.MessageTypeAdvertise)
	if err != nil {
		return nil, err
	}
	req, err := dhcpv6.NewRequestFromAdvertise(adv)
	if err != nil {
		return nil, err
	}
	return dhcpv6.EncapsulateRelay(req, dhcpv6.MessageTypeRelayForward, relay.LinkAddr(), relay.PeerAddr())
}
