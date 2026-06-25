# cch — console command helper

[![Tests](https://github.com/Goodmain/console-command-helper/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Goodmain/console-command-helper/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/Goodmain/console-command-helper/graph/badge.svg)](https://codecov.io/gh/Goodmain/console-command-helper)
[![Release](https://img.shields.io/github/v/release/Goodmain/console-command-helper)](https://github.com/Goodmain/console-command-helper/releases)

`cch` is an interactive helper for project console commands. Define your commands once in a `.cch.json` file, then pick, fill, confirm, and run them from a searchable picker — no need to remember exact flags or argument order.

## Why

Projects accumulate commands you run rarely and forget: migrations, seeders, deploy scripts, one-off maintenance tasks. `cch` keeps them documented and runnable in one place. Each command declares its required arguments and optional flags, so the picker walks you through filling them and shows the final line before it runs.

## Installation

### Homebrew

```bash
brew tap Goodmain/console-command-helper
brew install cch
```

### Build from source

- Go 1.25+

```bash
go build .
```

## Usage

```
cch            pick and run a command (interactive)
cch init       create a starter ./.cch.json in the current folder
cch schema     print the config spec (paste into an AI to fill the config)
cch version    print version information (aliases: -v, --version)
cch help       show help (aliases: -h, --help)
```

Run `cch` with no arguments to open the interactive picker. Choose a command, fill its arguments, set any optional flags, confirm the assembled line, and `cch` executes it.

## Config files

`cch` merges two config files:

| File          | Scope                                              |
|---------------|----------------------------------------------------|
| `~/.cch.json` | Global commands, available everywhere.             |
| `./.cch.json` | Local commands for the current folder.             |

Commands are keyed by `name`. A local command **overrides** a global command with the same name. In the picker, local commands are listed first, then global, each group sorted alphabetically.

## Config format

A `.cch.json` is a JSON object with a single `commands` array.

### Command fields

| Field         | Type   | Required | Meaning                                                    |
|---------------|--------|----------|------------------------------------------------------------|
| `name`        | string | yes      | Short identifier shown and filtered in the picker.         |
| `description` | string | yes      | One-line summary shown next to the name.                   |
| `command`     | string | yes      | Base shell command (may include subcommands/flags).        |
| `arguments`   | array  | no       | Mandatory, ordered positional inputs.                      |
| `parameters`  | array  | no       | Optional flags the user may set or skip.                   |

### `arguments[]` — mandatory, ordered

- `name` (string): label shown when prompting.
- `description` (string): explanation of the value.

Values are appended to the command as bare tokens, in declared order.

### `parameters[]` — optional flags

- `flag` (string): the flag, e.g. `--step` or `-f`.
- `description` (string): explanation of the flag.
- `valued` (boolean): `true` if the flag takes a value, `false` if bare boolean.

A valued parameter renders as `flag=value` (e.g. `--step=3`) when set, omitted when skipped. A boolean parameter is appended bare (e.g. `--force`) only when enabled.

## How the line is assembled

The final line is: `command`, then each argument value in order, then each chosen parameter. It runs through the system shell (`sh -c`), so pipes, `&&`, env vars, and globs in `command` work.

Example — `command` `php artisan migrate`, argument `env=prod`, `--step=3` set, boolean `--force` not set, produces:

```sh
php artisan migrate prod --step=3
```

## Example config

```json
{
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
```

## Generating a config with AI

`cch schema` prints a self-contained spec (Markdown guide + worked example + JSON Schema + an instruction). Pipe it to an AI agent to scaffold a `.cch.json` for your project:

```sh
cch schema | pbcopy   # paste into your AI of choice
```

## Project layout

```
main.go              CLI entry point and interactive flow
internal/config/     .cch.json schema, load/merge, init scaffolding
internal/exec/       command-line assembly and shell execution
internal/help/       help text and AI-facing config spec
internal/ui/         interactive picker and prompts (promptui)
```

## Development

Run the unit test suite with coverage:

```sh
go test -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Run the end-to-end smoke test (gated behind the `integration` build tag):

```sh
go test ./test/smoke -tags=integration -v
```

CI runs tests, coverage, `golangci-lint`, and a cross-platform build matrix
(Linux/macOS/Windows) on every push and PR to `main` — see
[.github/workflows/ci.yml](.github/workflows/ci.yml).

## Releases

Releases are automated with [GoReleaser](https://goreleaser.com). Pushing a
`v*` tag triggers [.github/workflows/release.yml](.github/workflows/release.yml),
which builds cross-platform binaries, publishes a GitHub Release with archives
and checksums, and updates the Homebrew tap.

```sh
git tag v0.1.0
git push origin v0.1.0
```

Required repository secrets:

- `HOMEBREW_TAP_GITHUB_TOKEN` — token with write access to the
  `Goodmain/homebrew-console-command-helper` tap repo (for the Homebrew formula update).
- `CODECOV_TOKEN` — optional, for coverage uploads in CI.

## License

[MIT](LICENSE)
