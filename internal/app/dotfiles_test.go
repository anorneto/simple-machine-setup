package app

import (
	"os"
	"path/filepath"
	"testing"
)

// expandPath lives in dotfiles.go. It is used by both the dotfile deploy
// phase and the status phase, so it earns its own test even though the
// file it's in is about dotfiles.

func TestExpandPathHome(t *testing.T) {
	var got string = expandHomePath("~/.config/fish/config.fish")
	if got == "~/.config/fish/config.fish" {
		t.Error("expected ~ to be expanded, got unchanged")
	}
	if got[len(got)-len(".config/fish/config.fish"):] != ".config/fish/config.fish" {
		t.Errorf("unexpected suffix: %q", got)
	}
}

func TestExpandPathAbsolute(t *testing.T) {
	got := expandHomePath("/etc/hosts")
	if got != "/etc/hosts" {
		t.Errorf("absolute path should be unchanged, got %q", got)
	}
}

func TestExpandPathJustHome(t *testing.T) {
	got := expandHomePath("~")
	if got == "~" {
		t.Error("expected ~ to be expanded to home dir, got unchanged")
	}
}

func TestRunDotfilesDirectory(t *testing.T) {
	// Create temp directories for source (configs) and target.
	srcDir := t.TempDir()
	targetDir := t.TempDir()

	// Write source files under configs/dotfiles/fish
	fishSrcDir := filepath.Join(srcDir, "dotfiles", "fish")
	if err := os.MkdirAll(fishSrcDir, 0o755); err != nil {
		t.Fatalf("failed to create src subdirs: %v", err)
	}

	file1 := filepath.Join(fishSrcDir, "config.fish")
	if err := os.WriteFile(file1, []byte("echo hello"), 0o644); err != nil {
		t.Fatalf("failed to write source file1: %v", err)
	}

	// Also test subdirectories
	funcSrcDir := filepath.Join(fishSrcDir, "functions")
	if err := os.MkdirAll(funcSrcDir, 0o755); err != nil {
		t.Fatalf("failed to create functions subdir: %v", err)
	}
	file2 := filepath.Join(funcSrcDir, "cls.fish")
	if err := os.WriteFile(file2, []byte("clear"), 0o644); err != nil {
		t.Fatalf("failed to write source file2: %v", err)
	}

	cfg := &Config{
		Dotfiles: []Dotfile{
			{
				ID:     "fish-configs",
				Source: "dotfiles/fish",
				Target: targetDir,
			},
		},
	}

	// Create Apply instance
	a := NewApply(cfg, true, true, srcDir)

	res, err := a.runDotfiles()
	if err != nil {
		t.Fatalf("runDotfiles returned error: %v", err)
	}

	if res.Created != 2 {
		t.Errorf("expected 2 created files, got %d", res.Created)
	}

	// Verify target files exist and match contents
	targetFile1 := filepath.Join(targetDir, "config.fish")
	targetFile2 := filepath.Join(targetDir, "functions", "cls.fish")

	c, err := os.ReadFile(targetFile1)
	if err != nil || string(c) != "echo hello" {
		t.Errorf("expected targetFile1 content to be 'echo hello', got error %v or content %q", err, string(c))
	}
	c, err = os.ReadFile(targetFile2)
	if err != nil || string(c) != "clear" {
		t.Errorf("expected targetFile2 content to be 'clear', got error %v or content %q", err, string(c))
	}

	// Run again, should skip
	res, err = a.runDotfiles()
	if err != nil {
		t.Fatalf("runDotfiles 2nd run returned error: %v", err)
	}
	if res.Skipped != 2 {
		t.Errorf("expected 2 skipped files on 2nd run, got %d", res.Skipped)
	}
}
