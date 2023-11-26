package dhcpv4

import (
	"context"
	"fmt"
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/nclient4"
	"github.com/pinterest/bender"
)

// ResponseValidator validates a DHCPv4 response.
type ResponseValidator func(request, response *dhcpv4.DHCPv4) error

// CreateExecutor creates a new DHCPv4 RequestExecutor.
//
// relayIP is the IP used as the gateway IP.
func CreateExecutor(client *nclient4.Client, relayIP net.IP, validator ResponseValidator) (bender.RequestExecutor, error) {
	return func(_ int64, request interface{}) (interface{}, error) {
		dis, ok := request.(*dhcpv4.DHCPv4)
		if !ok {
			return nil, fmt.Errorf("invalid request type %T, want: *dhcpv4.DHCPv4", request)
		}
		ctx := context.Background()
		relayMod := dhcpv4.WithRelay(relayIP)
		relayMod(dis)
		off, err := client.SendAndRead(ctx, client.RemoteAddr(), dis, nclient4.IsMessageType(dhcpv4.MessageTypeOffer))
		if err != nil {
			return nil, fmt.Errorf("error receiving DHCP offer: %w", err)
		}
		req, err := dhcpv4.NewRequestFromOffer(off, relayMod)
		if err != nil {
			return nil, err
		}
		ack, err := client.SendAndRead(ctx, client.RemoteAddr(), req, nclient4.IsMessageType(dhcpv4.MessageTypeAck))
		if err != nil {
			return nil, fmt.Errorf("error receiving DHCP ACK: %w", err)
		}
		err = validator(dis, ack)
		return ack, err
	}, nil
}
