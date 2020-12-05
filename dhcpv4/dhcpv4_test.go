package dhcpv4

import (
	"net"
	"testing"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/nclient4"
)

func validator(_, _ *dhcpv4.DHCPv4) error {
	return nil
}

func TestExecutorTypeCheck(t *testing.T) {
	executor, err := CreateExecutor(&nclient4.Client{}, net.IP{1, 1, 1, 1}, validator)
	if err != nil {
		t.Errorf("Expected no error when creating executor, got (%v)", err)
	}
	_, err = executor(0, 42)
	if err == nil || err.Error() != "invalid request type int, want: *dhcpv4.DHCPv4" {
		t.Errorf("Expected executor to fail with invalid request type error, got (%v)", err)
	}
}
