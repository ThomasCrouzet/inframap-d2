package util

import "regexp"

var jinjaVarPattern = regexp.MustCompile(`\{\{[^}]*\}\}`)

// StripJinja2 replaces Jinja2 {{ var }} expressions with a placeholder value
// so the YAML can be parsed by a standard YAML parser.
func StripJinja2(content string) string {
	return jinjaVarPattern.ReplaceAllString(content, "PLACEHOLDER")
}
