package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/anorneto/zimvor/internal/app"
)

// Version metadata, set via -ldflags at build time.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// Global flags. `apply=false` is the default — `install` is dry-run unless
// the user explicitly passes --apply.
// configDir is the explicit --config value, or empty to auto-detect.
var (
	applyFlag bool
	yesFlag   bool
	configDir string
)

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		// Cobra already printed the error; we just set the exit code.
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zimvor",
		Short: "Cross-platform config sync tool",
		Long: "Minimal cross-platform config sync CLI for dotfiles, tools, " +
			"and machine-specific settings. Loads a per-OS TOML manifest " +
			"and reconciles the machine against it.",
		SilenceUsage: true,
	}

	cmd.PersistentFlags().BoolVarP(&applyFlag, "apply", "a", false,
		"Actually execute changes (default: dry-run)")
	cmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false,
		"Skip all confirmation prompts")
	cmd.PersistentFlags().StringVarP(&configDir, "config", "c", "",
		"Config directory path (default: auto-detect)")

	cmd.AddCommand(newInstallCmd())
	cmd.AddCommand(statusCmd())
	cmd.AddCommand(diffCmd())
	cmd.AddCommand(versionCmd())
	cmd.AddCommand(initCmd())

	return cmd
}

// newInstallCmd is the main action. Without sub-commands it runs the full
// flow; with sub-commands it runs only the named phase.
func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Run setup (default: dry-run, pass --apply to execute)",
		Long: "Runs the full setup flow: pre-tasks, packages, dotfiles, " +
			"post-tasks. Without --apply it shows what would happen. " +
			"Use a sub-command to run only one phase.",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := newApply()
			if err != nil {
				return err
			}
			_, err = a.Run()
			return err
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "packages",
		Short: "Install system packages only",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := newApply()
			if err != nil {
				return err
			}
			_, err = a.Packages()
			return err
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "dotfiles",
		Short: "Sync dotfiles only",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := newApply()
			if err != nil {
				return err
			}
			_, err = a.Dotfiles()
			return err
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "tasks",
		Short: "Run tasks only (pre and post)",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := newApply()
			if err != nil {
				return err
			}
			_, err = a.Tasks()
			return err
		},
	})

	return cmd
}

// statusCmd shows the state of the machine vs the config. Read-only.
func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show what is installed and what is in sync",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := newApply()
			if err != nil {
				return err
			}
			return a.Status()
		},
	}
}

// diffCmd shows what `install` would change without changing anything.
func diffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Show dotfile diffs that install would apply",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := newApply()
			if err != nil {
				return err
			}
			return a.Diff()
		},
	}
}

// versionCmd prints build metadata.
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("zimvor %s (commit: %s, built: %s)\n", version, commit, date)
		},
	}
}

// newApply loads the config for the current OS and constructs an app.Apply.
// It bails out with a friendly error if the file is missing or invalid.
func newApply() (*app.Apply, error) {
	configsDir := app.ConfigDir(configDir)
	configFile := filepath.Join(configsDir, app.GetOSConfigFile())

	cfg, err := app.Load(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no config found at %s — run `zimvor init` to create one", configFile)
		}
		return nil, fmt.Errorf("failed to load config from %s: %w", configFile, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", configFile, err)
	}

	return app.NewApply(cfg, applyFlag, yesFlag, configsDir), nil
}
