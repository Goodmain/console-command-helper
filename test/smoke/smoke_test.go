//go:build integration

package smoke

import (
	"os/exec"
	"strings"
	"testing"
)

// TestNonInteractiveCommandsSmoke builds and exercises cch's non-interactive
// subcommands end to end. It runs the real binary via `go run` and asserts the
// commands succeed and emit their expected markers. Gated behind the
// `integration` build tag so it stays out of the default unit-test run.
func TestNonInteractiveCommandsSmoke(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{name: "help", args: []string{"help"}, want: "interactive helper"},
		{name: "schema", args: []string{"schema"}, want: ".cch.json"},
		{name: "version", args: []string{"version"}, want: "cch "},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"run", "../.."}, tc.args...)
			out, err := exec.Command("go", args...).CombinedOutput()
			if err != nil {
				t.Fatalf("cch %v failed: %v\n%s", tc.args, err, out)
			}
			if !strings.Contains(string(out), tc.want) {
				t.Fatalf("cch %v output missing %q:\n%s", tc.args, tc.want, out)
			}
		})
	}
}
