// Package ui drives the interactive flow: selection, detail view, input
// prompting, and the final confirmation, built on promptui.
package ui

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Goodmain/cch/internal/config"
	"github.com/Goodmain/cch/internal/exec"
	"github.com/manifoldco/promptui"
	"golang.org/x/term"
)

// terminalWidth returns the current terminal width in columns, defaulting to 80
// when it cannot be determined (e.g. output is not a TTY).
func terminalWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return 80
}

// Subtle ANSI styling: bold for primary text/headings, faint for secondary
// text, cyan as the single accent color.
const (
	ansiReset = "\033[0m"
	ansiBold  = "\033[1m"
	ansiFaint = "\033[2m"
	ansiCyan  = "\033[36m"
)

func bold(s string) string  { return ansiBold + s + ansiReset }
func faint(s string) string { return ansiFaint + s + ansiReset }
func cyan(s string) string  { return ansiCyan + s + ansiReset }

// ErrAborted is returned when the user declines the final confirmation.
var ErrAborted = errors.New("aborted by user")

// listItem is a row in the selector: either a selectable command or a
// non-selectable group header (when Header is non-empty).
type listItem struct {
	Cmd    config.Command
	Header string
}

// SelectCommand shows an interactive, filterable list of commands grouped into
// a local section then a global section (each pre-sorted a-z) and returns the
// chosen command.
func SelectCommand(commands []config.Command) (config.Command, error) {
	// Truncate the whole rendered line ("name — desc") to the terminal width so
	// an item never wraps to a second row. Wrapped rows break promptui's redraw
	// line accounting and make it stack new lines on every arrow key press.
	// Reserve 2 cols for the "> " / "  " prefix plus a 3-column safety margin
	// (some terminals soft-wrap when the final column is filled).
	limit := terminalWidth() - 5
	if limit < 10 {
		limit = 10
	}

	// Build display rows with a header before each non-empty group. commands is
	// already ordered local-first then global, each group a-z.
	var items []listItem
	firstCmd := -1
	addGroup := func(header string, src config.Source) {
		var added bool
		for _, c := range commands {
			if c.Source != src {
				continue
			}
			if !added {
				items = append(items, listItem{Header: header})
				added = true
			}
			if firstCmd < 0 {
				firstCmd = len(items)
			}
			items = append(items, listItem{Cmd: c})
		}
	}
	addGroup("── local commands ──", config.SourceLocal)
	addGroup("── global commands ──", config.SourceGlobal)
	if firstCmd < 0 {
		return config.Command{}, errors.New("no commands to select")
	}

	funcs := template.FuncMap{}
	for k, v := range promptui.FuncMap {
		funcs[k] = v
	}
	// row renders a command line truncated to the terminal width (computed on
	// the plain text), then styles the name bold and the description faint. When
	// accent is true the name also gets the cyan accent (used for the highlighted
	// row).
	row := func(it listItem, accent bool) string {
		full := it.Cmd.Name + " — " + it.Cmd.Description
		if r := []rune(full); len(r) > limit {
			full = string(r[:limit-1]) + "…"
		}
		rest := strings.TrimPrefix(full, it.Cmd.Name)
		name := it.Cmd.Name
		if accent {
			name = cyan(name)
		}
		return bold(name) + faint(rest)
	}
	funcs["rowActive"] = func(it listItem) string { return row(it, true) }
	funcs["rowInactive"] = func(it listItem) string { return row(it, false) }
	funcs["faint"] = faint

	prompt := promptui.Select{
		Label: "Select a command (type to filter)",
		Items: items,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "{{ if .Header }}  {{ .Header | faint }}{{ else }}> {{ rowActive . }}{{ end }}",
			Inactive: "{{ if .Header }}  {{ .Header | faint }}{{ else }}  {{ rowInactive . }}{{ end }}",
			Selected: "> {{ .Cmd.Name | cyan }}",
			FuncMap:  funcs,
		},
		Searcher: func(input string, index int) bool {
			// Empty query: show everything, including the group headers.
			if strings.TrimSpace(input) == "" {
				return true
			}
			// While filtering, headers are never matched.
			if items[index].Header != "" {
				return false
			}
			name := strings.ToLower(items[index].Cmd.Name)
			return strings.Contains(name, strings.ToLower(input))
		},
		// The default help line contains multibyte arrow glyphs whose display
		// width promptui miscounts, so its row is never cleared on redraw and
		// stacks on every key press. Hide it.
		HideHelp:  true,
		CursorPos: firstCmd,
		Size:      12,
		// Filter live as the user types instead of requiring "/" to enter
		// search mode first.
		StartInSearchMode: true,
	}

	// A header row is not selectable; if chosen, re-prompt.
	for {
		i, _, err := prompt.Run()
		if err != nil {
			return config.Command{}, err
		}
		if items[i].Header != "" {
			continue
		}
		return items[i].Cmd, nil
	}
}

// ShowDetail prints the full description and every argument and parameter with
// its description, before prompting for input.
func ShowDetail(cmd config.Command) {
	fmt.Printf("\n%s\n%s\n", bold(cyan(cmd.Name)), cmd.Description)
	fmt.Printf("%s %s\n", faint("Command:"), cmd.Command)
	if len(cmd.Arguments) > 0 {
		fmt.Printf("\n%s\n", bold("Arguments (required, in order):"))
		for _, a := range cmd.Arguments {
			fmt.Printf("  %s %s\n", cyan(a.Name), faint("— "+a.Description))
		}
	}
	if len(cmd.Parameters) > 0 {
		fmt.Printf("\n%s\n", bold("Parameters (optional):"))
		for _, p := range cmd.Parameters {
			kind := "boolean flag"
			if p.Valued {
				kind = "valued"
			}
			fmt.Printf("  %s %s\n", cyan(p.Flag), faint(fmt.Sprintf("(%s) — %s", kind, p.Description)))
		}
	}
	fmt.Println()
}

// CollectInputs prompts for arguments (mandatory, in order) then parameters
// (optional). Returns the collected Inputs.
func CollectInputs(cmd config.Command) (exec.Inputs, error) {
	var in exec.Inputs

	for _, a := range cmd.Arguments {
		prompt := promptui.Prompt{
			Label: fmt.Sprintf("%s (%s)", a.Name, a.Description),
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return errors.New("required")
				}
				return nil
			},
		}
		val, err := prompt.Run()
		if err != nil {
			return in, err
		}
		in.Args = append(in.Args, strings.TrimSpace(val))
	}

	for _, p := range cmd.Parameters {
		if p.Valued {
			prompt := promptui.Prompt{
				Label: fmt.Sprintf("%s (%s, Enter to skip)", p.Flag, p.Description),
			}
			val, err := prompt.Run()
			if err != nil {
				return in, err
			}
			if v := strings.TrimSpace(val); v != "" {
				in.Params = append(in.Params, fmt.Sprintf("%s=%s", p.Flag, v))
			}
			continue
		}
		// Boolean flag: y/N select.
		sel := promptui.Select{
			Label: fmt.Sprintf("Enable %s? (%s)", p.Flag, p.Description),
			Items: []string{"no", "yes"},
		}
		_, choice, err := sel.Run()
		if err != nil {
			return in, err
		}
		if choice == "yes" {
			in.Params = append(in.Params, p.Flag)
		}
	}

	return in, nil
}

// Confirm shows the assembled line and asks the user to confirm. Returns
// ErrAborted if the user declines.
func Confirm(line string) error {
	fmt.Printf("\n%s\n  %s\n", faint("Will run:"), bold(cyan(line)))
	prompt := promptui.Select{
		Label:    "Run this command?",
		Items:    []string{"No", "Yes"},
		HideHelp: true,
	}
	_, choice, err := prompt.Run()
	if err != nil || choice != "Yes" {
		// Error covers Ctrl-C / Esc; non-"Yes" covers an explicit decline.
		return ErrAborted
	}
	return nil
}
