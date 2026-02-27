package util

import (
	"regexp"
	"strings"
)

var nonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

// SanitizeID converts a string into a valid D2 identifier.
// D2 identifiers must be alphanumeric with hyphens/underscores.
func SanitizeID(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = nonAlphaNum.ReplaceAllString(s, "")
	if s == "" {
		return "unknown"
	}
	return s
}

// Quote wraps a string in double quotes for D2 labels.
func Quote(s string) string {
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
