package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// dotfileResult is the internal count struct for the dotfile phase.
type dotfileResult struct {
	Created int
	Updated int
	Skipped int
	Failed  int
}

// walkDotfile resolves source/target for a dotfile entry and calls fn for
// every file found — whether the entry is a single file or a directory.
func walkDotfile(df Dotfile, configsDir string, fn func(id, sourcePath, targetPath string) error) error {
	sourcePath := filepath.Join(configsDir, df.Source)
	targetPath := expandHomePath(df.Target)

	fi, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return filepath.WalkDir(sourcePath, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(sourcePath, path)
			if err != nil {
				return err
			}
			return fn(df.ID+"/"+rel, path, filepath.Join(targetPath, rel))
		})
	}

	return fn(df.ID, sourcePath, targetPath)
}

// runDotfiles walks every dotfile declaration and reconciles the source
// file/directory in the repo with the target on disk. The flow is:
//  1. resolve paths (source relative to ConfigsDir, target expands ~)
//  2. if source is a directory, recursively walk and sync all files within it
//  3. if target is missing → create
//  4. if contents match → skip
//  5. if contents differ → show diff, prompt, backup, write
func (a *Apply) runDotfiles() (*dotfileResult, error) {
	fmt.Println(Header.Render("==> Deploying dotfiles"))

	res := &dotfileResult{}
	for _, df := range a.Cfg.Dotfiles {
		err := walkDotfile(df, a.ConfigsDir, func(id, sourcePath, targetPath string) error {
			a.deployFile(id, sourcePath, targetPath, res)
			return nil // errors are tracked in res, not propagated to stop the walk
		})
		if err != nil {
			fmt.Printf("  %s %s (%v)\n", Error.Render("✗"), df.ID, err)
			res.Failed++
		}
	}

	return res, nil
}

// deployFile deploys a single file from sourcePath to targetPath.
func (a *Apply) deployFile(id, sourcePath, targetPath string, res *dotfileResult) {
	newContent, err := os.ReadFile(sourcePath)
	if err != nil {
		fmt.Printf("  %s %s (read source: %v)\n", Error.Render("✗"), id, err)
		res.Failed++
		return
	}

	existing, err := os.ReadFile(targetPath)
	if os.IsNotExist(err) {
		a.handleCreate(id, targetPath, newContent, res)
		return
	}
	if err != nil {
		fmt.Printf("  %s %s (read target: %v)\n", Error.Render("✗"), id, err)
		res.Failed++
		return
	}

	if string(existing) == string(newContent) {
		fmt.Printf("  %s %s (up to date)\n", Success.Render("✓"), id)
		res.Skipped++
		return
	}

	a.handleUpdate(id, targetPath, string(existing), string(newContent), res)
}

// handleCreate writes a new file at targetPath. Respects dry-run.
func (a *Apply) handleCreate(id, targetPath string, content []byte, res *dotfileResult) {
	if a.Runner.DryRun {
		fmt.Printf("  %s %s (would create: %s)\n", Warning.Render("⚠"), id, targetPath)
		res.Skipped++
		return
	}
	if err := writeFile(targetPath, content); err != nil {
		fmt.Printf("  %s %s (create: %v)\n", Error.Render("✗"), id, err)
		res.Failed++
		return
	}
	fmt.Printf("  %s %s (created: %s)\n", Success.Render("✓"), id, targetPath)
	res.Created++
}

// handleUpdate overwrites targetPath after showing a diff, prompting, and
// creating a backup. Respects dry-run.
func (a *Apply) handleUpdate(id, targetPath, existing, newContent string, res *dotfileResult) {
	if a.Runner.DryRun {
		fmt.Printf("  %s %s (would overwrite: %s)\n", Warning.Render("⚠"), id, targetPath)
		if text := ColorizeDiff(targetPath, existing, newContent); text != "" {
			fmt.Println(text)
		}
		res.Skipped++
		return
	}

	fmt.Printf("  %s %s (differs: %s)\n", Warning.Render("⚠"), id, targetPath)
	if text := ColorizeDiff(targetPath, existing, newContent); text != "" {
		fmt.Println(text)
	}

	if !Confirm(fmt.Sprintf("Overwrite %s?", id), a.AutoYes) {
		fmt.Printf("  %s %s (skipped)\n", Warning.Render("⚠"), id)
		res.Skipped++
		return
	}

	if err := backupFile(targetPath); err != nil {
		fmt.Printf("  %s %s (backup: %v)\n", Error.Render("✗"), id, err)
		res.Failed++
		return
	}

	if err := writeFile(targetPath, []byte(newContent)); err != nil {
		fmt.Printf("  %s %s (write: %v)\n", Error.Render("✗"), id, err)
		res.Failed++
		return
	}
	fmt.Printf("  %s %s (updated)\n", Success.Render("✓"), id)
	res.Updated++
}

// statusDotfiles prints whether each declared dotfile is in sync with the
// repo. Read-only; used by `status`.
func statusDotfiles(dotfiles []Dotfile, configsDir string) {
	fmt.Println(Header.Render("==> Dotfile status"))
	for _, df := range dotfiles {
		err := walkDotfile(df, configsDir, func(id, sourcePath, targetPath string) error {
			checkStatus(id, sourcePath, targetPath)
			return nil
		})
		if err != nil {
			fmt.Printf("  %s %s (%v)\n", Error.Render("✗"), df.ID, err)
		}
	}
}

// checkStatus prints the sync status of a single file.
func checkStatus(id, sourcePath, targetPath string) {
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		fmt.Printf("  %s %s (not deployed)\n", Error.Render("✗"), id)
		return
	}

	newContent, err := os.ReadFile(sourcePath)
	if err != nil {
		fmt.Printf("  %s %s (read source failed)\n", Error.Render("✗"), id)
		return
	}
	existing, err := os.ReadFile(targetPath)
	if err != nil {
		fmt.Printf("  %s %s (read target failed)\n", Error.Render("✗"), id)
		return
	}

	if string(existing) == string(newContent) {
		fmt.Printf("  %s %s\n", Success.Render("✓"), id)
	} else {
		fmt.Printf("  %s %s (differs)\n", Warning.Render("⚠"), id)
	}
}

// expandHomePath turns "~/.config/foo" into "/home/user/.config/foo".
// Absolute paths are returned unchanged. Anything else is left as-is
// (validation in Config.Validate should have caught it earlier).
func expandHomePath(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if p == "~" {
		return home
	}
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(home, p[2:])
	}
	return p
}

// writeFile creates parent directories as needed and writes content with
// mode 0644. Centralized so the create and update paths use the same logic.
func writeFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

// backupFile copies the existing file to "<path>.bak.<timestamp>" before
// the caller overwrites it. The timestamp makes repeated backups
// non-overwriting.
func backupFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	stamp := time.Now().Format("20060102-150405")
	return os.WriteFile(path+".bak."+stamp, content, 0o644)
}
