package exec

import (
	"errors"
	"os/exec"
	"testing"
)

func TestRunSuccess(t *testing.T) {
	if err := Run("exit 0"); err != nil {
		t.Fatalf("Run success: unexpected error %v", err)
	}
}

func TestRunPropagatesExitCode(t *testing.T) {
	err := Run("exit 3")
	if err == nil {
		t.Fatalf("expected error for non-zero exit")
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *exec.ExitError, got %T (%v)", err, err)
	}
	if code := exitErr.ExitCode(); code != 3 {
		t.Fatalf("expected exit code 3, got %d", code)
	}
}
