// Package render handles glamour-based rendering of parked notes.
package render

import (
	"fmt"
	"io"

	"charm.land/glamour/v2"

	"github.com/polymorcodeus/park/internal/note"
)

// ShowFile renders a parked note's frontmatter summary + body to w via
// glamour — the "look deeper" step after the synopsis in the list view
// earned a second look.
func ShowFile(path string, w io.Writer) error {
	fm, body, err := note.ParseFrontmatter(path)
	if err != nil {
		return fmt.Errorf("show %q: %w", path, err)
	}

	header := fmt.Sprintf(
		"**category:** %s &nbsp;&nbsp; **created:** %s &nbsp;&nbsp; **source:** %s\n\n> %s\n\n---\n\n",
		fm.Category, fm.Created, fm.Source, fm.Synopsis,
	)

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return fmt.Errorf("create glamour renderer: %w", err)
	}

	out, err := renderer.Render(header + body)
	if err != nil {
		return fmt.Errorf("render %q: %w", path, err)
	}
	_, err = fmt.Fprint(w, out)
	return err
}
