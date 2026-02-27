package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#DC2626")).Bold(true)
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#CA8A04"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#16A34A"))
	hintStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Italic(true)
	boldStyle    = lipgloss.NewStyle().Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
)

// FormatError returns a styled multi-line error message.
func FormatError(title, detail, suggestion string) string {
	out := errorStyle.Render("Error: "+title) + "\n"
	if detail != "" {
		out += "  " + detail + "\n"
	}
	if suggestion != "" {
		out += "  " + hintStyle.Render("Hint: "+suggestion) + "\n"
	}
	return out
}

// CollectorStarted prints a styled status when a collector begins work.
func CollectorStarted(name string) {
	fmt.Printf("  %s %s\n", dimStyle.Render("..."), name)
}

// CollectorDone prints a styled status when a collector finishes.
func CollectorDone(name, detail string) {
	msg := successStyle.Render("  OK ") + " " + name
	if detail != "" {
		msg += " " + dimStyle.Render(detail)
	}
	// overwrite the "started" line by moving up
	fmt.Printf("\033[1A\033[2K%s\n", msg)
}

// CollectorSkipped prints a styled status when a collector is not enabled.
func CollectorSkipped(name string) {
	fmt.Printf("  %s %s\n", dimStyle.Render("--"), dimStyle.Render(name+" (skipped)"))
}

// Success prints a green success message.
func Success(msg string) {
	fmt.Println(successStyle.Render(msg))
}

// Warn prints a yellow warning message.
func Warn(msg string) {
	fmt.Println(warnStyle.Render("Warning: " + msg))
}

// Bold renders text in bold.
func Bold(s string) string {
	return boldStyle.Render(s)
}

// Hint renders text in dim italic.
func Hint(s string) string {
	return hintStyle.Render(s)
}

// ValidationOK prints a green check for a valid field.
func ValidationOK(field, detail string) {
	fmt.Printf("  %s %s: %s\n", successStyle.Render("OK "), field, detail)
}

// ValidationErr prints a red error for an invalid field.
func ValidationErr(field, message, suggestion string) {
	fmt.Printf("  %s %s: %s\n", errorStyle.Render("ERR"), field, message)
	if suggestion != "" {
		fmt.Printf("      %s\n", hintStyle.Render("Hint: "+suggestion))
	}
}
