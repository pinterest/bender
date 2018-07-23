package dns

import (
	"testing"

	"github.com/miekg/dns"
)

func validator(_, _ *dns.Msg) error {
	return nil
}

func TestExecutorTypeCheck(t *testing.T) {
	executor := CreateExecutor(nil, validator, ":54321")
	_, err := executor(0, dns.Msg{})
	if err == nil || err.Error() != "invalid request type dns.Msg, want: *dns.Msg" {
		t.Errorf("Expected executor to fail with invalid request type error, got (%s)", err)
	}
}
