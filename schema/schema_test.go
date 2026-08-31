package schema

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestSchemaVersion(t *testing.T) {
	if SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", SchemaVersion)
	}
}

func TestCategories(t *testing.T) {
	want := []string{"inbox", "projects", "areas", "archive"}
	got := Categories()
	if !slices.Equal(got, want) {
		t.Errorf("Categories() = %v, want %v", got, want)
	}
}

func TestFrontmatterIsComplete(t *testing.T) {
	complete := Frontmatter{Category: "inbox", Created: "2026-08-31", Source: "test", Synopsis: "ok"}
	if !complete.IsComplete() {
		t.Error("complete frontmatter reported incomplete")
	}

	incomplete := Frontmatter{Category: "inbox", Source: "test", Synopsis: "ok"}
	if incomplete.IsComplete() {
		t.Error("incomplete frontmatter reported complete")
	}
}

func TestFrontmatterMissingFields(t *testing.T) {
	f := Frontmatter{Category: "inbox", Source: "test"}
	got := f.MissingFields()
	want := []string{"created", "synopsis"}
	if !slices.Equal(got, want) {
		t.Errorf("MissingFields() = %v, want %v", got, want)
	}
}

func TestRender(t *testing.T) {
	f := Frontmatter{
		Category: "projects",
		Created:  "2026-08-31",
		Source:   "terminal",
		Synopsis: "a note",
	}
	got := Render(f, "# body\n")
	want := "---\ncategory: projects\ncreated: 2026-08-31\nsource: terminal\nsynopsis: a note\n---\n\n# body\n\n"
	if got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestDescribe(t *testing.T) {
	s := Describe()

	if s.SchemaVersion != SchemaVersion {
		t.Errorf("Spec.SchemaVersion = %d, want %d", s.SchemaVersion, SchemaVersion)
	}

	if s.DateFormat != DateFormat {
		t.Errorf("Spec.DateFormat = %q, want %q", s.DateFormat, DateFormat)
	}

	if s.WriteTemplate != WriteTemplate {
		t.Errorf("Spec.WriteTemplate = %q, want %q", s.WriteTemplate, WriteTemplate)
	}

	if len(s.Fields) != 4 {
		t.Fatalf("len(Spec.Fields) = %d, want 4", len(s.Fields))
	}

	category := s.Fields[0]
	if category.Name != "category" || category.Kind != "enum" || !category.Required {
		t.Errorf("category field = %+v, want required enum", category)
	}
	if !slices.Equal(category.Values, Categories()) {
		t.Errorf("category values = %v, want %v", category.Values, Categories())
	}

	created := s.Fields[1]
	if created.Name != "created" || created.Kind != "date" || created.Format != DateFormat || !created.Auto {
		t.Errorf("created field = %+v, want date with format and auto", created)
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal(Spec) error = %v", err)
	}

	var round Spec
	if err := json.Unmarshal(data, &round); err != nil {
		t.Fatalf("json.Unmarshal(Spec) error = %v", err)
	}
	if round.SchemaVersion != SchemaVersion {
		t.Errorf("round-trip SchemaVersion = %d, want %d", round.SchemaVersion, SchemaVersion)
	}
}
