// Package app contains all of zimvor's non-CLI logic.
//
// It is intentionally one package. The previous layout split this across
// several tiny packages (config, apply, exec, diff, output, prompt, platform)
// but every one of them was consumed by every other one, which added
// import noise without any isolation benefit. If a future piece genuinely
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

// Package declares a system-level package to install.
// "ID" is a stable identifier used in logs and errors.
// "Install" is a list of shell commands to run, in order, to install the
// package. The user is responsible for writing commands appropriate to
// their OS — we just run them. This keeps the schema flat and explicit.
// "Binary" is optional; when set, it's what we look for on PATH to
// decide whether the package is already installed. When unset, we use
// the ID instead.
type Package struct {
	ID      string   `toml:"id"`
	Binary  string   `toml:"binary"`
	Install []string `toml:"install"`
}

// BinaryOrID returns the name we should use when checking PATH for the
// installed tool. It returns Binary if set, otherwise ID.
func (p Package) BinaryOrID() string {
	if p.Binary != "" {
		return p.Binary
	}
	return p.ID
}

// Dotfile declares a single file to sync from the repo to the machine.
// "Source" is relative to the configs/ directory.
// "Target" is an absolute path or ~/... path on the machine.
type Dotfile struct {
	ID     string `toml:"id"`
	Source string `toml:"source"`
	Target string `toml:"target"`
}

// Task declares a shell command to run before or after the main install flow.
// "Stage" must be "pre" or "post".
type Task struct {
	ID          string `toml:"id"`
	Description string `toml:"description"`
	Stage       string `toml:"stage"`
	Command     string `toml:"command"`
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
func (c *Config) Validate() error {
	var errs []string

	// Meta
	if strings.TrimSpace(c.Meta.OS) == "" {
		errs = append(errs, "meta.os is required")
	}
	if strings.TrimSpace(c.Meta.Description) == "" {
		errs = append(errs, "meta.description is required")
	}

	// Packages
	pkgIDs := make(map[string]bool)
	for i, p := range c.Packages {
		if strings.TrimSpace(p.ID) == "" {
			errs = append(errs, fmt.Sprintf("packages[%d].id is required", i))
		} else if pkgIDs[p.ID] {
			errs = append(errs, fmt.Sprintf("duplicate package id: %s", p.ID))
		} else {
			pkgIDs[p.ID] = true
		}
		if len(p.Install) == 0 {
			errs = append(errs, fmt.Sprintf("packages[%d] (id=%q).install must not be empty", i, p.ID))
		}
	}

	// Dotfiles
	dfIDs := make(map[string]bool)
	for i, d := range c.Dotfiles {
		if strings.TrimSpace(d.ID) == "" {
			errs = append(errs, fmt.Sprintf("dotfiles[%d].id is required", i))
		} else if dfIDs[d.ID] {
			errs = append(errs, fmt.Sprintf("duplicate dotfile id: %s", d.ID))
		} else {
			dfIDs[d.ID] = true
		}
		if strings.TrimSpace(d.Source) == "" {
			errs = append(errs, fmt.Sprintf("dotfiles[%d] (id=%q).source is required", i, d.ID))
		}
		if strings.TrimSpace(d.Target) == "" {
			errs = append(errs, fmt.Sprintf("dotfiles[%d] (id=%q).target is required", i, d.ID))
		} else if !strings.HasPrefix(d.Target, "/") && !strings.HasPrefix(d.Target, "~/") {
			errs = append(errs, fmt.Sprintf("dotfiles[%d] (id=%q).target must be an absolute path", i, d.ID))
		}
	}

	// Tasks
	taskIDs := make(map[string]bool)
	for i, t := range c.Tasks {
		if strings.TrimSpace(t.ID) == "" {
			errs = append(errs, fmt.Sprintf("tasks[%d].id is required", i))
		} else if taskIDs[t.ID] {
			errs = append(errs, fmt.Sprintf("duplicate task id: %s", t.ID))
		} else {
			taskIDs[t.ID] = true
		}
		if strings.TrimSpace(t.Command) == "" {
			errs = append(errs, fmt.Sprintf("tasks[%d] (id=%q).command is required", i, t.ID))
		}
		if strings.TrimSpace(t.Stage) == "" {
			errs = append(errs, fmt.Sprintf("tasks[%d] (id=%q).stage is required", i, t.ID))
		} else if t.Stage != "pre" && t.Stage != "post" {
			errs = append(errs, fmt.Sprintf("tasks[%d] (id=%q).stage must be 'pre' or 'post'", i, t.ID))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
}

// FindConfigPath locates the configs/ directory.
// It checks: next to the binary, then in the current working directory.
// Returns just "configs" as a last resort so the caller can still produce
// a useful error message about which path was tried.
func FindConfigPath() string {
	if execPath, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(execPath), "configs")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(cwd, "configs")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "configs"
}
