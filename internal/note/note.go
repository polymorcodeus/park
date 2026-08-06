// Package note handles parked-note content: frontmatter, creation, ingestion,
// and the headless/interactive decision path.
package note

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/polymorcodeus/park/internal/config"
	"github.com/polymorcodeus/park/internal/fs"
)

// Frontmatter is the fixed schema for every parked note. Deliberately flat
// (no nesting) so it can be parsed line-by-line without a YAML dependency.
type Frontmatter struct {
	Category string
	Created  string
	Source   string
	Synopsis string
}

// MissingFields returns the required frontmatter fields that are empty.
func (fm Frontmatter) MissingFields() []string {
	var missing []string
	if strings.TrimSpace(fm.Category) == "" {
		missing = append(missing, "category")
	}
	if strings.TrimSpace(fm.Created) == "" {
		missing = append(missing, "created")
	}
	if strings.TrimSpace(fm.Source) == "" {
		missing = append(missing, "source")
	}
	if strings.TrimSpace(fm.Synopsis) == "" {
		missing = append(missing, "synopsis")
	}
	return missing
}

// ParseFrontmatter reads the leading `---` delimited block from a markdown
// file and returns the parsed fields plus the raw body that follows.
func ParseFrontmatter(path string) (Frontmatter, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return Frontmatter{}, "", fmt.Errorf("open %q: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	fm, body, _, err := parseFrontmatter(newScanner(f))
	if err != nil {
		return fm, "", fmt.Errorf("parse %q: %w", path, err)
	}
	return fm, body, nil
}

// ParseFrontmatterString parses the leading `---` delimited block from a
// markdown string and returns the parsed fields, the raw body that follows,
// and a flag indicating whether a frontmatter block was found.
func ParseFrontmatterString(content string) (Frontmatter, string, bool) {
	fm, body, found, _ := parseFrontmatter(newStringScanner(content))
	return fm, body, found
}

func newScanner(f *os.File) *bufio.Scanner { return bufio.NewScanner(f) }

func newStringScanner(s string) *bufio.Scanner { return bufio.NewScanner(strings.NewReader(s)) }

func parseFrontmatter(scanner *bufio.Scanner) (Frontmatter, string, bool, error) {
	fm := Frontmatter{}

	// Frontmatter must start on the very first line of the file.
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fm, "", false, err
		}
		return fm, "", false, nil
	}
	firstLine := scanner.Text()
	if strings.TrimSpace(firstLine) != "---" {
		bodyLines := []string{firstLine}
		for scanner.Scan() {
			bodyLines = append(bodyLines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return fm, "", false, err
		}
		body := strings.TrimLeft(strings.Join(bodyLines, "\n"), "\n")
		return fm, body, false, nil
	}

	var bodyLines []string
	inBlock := true
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			inBlock = false
			continue
		}
		if inBlock {
			key, val, ok := splitKV(line)
			if !ok {
				continue
			}
			switch key {
			case "category":
				fm.Category = val
			case "created":
				fm.Created = val
			case "source":
				fm.Source = val
			case "synopsis":
				fm.Synopsis = val
			}
			continue
		}
		bodyLines = append(bodyLines, line)
	}
	if err := scanner.Err(); err != nil {
		return fm, "", false, err
	}
	return fm, strings.TrimLeft(strings.Join(bodyLines, "\n"), "\n"), true, nil
}

func splitKV(line string) (key, val string, ok bool) {
	key, val, ok = strings.Cut(line, ":")
	if !ok {
		return "", "", false
	}
	key = strings.TrimSpace(key)
	val = strings.TrimSpace(val)
	return key, val, key != ""
}

// WriteFrontmatter writes (or overwrites) a markdown file with the given
// frontmatter and body. Writes to a temp file and renames into place so a
// crash mid-write never leaves a partially-written note.
func WriteFrontmatter(path string, fm Frontmatter, body string) (err error) {
	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp %q: %w", tmpPath, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err = fmt.Fprintf(f, "---\ncategory: %s\ncreated: %s\nsource: %s\nsynopsis: %s\n---\n\n%s\n",
		fm.Category, fm.Created, fm.Source, fm.Synopsis, body); err != nil {
		return fmt.Errorf("write temp %q: %w", tmpPath, err)
	}

	if err = os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename %q to %q: %w", tmpPath, path, err)
	}
	return nil
}

// Today returns the current date in the frontmatter's date format.
func Today() string {
	return time.Now().Format("2006-01-02")
}

// NewWithBody creates a fresh parked note with an explicit body. The filename
// is slugified to form the note's filename.
func NewWithBody(cfg *config.Config, filename, synopsis, source, targetCategory, body string) (string, error) {
	cl, ok := cfg.CategoryByName(targetCategory)
	if !ok {
		return "", fmt.Errorf("unknown category %q — valid: %s", targetCategory, strings.Join(cfg.CategoryNames(), ", "))
	}

	slug := slugify(filename)
	if slug == "" {
		slug = "note-" + Today()
	}
	path := filepath.Join(cl.Path, slug+".md")

	if _, err := os.Stat(cl.Path); os.IsNotExist(err) {
		return "", fmt.Errorf("category folder %q does not exist; run `park init` to create it", cl.Path)
	}

	fm := Frontmatter{
		Category: targetCategory,
		Created:  Today(),
		Source:   source,
		Synopsis: synopsis,
	}

	if err := WriteFrontmatter(path, fm, body); err != nil {
		return "", fmt.Errorf("write note %q: %w", path, err)
	}
	return path, nil
}

// IngestFile moves an existing markdown file into the park, rewriting its
// frontmatter. If body is empty, the original file content is preserved;
// otherwise the supplied body is used. The source file is removed after a
// successful write.
func IngestFile(cfg *config.Config, srcPath, filename, synopsis, source, targetCategory, body string) (string, error) {
	cl, ok := cfg.CategoryByName(targetCategory)
	if !ok {
		return "", fmt.Errorf("unknown category %q — valid: %s", targetCategory, strings.Join(cfg.CategoryNames(), ", "))
	}

	srcPath = fs.ExpandPath(srcPath)

	info, err := os.Stat(srcPath)
	if err != nil {
		return "", fmt.Errorf("stat source file %q: %w", srcPath, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("source path %q is a directory", srcPath)
	}

	if body == "" {
		bodyBytes, err := os.ReadFile(srcPath)
		if err != nil {
			return "", fmt.Errorf("read source file %q: %w", srcPath, err)
		}
		body = string(bodyBytes)
	}

	slug := slugify(filename)
	if slug == "" {
		slug = "note-" + Today()
	}
	dstPath := filepath.Join(cl.Path, slug+".md")

	if _, err := os.Stat(cl.Path); os.IsNotExist(err) {
		return "", fmt.Errorf("category folder %q does not exist; run `park init` to create it", cl.Path)
	}

	fm := Frontmatter{
		Category: targetCategory,
		Created:  Today(),
		Source:   source,
		Synopsis: synopsis,
	}

	if err := WriteFrontmatter(dstPath, fm, body); err != nil {
		return "", fmt.Errorf("write ingested note %q: %w", dstPath, err)
	}
	if err := os.Remove(srcPath); err != nil {
		return "", fmt.Errorf("remove source file %q: %w", srcPath, err)
	}
	return dstPath, nil
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// extractHeading scans a markdown body for a single-line H1 heading, skipping
// any leading frontmatter block. If found, it returns the heading text and
// the body with both the frontmatter and the heading removed so the title is
// not duplicated in the rendered note.
func extractHeading(body string) (title, remaining string, ok bool) {
	_, body, _ = ParseFrontmatterString(body)

	lines := strings.Split(body, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if after, ok := strings.CutPrefix(trimmed, "# "); ok {
			title = strings.TrimSpace(after)
			remainingLines := append(lines[:i], lines[i+1:]...)
			remaining = strings.TrimLeft(strings.Join(remainingLines, "\n"), "\n")
			return title, remaining, true
		}
		break
	}
	return "", body, false
}

// ExtractH1 scans a markdown body for a single-line H1 heading and returns
// the heading text. It stops at the first non-empty line that is not an H1.
func ExtractH1(body string) (string, bool) {
	for line := range strings.SplitSeq(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if after, ok := strings.CutPrefix(trimmed, "# "); ok {
			return strings.TrimSpace(after), true
		}
		break
	}
	return "", false
}

// NoteInput captures the raw inputs for creating a parked note.
type NoteInput struct {
	Filename string
	Synopsis string
	Source   string
	Category string // explicit target category; empty means the config default
	Body     string // raw content from stdin or a file read by the caller
	FromFile string // path to the original file when ingesting
}

// NoteOutcome is the result of attempting to create a note headlessly.
// Exactly one of Path or Form is set.
type NoteOutcome struct {
	Path string
	Form *NoteForm
}

// NoteForm holds the starting values for the interactive note form.
type NoteForm struct {
	Filename string
	Synopsis string
	Source   string
	Category string
	Body     string
	FromFile string
}

// AddNote attempts to create a note without user interaction. If required
// metadata is missing, it returns a NoteOutcome with Form populated and no
// error.
func AddNote(cfg *config.Config, in NoteInput) (NoteOutcome, error) {
	target := cfg.DefaultCategory
	if in.Category != "" {
		target = in.Category
	}

	hasInput := in.Body != "" || in.FromFile != ""
	hasMetadata := fieldSet(in.Filename) && fieldSet(in.Synopsis) && fieldSet(in.Source)

	if hasInput && hasMetadata {
		// Caller supplied all required metadata and a body; create the note
		// directly. Strip any frontmatter in the body so it isn't duplicated
		// by WriteFrontmatter; if there is no frontmatter, keep the body as-is.
		body := in.Body
		if _, parsed, hasFM := ParseFrontmatterString(in.Body); hasFM {
			body = parsed
		}
		var path string
		var err error
		if in.FromFile != "" {
			path, err = IngestFile(cfg, in.FromFile, in.Filename, in.Synopsis, in.Source, target, body)
		} else {
			path, err = NewWithBody(cfg, in.Filename, in.Synopsis, in.Source, target, body)
		}
		if err != nil {
			return NoteOutcome{}, err
		}
		return NoteOutcome{Path: path}, nil
	}

	if hasInput {
		return addFromInput(cfg, in, target)
	}

	if hasMetadata {
		path, err := NewWithBody(cfg, in.Filename, in.Synopsis, in.Source, target, "")
		if err != nil {
			return NoteOutcome{}, err
		}
		return NoteOutcome{Path: path}, nil
	}

	return NoteOutcome{Form: &NoteForm{
		Filename: in.Filename,
		Synopsis: in.Synopsis,
		Source:   in.Source,
		Category: target,
		FromFile: in.FromFile,
	}}, nil
}

func addFromInput(cfg *config.Config, in NoteInput, target string) (NoteOutcome, error) {
	fm, body, hasFM := ParseFrontmatterString(in.Body)
	if !hasFM {
		return NoteOutcome{Form: &NoteForm{
			Filename: in.Filename,
			Synopsis: in.Synopsis,
			Source:   in.Source,
			Category: target,
			Body:     in.Body,
			FromFile: in.FromFile,
		}}, nil
	}

	if missing := fm.MissingFields(); len(missing) > 0 {
		return NoteOutcome{}, fmt.Errorf("incomplete frontmatter: missing %s", strings.Join(missing, ", "))
	}
	cl, ok := cfg.CategoryByName(fm.Category)
	if !ok {
		return NoteOutcome{}, fmt.Errorf("unknown category %q in frontmatter — valid: %s", fm.Category, strings.Join(cfg.CategoryNames(), ", "))
	}
	target = cl.Name

	filename := strings.TrimSpace(in.Filename)
	if filename == "" {
		h1, hasH1 := ExtractH1(body)
		if !hasH1 {
			return NoteOutcome{}, fmt.Errorf("missing filename, retry with --filename")
		}
		filename = h1
	}

	var path string
	var err error
	if in.FromFile != "" {
		path, err = IngestFile(cfg, in.FromFile, filename, fm.Synopsis, fm.Source, target, body)
	} else {
		path, err = NewWithBody(cfg, filename, fm.Synopsis, fm.Source, target, body)
	}
	if err != nil {
		return NoteOutcome{}, err
	}
	return NoteOutcome{Path: path}, nil
}

func fieldSet(s string) bool {
	return strings.TrimSpace(s) != ""
}
