// Package fs provides filesystem helpers for park.
package fs

import (
	"os"
	"strings"
)

func homeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		dir = os.Getenv("HOME")
	}
	return dir
}

func ExpandPath(path string) string {
	home := homeDir()

	// Expand ~
	if after, ok := strings.CutPrefix(path, "~"); ok {
		return home + after
	}
	// Expand $HOME
	if after, ok := strings.CutPrefix(path, "$HOME"); ok {
		return home + after
	}

	return path
}
