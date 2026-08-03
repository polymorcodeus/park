package config

import (
	"testing"
)

func TestValidate(t *testing.T) {
	base := Config{
		DefaultCategory: "inbox",
		Categories: []Category{
			{Name: "inbox", Path: "/tmp/park/_inbox", Key: "i"},
			{Name: "projects", Path: "/tmp/park/_projects", Key: "p"},
		},
	}

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "valid",
			cfg:     base,
			wantErr: false,
		},
		{
			name: "missing default category",
			cfg: Config{
				Categories: base.Categories,
			},
			wantErr: true,
		},
		{
			name: "default category not found",
			cfg: Config{
				DefaultCategory: "areas",
				Categories:      base.Categories,
			},
			wantErr: true,
		},
		{
			name: "empty category name",
			cfg: Config{
				DefaultCategory: "inbox",
				Categories: []Category{
					{Name: "inbox", Path: "/tmp/park/_inbox", Key: "i"},
					{Name: "", Path: "/tmp/park/_empty", Key: "e"},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate category name",
			cfg: Config{
				DefaultCategory: "inbox",
				Categories: []Category{
					{Name: "inbox", Path: "/tmp/park/_inbox", Key: "i"},
					{Name: "inbox", Path: "/tmp/park/_inbox2", Key: "p"},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate key",
			cfg: Config{
				DefaultCategory: "inbox",
				Categories: []Category{
					{Name: "inbox", Path: "/tmp/park/_inbox", Key: "i"},
					{Name: "projects", Path: "/tmp/park/_projects", Key: "i"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHasCategory(t *testing.T) {
	cfg := Config{
		DefaultCategory: "inbox",
		Categories: []Category{
			{Name: "inbox"},
			{Name: "projects"},
		},
	}

	if !cfg.HasCategory("inbox") {
		t.Error("HasCategory(inbox) = false, want true")
	}
	if cfg.HasCategory("missing") {
		t.Error("HasCategory(missing) = true, want false")
	}
}

func TestCategoryByName(t *testing.T) {
	cfg := Config{
		DefaultCategory: "inbox",
		Categories: []Category{
			{Name: "inbox", Path: "/tmp/inbox", Key: "i"},
		},
	}

	cat, ok := cfg.CategoryByName("inbox")
	if !ok || cat.Name != "inbox" {
		t.Errorf("CategoryByName(inbox) = %+v, %v; want inbox category", cat, ok)
	}

	_, ok = cfg.CategoryByName("nope")
	if ok {
		t.Error("CategoryByName(nope) found unexpected category")
	}
}

func TestCategoryByKey(t *testing.T) {
	cfg := Config{
		DefaultCategory: "inbox",
		Categories: []Category{
			{Name: "inbox", Key: "i"},
			{Name: "projects", Key: "p"},
		},
	}

	cat, ok := cfg.CategoryByKey("p")
	if !ok || cat.Name != "projects" {
		t.Errorf("CategoryByKey(p) = %+v, %v; want projects", cat, ok)
	}

	_, ok = cfg.CategoryByKey("z")
	if ok {
		t.Error("CategoryByKey(z) found unexpected category")
	}
}

func TestCategoryNames(t *testing.T) {
	cfg := Config{
		Categories: []Category{
			{Name: "inbox"},
			{Name: "projects"},
		},
	}

	got := cfg.CategoryNames()
	want := []string{"inbox", "projects"}
	if len(got) != len(want) {
		t.Fatalf("CategoryNames() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("CategoryNames()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDefaultConfigUsesProvidedRoot(t *testing.T) {
	cfg := DefaultConfig("/custom/root")
	if cfg.Categories[0].Path != "/custom/root/_inbox" {
		t.Errorf("DefaultConfig path = %q, want %q", cfg.Categories[0].Path, "/custom/root/_inbox")
	}
}
