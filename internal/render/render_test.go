package render

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/polymorcodeus/park/internal/config"
	"github.com/polymorcodeus/park/internal/note"
)

var ansiEscape = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

func TestShowFileRendersNote(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)
	if err := os.MkdirAll(cfg.Categories[0].Path, 0o755); err != nil {
		t.Fatalf("create category folder: %v", err)
	}

	path := filepath.Join(cfg.Categories[0].Path, "test-note.md")
	n := note.Note{
		Body: "# Hello\n\nbody content\n",
		Metadata: note.Metadata{
			Category: "inbox",
			Created:  "2026-08-09",
			Source:   "terminal",
			Synopsis: "a rendered note",
		},
	}
	if err := note.Write(path, n); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var buf bytes.Buffer
	if err := ShowFile(path, &buf); err != nil {
		t.Fatalf("ShowFile() error = %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Fatal("ShowFile() produced empty output")
	}
	plain := stripANSI(out)
	if !strings.Contains(plain, "a rendered note") {
		t.Errorf("output missing synopsis; got %q", plain)
	}
	if !strings.Contains(plain, "Hello") {
		t.Errorf("output missing body; got %q", plain)
	}
}

func TestShowFileMissing(t *testing.T) {
	var buf bytes.Buffer
	err := ShowFile(filepath.Join(t.TempDir(), "missing.md"), &buf)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
