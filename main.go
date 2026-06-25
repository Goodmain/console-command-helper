// Command cch is an interactive helper for project console commands defined in
// .cch.json config files (global ~/.cch.json + local ./.cch.json).
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/Goodmain/cch/internal/config"
	cexec "github.com/Goodmain/cch/internal/exec"
	"github.com/Goodmain/cch/internal/help"
	"github.com/Goodmain/cch/internal/ui"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		if err := runInteractive(); err != nil {
			fail(err)
		}
		return
	}

	switch args[0] {
	case "init":
		if err := runInit(); err != nil {
			fail(err)
		}
	case "help", "-h", "--help":
		fmt.Print(help.Help())
	case "schema":
		fmt.Print(help.Schema())
	default:
		fmt.Fprintf(os.Stderr, "cch: unknown command %q\n\n", args[0])
		fmt.Fprint(os.Stderr, help.Help())
		os.Exit(1)
	}
}

func runInit() error {
	res, err := config.Init(false)
	if err != nil {
		return err
	}
	if res.Exists {
		fmt.Printf("%s already exists. Overwrite? [y/N]: ", res.Path)
		var ans string
		fmt.Scanln(&ans)
		if ans != "y" && ans != "Y" {
			fmt.Println("Left existing config untouched.")
			return nil
		}
		if _, err := config.Init(true); err != nil {
			return err
		}
	}
	fmt.Printf("Wrote %s\n", res.Path)
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

func fail(err error) {
	fmt.Fprintln(os.Stderr, "cch: "+err.Error())
	os.Exit(1)
}
