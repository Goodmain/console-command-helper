// Package help provides the human-facing usage text (`cch help`) and the
// framework-agnostic config specification (`cch schema`) intended to be pasted
// into an AI prompt so an agent can generate a .cch.json for a project.
package help

// helpText is the short, human-facing usage shown by `cch help` / -h / --help.
const helpText = `cch — interactive helper for project console commands.

Commands defined in .cch.json files are listed in an interactive picker; you
choose one, fill its arguments, confirm, and cch runs it.

Usage:
  cch            pick and run a command (interactive)
  cch init       create a starter ./.cch.json in the current folder
  cch schema     print the config spec (paste into an AI to fill the config)
  cch version    print version information (aliases: -v, --version)
  cch help       show this help (aliases: -h, --help)

Config files:
  ~/.cch.json    global commands (available everywhere)
  ./.cch.json    local commands for the current folder (override global by name)
`

// schemaText is the self-contained, framework-agnostic specification of the
// .cch.json format: a Markdown guide, a worked example, a formal JSON Schema,
// and a final instruction for an AI agent.
const schemaText = "# cch configuration (.cch.json)\n" + `
cch reads commands from a JSON file named .cch.json. This document fully
describes that format. Use it to generate a .cch.json for a project.

## File structure

A .cch.json file is a JSON object with a single key "commands", an array of
command objects.

## Command fields

| Field        | Type    | Required | Meaning                                                        |
|--------------|---------|----------|----------------------------------------------------------------|
| name         | string  | yes      | Short identifier shown and filtered in the picker.             |
| description  | string  | yes      | One-line summary shown next to the name.                       |
| command      | string  | yes      | Base shell command to run (may include subcommands/flags).     |
| arguments    | array   | no       | Mandatory, ordered positional inputs (see below).              |
| parameters   | array   | no       | Optional flags the user may set or skip (see below).           |

### arguments[] (mandatory, ordered)

Each argument object has:
- name (string, required): label shown when prompting.
- description (string, required): explanation of the value.

Arguments are required and order matters. Their entered values are appended to
the command as bare tokens, in declared order.

### parameters[] (optional flags)

Each parameter object has:
- flag (string, required): the flag itself, e.g. "--step" or "-f".
- description (string, required): explanation of the flag.
- valued (boolean, required): true if the flag takes a value, false if it is a
  bare boolean flag.

A valued parameter (valued: true) is rendered as flag=value (e.g. --step=3) when
the user provides a value, and omitted when skipped. A boolean parameter
(valued: false) is appended bare (e.g. --force) only when the user enables it.

## How the command line is assembled

The final line is: command, then each argument value in order, then each chosen
parameter. It is executed with the system shell (sh -c), so pipes, &&, env vars,
and globs in "command" work.

Example: command "php artisan migrate", argument env="prod", parameter
--step=3 set, boolean --force not set, produces:

    php artisan migrate prod --step=3

## Worked example

` + "```json\n" + workedExample + "```\n" + `
## JSON Schema (draft-07)

` + "```json\n" + jsonSchema + "```\n" + `
## Instruction

Explore the project and produce a single valid JSON document in the format above
and write it to ./.cch.json. Output JSON only — no commentary. Use clear,
human-readable names and descriptions. Mark genuinely required inputs as
arguments and optional flags as parameters.
`

// workedExample is a complete, valid .cch.json. It is round-tripped through
// config.Config in a test to guard against drift from the Go structs.
const workedExample = `{
  "commands": [
    {
      "name": "migrate",
      "description": "Run database migrations for an environment",
      "command": "php artisan migrate",
      "arguments": [
        { "name": "env", "description": "Target environment, e.g. prod or staging" }
      ],
      "parameters": [
        { "flag": "--step", "description": "Number of migrations to run", "valued": true },
        { "flag": "--force", "description": "Force the operation in production", "valued": false }
      ]
    },
    {
      "name": "test",
      "description": "Run the test suite (no arguments)",
      "command": "npm test"
    }
  ]
}
`

// jsonSchema is a draft-07 JSON Schema for .cch.json.
const jsonSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "cch configuration",
  "type": "object",
  "required": ["commands"],
  "properties": {
    "commands": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name", "description", "command"],
        "properties": {
          "name": { "type": "string" },
          "description": { "type": "string" },
          "command": { "type": "string" },
          "arguments": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["name", "description"],
              "properties": {
                "name": { "type": "string" },
                "description": { "type": "string" }
              },
              "additionalProperties": false
            }
          },
          "parameters": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["flag", "description", "valued"],
              "properties": {
                "flag": { "type": "string" },
                "description": { "type": "string" },
                "valued": { "type": "boolean" }
              },
              "additionalProperties": false
            }
          }
        },
        "additionalProperties": false
      }
    }
  },
  "additionalProperties": false
}
`

// Help returns the human-facing usage text.
func Help() string { return helpText }

// Schema returns the framework-agnostic config specification for AI consumption.
func Schema() string { return schemaText }

// Example returns the worked-example .cch.json (exposed for drift testing).
func Example() string { return workedExample }
