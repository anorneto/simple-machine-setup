package app

import "testing"

func TestRunSuccess(t *testing.T) {
	r := &Runner{}
	res := r.Run("true")
	if res.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d", res.ExitCode)
	}
	if res.Err != nil {
		t.Errorf("expected no error, got %v", res.Err)
	}
}

func TestRunFailure(t *testing.T) {
	r := &Runner{}
	res := r.Run("false")
	if res.ExitCode == 0 {
		t.Error("expected non-zero exit, got 0")
	}
}

func TestRunCapturesOutput(t *testing.T) {
	r := &Runner{}
	res := r.Run("echo hello")
	if res.Output != "hello" {
		t.Errorf("expected output 'hello', got %q", res.Output)
	}
}

func TestDryRunDoesNotExecute(t *testing.T) {
	r := &Runner{DryRun: true}
	res := r.Run("this-command-does-not-exist")
	if res.ExitCode != 0 {
		t.Errorf("dry-run should report exit 0, got %d", res.ExitCode)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true on CommandResult")
	}
}

func TestCheckReturnsTrueForZeroExit(t *testing.T) {
	r := &Runner{}
	if !r.Check("true") {
		t.Error("expected Check('true') to return true")
	}
}

func TestCheckReturnsFalseForNonZeroExit(t *testing.T) {
	r := &Runner{}
	if r.Check("false") {
		t.Error("expected Check('false') to return false")
	}
}

func TestCheckInDryRunAlwaysTrue(t *testing.T) {
	r := &Runner{DryRun: true}
	if !r.Check("this-command-does-not-exist") {
		t.Error("Check() in dry-run mode should always return true")
	}
}

func TestRunInteractiveDryRun(t *testing.T) {
	r := &Runner{DryRun: true}
	if err := r.RunInteractive("echo hi"); err != nil {
		t.Errorf("expected no error in dry-run, got %v", err)
	}
}
