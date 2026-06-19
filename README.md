# zimvor

A minimal cross-platform CLI for syncing dotfiles and installing dev tools
across Linux and macOS machines. One TOML manifest per OS, no templates, no
magic, no secrets.

## What it does

- **Installs system packages** by running the shell commands you list in the manifest
- **Syncs dotfiles** from this repo to your home directory, with colored
  diffs shown before any overwrite and timestamped backups kept
- **Runs tasks** (arbitrary shell commands) before and after the main flow
- **Reports status** of every declared package and dotfile
- **Dry-runs by default** — you must pass `--apply` to make changes

## What it does NOT do (v1)

- No template rendering in dotfiles
- No secrets management
- No git pull/push against a remote
- No Windows-specific commands tested (the schema is platform-agnostic; you can write PowerShell commands in `install`, but it's untested)
- No directory sync (single files only)
- No auto-detection of installed tools

See [PROJECT_IDEAS.md](PROJECT_IDEAS.md) for v2+ ideas.

## Build & install

Build is driven by [mise](https://mise.jdx.dev) tasks in `mise.toml`.

```bash
# Build for the current platform
mise run build

# Cross-compile for all targets (linux/darwin, amd64/arm64)
mise run build-all

# Binaries are written to dist/
./dist/zimvor --help
```

You can also build with plain `go`:

```bash
go build -o dist/zimvor ./cmd/zimvor
```

## Usage

### First run

```bash
# Scaffold a starter config for your OS
zimvor init

# See what's currently installed and in sync
zimvor status

# See what would change (dry-run)
zimvor diff
```

### Install (dry-run is the default)

```bash
# Full flow: pre-tasks, packages, dotfiles, post-tasks
zimvor install            # dry-run — prints what would happen
zimvor install --apply    # actually execute
zimvor install --apply --yes   # skip all confirmations

# Sub-commands for partial runs
zimvor install packages   # packages only
zimvor install dotfiles   # dotfiles only
zimvor install tasks      # tasks only (pre + post)
```

### Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--apply` | `-a` | `false` | Actually execute changes (without this, `install` is dry-run) |
| `--yes` | `-y` | `false` | Skip all confirmation prompts |
| `--config` | `-c` | auto | Path to the config directory (default: `configs/` next to the binary or in CWD) |

## Configuration

### Layout

```
configs/
├── linux.toml          # Linux/WSL manifest
├── darwin.toml         # macOS manifest
└── dotfiles/           # source files referenced by [[dotfiles]] entries
    ├── fish/
    │   ├── config.fish
    │   └── aliases.fish
    └── mise/
        └── config.toml
```

The manifest filename is auto-selected from `runtime.GOOS` (linux → `linux.toml`,
darwin → `darwin.toml`). You can override with `--config`.

### Manifest schema

```toml
[meta]
os = "linux"
description = "WSL / Linux workstation"

# Packages to install. `id` is a stable identifier shown in logs.
# `install` is a list of shell commands to run, in order, to install
# the package. You write the command for your OS — we just run it.
# `binary` is optional; when set, it's the name we look for on PATH to
# decide whether the package is already installed. When unset, we use
# the `id` instead.
[[packages]]
id = "git"
install = ["sudo apt install -y git"]

# Example with a different binary name than the id (e.g. `ripgrep`
# installs the `rg` binary).
[[packages]]
id = "ripgrep"
binary = "rg"
install = ["sudo apt install -y ripgrep"]

# Dotfiles to sync. `source` is relative to the configs/ directory;
# `target` is an absolute path or starts with `~/`.
[[dotfiles]]
id = "fish-config"
source = "dotfiles/fish/config.fish"
target = "~/.config/fish/config.fish"

# Shell commands to run. `stage` is "pre" (before packages/dotfiles) or
# "post" (after).
[[tasks]]
id = "set-fish-default"
description = "Set fish as default shell"
stage = "post"
command = "chsh -s $(which fish)"
```

The TOML schema is validated on load; you'll get a list of every problem in
one error if the file is malformed.

## Behavior

- **Backup before overwrite:** when a dotfile would change, the existing file
  is copied to `<target>.bak.<timestamp>` first.
- **Skip on match:** identical files are reported as `up to date` and not touched.
- **Confirm before destructive changes:** every overwrite (and every package
  install) asks `[y/n]` unless `--yes` is set.
- **Diff before prompt:** when a dotfile would be overwritten, the colored
  unified diff is printed before the prompt.
- **Idempotent:** safe to re-run any time; only changed or missing items are
  acted on.

## Project layout

```
cmd/zimvor/        # CLI entry point (cobra commands)
internal/app/      # all non-CLI logic, in one package
  config.go        # TOML structs, Load, Validate
  platform.go      # GOOS detection
  exec.go          # shell command runner with dry-run support
  styles.go        # all lipgloss styles in one place
  diff.go          # unified diff with lipgloss coloring
  prompt.go        # y/n confirmation helper
  apply.go         # orchestrator (Apply type, full flow, summary)
  packages.go      # package install phase
  dotfiles.go      # dotfile sync phase (diff/confirm/backup/write)
  tasks.go         # pre/post task runner
  *_test.go        # tests live next to the code they cover
configs/           # manifests + dotfile source tree
```

We keep everything under `internal/app` in one package. The earlier layout
split this across seven tiny packages (`config`, `apply`, `exec`, `diff`,
`output`, `prompt`, `platform`), but every one of them was used by every
other one. Splitting added import noise without any isolation benefit. If a
future piece genuinely needs to be reused outside this binary, it can be
moved into its own subpackage then.

## Development

```bash
# Run all tests
mise run test

# Cross-compile for all platforms
mise run build-all

# Static analysis
go vet ./...
```

## License

MIT
