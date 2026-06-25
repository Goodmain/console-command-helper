package config

import (
	"path/filepath"
	"testing"
)

func TestGlobalPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := GlobalPath()
	if err != nil {
		t.Fatalf("GlobalPath: %v", err)
	}
	if want := filepath.Join(home, FileName); got != want {
		t.Fatalf("GlobalPath = %q, want %q", got, want)
	}
}

func TestLoadMergedLocalOverridesGlobalAndSorts(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	t.Setenv("HOME", home)
	chdir(t, work)

	// Global: two commands.
	writeFile(t, filepath.Join(home, FileName), `{
	  "commands": [
	    { "name": "deploy", "description": "global deploy", "command": "echo global-deploy" },
	    { "name": "build",  "description": "global build",  "command": "echo build" }
	  ]
	}`)
	// Local: overrides "deploy", adds "apply".
	writeFile(t, filepath.Join(work, FileName), `{
	  "commands": [
	    { "name": "deploy", "description": "local deploy", "command": "echo local-deploy" },
	    { "name": "apply",  "description": "local apply",  "command": "echo apply" }
	  ]
	}`)

	cmds, err := LoadMerged()
	if err != nil {
		t.Fatalf("LoadMerged: %v", err)
	}
	if len(cmds) != 3 {
		t.Fatalf("want 3 merged commands, got %d: %+v", len(cmds), cmds)
	}

	// Local group first (sorted a-z), then global group (sorted a-z).
	wantOrder := []struct {
		name   string
		source Source
		desc   string
	}{
		{"apply", SourceLocal, "local apply"},
		{"deploy", SourceLocal, "local deploy"}, // local overrides global
		{"build", SourceGlobal, "global build"},
	}
	for i, w := range wantOrder {
		if cmds[i].Name != w.name || cmds[i].Source != w.source || cmds[i].Description != w.desc {
			t.Fatalf("cmds[%d] = {%s, src=%d, %q}, want {%s, src=%d, %q}",
				i, cmds[i].Name, cmds[i].Source, cmds[i].Description, w.name, w.source, w.desc)
		}
	}
}

func TestLoadMergedEmptyWhenNoConfigs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	chdir(t, t.TempDir())

	cmds, err := LoadMerged()
	if err != nil {
		t.Fatalf("LoadMerged: %v", err)
	}
	if len(cmds) != 0 {
		t.Fatalf("want 0 commands, got %d", len(cmds))
	}
}

func TestLoadMergedReportsMalformedLocal(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	work := t.TempDir()
	chdir(t, work)
	writeFile(t, filepath.Join(work, FileName), `{ not json`)

	if _, err := LoadMerged(); err == nil {
		t.Fatalf("expected error for malformed local config")
	}
}
