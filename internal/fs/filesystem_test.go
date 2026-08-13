package fs

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandPathNoExpansion(t *testing.T) {
	got, err := ExpandPath("/absolute/path")
	if err != nil {
		t.Fatalf("ExpandPath() error = %v", err)
	}
	if got != "/absolute/path" {
		t.Errorf("ExpandPath() = %q, want %q", got, "/absolute/path")
	}
}

func TestExpandPathTilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := ExpandPath("~/notes")
	if err != nil {
		t.Fatalf("ExpandPath() error = %v", err)
	}
	want := filepath.Join(home, "notes")
	if got != want {
		t.Errorf("ExpandPath() = %q, want %q", got, want)
	}
}

func TestExpandPathHomeVar(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := ExpandPath("$HOME/notes")
	if err != nil {
		t.Fatalf("ExpandPath() error = %v", err)
	}
	want := filepath.Join(home, "notes")
	if got != want {
		t.Errorf("ExpandPath() = %q, want %q", got, want)
	}
}

func TestExpandPathEmpty(t *testing.T) {
	got, err := ExpandPath("")
	if err != nil {
		t.Fatalf("ExpandPath() error = %v", err)
	}
	if got != "" {
		t.Errorf("ExpandPath() = %q, want empty", got)
	}
}

func TestExpandPathOnlyHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := ExpandPath("~")
	if err != nil {
		t.Fatalf("ExpandPath() error = %v", err)
	}
	if got != home {
		t.Errorf("ExpandPath() = %q, want %q", got, home)
	}
}

func TestExpandPathPreservesSuffix(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := ExpandPath("~/a/b/c")
	if err != nil {
		t.Fatalf("ExpandPath() error = %v", err)
	}
	if !strings.HasSuffix(got, "/a/b/c") {
		t.Errorf("ExpandPath() = %q, want suffix /a/b/c", got)
	}
}
