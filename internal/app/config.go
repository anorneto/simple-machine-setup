// Package app contains all of zimvor's non-CLI logic.
//
// It is intentionally one package. If a future piece genuinely
// needs to be reused outside this binary, it can be split out then.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config is the top-level TOML structure for one OS manifest.
type Config struct {
	Meta     Meta      `toml:"meta"`
	Packages []Package `toml:"packages"`
	Dotfiles []Dotfile `toml:"dotfiles"`
	Tasks    []Task    `toml:"tasks"`
}

// Meta holds metadata about the config file itself.
type Meta struct {
	OS          string `toml:"os"`
	Description string `toml:"description"`
}

// Validate checks the Meta section for required fields and returns a list of errors.
func (m Meta) Validate() []string {
	var errs []string
	if strings.TrimSpace(m.OS) == "" {
		errs = append(errs, "meta.os is required")
	}
	if strings.TrimSpace(m.Description) == "" {
		errs = append(errs, "meta.description is required")
	}
	return errs
}

// Package declares a system-level package to install:
//   - "ID" : stable identifier used in logs and errors
//   - "Install" : shell commands to run, to install the package.
//     The user writes commands appropriate to their OS; we just run them.
//   - "Binary" : optional; when set we look for this on PATH to decide
//     if the package is already installed. Defaults to ID.
type Package struct {
	ID      string   `toml:"id"`
	Binary  string   `toml:"binary"`
	Install []string `toml:"install"`
}

// GetID returns the stable identifier for the package.
func (p Package) GetID() string { return p.ID }

// BinaryOrID returns the name we should use when checking PATH for the
// installed tool. It returns Binary if set, otherwise ID.
func (p Package) BinaryOrID() string {
	if p.Binary != "" {
		return p.Binary
	}
	return p.ID
}

func (p Package) Validate() []string {
	var errs []string
	if strings.TrimSpace(p.ID) == "" {
		errs = append(errs, "packages.id is required")
	}
	if len(p.Install) == 0 {
		errs = append(errs, fmt.Sprintf("packages(id=%q).install must not be empty", p.ID))
	}
	return errs
}

// Dotfile declares a single file to sync from the repo to the machine:
//   - "ID" : stable identifier used in logs and errors
//   - "Source" is relative to the configs/ directory.
//   - "Target" is an absolute path or ~/... path on the machine.
type Dotfile struct {
	ID     string `toml:"id"`
	Source string `toml:"source"`
	Target string `toml:"target"`
}

// GetID returns the stable identifier for the dotfile.
func (d Dotfile) GetID() string { return d.ID }

// Validate checks the dotfile for required fields and returns a list of errors.
func (d Dotfile) Validate() []string {
	var errs []string

	if strings.TrimSpace(d.ID) == "" {
		errs = append(errs, "dotfiles.id is required")
	}
	if strings.TrimSpace(d.Source) == "" {
		errs = append(errs, fmt.Sprintf("dotfiles(id=%q).source is required", d.ID))
	}
	if strings.TrimSpace(d.Target) == "" {
		errs = append(errs, fmt.Sprintf("dotfiles(id=%q).target is required", d.ID))
	} else if !strings.HasPrefix(d.Target, "/") && !strings.HasPrefix(d.Target, "~/") {
		errs = append(errs, fmt.Sprintf("dotfiles(id=%q).target must be an absolute or home path", d.ID))
	}
	return errs
}

// Task declares a shell command to run before or after the main install flow:
//   - "ID" : stable identifier used in logs and errors
//   - "Description" : human-readable description of what the task does
//   - "Stage" : either "pre" or "post", to run before or after the main install flow
//   - "Command" : shell command to run. The user writes commands appropriate to their OS; we just run them.
type Task struct {
	ID          string `toml:"id"`
	Description string `toml:"description"`
	Stage       string `toml:"stage"`
	Command     string `toml:"command"`
}

// GetID returns the stable identifier for the task.
func (t Task) GetID() string { return t.ID }

// Validate checks the task for required fields and returns a list of errors.
func (t Task) Validate() []string {
	var errs []string

	if strings.TrimSpace(t.ID) == "" {
		errs = append(errs, "tasks.id is required")
	}
	if strings.TrimSpace(t.Command) == "" {
		errs = append(errs, "tasks.command is required")
	}
	if strings.TrimSpace(t.Stage) == "" {
		errs = append(errs, "tasks.stage is required")
	} else {
		switch t.Stage {
		case "pre", "post": // TODO(anor): maybe add enum later
			// valid
		default:
			errs = append(errs, "tasks.stage must be 'pre' or 'post'")
		}
	}
	return errs
}

// Load reads and parses a TOML config file from the given path.
// It does not validate; call Validate() for that.
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks the config against all rules and returns a joined error
// listing every problem found, or nil if the config is valid.
// We collect every error before returning so the user can fix the whole
// file in one pass instead of playing whack-a-mole.
func (cfg *Config) Validate() error {
	var errs []string

	// Meta
	errs = append(errs, cfg.Meta.Validate()...)
	// Packages
	pkgIDs := make(map[string]bool)
	for _, p := range cfg.Packages {
		errs = append(errs, p.Validate()...)
		if pkgIDs[p.ID] {
			errs = append(errs, fmt.Sprintf("duplicate package id: %s", p.ID))
		} else {
			pkgIDs[p.ID] = true
		}
	}

	// Dotfiles
	dfIDs := make(map[string]bool)
	for _, d := range cfg.Dotfiles {
		errs = append(errs, d.Validate()...)
		if dfIDs[d.ID] {
			errs = append(errs, fmt.Sprintf("duplicate dotfile id: %s", d.ID))
		} else {
			dfIDs[d.ID] = true
		}
	}

	// Tasks
	taskIDs := make(map[string]bool)
	for _, t := range cfg.Tasks {
		errs = append(errs, t.Validate()...)
		if taskIDs[t.ID] {
			errs = append(errs, fmt.Sprintf("duplicate task id: %s", t.ID))
		} else {
			taskIDs[t.ID] = true
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
}

// ConfigDir returns the explicit path if set, otherwise auto-detects
// via FindConfigPath. This keeps the "explicit or auto-detect" logic in
// one place so CLI flags just pass the value through.
func ConfigDir(explicit string) string {
	if explicit != "" {
		return explicit
	}
	return FindConfigPath()
}

// FindConfigPath locates the configs/ directory.
// It checks: next to the binary, then in the current working directory.
// Returns just "configs" as a last resort so the caller can still produce
// a useful error message about which path was tried.
func FindConfigPath() string {
	if execPath, err := os.Executable(); err == nil {
		cfgCandidate := filepath.Join(filepath.Dir(execPath), "configs")
		if _, err := os.Stat(cfgCandidate); err == nil {
			return cfgCandidate
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		cfgCandidate := filepath.Join(cwd, "configs")
		if _, err := os.Stat(cfgCandidate); err == nil {
			return cfgCandidate
		}
	}
	return "configs"
}
