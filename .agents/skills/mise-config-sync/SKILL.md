---
name: mise-config-sync
description: Use when organizing, categorizing, or syncing the mise tool list in configs/dotfiles/mise.toml or ~/.config/mise/config.toml. Triggers: "mise config", "organize mise tools", "mise sections", "mise tool list", "sync mise config". Groups tools into labeled sections with concise per-tool comments.
---

# mise-config-sync

Keep `configs/dotfiles/mise.toml` organized: tools grouped into labeled
sections, each with a concise comment. This file is the source for
`~/.config/mise/config.toml` (see the `mise-config` dotfile in
`configs/linux.toml` / `configs/darwin.toml`).

## Workflow

1. Run `mise list`. Columns: `name  installed  source  requested`. Keep rows
   where `source` is `~/.config/mise/config.toml`. Ignore rows with no source.

2. Use the `requested` column from `mise list` for each tool's version. Never
   pin to the installed version — keep `python = "3.14"`, not `"3.14.6"`.

3. Read `configs/dotfiles/mise.toml`. Preserve `[settings]` and every
   non-`[tools]` section verbatim. Rewrite only `[tools]`.

4. If a tool in the file is missing from the filtered `mise list` output,
   comment out its line instead of removing it. Preserves the user's intent
   for tools that are temporarily not installed.

5. Group tools under `# Section` comments. Reuse existing sections first.
   Starting taxonomy:
   - `# Programming languages & runtimes`
   - `# Package managers`
   - `# Linters and code analysis`
   - `# Shell and terminal UX`
   - `# AI assistants`
   - `# Security`
   - `# Other` (always last; fallback)

6. One concise aligned `#` comment per tool line. Comment every tool, even
   well-known ones.

7. Write the file: preserved sections, blank line, `[tools]`, blank line,
   grouped tool lines, one blank line between sections, aligned comments.

## Notes

- Preserve `[settings]` / `idiomatic_version_file_enable_tools` exactly.
- Data file under `configs/` — hand-editable, no code/templates.
- Version-controlled; `git diff` is the review step.
