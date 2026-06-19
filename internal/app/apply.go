package app

import (
	"fmt"
	"path/filepath"
)

// Result summarizes the outcome of one full install run (or a sub-phase).
// Counts are useful for the summary printed at the end of `install --apply`.
type Result struct {
	// Package counts
	PackagesInstalled int
	PackagesSkipped   int
	PackagesFailed    int

	// Dotfile counts
	DotfilesCreated int
	DotfilesUpdated int
	DotfilesSkipped int
	DotfilesFailed  int

	// Task counts
	TasksSucceeded int
	TasksSkipped   int
	TasksFailed    int
}

// Apply is the top-level orchestrator. It owns the runner, the loaded config,
// and the flags. Each sub-phase (runPackages, runDotfiles, runTasks) reads
// from the same Apply.
type Apply struct {
	Cfg        *Config
	Runner     *Runner
	ApplyMode  bool // false = dry-run, true = actually execute
	AutoYes    bool // skip all confirmations
	ConfigsDir string
}

// NewApply builds an Apply from a validated config.
// apply=false means "dry run": commands are described, not executed.
func NewApply(cfg *Config, apply, autoYes bool, configsDir string) *Apply {
	return &Apply{
		Cfg:        cfg,
		Runner:     &Runner{DryRun: !apply},
		ApplyMode:  apply,
		AutoYes:    autoYes,
		ConfigsDir: configsDir,
	}
}

// Run executes the full install flow:
//  1. pre-stage tasks
//  2. package installations
//  3. dotfile sync
//  4. post-stage tasks
//
// The order is intentional: tasks may depend on tooling being in place
// (post), and dotfile deploys may also depend on packages (e.g. ~/.config/fish
// exists after fish is installed).
func (a *Apply) Run() (*Result, error) {
	if err := a.absConfigsDir(); err != nil {
		return nil, err
	}

	res := &Result{}

	// Pre-stage tasks
	if r, err := a.runTasks("pre"); err != nil {
		return res, err
	} else {
		res.TasksSucceeded += r.Succeeded
		res.TasksSkipped += r.Skipped
		res.TasksFailed += r.Failed
	}

	// Packages
	if r, err := a.runPackages(); err != nil {
		return res, err
	} else {
		res.PackagesInstalled = r.Installed
		res.PackagesSkipped = r.Skipped
		res.PackagesFailed = r.Failed
	}

	// Dotfiles
	if r, err := a.runDotfiles(); err != nil {
		return res, err
	} else {
		res.DotfilesCreated = r.Created
		res.DotfilesUpdated = r.Updated
		res.DotfilesSkipped = r.Skipped
		res.DotfilesFailed = r.Failed
	}

	// Post-stage tasks
	if r, err := a.runTasks("post"); err != nil {
		return res, err
	} else {
		res.TasksSucceeded += r.Succeeded
		res.TasksSkipped += r.Skipped
		res.TasksFailed += r.Failed
	}

	a.printSummary(res)
	return res, nil
}

// Packages runs only the package install phase. Used by `install packages`.
func (a *Apply) Packages() (*Result, error) {
	if err := a.absConfigsDir(); err != nil {
		return nil, err
	}

	r, err := a.runPackages()
	if err != nil {
		return nil, err
	}
	return &Result{
		PackagesInstalled: r.Installed,
		PackagesSkipped:   r.Skipped,
		PackagesFailed:    r.Failed,
	}, nil
}

// Dotfiles runs only the dotfile sync phase. Used by `install dotfiles`.
func (a *Apply) Dotfiles() (*Result, error) {
	if err := a.absConfigsDir(); err != nil {
		return nil, err
	}

	r, err := a.runDotfiles()
	if err != nil {
		return nil, err
	}
	return &Result{
		DotfilesCreated: r.Created,
		DotfilesUpdated: r.Updated,
		DotfilesSkipped: r.Skipped,
		DotfilesFailed:  r.Failed,
	}, nil
}

// Tasks runs only the tasks of a given stage. Used by `install tasks`.
// We run both pre and post because there's no per-stage flag yet; v1 keeps
// the CLI simple.
func (a *Apply) Tasks() (*Result, error) {
	if err := a.absConfigsDir(); err != nil {
		return nil, err
	}

	res := &Result{}
	for _, stage := range []string{"pre", "post"} {
		r, err := a.runTasks(stage)
		if err != nil {
			return res, err
		}
		res.TasksSucceeded += r.Succeeded
		res.TasksSkipped += r.Skipped
		res.TasksFailed += r.Failed
	}
	return res, nil
}

// Status reports on the state of the machine vs the config.
// Read-only: no changes, no prompts.
func (a *Apply) Status() error {
	if err := a.absConfigsDir(); err != nil {
		return err
	}

	statusPackages(a.Cfg.Packages, a.Runner)
	statusDotfiles(a.Cfg.Dotfiles, a.ConfigsDir)
	return nil
}

// Diff shows what `install` would change without actually changing anything.
// It always runs in dry-run mode and never prompts.
func (a *Apply) Diff() error {
	if err := a.absConfigsDir(); err != nil {
		return err
	}

	// Force dry-run regardless of how the caller configured Apply.
	// We snapshot and restore so a caller that re-uses the Apply for a
	// real run later doesn't accidentally inherit the dry-run override.
	original := a.Runner.DryRun
	a.Runner.DryRun = true
	defer func() { a.Runner.DryRun = original }()

	_, _ = a.runDotfiles()
	return nil
}

// absConfigsDir resolves ConfigsDir to an absolute path. We do this once
// per public method so that all the relative joins in runDotfiles and
// statusDotfiles work the same way regardless of where the binary was
// invoked from.
func (a *Apply) absConfigsDir() error {
	abs, err := filepath.Abs(a.ConfigsDir)
	if err != nil {
		return fmt.Errorf("failed to resolve configs directory: %w", err)
	}
	a.ConfigsDir = abs
	return nil
}

// printSummary renders a one-line summary of what happened.
func (a *Apply) printSummary(r *Result) {
	if !a.ApplyMode {
		// No summary needed in dry-run; phase headers are enough.
		return
	}
	fmt.Println()
	fmt.Println(Header.Render("==> Summary"))
	if r.PackagesInstalled+r.PackagesSkipped+r.PackagesFailed > 0 {
		fmt.Printf("  Packages: %s installed, %s skipped, %s failed\n",
			Success.Render(fmt.Sprintf("%d", r.PackagesInstalled)),
			Warning.Render(fmt.Sprintf("%d", r.PackagesSkipped)),
			Error.Render(fmt.Sprintf("%d", r.PackagesFailed)),
		)
	}
	if r.DotfilesCreated+r.DotfilesUpdated+r.DotfilesSkipped+r.DotfilesFailed > 0 {
		fmt.Printf("  Dotfiles: %s created, %s updated, %s skipped, %s failed\n",
			Success.Render(fmt.Sprintf("%d", r.DotfilesCreated)),
			Success.Render(fmt.Sprintf("%d", r.DotfilesUpdated)),
			Warning.Render(fmt.Sprintf("%d", r.DotfilesSkipped)),
			Error.Render(fmt.Sprintf("%d", r.DotfilesFailed)),
		)
	}
	if r.TasksSucceeded+r.TasksSkipped+r.TasksFailed > 0 {
		fmt.Printf("  Tasks:    %s succeeded, %s skipped, %s failed\n",
			Success.Render(fmt.Sprintf("%d", r.TasksSucceeded)),
			Warning.Render(fmt.Sprintf("%d", r.TasksSkipped)),
			Error.Render(fmt.Sprintf("%d", r.TasksFailed)),
		)
	}
}
