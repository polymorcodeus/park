// Package schema describes the canonical frontmatter contract for parked
// notes. It is a pure leaf package: no internal dependencies, stdlib only, so
// downstream tooling can import the contract instead of re-deriving it.
package schema

import (
	"strings"
)

// SchemaVersion is the current version of the frontmatter contract.
// It increments on any breaking change to the contract.
const SchemaVersion = 1

// DateFormat is the canonical date layout used for the `created` field.
const DateFormat = "2006-01-02"

// Category is a canonical parked-note category. The set below is the
// contract's enum and is independent of any loaded TOML config.
type Category string

// Canonical categories.
const (
	CategoryInbox    Category = "inbox"
	CategoryProjects Category = "projects"
	CategoryAreas    Category = "areas"
	CategoryArchive  Category = "archive"
)

// Categories returns the canonical category values in order.
func Categories() []string {
	return []string{
		string(CategoryInbox),
		string(CategoryProjects),
		string(CategoryAreas),
		string(CategoryArchive),
	}
}

// Frontmatter is the set of fields persisted as frontmatter in every note.
type Frontmatter struct {
	Category string
	Created  string
	Source   string
	Synopsis string
}

// IsComplete reports whether all frontmatter fields are populated.
func (f Frontmatter) IsComplete() bool {
	return fieldSet(f.Category) && fieldSet(f.Created) && fieldSet(f.Source) && fieldSet(f.Synopsis)
}

// MissingFields returns the frontmatter fields that are empty.
func (f Frontmatter) MissingFields() []string {
	var missing []string
	if !fieldSet(f.Category) {
		missing = append(missing, "category")
	}
	if !fieldSet(f.Created) {
		missing = append(missing, "created")
	}
	if !fieldSet(f.Source) {
		missing = append(missing, "source")
	}
	if !fieldSet(f.Synopsis) {
		missing = append(missing, "synopsis")
	}
	return missing
}

// WriteTemplate is the canonical frontmatter+body layout. Placeholders are
// {category}, {created}, {source}, {synopsis}, and {body}.
const WriteTemplate = "---\ncategory: {category}\ncreated: {created}\nsource: {source}\nsynopsis: {synopsis}\n---\n\n{body}\n"

// Render returns the frontmatter block and body with the given values
// substituted into WriteTemplate.
func Render(f Frontmatter, body string) string {
	return strings.NewReplacer(
		"{category}", f.Category,
		"{created}", f.Created,
		"{source}", f.Source,
		"{synopsis}", f.Synopsis,
		"{body}", body,
	).Replace(WriteTemplate)
}

func fieldSet(s string) bool {
	return strings.TrimSpace(s) != ""
}

// Field describes a single frontmatter field in the JSON schema export.
type Field struct {
	Name     string   `json:"name"`
	Required bool     `json:"required"`
	Kind     string   `json:"kind"`
	Values   []string `json:"values,omitempty"`
	Format   string   `json:"format,omitempty"`
	Auto     bool     `json:"auto,omitempty"`
}

// Spec is the JSON-serializable schema description.
type Spec struct {
	SchemaVersion int     `json:"schema_version"`
	Fields        []Field `json:"fields"`
	DateFormat    string  `json:"date_format"`
	WriteTemplate string  `json:"write_template"`
}

// Describe returns the current schema contract as a serializable value.
func Describe() Spec {
	return Spec{
		SchemaVersion: SchemaVersion,
		Fields: []Field{
			{Name: "category", Required: true, Kind: "enum", Values: Categories()},
			{Name: "created", Required: true, Kind: "date", Format: DateFormat, Auto: true},
			{Name: "source", Required: true, Kind: "string"},
			{Name: "synopsis", Required: true, Kind: "string"},
		},
		DateFormat:    DateFormat,
		WriteTemplate: WriteTemplate,
	}
}
