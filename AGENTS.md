# AGENTS.md — rules for AI agents working on zimvor

This file is read by AI agents before they touch the project. Human
contributors should read it too — same rules apply. The goal is to keep
the codebase easy to read and easy to learn from.

**AGENTS.md is a first-class file in this repository.** It is committed,
versioned, and maintained alongside the code. Any agent or contributor
who changes project structure, layout rules, or workflow expectations
must update this file in the same change so the rules stay in sync with
the code they describe. Do not gitignore it, do not move it out of the
repo root, do not treat it as scratch.

## Project goal

A minimal cross-platform CLI for syncing dotfiles and installing dev tools
across Linux and macOS. One TOML manifest per OS, no templates, no
secrets, no magic. Read `README.md` for the user-facing docs and
`PROJECT_IDEAS.md` for the v2+ backlog.

## Layout rules

- **`cmd/zimvor/`** is the only place with `package main`. It contains the
  cobra command tree and the global flags. Keep CLI plumbing here, not
  in `internal/`.
- **`internal/app/`** is the only internal package. All non-CLI logic
  lives here. Put new logic in this package unless you have a strong
  reason to do otherwise.
- **`configs/`** holds data, not code. It is the manifest tree that ships
  with the binary.
- **Tests live next to the code they cover**, with the same package name
  and a `_test.go` suffix. Use `t.TempDir()` for filesystem fixtures.

### When to create a new package

Do not create a new package by default. Only do it when at least one of
the following is true:

- The package is intended to be imported by a future separate binary
  (e.g. a planned `zimvor-doctor` or `zimvor-scan`).
- The package wraps a heavy third-party dependency and we want the rest
  of the code to compile without it (e.g. optional `go-git`).
- The package is a stable, versioned API surface (we don't have one yet).

If you create a new package, write a top-of-file comment explaining *why*
this exists as its own package, not as a file in `internal/app`. Future
agents need to see the reasoning to avoid reverting the decision.

### When to create a new file in `internal/app/`

- A file should focus on one concern. A 2000-line `app.go` is wrong.
- Use a file name that matches its single export (e.g. `Apply` lives in
  `apply.go`, `Runner` lives in `exec.go`).
- Helpers that are only used by one file should be unexported and live
  in that same file.

## Code style

- **Prefer small, readable functions** over clever ones. If a function
  doesn't fit on one screen, look for a way to split it.
- **Add comments to non-obvious code.** The user is learning Go from
  this project. A line that explains *why* a particular approach was
  taken is more valuable than a line that restates *what* the code does.
- **Do not comment obvious code.** `// increment i` above `i++` is noise.
- **Validate at the boundary.** `Config.Validate()` collects every
  problem before returning so users can fix the file in one pass — keep
  that pattern when adding new config sections.
- **Default to dry-run.** Every command that can change the system must
  accept `--apply` to opt in. `install` without `--apply` must never
  modify the machine.
- **Confirm before destructive actions.** Every overwrite, install, or
  side-effecting command shows a `[y/n]` prompt unless `--yes` is set.
- **Idempotent.** Running `install` twice should not break the second
  run. Always check before acting.

## Schema rules

- Every repeated config item has a stable `id` field. Use the id in
  output (logs, error messages) instead of the index.
- `target` paths in dotfiles must be absolute or start with `~/`.
  Anything else fails `Config.Validate()`.
- The `[[packages]]` schema is intentionally flat: `id` required,
  `install` required (a list of shell commands to run in order),
  `binary` optional (defaults to `id`). Do not reintroduce a `manager`
  field, an apt/brew/mise abstraction, or a separate "package names"
  list — the user writes the install command for their OS. This keeps
  the code simple and the manifest readable.
- New `stage` values must be added to `Config.Validate()` in the same
  commit.

## Error handling

- Wrap errors with `fmt.Errorf("...: %w", err)` so callers can `errors.Is`
  on them when needed.
- For user-facing load/validate errors, prefer one joined error that
  lists every problem over bailing on the first one.
- Never silently swallow an error. Log it, count it, or return it.

## Build & test

- `mise run build` builds the binary into `dist/zimvor`.
- `mise run test` runs `rtk go test ./...`.
- `mise run lint` runs `rtk golangci-lint run`.
- Both `rtk go vet ./...` and `mise run lint` should always pass for code changes.
- Add a unit test for any new exported function or any new branch in an
  existing one.

## Things to avoid

- No new top-level packages inside `internal/` without a written reason.
- No interface abstractions until there are at least two concrete
  implementations that would benefit.
- No template engines, secret stores, plugin systems, or external
  config formats. (All are listed in `PROJECT_IDEAS.md` and intentionally
  deferred.) The one exception is `text/template` in the `init` command
  for one-shot scaffolding of the starter config — see
  `cmd/zimvor/template.toml`. Template rendering inside dotfile content
  is still v2+.
- No commits that change behavior without updating `README.md` if the
  user-facing surface changes.
