package note

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/polymorcodeus/park/internal/config"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		want     Frontmatter
		wantBody string
	}{
		{
			name:    "all fields",
			content: "---\ncategory: inbox\ncreated: 2026-07-28\nsource: terminal\nsynopsis: a test note\n---\n\n# hello\n",
			want: Frontmatter{
				Category: "inbox",
				Created:  "2026-07-28",
				Source:   "terminal",
				Synopsis: "a test note",
			},
			wantBody: "# hello",
		},
		{
			name:    "empty body",
			content: "---\ncategory: archive\ncreated: 2026-07-28\nsource: chat\nsynopsis:\n---\n",
			want: Frontmatter{
				Category: "archive",
				Created:  "2026-07-28",
				Source:   "chat",
				Synopsis: "",
			},
			wantBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			path := filepath.Join(tmp, "note.md")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("write test file: %v", err)
			}

			got, body, err := ParseFrontmatter(path)
			if err != nil {
				t.Fatalf("ParseFrontmatter() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("ParseFrontmatter() = %+v, want %+v", got, tt.want)
			}
			if body != tt.wantBody {
				t.Errorf("ParseFrontmatter() body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

func TestParseFrontmatterMissingFile(t *testing.T) {
	_, _, err := ParseFrontmatter(filepath.Join(t.TempDir(), "nope.md"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestWriteFrontmatterRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "note.md")
	fm := Frontmatter{
		Category: "projects",
		Created:  "2026-07-28",
		Source:   "test",
		Synopsis: "round trip",
	}

	if err := WriteFrontmatter(path, fm, "# title\n"); err != nil {
		t.Fatalf("WriteFrontmatter() error = %v", err)
	}

	got, body, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("ParseFrontmatter() error = %v", err)
	}
	if got != fm {
		t.Errorf("frontmatter mismatch: %+v, want %+v", got, fm)
	}
	if body != "# title\n" {
		t.Errorf("body = %q, want %q", body, "# title\n")
	}
}

func TestToday(t *testing.T) {
	got := Today()
	if len(got) != 10 {
		t.Errorf("Today() = %q, expected YYYY-MM-DD format", got)
	}
}

func TestNewWithBody(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)
	if err := os.MkdirAll(cfg.Categories[0].Path, 0o755); err != nil {
		t.Fatalf("create category folder: %v", err)
	}

	body := "## heading\n\nparagraph\n"
	path, err := NewWithBody(cfg, "Body Note", "with body", "test", "inbox", body)
	if err != nil {
		t.Fatalf("NewWithBody() error = %v", err)
	}

	fm, parsedBody, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("ParseFrontmatter() error = %v", err)
	}
	if fm.Category != "inbox" {
		t.Errorf("category = %q, want inbox", fm.Category)
	}
	if parsedBody != body {
		t.Errorf("body = %q, want %q", parsedBody, body)
	}
}

func TestNewWithBodyMissingFolder(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)

	_, err := NewWithBody(cfg, "Note", "synopsis", "test", "inbox", "")
	if err == nil {
		t.Fatal("expected error when category folder does not exist")
	}
}

func TestIngestFile(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)
	if err := os.MkdirAll(cfg.Categories[1].Path, 0o755); err != nil {
		t.Fatalf("create category folder: %v", err)
	}

	src := filepath.Join(tmp, "draft.md")
	content := "# Draft\n\nsome content\n"
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	path, err := IngestFile(cfg, src, "Draft Note", "ingested", "agent", "projects", "")
	if err != nil {
		t.Fatalf("IngestFile() error = %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source file was not removed: %v", err)
	}

	fm, body, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("ParseFrontmatter() error = %v", err)
	}
	if fm.Category != "projects" {
		t.Errorf("category = %q, want projects", fm.Category)
	}
	if fm.Synopsis != "ingested" {
		t.Errorf("synopsis = %q, want ingested", fm.Synopsis)
	}
	if fm.Source != "agent" {
		t.Errorf("source = %q, want agent", fm.Source)
	}
	if body != content {
		t.Errorf("body = %q, want %q", body, content)
	}
}

func TestIngestFileMissingSource(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)

	_, err := IngestFile(cfg, filepath.Join(tmp, "missing.md"), "Filename", "synopsis", "test", "inbox", "")
	if err == nil {
		t.Fatal("expected error for missing source file")
	}
}

func TestIngestFileDirectory(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)

	dir := filepath.Join(tmp, "adir")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	_, err := IngestFile(cfg, dir, "Filename", "synopsis", "test", "inbox", "")
	if err == nil {
		t.Fatal("expected error for directory source path")
	}
}

func TestExtractHeading(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		wantHeading   string
		wantRemaining string
		wantOK        bool
	}{
		{
			name:          "h1 at start",
			body:          "# My Title\n\nbody content\n",
			wantHeading:   "My Title",
			wantRemaining: "body content",
			wantOK:        true,
		},
		{
			name:          "h1 after blank lines",
			body:          "\n\n# Another Title\ncontent\n",
			wantHeading:   "Another Title",
			wantRemaining: "content",
			wantOK:        true,
		},
		{
			name:          "no h1",
			body:          "just some text\n",
			wantHeading:   "",
			wantRemaining: "just some text",
			wantOK:        false,
		},
		{
			name:          "h2 not extracted",
			body:          "## Section\n",
			wantHeading:   "",
			wantRemaining: "## Section",
			wantOK:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			heading, remaining, ok := extractHeading(tt.body)
			if ok != tt.wantOK {
				t.Fatalf("extractHeading() ok = %v, want %v", ok, tt.wantOK)
			}
			if heading != tt.wantHeading {
				t.Errorf("heading = %q, want %q", heading, tt.wantHeading)
			}
			if remaining != tt.wantRemaining {
				t.Errorf("remaining = %q, want %q", remaining, tt.wantRemaining)
			}
		})
	}
}

func TestExtractH1(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantHeading string
		wantOK      bool
	}{
		{
			name:        "h1 present",
			body:        "# My Title\n\nbody\n",
			wantHeading: "My Title",
			wantOK:      true,
		},
		{
			name:        "h1 after blank lines",
			body:        "\n\n# Another Title\ncontent\n",
			wantHeading: "Another Title",
			wantOK:      true,
		},
		{
			name:        "no h1",
			body:        "just some text\n",
			wantHeading: "",
			wantOK:      false,
		},
		{
			name:        "h2 not extracted",
			body:        "## Section\n",
			wantHeading: "",
			wantOK:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			heading, ok := ExtractH1(tt.body)
			if ok != tt.wantOK {
				t.Fatalf("ExtractH1() ok = %v, want %v", ok, tt.wantOK)
			}
			if heading != tt.wantHeading {
				t.Errorf("ExtractH1() heading = %q, want %q", heading, tt.wantHeading)
			}
		})
	}
}
