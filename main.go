// Command cch is an interactive helper for project console commands defined in
// .cch.json config files (global ~/.cch.json + local ./.cch.json).
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/Goodmain/cch/internal/config"
	cexec "github.com/Goodmain/cch/internal/exec"
	"github.com/Goodmain/cch/internal/help"
	"github.com/Goodmain/cch/internal/ui"
)

// Build information, injected at release time via -ldflags by GoReleaser.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run dispatches a single invocation and returns the process exit code. It is
// the testable core of main: all output goes to the provided writers and it
// never calls os.Exit itself.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		if err := runInteractive(); err != nil {
			fmt.Fprintln(stderr, "cch: "+err.Error())
			return 1
		}
		return 0
	}

	switch args[0] {
	case "init":
		if err := runInit(stdout); err != nil {
			fmt.Fprintln(stderr, "cch: "+err.Error())
			return 1
		}
	case "version", "-v", "--version":
		fmt.Fprintln(stdout, versionString())
	case "help", "-h", "--help":
		fmt.Fprint(stdout, help.Help())
	case "schema":
		fmt.Fprint(stdout, help.Schema())
	default:
		fmt.Fprintf(stderr, "cch: unknown command %q\n\n", args[0])
		fmt.Fprint(stderr, help.Help())
		return 1
	}
	return 0
}

func runInit(stdout io.Writer) error {
	res, err := config.Init(false)
	if err != nil {
		return err
	}
	if res.Exists {
		fmt.Fprintf(stdout, "%s already exists. Overwrite? [y/N]: ", res.Path)
		var ans string
		_, _ = fmt.Scanln(&ans)
		if ans != "y" && ans != "Y" {
			fmt.Fprintln(stdout, "Left existing config untouched.")
			return nil
		}
		if _, err := config.Init(true); err != nil {
			return err
		}
	}
	fmt.Fprintf(stdout, "Wrote %s\n", res.Path)
	return nil
}

func runInteractive() error {
	commands, err := config.LoadMerged()
	if err != nil {
		return err
	}
	if len(commands) == 0 {
		fmt.Println("No commands available. Run `cch init` to create a .cch.json.")
		return nil
	}

	cmd, err := ui.SelectCommand(commands)
	if err != nil {
		// User interrupted the picker (Ctrl-C / Esc).
		return nil
	}

	ui.ShowDetail(cmd)

	in, err := ui.CollectInputs(cmd)
	if err != nil {
		return nil
	}

	line := cexec.Assemble(cmd, in)

	if err := ui.Confirm(line); err != nil {
		if errors.Is(err, ui.ErrAborted) {
			fmt.Println("Aborted.")
			return nil
		}
		return err
	}

	if err := cexec.Run(line); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("command failed: exit status %d", exitErr.ExitCode())
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func versionString() string {
	return fmt.Sprintf("cch %s (commit %s, built %s)", version, commit, date)
}
