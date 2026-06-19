package app

import (
	"fmt"
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

// runDotfiles walks every dotfile declaration and reconciles the source
// file in the repo with the target file on disk. The flow is:
//  1. resolve paths (source relative to ConfigsDir, target expands ~)
//  2. if target is missing → create
//  3. if contents match → skip
//  4. if contents differ → show diff, prompt, backup, write
func (a *Apply) runDotfiles() (*dotfileResult, error) {
	fmt.Println(Header.Render("==> Deploying dotfiles"))

	res := &dotfileResult{}
	for _, df := range a.Cfg.Dotfiles {
		sourcePath := filepath.Join(a.ConfigsDir, df.Source)
		targetPath := expandPath(df.Target)

		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			fmt.Printf("  %s %s (source missing: %s)\n", Error.Render("✗"), df.ID, df.Source)
			res.Failed++
			continue
		}

		newContent, err := os.ReadFile(sourcePath)
		if err != nil {
			fmt.Printf("  %s %s (read source: %v)\n", Error.Render("✗"), df.ID, err)
			res.Failed++
			continue
		}

		// Target doesn't exist yet → create.
		existing, err := os.ReadFile(targetPath)
		if os.IsNotExist(err) {
			if a.Runner.DryRun {
				fmt.Printf("  %s %s (would create: %s)\n", Warning.Render("⚠"), df.ID, targetPath)
				res.Skipped++
				continue
			}
			if err := writeFile(targetPath, newContent); err != nil {
				fmt.Printf("  %s %s (create: %v)\n", Error.Render("✗"), df.ID, err)
				res.Failed++
				continue
			}
			fmt.Printf("  %s %s (created: %s)\n", Success.Render("✓"), df.ID, targetPath)
			res.Created++
			continue
		}
		if err != nil {
			fmt.Printf("  %s %s (read target: %v)\n", Error.Render("✗"), df.ID, err)
			res.Failed++
			continue
		}

		// Contents match → nothing to do.
		if string(existing) == string(newContent) {
			fmt.Printf("  %s %s (up to date)\n", Success.Render("✓"), df.ID)
			res.Skipped++
			continue
		}

		// Contents differ → diff, confirm, backup, write.
		if a.Runner.DryRun {
			fmt.Printf("  %s %s (would overwrite: %s)\n", Warning.Render("⚠"), df.ID, targetPath)
			if text := Unified(targetPath, string(existing), string(newContent)); text != "" {
				fmt.Println(text)
			}
			res.Skipped++
			continue
		}

		fmt.Printf("  %s %s (differs: %s)\n", Warning.Render("⚠"), df.ID, targetPath)
		if text := Unified(targetPath, string(existing), string(newContent)); text != "" {
			fmt.Println(text)
		}

		if !Confirm(fmt.Sprintf("Overwrite %s?", df.ID), a.AutoYes) {
			fmt.Printf("  %s %s (skipped)\n", Warning.Render("⚠"), df.ID)
			res.Skipped++
			continue
		}

		if err := backupFile(targetPath); err != nil {
			fmt.Printf("  %s %s (backup: %v)\n", Error.Render("✗"), df.ID, err)
			res.Failed++
			continue
		}

		if err := writeFile(targetPath, newContent); err != nil {
			fmt.Printf("  %s %s (write: %v)\n", Error.Render("✗"), df.ID, err)
			res.Failed++
			continue
		}
		fmt.Printf("  %s %s (updated)\n", Success.Render("✓"), df.ID)
		res.Updated++
	}

	return res, nil
}

// statusDotfiles prints whether each declared dotfile is in sync with the
// repo. Read-only; used by `status`.
func statusDotfiles(dotfiles []Dotfile, configsDir string) {
	fmt.Println(Header.Render("==> Dotfile status"))
	for _, df := range dotfiles {
		sourcePath := filepath.Join(configsDir, df.Source)
		targetPath := expandPath(df.Target)

		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			fmt.Printf("  %s %s (source missing)\n", Error.Render("✗"), df.ID)
			continue
		}
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			fmt.Printf("  %s %s (not deployed)\n", Error.Render("✗"), df.ID)
			continue
		}

		newContent, err := os.ReadFile(sourcePath)
		if err != nil {
			fmt.Printf("  %s %s (read source failed)\n", Error.Render("✗"), df.ID)
			continue
		}
		existing, err := os.ReadFile(targetPath)
		if err != nil {
			fmt.Printf("  %s %s (read target failed)\n", Error.Render("✗"), df.ID)
			continue
		}

		if string(existing) == string(newContent) {
			fmt.Printf("  %s %s\n", Success.Render("✓"), df.ID)
		} else {
			fmt.Printf("  %s %s (differs)\n", Warning.Render("⚠"), df.ID)
		}
	}
}

// expandPath turns "~/.config/foo" into "/home/user/.config/foo".
// Absolute paths are returned unchanged. Anything else is left as-is
// (validation in Config.Validate should have caught it earlier).
func expandPath(p string) string {
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
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
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
	return os.WriteFile(path+".bak."+stamp, content, 0644)
}
