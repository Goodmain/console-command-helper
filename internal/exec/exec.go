// Package exec assembles the final command line and runs it via the system shell.
package exec

import (
	"os"
	"os/exec"
	"strings"

	"github.com/Goodmain/cch/internal/config"
)

// Inputs holds the user-collected values for a command.
type Inputs struct {
	// Args are mandatory argument values, in declared order.
	Args []string
	// Params are the chosen optional parameter tokens, already rendered
	// (e.g. "--step=3" or bare "--force").
	Params []string
}

// Assemble builds the full shell command line: base + ordered args + params.
func Assemble(cmd config.Command, in Inputs) string {
	parts := []string{strings.TrimSpace(cmd.Command)}
	for _, a := range in.Args {
		if a != "" {
			parts = append(parts, a)
		}
	}
	parts = append(parts, in.Params...)
	return strings.Join(parts, " ")
}

// Run executes the assembled line through `sh -c`, wiring stdio to the current
// process so the command runs interactively and streams output. It returns the
// command's error (including a non-zero *exec.ExitError) to the caller.
func Run(line string) error {
	c := exec.Command("sh", "-c", line)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
