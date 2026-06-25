package ui

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Goodmain/cch/internal/config"
)

func TestStylingHelpers(t *testing.T) {
	cases := []struct {
		fn  func(string) string
		seq string
	}{
		{bold, ansiBold},
		{faint, ansiFaint},
		{cyan, ansiCyan},
	}
	for _, c := range cases {
		out := c.fn("x")
		if !strings.Contains(out, c.seq) || !strings.HasSuffix(out, ansiReset) || !strings.Contains(out, "x") {
			t.Fatalf("styling helper produced %q", out)
		}
	}
}

func TestTerminalWidthDefaults(t *testing.T) {
	// In the test process stdout is not a TTY, so this falls back to 80.
	if w := terminalWidth(); w <= 0 {
		t.Fatalf("terminalWidth = %d, want > 0", w)
	}
}

// captureStdout redirects os.Stdout for the duration of fn and returns what was
// written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	_ = w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestShowDetailRendersAllSections(t *testing.T) {
	cmd := config.Command{
		Name:        "deploy",
		Description: "Deploy the app",
		Command:     "echo deploy",
		Arguments:   []config.Argument{{Name: "env", Description: "target env"}},
		Parameters: []config.Parameter{
			{Flag: "--force", Description: "force it", Valued: false},
			{Flag: "--step", Description: "how many", Valued: true},
		},
	}
	out := captureStdout(t, func() { ShowDetail(cmd) })

	for _, want := range []string{"deploy", "Deploy the app", "echo deploy", "env", "--force", "boolean flag", "--step", "valued"} {
		if !strings.Contains(out, want) {
			t.Fatalf("ShowDetail output missing %q:\n%s", want, out)
		}
	}
}

func TestCollectInputsNoArgsOrParams(t *testing.T) {
	cmd := config.Command{Name: "noop", Command: "echo hi"}
	in, err := CollectInputs(cmd)
	if err != nil {
		t.Fatalf("CollectInputs: %v", err)
	}
	if len(in.Args) != 0 || len(in.Params) != 0 {
		t.Fatalf("expected empty inputs, got %+v", in)
	}
}
