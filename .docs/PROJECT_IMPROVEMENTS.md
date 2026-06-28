# Project Improvements

This document lists structural and logical improvements proposed for Zimvor, focusing on Golang best practices, SOLID principles, KISS design, and overall code quality.

> [!NOTE]
> **Checklist Usage:** Use this file as a checklist to track which of the improvements have already been implemented by checking the boxes (`[x]`).

---

## Proposed Improvements Checklist

- [ ] [Improvement 1: Fix Non-Portable Package Status Check (`which` -> `command -v`)](#1-fix-non-portable-package-status-check-which---command--v)
- [ ] [Improvement 2: Fix Broken Interactive Shell Execution (Inherit Stdio)](#2-fix-broken-interactive-shell-execution-inherit-stdio)
- [ ] [Improvement 3: Decouple Console UI/Presentation from Execution Logic](#3-decouple-console-uipresentation-from-execution-logic)
- [ ] [Improvement 4: Centralize Backups and Use Safe Streaming](#4-centralize-backups-and-use-safe-streaming)
- [ ] [Improvement 5: Add Strict TOML Validation for Typos/Unknown Keys](#5-add-strict-toml-validation-for-typosunknown-keys)
- [ ] [Improvement 6: Implement Integration Tests with `t.TempDir()`](#6-implement-integration-tests-with-ttempdir)

---

## Detailed Improvements

### 1. Fix Non-Portable Package Status Check (`which` -> `command -v`)

* **Category**: Reliability / Portability / DRY
* **Description**: 
  In [internal/app/packages.go](file:///home/anorneto/DEV/simple-machine-setup/internal/app/packages.go#L80), `statusPackages` checks whether a package is installed by running `which <binary>`. However, `which` is not portable and is missing from minimal environments (like Alpine/busybox). 
  In contrast, `runPackages` correctly uses the portable built-in shell tool `command -v`. Furthermore, the string in `statusPackages` has a typo with double spaces: `"which  "`.
* **KISS/SOLID Alignment**: Single Responsibility Principle & DRY. The check logic should be defined once, or at least share the same underlying command format.
* **Plan Prompt**:
  ```text
  Refactor package status checking in internal/app/packages.go by replacing the non-portable "which" check in statusPackages with "command -v", matching the logic in runPackages. Optionally, extract this logic into a method (p Package) CheckCommand() string to enforce DRY.
  ```

---

### 2. Fix Broken Interactive Shell Execution (Inherit Stdio)

* **Category**: Correctness / Bug Fix
* **Description**: 
  In [internal/app/exec.go](file:///home/anorneto/DEV/simple-machine-setup/internal/app/exec.go#L86-L97), `RunInteractive` executes shell commands but sets `cmd.Stdin`, `cmd.Stdout`, and `cmd.Stderr` to `nil`. In Go's `os/exec` package, assigning `nil` to stdio redirect them to `os.DevNull`. 
  This means users cannot see installation progress or interact with commands (e.g. typing a `sudo` password or answering yes/no prompts during `apt` installs).
* **KISS/SOLID Alignment**: Bug fix to align implementation with the interface contract/documented purpose of `RunInteractive`.
* **Plan Prompt**:
  ```text
  Fix the shell runner in internal/app/exec.go. In the RunInteractive function, set the exec.Cmd fields Stdin, Stdout, and Stderr to os.Stdin, os.Stdout, and os.Stderr respectively, enabling interactive processes (such as package managers prompting for sudo password) to correctly inherit standard streams.
  ```

---

### 3. Decouple Console UI/Presentation from Execution Logic

* **Category**: SOLID (Single Responsibility / Dependency Inversion)
* **Description**: 
  Currently, packages, dotfiles, and tasks files in `internal/app` print success/failure symbols, format headings, and apply lipgloss styles directly to stdout. This makes the logic difficult to unit test cleanly (as stdout gets polluted) and limits future extensions like a `--quiet` flag, `--json` reporter, or a custom logger.
* **KISS/SOLID Alignment**: Single Responsibility (core logic shouldn't care about rendering colors/symbols) and Dependency Inversion (programming to a presenter interface).
* **Plan Prompt**:
  ```text
  Introduce a UI interface in internal/app/ (e.g., type UI interface { Header(string); Success(id, msg string); Warning(id, msg string); Error(id, msg string) }) and a default ConsoleUI implementation using lipgloss. Inject this interface into the Apply struct, replacing all direct fmt.Print and lipgloss usages in packages.go, dotfiles.go, and tasks.go to decouple execution logic from UI rendering.
  ```

---

### 4. Centralize Backups and Use Safe Streaming

* **Category**: Usability / Robustness / KISS
* **Description**: 
  Currently, `dotfiles.go` backs up overwritten dotfiles directly in the target directory (e.g. `~/.gitconfig.bak.20260624-172047`), polluting the home directory. 
  In addition, `backupFile` reads the entire source file into memory via `os.ReadFile`. If a user accidentally registers a very large file, this can cause out-of-memory errors.
* **KISS/SOLID Alignment**: KISS (keep the user home directory clean) and Robustness (use streaming copies for files).
* **Plan Prompt**:
  ```text
  Refactor backupFile in internal/app/dotfiles.go to store backups in a centralized directory (e.g., ~/.zimvor/backups/) instead of polluting the home folder. Use io.Copy and os.Open/os.Create to stream bytes safely from the source to the backup location rather than reading the entire file into memory with os.ReadFile.
  ```

---

### 5. Add Strict TOML Validation for Typos/Unknown Keys

* **Category**: Robustness / Validation
* **Description**: 
  `BurntSushi/toml` by default silently discards keys that do not map to the Go struct field tags. If a user writes `installs` instead of `install`, or `binary_name` instead of `binary`, the key is ignored, and the validation succeeds but the program behaves unexpectedly.
* **KISS/SOLID Alignment**: Fail-fast validation. The configuration validation boundary should check that all provided properties are valid configuration keys.
* **Plan Prompt**:
  ```text
  Modify Load in internal/app/config.go to catch unrecognized TOML keys. Decode using toml.NewDecoder and check metadata.Undecoded() for keys present in the input file but not mapped to the Config struct. Report these unknown fields as validation errors so typos are caught immediately.
  ```

---

### 6. Implement Integration Tests with `t.TempDir()`

* **Category**: Testability / Robustness
* **Description**: 
  The project currently has unit tests for loading/validation, but lacks tests for `Apply.Run()`, package installations, dotfiles deployments, and task steps, as they interface directly with the real system.
* **KISS/SOLID Alignment**: Testability and regression safety.
* **Plan Prompt**:
  ```text
  Write an integration test in internal/app/apply_test.go (or integration_test.go) that constructs a temporary filesystem using t.TempDir(). Configure a mock configuration file with packages, dotfiles, and tasks, and call Apply.Run() to verify the creation, skip, and update stages of dotfiles, and the execution order of packages and tasks.
  ```
