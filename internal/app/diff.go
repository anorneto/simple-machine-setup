package app

import (
	"fmt"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

// TODO(anor): this difflib is old.. should we make one?

// ColorizeDiff produces a colored unified diff between two strings.
// `label` is used as the filename in the diff header (e.g. the dotfile path).
// Returns an empty string if the contents are identical.
func ColorizeDiff(label, oldContent, newContent string) string {
	d := difflib.UnifiedDiff{
		A:        difflib.SplitLines(oldContent),
		B:        difflib.SplitLines(newContent),
		FromFile: label,
		ToFile:   label + " (new)",
		Context:  3,
	}

	text, err := difflib.GetUnifiedDiffString(d)
	if err != nil {
		return fmt.Sprintf("error generating diff: %v", err)
	}

	if text == "" {
		return ""
	}

	return colorize(text)
}

// colorize applies lipgloss styles to each line of a unified diff.
// + lines are green, - lines are red, @@/+++/--- are orange headers,
// and context lines are dimmed.
func colorize(diffText string) string {
	lines := strings.Split(diffText, "\n")
	var result []string

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+++"), strings.HasPrefix(line, "---"):
			result = append(result, Header.Render(line))
		case strings.HasPrefix(line, "@@"):
			result = append(result, Header.Render(line))
		case strings.HasPrefix(line, "+"):
			result = append(result, DiffAdded.Render(line))
		case strings.HasPrefix(line, "-"):
			result = append(result, DiffRemoved.Render(line))
		case line == "":
			result = append(result, line)
		default:
			result = append(result, DiffContext.Render(line))
		}
	}

	return strings.Join(result, "\n")
}
