package dns

import (
	"github.com/miekg/dns"
	"github.com/pinterest/bender"
)

// ResponseValidator validates a DNS response.
type ResponseValidator func(request interface{}, response *dns.Msg) error

// CreateExecutor creates a new DNS RequestExecutor.
func CreateExecutor(client *dns.Client, responseValidator ResponseValidator, hosts ...string) bender.RequestExecutor {
	if client == nil {
		client = new(dns.Client)
	}

	var i int
	return func(_ int64, request interface{}) (interface{}, error) {
		msg := request.(*dns.Msg)
		addr := hosts[i]
		i = (i + 1) % len(hosts)
		resp, _, err := client.Exchange(msg, addr)
		if err != nil {
			return nil, err
		}
		if err = responseValidator(request, resp); err != nil {
			return nil, err
		}
		return resp, nil
	}
}
