package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// chdir switches into dir for the duration of the test and restores the
// previous working directory afterwards.
func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
}

func TestInitWritesStarterConfig(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	res, err := Init(false)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if res.Exists {
		t.Fatalf("expected Exists=false in empty dir")
	}
	if filepath.Base(res.Path) != FileName {
		t.Fatalf("unexpected path %q", res.Path)
	}

	// The written file must be valid JSON that parses into the config structs.
	data, err := os.ReadFile(res.Path)
	if err != nil {
		t.Fatalf("read written config: %v", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("starter config does not parse: %v", err)
	}
	if len(cfg.Commands) != 1 || cfg.Commands[0].Name != "example" {
		t.Fatalf("unexpected starter commands: %+v", cfg.Commands)
	}
}

func TestInitDoesNotOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	path := filepath.Join(dir, FileName)
	if err := os.WriteFile(path, []byte(`{"commands":[]}`), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	res, err := Init(false)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if !res.Exists {
		t.Fatalf("expected Exists=true when file present")
	}

	// Existing content must be untouched.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != `{"commands":[]}` {
		t.Fatalf("file was overwritten: %q", data)
	}
}

func TestInitForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	path := filepath.Join(dir, FileName)
	if err := os.WriteFile(path, []byte(`{"commands":[]}`), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	res, err := Init(true)
	if err != nil {
		t.Fatalf("init force: %v", err)
	}
	if res.Exists {
		t.Fatalf("force result should not report Exists")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) == `{"commands":[]}` {
		t.Fatalf("file was not overwritten")
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("overwritten config does not parse: %v", err)
	}
}
