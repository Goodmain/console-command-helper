package help

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Goodmain/cch/internal/config"
)

// TestExampleParses guards against the worked example in the schema drifting
// from the real config structs: it must unmarshal into config.Config cleanly
// and produce the expected commands, arguments, and parameters.
func TestExampleParses(t *testing.T) {
	var cfg config.Config
	if err := json.Unmarshal([]byte(Example()), &cfg); err != nil {
		t.Fatalf("worked example does not parse into config.Config: %v", err)
	}
	if len(cfg.Commands) != 2 {
		t.Fatalf("want 2 commands, got %d", len(cfg.Commands))
	}

	migrate := cfg.Commands[0]
	if migrate.Name != "migrate" || migrate.Command != "php artisan migrate" {
		t.Fatalf("unexpected first command: %+v", migrate)
	}
	if len(migrate.Arguments) != 1 || migrate.Arguments[0].Name != "env" {
		t.Fatalf("unexpected arguments: %+v", migrate.Arguments)
	}
	if len(migrate.Parameters) != 2 {
		t.Fatalf("want 2 parameters, got %d", len(migrate.Parameters))
	}
	if !migrate.Parameters[0].Valued || migrate.Parameters[1].Valued {
		t.Fatalf("valued flags wrong: %+v", migrate.Parameters)
	}

	if cfg.Commands[1].Name != "test" || len(cfg.Commands[1].Arguments) != 0 {
		t.Fatalf("unexpected second command: %+v", cfg.Commands[1])
	}
}

// TestSchemaEmbedsExampleAndJSONSchema sanity-checks that the schema output
// actually contains both sections.
func TestSchemaEmbedsExampleAndJSONSchema(t *testing.T) {
	s := Schema()
	if !strings.Contains(s, "JSON Schema") {
		t.Error("schema output missing JSON Schema section")
	}
	if !strings.Contains(s, "draft-07") {
		t.Error("schema output missing draft-07 JSON Schema")
	}
	if !strings.Contains(s, "php artisan migrate") {
		t.Error("schema output missing worked example")
	}
}
