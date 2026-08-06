package model

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/polymorcodeus/park/internal/config"
	"github.com/polymorcodeus/park/internal/note"
	"github.com/polymorcodeus/park/internal/store"
)

func TestNewNoteFormModel(t *testing.T) {
	cfg := config.DefaultConfig(t.TempDir())
	m, err := NewNoteFormModel(cfg, "filename", "synopsis", "source", "inbox", "body", "")
	if err != nil {
		t.Fatalf("NewNoteFormModel() error = %v", err)
	}

	if m.inputs[0].Value() != "filename" {
		t.Errorf("filename = %q, want filename", m.inputs[0].Value())
	}
	if m.inputs[1].Value() != "synopsis" {
		t.Errorf("synopsis = %q, want synopsis", m.inputs[1].Value())
	}
	if m.inputs[2].Value() != "source" {
		t.Errorf("source = %q, want source", m.inputs[2].Value())
	}
	if m.categoryName() != "inbox" {
		t.Errorf("category = %q, want inbox", m.categoryName())
	}
	if m.bodyInput.Value() != "body" {
		t.Errorf("body = %q, want body", m.bodyInput.Value())
	}
}

func TestNewNoteFormModelUnknownCategory(t *testing.T) {
	cfg := config.DefaultConfig(t.TempDir())
	_, err := NewNoteFormModel(cfg, "", "", "", "nope", "", "")
	if err == nil {
		t.Fatal("expected error for unknown category")
	}
}

func TestNoteFormModelSubmission(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)
	if _, _, err := store.Init(cfg); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	m, err := NewNoteFormModel(cfg, "Form Note", "form synopsis", "test", "inbox", "# Form Note\n", "")
	if err != nil {
		t.Fatalf("NewNoteFormModel() error = %v", err)
	}

	updated, cmd := m.Update(submitMsg{})
	final, ok := updated.(NoteFormModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if final.err != nil {
		t.Fatalf("submit error = %v", final.err)
	}
	if cmd == nil {
		t.Fatal("expected create command")
	}

	msg := cmd()
	created, ok := msg.(formCreatedMsg)
	if !ok {
		t.Fatalf("expected formCreatedMsg, got %T", msg)
	}
	if created.path == "" {
		t.Fatal("expected non-empty path")
	}
}

func TestNoteFormModelRequiresSource(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)
	if _, _, err := store.Init(cfg); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	m, err := NewNoteFormModel(cfg, "Filename", "Synopsis", "", "inbox", "", "")
	if err != nil {
		t.Fatalf("NewNoteFormModel() error = %v", err)
	}

	updated, cmd := m.Update(submitMsg{})
	final, ok := updated.(NoteFormModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if final.err == nil {
		t.Fatal("expected error for missing source")
	}
	if cmd != nil {
		t.Fatal("expected no command when validation fails")
	}
}

func TestNoteFormModelView(t *testing.T) {
	cfg := config.DefaultConfig(t.TempDir())
	m, err := NewNoteFormModel(cfg, "", "", "", "inbox", "", "")
	if err != nil {
		t.Fatalf("NewNoteFormModel() error = %v", err)
	}

	_ = m.View()
}

func TestNoteFormModelCategoryNavigation(t *testing.T) {
	cfg := config.DefaultConfig(t.TempDir())
	m, err := NewNoteFormModel(cfg, "", "", "", "inbox", "", "")
	if err != nil {
		t.Fatalf("NewNoteFormModel() error = %v", err)
	}

	m.focusIndex = m.categoryIndex()
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	final, ok := updated.(NoteFormModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if final.categoryName() != "projects" {
		t.Errorf("category after right = %q, want projects", final.categoryName())
	}
}

func TestNoteFormModelBodyCursorStartsAtTop(t *testing.T) {
	cfg := config.DefaultConfig(t.TempDir())
	body := strings.Repeat("line\n", 20)
	m, err := NewNoteFormModel(cfg, "", "", "", "inbox", body, "")
	if err != nil {
		t.Fatalf("NewNoteFormModel() error = %v", err)
	}

	m.focusIndex = m.bodyIndex()
	m, _ = m.updateFocus()

	if m.bodyInput.ScrollYOffset() != 0 {
		t.Errorf("body scroll offset = %d, want 0", m.bodyInput.ScrollYOffset())
	}
}

func TestNoteFormModelFilePreview(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.DefaultConfig(tmp)
	path := tmp + "/draft.md"
	if err := note.WriteFrontmatter(path, note.Frontmatter{}, "# Draft\n\ncontent\n"); err != nil {
		t.Fatalf("WriteFrontmatter() error = %v", err)
	}

	m, err := NewNoteFormModel(cfg, "", "", "", "inbox", "", path)
	if err != nil {
		t.Fatalf("NewNoteFormModel() error = %v", err)
	}
	if m.hasBodyField() {
		t.Fatal("expected body field to be hidden when from-file is set")
	}
	if m.filePreview == "" {
		t.Fatal("expected file preview to be populated")
	}
}
