package help

import (
	"strings"
	"testing"
)

func TestHelpText(t *testing.T) {
	h := Help()
	for _, want := range []string{"cch", "cch init", "cch schema", "cch version", "cch help"} {
		if !strings.Contains(h, want) {
			t.Fatalf("help text missing %q", want)
		}
	}
}

func TestSchemaText(t *testing.T) {
	s := Schema()
	for _, want := range []string{".cch.json", "JSON Schema", "arguments", "parameters"} {
		if !strings.Contains(s, want) {
			t.Fatalf("schema text missing %q", want)
		}
	}
}
