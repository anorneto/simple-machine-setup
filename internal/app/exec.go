package app

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Runner executes shell commands. It supports a dry-run mode that records
// what would have been run without actually invoking a subprocess.
//
// We use it for two reasons:
//  1. To respect zimvor's "default is dry-run" behavior
//  2. To keep all subprocess execution in one place so it's easy to audit
type Runner struct {
	// DryRun, when true, makes Run/RunInteractive return success without
	// executing the command. The command is captured in CommandResult.
	DryRun bool
}

// CommandResult describes the outcome of a single command invocation.
// It's named CommandResult (not Result) so it doesn't collide with the
// Apply-phase Result struct also defined in this package.
type CommandResult struct {
	Command  string
	Output   string
	ExitCode int
	Err      error
	DryRun   bool
}

// Run executes `command` via `sh -c` and captures stdout+stderr.
// In dry-run mode it returns a successful CommandResult without executing.
func (r *Runner) Run(command string) *CommandResult {
	if r.DryRun {
		return &CommandResult{
			Command:  command,
			ExitCode: 0,
			DryRun:   true,
		}
	}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	exitCode := 0
	if err != nil {
		// exec.ExitError is the error you get when the command itself
		// failed (non-zero exit). Other errors (e.g. "command not found")
		// come back as a plain error and have no exit code, so we map them
		// to 1 ourselves.
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return &CommandResult{
		Command:  command,
		Output:   strings.TrimSpace(output),
		ExitCode: exitCode,
		Err:      err,
	}
}

// RunInteractive executes `command` inheriting the parent's stdio so the user
// sees live output and can respond to prompts (e.g., sudo password).
// In dry-run mode it prints what would run and returns nil.
func (r *Runner) RunInteractive(command string) error {
	if r.DryRun {
		fmt.Printf("[dry-run] would execute: %s\n", command)
		return nil
	}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// Check returns true if `command` exits with status 0.
// In dry-run mode it returns true (assume the tool is present) so the
// dry-run output mirrors what an "everything installed" state looks like.
func (r *Runner) Check(command string) bool {
	if r.DryRun {
		return true
	}
	return r.Run(command).ExitCode == 0
}
