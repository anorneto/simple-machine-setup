package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anorneto/zimvor/internal/app"
	"github.com/spf13/cobra"
)

// initCmd scaffolds a starter TOML config in configs/ for the current OS.
// It refuses to overwrite an existing file.
func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a starter TOML config for the current OS",
		Long: "Scaffolds configs/<os>.toml with a minimal template and creates " +
			"the configs/dotfiles/ directory. Refuses to overwrite an existing file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configsDir := getConfigDir()
			if err := os.MkdirAll(configsDir, 0755); err != nil {
				return fmt.Errorf("failed to create configs dir: %w", err)
			}
			if err := os.MkdirAll(filepath.Join(configsDir, "dotfiles"), 0755); err != nil {
				return fmt.Errorf("failed to create dotfiles dir: %w", err)
			}

			osName := app.Detect()
			configPath := filepath.Join(configsDir, app.ConfigFileName())

			if _, err := os.Stat(configPath); err == nil {
				return fmt.Errorf("config already exists: %s (refusing to overwrite)", configPath)
			}

			tmpl := starterTemplate(osName)
			if err := os.WriteFile(configPath, []byte(tmpl), 0644); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			fmt.Printf("Created %s\n", configPath)
			fmt.Println("Edit it to declare your packages, dotfiles, and tasks, then run `zimvor status`.")
			return nil
		},
	}
}

// starterTemplate returns a minimal TOML config with install commands
// appropriate for the OS. Each OS gets the right package manager command
// in the `install` list so the user can run `zimvor install --apply`
// immediately to see something happen.
func starterTemplate(osName string) string {
	fishInstall := "sudo apt install -y fish"
	gitInstall := "sudo apt install -y git"
	if osName == "darwin" {
		fishInstall = "brew install fish"
		gitInstall = "brew install git"
	}

	return fmt.Sprintf(`[meta]
os = %q
description = "My %s setup"

[[packages]]
id = "git"
install = [%q]

[[packages]]
id = "fish"
install = [%q]

[[dotfiles]]
id = "gitconfig"
source = "dotfiles/git/config"
target = "~/.gitconfig"

[[tasks]]
id = "example-task"
description = "Example post-install task"
stage = "post"
command = "echo done"
`, osName, osName, gitInstall, fishInstall)
}
