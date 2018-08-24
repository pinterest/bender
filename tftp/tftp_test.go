package tftp

import (
	"testing"
)

func TestExecutorTypeCheck(t *testing.T) {
	executor := CreateExecutor(nil, DiscardingValidator)
	_, err := executor(0, 42)
	if err == nil || err.Error() != "invalid request type int, want: *tftp.Request" {
		t.Errorf("Expected executor to fail with invalid request type error, got (%s)", err)
	}
}
