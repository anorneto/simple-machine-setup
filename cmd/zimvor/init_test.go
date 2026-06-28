package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/anorneto/zimvor/internal/app"
)

func TestInstallCmdForOS(t *testing.T) {
	tests := []struct {
		os   string
		pkg  string
		want string
	}{
		{"darwin", "git", "brew install git"},
		{"darwin", "fish", "brew install fish"},
		{"linux", "git", "sudo apt install -y git"},
		{"linux", "fish", "sudo apt install -y fish"},
		{"freebsd", "git", "sudo apt install -y git"},
	}
	for _, tt := range tests {
		t.Run(tt.os+"/"+tt.pkg, func(t *testing.T) {
			got := installCmdForOS(tt.os, tt.pkg)
			if got != tt.want {
				t.Errorf("installCmd(%q, %q) = %q, want %q", tt.os, tt.pkg, got, tt.want)
			}
		})
	}
}

func TestRenderStarterTemplate_Linux(t *testing.T) {
	vars := starterVars{
		OS:          "linux",
		Description: "My linux setup",
		GitInstall:  "sudo apt install -y git",
		FishInstall: "sudo apt install -y fish",
	}
	got, err := renderStarterTemplate(vars)
	if err != nil {
		t.Fatalf("renderStarterTemplate: %v", err)
	}
	for _, want := range []string{
		`os = "linux"`,
		`description = "My linux setup"`,
		`install = ["sudo apt install -y git"]`,
		`install = ["sudo apt install -y fish"]`,
		`target = "~/.gitconfig"`,
		`stage = "post"`,
	} {
		if !bytes.Contains(got, []byte(want)) {
			t.Errorf("rendered template missing %q\n--- got ---\n%s", want, got)
		}
	}
	for _, bad := range []string{"{{", "}}"} {
		if bytes.Contains(got, []byte(bad)) {
			t.Errorf("rendered template still contains placeholder %q", bad)
		}
	}
}

func TestRenderStarterTemplate_Darwin(t *testing.T) {
	vars := starterVars{
		OS:          "darwin",
		Description: "My darwin setup",
		GitInstall:  "brew install git",
		FishInstall: "brew install fish",
	}
	got, err := renderStarterTemplate(vars)
	if err != nil {
		t.Fatalf("renderStarterTemplate: %v", err)
	}
	for _, want := range []string{
		`os = "darwin"`,
		`install = ["brew install git"]`,
		`install = ["brew install fish"]`,
	} {
		if !bytes.Contains(got, []byte(want)) {
			t.Errorf("rendered template missing %q\n--- got ---\n%s", want, got)
		}
	}
}

// TestRenderedTemplateIsValidTOML parses the rendered output through the
// real app loader and validates it. This catches drift between the
// template and the schema in internal/app.
func TestRenderedTemplateIsValidTOML(t *testing.T) {
	vars := starterVars{
		OS:          "linux",
		Description: "My linux setup",
		GitInstall:  "sudo apt install -y git",
		FishInstall: "sudo apt install -y fish",
	}
	rendered, err := renderStarterTemplate(vars)
	if err != nil {
		t.Fatalf("renderStarterTemplate: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "linux.toml")
	if err := os.WriteFile(path, rendered, 0644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	cfg, err := app.Load(path)
	if err != nil {
		t.Fatalf("app.Load: %v\n--- rendered ---\n%s", err, rendered)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("cfg.Validate: %v", err)
	}
}
