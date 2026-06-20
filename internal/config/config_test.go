package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestLoadFileMissingIsEmpty(t *testing.T) {
	cfg, err := LoadFile(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatalf("missing file should not error, got %v", err)
	}
	if len(cfg.Commands) != 0 {
		t.Fatalf("expected empty config, got %d commands", len(cfg.Commands))
	}
}

func TestLoadFileMalformedErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	writeFile(t, path, "{ not json")
	if _, err := LoadFile(path); err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestLoadFileParsesCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	writeFile(t, path, `{"commands":[{"name":"migrate","description":"d","command":"php artisan migrate","arguments":[{"name":"env","description":"e"}],"parameters":[{"flag":"--step","description":"s","valued":true}]}]}`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Commands) != 1 {
		t.Fatalf("want 1 command, got %d", len(cfg.Commands))
	}
	c := cfg.Commands[0]
	if c.Name != "migrate" || len(c.Arguments) != 1 || len(c.Parameters) != 1 || !c.Parameters[0].Valued {
		t.Fatalf("unexpected parse: %+v", c)
	}
}

// mergeFrom replicates LoadMerged's merge over explicit global/local slices so
// the rule can be tested without touching $HOME or cwd.
func mergeFrom(global, local []Command) []Command {
	merged := map[string]Command{}
	for _, c := range global {
		c.Source = SourceGlobal
		merged[c.Name] = c
	}
	for _, c := range local {
		c.Source = SourceLocal
		merged[c.Name] = c
	}
	out := make([]Command, 0, len(merged))
	for _, c := range merged {
		out = append(out, c)
	}
	sortGrouped(out)
	return out
}

func TestMergeAndOverrideAndSort(t *testing.T) {
	global := []Command{
		{Name: "zeta", Command: "z-global"},
		{Name: "deploy", Command: "deploy-global"},
	}
	local := []Command{
		{Name: "deploy", Command: "deploy-local"},
		{Name: "alpha", Command: "a-local"},
	}
	got := mergeFrom(global, local)

	if len(got) != 3 {
		t.Fatalf("want 3 merged commands, got %d", len(got))
	}
	// Local group first (a-z within), then global group (a-z within).
	// Local has alpha + deploy (deploy overrides global); global has zeta.
	if got[0].Name != "alpha" || got[1].Name != "deploy" || got[2].Name != "zeta" {
		t.Fatalf("unexpected order: %v", []string{got[0].Name, got[1].Name, got[2].Name})
	}
	if got[0].Source != SourceLocal || got[1].Source != SourceLocal || got[2].Source != SourceGlobal {
		t.Fatalf("unexpected sources: %v", []Source{got[0].Source, got[1].Source, got[2].Source})
	}
	// Local overrides global on name collision.
	if got[1].Command != "deploy-local" {
		t.Fatalf("local should override global, got %q", got[1].Command)
	}
}
