package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/polymorcodeus/park/internal/config"
	"github.com/polymorcodeus/park/internal/note"
)

func TestInitCreatesFolders(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)

	if _, _, err := Init(cfg); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	for _, cat := range cfg.Categories {
		if _, err := os.Stat(cat.Path); err != nil {
			t.Errorf("missing category folder %q: %v", cat.Path, err)
		}
	}
}

func TestInitIdempotent(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)

	createdFirst, _, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() first error = %v", err)
	}
	if len(createdFirst) != len(cfg.Categories) {
		t.Fatalf("Init() first created = %d, want %d", len(createdFirst), len(cfg.Categories))
	}

	createdSecond, existedSecond, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init() second error = %v", err)
	}
	if len(createdSecond) != 0 {
		t.Fatalf("Init() second created = %d, want 0", len(createdSecond))
	}
	if len(existedSecond) != len(cfg.Categories) {
		t.Fatalf("Init() second existed = %d, want %d", len(existedSecond), len(cfg.Categories))
	}
}

func TestCheckAllExist(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)
	if _, _, err := Init(cfg); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	missing, err := Check(cfg)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(missing) > 0 {
		t.Fatalf("Check() = %v, want none", missing)
	}
}

func TestCheckMissing(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)

	// Only create one folder
	_ = os.MkdirAll(cfg.Categories[0].Path, 0o755)

	missing, err := Check(cfg)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(missing) != 3 {
		t.Fatalf("Check() = %d missing, want 3", len(missing))
	}
}

func TestCheckPathIsFile(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)

	// Create a file where a category folder is expected.
	if err := os.WriteFile(cfg.Categories[0].Path, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Check(cfg)
	if err == nil {
		t.Fatal("Check() expected error when category path is a file")
	}
}

func TestNewCreatesNote(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)
	if _, _, err := Init(cfg); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	path, err := note.NewWithBody(cfg, "My Note", "a synopsis", "test", "inbox", "")
	if err != nil {
		t.Fatalf("NewWithBody() error = %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("note file missing: %v", err)
	}

	fm, _, err := note.ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("ParseFrontmatter() error = %v", err)
	}
	if fm.Category != "inbox" {
		t.Errorf("category = %q, want inbox", fm.Category)
	}
	if fm.Synopsis != "a synopsis" {
		t.Errorf("synopsis = %q, want %q", fm.Synopsis, "a synopsis")
	}
	if fm.Source != "test" {
		t.Errorf("source = %q, want test", fm.Source)
	}
}

func TestReclassifyMovesFile(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)
	if _, _, err := Init(cfg); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	path, err := note.NewWithBody(cfg, "Move Me", "synopsis", "test", "inbox", "")
	if err != nil {
		t.Fatalf("NewWithBody() error = %v", err)
	}
	filename := filepath.Base(path)

	if err := Reclassify(cfg, filename, "projects"); err != nil {
		t.Fatalf("Reclassify() error = %v", err)
	}

	inboxPath := filepath.Join(tmp, "_inbox", filename)
	projectsPath := filepath.Join(tmp, "_projects", filename)

	if _, err := os.Stat(inboxPath); !os.IsNotExist(err) {
		t.Errorf("file still exists in inbox: %v", err)
	}
	if _, err := os.Stat(projectsPath); err != nil {
		t.Errorf("file missing in projects: %v", err)
	}

	fm, _, err := note.ParseFrontmatter(projectsPath)
	if err != nil {
		t.Fatalf("ParseFrontmatter() error = %v", err)
	}
	if fm.Category != "projects" {
		t.Errorf("category = %q, want projects", fm.Category)
	}
}

func TestReclassifySameCategory(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)
	if _, _, err := Init(cfg); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	path, err := note.NewWithBody(cfg, "Stay Put", "synopsis", "test", "inbox", "")
	if err != nil {
		t.Fatalf("NewWithBody() error = %v", err)
	}
	filename := filepath.Base(path)

	err = Reclassify(cfg, filename, "inbox")
	if err == nil {
		t.Fatal("expected error when reclassifying to the same category")
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("original file was moved or removed: %v", err)
	}
}

func TestReclassifyUnknownCategory(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)

	err := Reclassify(cfg, "note.md", "nope")
	if err == nil {
		t.Fatal("expected error for unknown category")
	}
}

func TestReclassifyMissingFile(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)
	if _, _, err := Init(cfg); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err := Reclassify(cfg, "missing.md", "projects")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestScan(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)
	if _, _, err := Init(cfg); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if _, err := note.NewWithBody(cfg, "First", "oldest", "test", "inbox", ""); err != nil {
		t.Fatalf("NewWithBody() error = %v", err)
	}
	if _, err := note.NewWithBody(cfg, "Second", "newer", "test", "inbox", ""); err != nil {
		t.Fatalf("NewWithBody() error = %v", err)
	}

	items, err := Scan(cfg, "inbox")
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("Scan() returned %d items, want 2", len(items))
	}
}

func TestScanUnknownCategory(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)

	_, err := Scan(cfg, "nope")
	if err == nil {
		t.Fatal("expected error for unknown category")
	}
}

func TestResolvePath(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)
	if _, _, err := Init(cfg); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	path, err := note.NewWithBody(cfg, "Resolve Me", "synopsis", "test", "inbox", "")
	if err != nil {
		t.Fatalf("NewWithBody() error = %v", err)
	}
	filename := filepath.Base(path)

	got := ResolvePath(cfg, filename)
	if got != path {
		t.Errorf("ResolvePath(%q) = %q, want %q", filename, got, path)
	}

	fullPath := ResolvePath(cfg, path)
	if fullPath != path {
		t.Errorf("ResolvePath(%q) = %q, want %q", path, fullPath, path)
	}
}
