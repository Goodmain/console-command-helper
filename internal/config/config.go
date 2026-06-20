// Package config defines the .cch.json schema and handles loading, merging,
// and scaffolding of command configuration files.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// FileName is the hidden config file name used both globally (in $HOME) and
// locally (in the current folder).
const FileName = ".cch.json"

// Config is the top-level structure of a .cch.json file.
type Config struct {
	Commands []Command `json:"commands"`
}

// Source identifies which config file a command came from.
type Source int

const (
	// SourceLocal is the current-folder config (./.cch.json).
	SourceLocal Source = iota
	// SourceGlobal is the home-folder config (~/.cch.json).
	SourceGlobal
)

// Command is a single runnable entry: a base shell command plus the arguments
// and parameters the user is prompted for.
type Command struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Command     string      `json:"command"`
	Arguments   []Argument  `json:"arguments,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty"`

	// Source is set during loading (not read from JSON) to indicate origin.
	Source Source `json:"-"`
}

// Argument is a mandatory, positional input appended to the command in order.
type Argument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Parameter is an optional flag. When Valued is true the user is prompted for a
// value and it renders as --flag=value; otherwise it is a boolean flag rendered
// bare as --flag when enabled.
type Parameter struct {
	Flag        string `json:"flag"`
	Description string `json:"description"`
	Valued      bool   `json:"valued"`
}

// LoadFile reads and parses a single config file. A missing file yields an
// empty Config and no error. Malformed JSON yields an error naming the file.
func LoadFile(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing %s: %w", path, err)
	}
	return cfg, nil
}

// GlobalPath returns the path to the global config (~/.cch.json).
func GlobalPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, FileName), nil
}

// LocalPath returns the path to the current-folder config (./.cch.json).
func LocalPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, FileName), nil
}

// LoadMerged loads the global and local configs and merges them. Commands are
// keyed by name; a local command overrides a global command with the same name.
// The result is sorted alphabetically (a-z) by name.
func LoadMerged() ([]Command, error) {
	merged := map[string]Command{}

	globalPath, err := GlobalPath()
	if err != nil {
		return nil, err
	}
	global, err := LoadFile(globalPath)
	if err != nil {
		return nil, err
	}
	for _, c := range global.Commands {
		c.Source = SourceGlobal
		merged[c.Name] = c
	}

	localPath, err := LocalPath()
	if err != nil {
		return nil, err
	}
	local, err := LoadFile(localPath)
	if err != nil {
		return nil, err
	}
	for _, c := range local.Commands {
		c.Source = SourceLocal
		merged[c.Name] = c
	}

	commands := make([]Command, 0, len(merged))
	for _, c := range merged {
		commands = append(commands, c)
	}
	sortGrouped(commands)
	return commands, nil
}

// sortGrouped orders commands local-first then global, each group sorted
// alphabetically (a-z) by name, in place.
func sortGrouped(commands []Command) {
	sort.Slice(commands, func(i, j int) bool {
		if commands[i].Source != commands[j].Source {
			return commands[i].Source < commands[j].Source
		}
		return commands[i].Name < commands[j].Name
	})
}
