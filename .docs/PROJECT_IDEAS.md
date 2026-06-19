# Project Ideas - Future Enhancements

## 5. `--config` flag
Override config file path via `--config` / `-c` flag. Useful for:
- Testing different configurations
- Multiple machine profiles (work vs personal)
- CI/CD scenarios

## 6. Structured section output
Use lipgloss to render clear phase headers and status indicators:
```
==> Installing packages
  ✓ fish (already installed)
  ⚠ git (installing...)
  ✗ mise (failed: network error)
```

## 7. Continue-on-error
If one package or step fails, report it and continue with remaining steps. Show summary at end:
```
Setup complete:
  Packages: 4 succeeded, 1 failed
  Dotfiles: 3 updated, 1 skipped
  Mise tools: 12 installed
```

## 8. Version embed
`zimvor --version` with build metadata via ldflags:
```
zimvor v0.1.0 (2026-01-18, commit: abc1234)
```

Build task in mise.toml:
```toml
[tasks.build]
run = """
VERSION=${VERSION:-dev}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
go build -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
  -o dist/zimvor ./cmd/zimvor
"""
```
