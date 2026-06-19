package app

import "github.com/charmbracelet/lipgloss"

// Centralized lipgloss styles used across the CLI.
// Keeping them here makes the look-and-feel easy to change in one place.
var (
	// Success — green checkmarks and positive outcomes.
	Success = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

	// Warning — yellow/orange, used for "would change" / "differs" hints.
	Warning = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	// Error — red, used for failures.
	Error = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	// Header — bold orange, used for section titles like "==> Installing packages".
	Header = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)

	// Prompt — bold orange, used for "[y/n]" confirmations.
	Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)

	// DiffAdded — green, lines starting with + in diffs.
	DiffAdded = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

	// DiffRemoved — red, lines starting with - in diffs.
	DiffRemoved = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	// DiffContext — grey, context lines in diffs.
	DiffContext = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)
