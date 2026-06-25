package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Goodmain/cch/internal/config"
)

func TestVersionString(t *testing.T) {
	version = "1.2.3"
	commit = "abc123"
	date = "2026-06-25"
	defer func() {
		version, commit, date = "dev", "none", "unknown"
	}()

	got := versionString()
	for _, want := range []string{"cch", "1.2.3", "abc123", "2026-06-25"} {
		if !strings.Contains(got, want) {
			t.Fatalf("versionString = %q, missing %q", got, want)
		}
	}
}

// runArgs invokes run() capturing stdout/stderr.
func runArgs(args ...string) (code int, out, errOut string) {
	var o, e bytes.Buffer
	code = run(args, &o, &e)
	return code, o.String(), e.String()
}

func TestRunVersion(t *testing.T) {
	code, out, _ := runArgs("version")
	if code != 0 || !strings.Contains(out, "cch ") {
		t.Fatalf("version: code=%d out=%q", code, out)
	}
}

func TestRunHelpAndSchema(t *testing.T) {
	if code, out, _ := runArgs("help"); code != 0 || !strings.Contains(out, "interactive helper") {
		t.Fatalf("help: code=%d out=%q", code, out)
	}
	if code, out, _ := runArgs("schema"); code != 0 || !strings.Contains(out, ".cch.json") {
		t.Fatalf("schema: code=%d out=%q", code, out)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	code, _, errOut := runArgs("bogus")
	if code != 1 {
		t.Fatalf("unknown: expected code 1, got %d", code)
	}
	if !strings.Contains(errOut, `unknown command "bogus"`) {
		t.Fatalf("unknown: stderr missing message: %q", errOut)
	}
}

func TestRunInitWritesConfig(t *testing.T) {
	dir := t.TempDir()
	prev, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	code, out, _ := runArgs("init")
	if code != 0 || !strings.Contains(out, "Wrote") {
		t.Fatalf("init: code=%d out=%q", code, out)
	}

	data, err := os.ReadFile(filepath.Join(dir, config.FileName))
	if err != nil {
		t.Fatalf("read written config: %v", err)
	}
	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("written config does not parse: %v", err)
	}
}
