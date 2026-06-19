package app

import "fmt"

// packageResult is the internal count struct for the package phase.
// It's unexported because callers only ever see the aggregated Result.
type packageResult struct {
	Installed int
	Skipped   int
	Failed    int
}

// runPackages walks every package in the config, checks for installation,
// and either skips, reports a dry-run, or installs (after confirmation).
func (a *Apply) runPackages() (*packageResult, error) {
	fmt.Println(Header.Render("==> Installing packages"))

	res := &packageResult{}
	for _, pkg := range a.Cfg.Packages {
		// `command -v` is the portable way to check whether a binary is on
		// PATH. `which` works on most systems but not on minimal Alpine /
		// busybox environments, so we use the built-in.
		checkCmd := "command -v " + pkg.BinaryOrID()
		if a.Runner.Check(checkCmd) {
			fmt.Printf("  %s %s\n", Success.Render("✓"), pkg.ID)
			res.Installed++
			continue
		}

		if a.Runner.DryRun {
			fmt.Printf("  %s %s (would install: %s)\n", Warning.Render("⚠"), pkg.ID, pkg.Install)
			res.Skipped++
			continue
		}

		fmt.Printf("  %s %s (installing...)\n", Warning.Render("⚠"), pkg.ID)

		if !Confirm(fmt.Sprintf("Install %s?", pkg.ID), a.AutoYes) {
			fmt.Printf("  %s %s (skipped)\n", Warning.Render("⚠"), pkg.ID)
			res.Skipped++
			continue
		}

		// Run each install command in order. Stop on the first failure so
		// the user sees a clear "command X failed" message instead of
		// cascading errors from a half-set-up system.
		failed := false
		for _, cmd := range pkg.Install {
			if err := a.Runner.RunInteractive(cmd); err != nil {
				fmt.Printf("  %s %s (failed: %v)\n", Error.Render("✗"), pkg.ID, err)
				failed = true
				break
			}
		}
		if failed {
			res.Failed++
			continue
		}

		// Verify the install actually left the binary on PATH.
		// This catches installs that "succeed" but don't put the binary
		// where we expected (e.g. PATH not updated yet).
		if a.Runner.Check(checkCmd) {
			fmt.Printf("  %s %s (installed)\n", Success.Render("✓"), pkg.ID)
			res.Installed++
		} else {
			fmt.Printf("  %s %s (command failed to verify)\n", Error.Render("✗"), pkg.ID)
			res.Failed++
		}
	}

	return res, nil
}

// statusPackages prints whether each declared package is installed.
// Read-only; used by `status`.
func statusPackages(pkgs []Package, runner *Runner) {
	fmt.Println(Header.Render("==> Package status"))
	for _, pkg := range pkgs {
		if runner.Check("which  " + pkg.BinaryOrID()) {
			fmt.Printf("  %s %s\n", Success.Render("✓"), pkg.ID)
		} else {
			fmt.Printf("  %s %s\n", Error.Render("✗"), pkg.ID)
		}
	}
	fmt.Println()
}
