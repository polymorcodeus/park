// Package fs provides filesystem helpers for park.
package fs

import (
	"fmt"
	"os"
	"strings"
)

// ExpandPath expands a leading `~` or `$HOME` in path using the user's home
// directory. It returns the path unchanged if no expansion is needed.
func ExpandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") && !strings.HasPrefix(path, "$HOME") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
		if home == "" {
			return "", fmt.Errorf("cannot expand home in path %q: %w", path, err)
		}
	}

	if after, ok := strings.CutPrefix(path, "~"); ok {
		return home + after, nil
	}
	if after, ok := strings.CutPrefix(path, "$HOME"); ok {
		return home + after, nil
	}
	return path, nil
}
