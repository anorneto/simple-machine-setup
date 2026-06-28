# Zimvor — Future Ideas (v2+)

These are features that are intentionally out of v1 scope. They are listed
here so we can revisit them later without losing the context.

## Commands

### `zimvor scan`

Auto-detect installed tools on the current machine and suggest a TOML
config. Would run `command -v`, `mise ls`, `brew list`, etc. and emit a
starter config the user can review and commit. Fragile by nature — needs
heuristics per manager — so deferred until the core config format is
stable.

### `zimvor pull`

Pull latest config from a configured git remote. Requires `go-git` or
shelling out to `git`. Needs a conflict resolution strategy (ours?
theirs? prompt?).

### `zimvor push`

Push local config changes to a git remote. Same dependency as `pull`.

### `zimvor run <task-id>`

Run a single named task by ID instead of all pre/post tasks. Useful for
re-running one-off setup steps without the full install flow.

## Config features

### `kind = "directory"` for dotfiles

Sync an entire directory recursively instead of a single file. Needs logic
for recursive copy, per-file conflict detection, and ignore patterns
(`.zimvorignore` style).

### Template rendering in dotfiles

Substitute variables (e.g. `{{ .hostname }}`, `{{ .email }}`) inside
dotfile content before deploying. Explicitly out of v1 to avoid
chezmoi-style complexity, but useful for machine-specific differences
inside the same dotfile.

### Secrets management

Encrypt sensitive values (API keys, tokens) in the config and decrypt at
deploy time. Needs keyring / `age` / `gpg` integration. Separate concern
from the core sync flow.

### Pre/post/per-dotfile/per-package tasks

Right now tasks are global. It would be useful to attach a pre/post task
to a single package or dotfile (e.g., "after installing fish, fetch my
fisher plugins"). Could be added as optional fields on the existing
structs without breaking the schema.

## Package managers

### Windows support (winget, scoop, choco)

The `manager` field already supports extensibility. Adding `winget`,
`scoop`, or `choco` is structurally straightforward (one new case in
`buildInstallCommand`) but needs testing on real Windows machines.

### `Manager` interface

Replace the `switch/case` in `internal/apply/packages.go` with a small
interface:

```go
type Manager interface {
    IsInstalled(binary string) bool
    Install(packages []string) (string, error) // returns the command to run
}
```

Each manager (apt, brew, mise, curl) becomes its own type implementing
the interface. Cleaner abstraction, easier to test, but overkill for 4
managers in v1.

## UX

### `charmbracelet/huh` for interactive prompts

Replace the simple `bufio.Scanner` y/n prompt with styled, interactive
confirmations. Pulls in `bubbletea` as a dependency — nice polish but not
essential.

### Continue-on-error with `--continue` flag

Currently `install` stops on the first failure within a phase and moves
on to the next phase. A `--continue` flag would let the user opt into
"run everything, report all errors at the end" behavior.

### `zimvor doctor`

Check that required system dependencies are available (`git`, `curl`,
`sudo`, the package manager binary, etc.) and report any issues before
running `install`.

### `pushd` for dotfile sources

A `~/.gitconfig` and `~/dotfiles` are typically the same file. A
`type = "link"` mode could symlink instead of copying, so manual edits
on the machine show up in the repo's working tree. Risky — needs care
to avoid breaking workflows.

## Git sync

### Treat `configs/` as a git repo

`pull` and `push` become thin wrappers around `git pull` / `git push`.
Could also support branches for different machine profiles (work vs
personal), with `zimvor switch <profile>` switching which manifest is
active.

### Merge conflict handling

When `pull` produces a merge conflict in a manifest TOML, zimvor could
detect it and refuse to run `install` until the user resolves the
conflict manually.

## Testing

### End-to-end tests with a fake home dir

v1 has unit tests for command generation and validation, but no test
that exercises a full install against a temp directory tree. A
`internal/apply/integration_test.go` using `t.TempDir()` for both the
configs/ root and a fake home would catch a lot of regressions.

### Golden output tests for the diff renderer

The diff output is colorized via lipgloss. Snapshot tests would catch
unintended formatting changes.

## Repository maintenance

These were reviewed and intentionally deferred — relevant once the
project grows beyond solo development or ships binaries to users.

### CI workflow (`.github/workflows/ci.yml`)

Run `go test ./...`, `golangci-lint run`, `betterleaks git .` (full
history), and a build matrix for the four targets defined in
`mise.toml`. Without CI, lefthook is the only gate and runs only on
the developer's machine.

### `goreleaser` for releases

Cross-compile to all four targets defined in `mise.toml`, package
binaries, sign them, and publish a GitHub Release. Pairs naturally
with `cog bump --auto` to cut releases from conventional commits.

### Renovate or Dependabot

Automated dependency updates. Useful for keeping `lefthook`,
`cocogitto`, `golangci-lint`, and Go modules current. Renovate has
better support for mise tool versions and groups PRs by scope.

### `actionlint` in pre-commit

Lints GitHub Actions workflow YAML for subtle mistakes. Only useful
once a CI workflow exists.

### `betterleaks` in CI (full-history scan)

The pre-commit hook uses `--staged`; in CI the goal is to catch
secrets that landed before the hook was set up. Run
`betterleaks git .` (no `--staged`) as a CI step.

### `.editorconfig`

5 lines to prevent tab/space/EOF wars across editors. Useful once
multiple people contribute.

### `.gitattributes`

At minimum, mark `configs/` as code for GitHub's language stats and
to ensure YAML/TOML files render correctly in the diff view.

### `CODEOWNERS`

Auto-assign reviews. Worth adding when the project has multiple
contributors or wants to protect `configs/` and `cog.toml` from
drive-by changes.

### Issue and PR templates

`.github/ISSUE_TEMPLATE/` and `.github/PULL_REQUEST_TEMPLATE.md` to
nudge contributors toward cog-compliant messages and mention running
`mise run test lint` before opening a PR.

### `SECURITY.md`

Vulnerability disclosure policy. Pairs with the betterleaks setup;
tells people how to report a vulnerability rather than filing a
public issue.
