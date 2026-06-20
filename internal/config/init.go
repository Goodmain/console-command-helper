package config

import (
	"fmt"
	"os"
)

// exampleConfig is the self-documenting starter written by `cch init`.
var exampleConfig = `{
  "commands": [
    {
      "name": "example",
      "description": "Example command - edit or replace me",
      "command": "echo hello",
      "arguments": [
        { "name": "target", "description": "Who to greet (mandatory, appended in order)" }
      ],
      "parameters": [
        { "flag": "--loud", "description": "Boolean flag, appended bare when enabled", "valued": false },
        { "flag": "--times", "description": "Valued flag, rendered as --times=value", "valued": true }
      ]
    }
  ]
}
`

// InitResult describes the outcome of an init attempt for the caller to report.
type InitResult struct {
	Path   string
	Exists bool // a config already existed at Path
}

// Init writes the example config to ./.cch.json. If a file already exists it is
// only overwritten when force is true; otherwise InitResult.Exists is true and
// no write occurs.
func Init(force bool) (InitResult, error) {
	path, err := LocalPath()
	if err != nil {
		return InitResult{}, err
	}
	if _, err := os.Stat(path); err == nil && !force {
		return InitResult{Path: path, Exists: true}, nil
	} else if err != nil && !os.IsNotExist(err) {
		return InitResult{Path: path}, fmt.Errorf("checking %s: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(exampleConfig), 0o644); err != nil {
		return InitResult{Path: path}, fmt.Errorf("writing %s: %w", path, err)
	}
	return InitResult{Path: path}, nil
}
