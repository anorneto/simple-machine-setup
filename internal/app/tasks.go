package app

import "fmt"

// taskResult is the internal count struct for the task phase.
type taskResult struct {
	Succeeded int
	Skipped   int
	Failed    int
}

// runTasks executes every task in the config whose Stage matches `stage`
// ("pre" or "post"). Tasks are run in declaration order.
// In dry-run mode the command is printed and counted as skipped.
func (a *Apply) runTasks(stage string) (*taskResult, error) {
	res := &taskResult{}
	tasks := a.tasksForStage(stage)
	if len(tasks) == 0 {
		return res, nil
	}

	fmt.Println(Header.Render(fmt.Sprintf("==> Running %s-stage tasks", stage)))

	for _, t := range tasks {
		if a.Runner.DryRun {
			fmt.Printf("  %s %s (would run: %s)\n", Warning.Render("⚠"), t.Description, t.Command)
			res.Skipped++
			continue
		}

		fmt.Printf("  %s %s\n", Warning.Render("⚠"), t.Description)

		if !Confirm(fmt.Sprintf("Run: %s?", t.Description), a.AutoYes) {
			fmt.Printf("  %s %s (skipped)\n", Warning.Render("⚠"), t.Description)
			res.Skipped++
			continue
		}

		if err := a.Runner.RunInteractive(t.Command); err != nil {
			fmt.Printf("  %s %s (failed: %v)\n", Error.Render("✗"), t.Description, err)
			res.Failed++
			continue
		}
		fmt.Printf("  %s %s (done)\n", Success.Render("✓"), t.Description)
		res.Succeeded++
	}
	return res, nil
}

// tasksForStage filters the config's tasks to those with the given stage.
func (a *Apply) tasksForStage(stage string) []Task {
	out := make([]Task, 0, len(a.Cfg.Tasks))
	for _, t := range a.Cfg.Tasks {
		if t.Stage == stage {
			out = append(out, t)
		}
	}
	return out
}
