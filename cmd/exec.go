package cmd

import (
	"os/exec"
)

// findExecutable wraps exec.LookPath for testability.
func findExecutable(name string) (string, error) {
	return exec.LookPath(name)
}

// execCommand wraps exec.Command for testability.
func execCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}
