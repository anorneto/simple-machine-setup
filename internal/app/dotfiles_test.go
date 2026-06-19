package app

import "testing"

// expandPath lives in dotfiles.go. It is used by both the dotfile deploy
// phase and the status phase, so it earns its own test even though the
// file it's in is about dotfiles.

func TestExpandPathHome(t *testing.T) {
	got := expandPath("~/.config/fish/config.fish")
	if got == "~/.config/fish/config.fish" {
		t.Error("expected ~ to be expanded, got unchanged")
	}
	if got[len(got)-len(".config/fish/config.fish"):] != ".config/fish/config.fish" {
		t.Errorf("unexpected suffix: %q", got)
	}
}

func TestExpandPathAbsolute(t *testing.T) {
	got := expandPath("/etc/hosts")
	if got != "/etc/hosts" {
		t.Errorf("absolute path should be unchanged, got %q", got)
	}
}

func TestExpandPathJustHome(t *testing.T) {
	got := expandPath("~")
	if got == "~" {
		t.Error("expected ~ to be expanded to home dir, got unchanged")
	}
}
