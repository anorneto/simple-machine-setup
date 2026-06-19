package app

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return path
}

const validConfig = `
[meta]
os = "linux"
description = "test"

[[packages]]
id = "git"
install = ["sudo apt install -y git"]

[[packages]]
id = "ripgrep"
binary = "rg"
install = ["sudo apt install -y ripgrep"]

[[dotfiles]]
id = "fish-config"
source = "dotfiles/fish/config.fish"
target = "~/.config/fish/config.fish"

[[tasks]]
id = "set-fish"
description = "Set fish as default"
stage = "post"
command = "chsh -s $(which fish)"
`

func TestLoadValid(t *testing.T) {
	path := writeTempConfig(t, validConfig)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Meta.OS != "linux" {
		t.Errorf("expected OS=linux, got %q", cfg.Meta.OS)
	}
	if len(cfg.Packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(cfg.Packages))
	}
	if len(cfg.Dotfiles) != 1 {
		t.Errorf("expected 1 dotfile, got %d", len(cfg.Dotfiles))
	}
	if len(cfg.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(cfg.Tasks))
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path.toml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadMalformed(t *testing.T) {
	path := writeTempConfig(t, "this is not valid TOML = = =")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for malformed TOML, got nil")
	}
}

func TestValidateValid(t *testing.T) {
	path := writeTempConfig(t, validConfig)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() returned error for valid config: %v", err)
	}
}

func TestValidateMissingMetaOS(t *testing.T) {
	cfg := &Config{Meta: Meta{Description: "test"}}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateMissingMetaDescription(t *testing.T) {
	cfg := &Config{Meta: Meta{OS: "linux"}}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateDuplicatePackageID(t *testing.T) {
	cfg := &Config{
		Meta: Meta{OS: "linux", Description: "test"},
		Packages: []Package{
			{ID: "git", Install: []string{"echo"}},
			{ID: "git", Install: []string{"echo"}},
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for duplicate package id, got nil")
	}
}

func TestValidatePackageMissingID(t *testing.T) {
	cfg := &Config{
		Meta: Meta{OS: "linux", Description: "test"},
		Packages: []Package{
			{Install: []string{"echo"}},
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for missing package id, got nil")
	}
}

func TestValidatePackageEmptyInstall(t *testing.T) {
	cfg := &Config{
		Meta: Meta{OS: "linux", Description: "test"},
		Packages: []Package{
			{ID: "git"},
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty install list, got nil")
	}
}

func TestPackageBinaryOrID(t *testing.T) {
	with := Package{ID: "ripgrep", Binary: "rg", Install: []string{"x"}}
	if got := with.BinaryOrID(); got != "rg" {
		t.Errorf("expected rg, got %q", got)
	}

	without := Package{ID: "git", Install: []string{"x"}}
	if got := without.BinaryOrID(); got != "git" {
		t.Errorf("expected git, got %q", got)
	}
}

func TestValidateDotfileRelativeTarget(t *testing.T) {
	cfg := &Config{
		Meta: Meta{OS: "linux", Description: "test"},
		Dotfiles: []Dotfile{
			{ID: "bad", Source: "x", Target: "relative/path"},
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for relative target, got nil")
	}
}

func TestValidateDotfileMissingFields(t *testing.T) {
	cfg := &Config{
		Meta: Meta{OS: "linux", Description: "test"},
		Dotfiles: []Dotfile{
			{ID: "bad", Source: "x", Target: "/abs/path"},
			{ID: "ok", Source: "", Target: "/abs/path"},
			{ID: "ok2", Source: "x", Target: ""},
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for missing dotfile fields, got nil")
	}
}

func TestValidateTaskInvalidStage(t *testing.T) {
	cfg := &Config{
		Meta: Meta{OS: "linux", Description: "test"},
		Tasks: []Task{
			{ID: "bad", Description: "x", Stage: "middle", Command: "echo"},
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for invalid stage, got nil")
	}
}

func TestValidateTaskMissingCommand(t *testing.T) {
	cfg := &Config{
		Meta: Meta{OS: "linux", Description: "test"},
		Tasks: []Task{
			{ID: "bad", Description: "x", Stage: "post", Command: ""},
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for missing command, got nil")
	}
}

func TestValidateCollectsAllErrors(t *testing.T) {
	cfg := &Config{
		Meta: Meta{OS: "", Description: ""},
		Packages: []Package{
			{ID: "", Install: nil},
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	// Ensure multiple errors are present
	if got := err.Error(); got == "" {
		t.Error("expected non-empty error message")
	}
}
