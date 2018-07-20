package dns

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/pinterest/bender"
)

// ResponseValidator validates a DNS response.
type ResponseValidator func(request, response *dns.Msg) error

// CreateExecutor creates a new DNS RequestExecutor.
func CreateExecutor(client *dns.Client, responseValidator ResponseValidator, hosts ...string) bender.RequestExecutor {
	if client == nil {
		client = new(dns.Client)
	}

	var i int
	return func(_ int64, request interface{}) (interface{}, error) {
		msg, ok := request.(*dns.Msg)
		if !ok {
			return nil, fmt.Errorf("invalid request type %T, want: *dns.Msg", request)
		}
		addr := hosts[i]
		i = (i + 1) % len(hosts)
		resp, _, err := client.Exchange(msg, addr)
		if err != nil {
			return nil, err
		}
		if err = responseValidator(msg, resp); err != nil {
			return nil, err
		}
		return resp, nil
	}
}
