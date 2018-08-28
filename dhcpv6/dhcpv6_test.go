package dhcpv6

import (
	"testing"

	"github.com/insomniacslk/dhcp/dhcpv6"
)

func validator(_, _ dhcpv6.DHCPv6) error {
	return nil
}

func TestExecutorTypeCheck(t *testing.T) {
	executor := CreateExecutor(nil, validator)
	_, err := executor(0, 42)
	if err == nil || err.Error() != "invalid request type int, want: dhcpv6.DHCPv6" {
		t.Errorf("Expected executor to fail with invalid request type error, got (%s)", err)
	}
}
