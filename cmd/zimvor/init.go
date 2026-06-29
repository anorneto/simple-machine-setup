package main

import (
	"bytes"
	_ "embed" // the _ before embed is required to use the //go:embed directive
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/anorneto/zimvor/internal/app"
)

//go:embed template.toml
var starterTemplate string

// starterVars holds the template variables substituted at init time.
type starterVars struct {
	OS          string
	Description string
	GitInstall  string
	FishInstall string
}

// initCmd scaffolds a starter TOML config for the current OS.
func initCmd() *cobra.Command {
	runInit := func(cmd *cobra.Command, args []string) error {
		// TODO(anor): maybe have an env var to set config dir?
		configsDir := app.ConfigDir(configDir)
		if err := os.MkdirAll(configsDir, 0o755); err != nil {
			return fmt.Errorf("failed to create configs dir: %w", err)
		}
		if err := os.MkdirAll(filepath.Join(configsDir, "dotfiles"), 0o755); err != nil {
			return fmt.Errorf("failed to create dotfiles dir: %w", err)
		}

		osName := app.DetectOS()
		configPath := filepath.Join(configsDir, app.GetOSConfigFile())

		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("config already exists: %s (refusing to overwrite)", configPath)
		}

		vars := starterVars{
			OS:          osName,
			Description: fmt.Sprintf("My %s setup", osName),
			GitInstall:  installCmdForOS(osName, "git"),
			FishInstall: installCmdForOS(osName, "fish"),
		}

		rendered, err := renderStarterTemplate(vars)
		if err != nil {
			return fmt.Errorf("failed to render starter config: %w", err)
		}

		if err := os.WriteFile(configPath, rendered, 0o644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		fmt.Printf("Created %s\n", configPath)
		fmt.Println("Edit it to declare your packages, dotfiles, and tasks, then run `zimvor status`.")
		return nil
	}

	return &cobra.Command{
		Use:   "init",
		Short: "Create a starter TOML config for the current OS",
		Long:  "Scaffolds configs/<os>.toml with a minimal template and creates the configs/dotfiles/ directory. Refuses to overwrite an existing file.",
		RunE:  runInit,
	}
}

// renderStarterTemplate executes the embedded template against vars.
func renderStarterTemplate(vars starterVars) ([]byte, error) {
	tmpl, err := template.New("starter").Parse(starterTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	// bytes.Buffer implements io.Writer so Execute can write into it,
	// then we read the result via buf.Bytes().
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	return buf.Bytes(), nil
}

// installCmdForOS returns the install command for a package on the given OS.
func installCmdForOS(osName, pkg string) string {
	if osName == "darwin" {
		return fmt.Sprintf("brew install %s", pkg)
	}
	return fmt.Sprintf("sudo apt install -y %s", pkg)
}
