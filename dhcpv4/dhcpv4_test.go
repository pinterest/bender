package dhcpv4

import (
	"net"
	"testing"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/async"
)

func validator(_, _ *dhcpv4.DHCPv4) error {
	return nil
}

func TestCreateExecutorAddressCheck(t *testing.T) {
	client := async.NewClient()
	_, err := CreateExecutor(client, validator)
	if err == nil || err.Error() != "invalid local address <nil>, want *net.UDPAddr" {
		t.Errorf("Expected CreateExecutor to fail with invalid address error, got (%s)", err)
	}
}

func TestExecutorTypeCheck(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:54321")
	if err != nil {
		t.Errorf("Expected no error when resolving udp address, got (%v)", err)
	}
	executor, err := CreateExecutor(&async.Client{LocalAddr: addr}, validator)
	if err != nil {
		t.Errorf("Expected no error when creating executor, got (%v)", err)
	}
	_, err = executor(0, 42)
	if err == nil || err.Error() != "invalid request type int, want: *dhcpv4.DHCPv4" {
		t.Errorf("Expected executor to fail with invalid request type error, got (%v)", err)
	}
}
