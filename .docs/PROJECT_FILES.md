# Project Files

This repository is a small Go CLI for cross-platform machine setup and dotfile deployment.

## Top-level files

- `README.md`
  - Project overview, usage examples, configuration format, and development commands.
- `go.mod`, `go.sum`
  - Go module dependencies and versioning.
- `mise.toml`
  - Configuration for the `mise` tool used by the project.

## `cmd/zimvor/main.go`

Entry point for the `zimvor` CLI.
- Defines global flags: `--dry-run`, `--yes`, and `--config`.
- Registers commands:
  - `run` - full setup sequence
  - `packages` - install packages only
  - `dotfiles` - deploy dotfiles only
  - `tools` - run post-install commands only
  - `diff` - preview dotfiles changes
  - `status` - report package and dotfile status
- Loads configuration via `internal/config` and delegates execution to `internal/setup`.

## `internal/config/config.go`

Configuration loading and model definitions.
- Parses TOML config files into `Config`, `Package`, `Dotfile`, and `PostInstall` structs.
- Provides `FindConfigPath()` to auto-locate the `configs/` directory.

## `internal/platform/platform.go`

Platform detection helper.
- Uses Go runtime to identify OS.
- Selects config filename like `linux.toml` or `darwin.toml`.

## `internal/setup/setup.go`

Central orchestration layer.
- Creates `Setup` objects with config, runner, and options.
- Implements high-level operations:
  - `Run()` → packages, dotfiles, post-install
  - `Packages()` → package installation
  - `Dotfiles()` → dotfile deployment
  - `PostInstall()` → post-install commands
  - `Status()` → status checks
  - `Diff()` → dry-run dotfile diffing

## `internal/setup/packages.go`

Package installer logic.
- Checks whether each package is already installed.
- Runs install commands interactively.
- Supports dry-run and auto-yes confirmation.

## `internal/setup/dotfiles.go`

Dotfile deployment logic.
- Reads source dotfiles from `configs/`.
- Expands destination paths like `~/...`.
- Compares existing destination content.
- Shows diffs and backs up files before overwriting.
- Supports dry-run mode and auto-yes confirmation.

## `internal/setup/mise.go`

Post-install command executor.
- Runs configured post-install commands from `post_install`.
- Supports dry-run and auto-yes.

## `internal/setup/status.go`

Status reporting.
- Verifies package installed state using check commands.
- Compares deployed dotfiles with source content.
- Reports missing or differing files.

## `internal/runner/runner.go`

Shell command execution helper.
- Runs shell commands with optional `DryRun` mode.
- Captures output and exit codes.
- Provides `RunInteractive()` for interactive install commands.
- Provides `Check()` to determine command success.

## `internal/diff/diff.go`

Diff generation and styling.
- Generates unified diffs for old/new file contents.
- Adds colored output for additions, removals, and context lines.

## `internal/prompt/prompt.go`

Interactive confirmation helper.
- Prompts the user for `y/n` confirmation.
- Supports auto-yes bypass.
