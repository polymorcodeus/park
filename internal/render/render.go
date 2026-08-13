// Package render handles glamour-based rendering of parked notes.
package render

import (
	"fmt"
	"io"
	"sync"

	"charm.land/glamour/v2"

	"github.com/polymorcodeus/park/internal/note"
)

var (
	renderer     *glamour.TermRenderer
	rendererOnce sync.Once
	rendererErr  error
)

func termRenderer() (*glamour.TermRenderer, error) {
	rendererOnce.Do(func() {
		renderer, rendererErr = glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(100),
		)
	})
	if rendererErr != nil {
		return nil, rendererErr
	}
	return renderer, nil
}

// ShowFile renders a parked note's frontmatter summary + body to w via
// glamour: the "look deeper" step after the synopsis in the list view
// earned a second look.
func ShowFile(path string, w io.Writer) error {
	n, err := note.Parse(path)
	if err != nil {
		return fmt.Errorf("show %q: %w", path, err)
	}

	header := fmt.Sprintf(
		"**category:** %s &nbsp;&nbsp; **created:** %s &nbsp;&nbsp; **source:** %s\n\n> %s\n\n---\n\n",
		n.Category, n.Created, n.Source, n.Synopsis,
	)

	renderer, err := termRenderer()
	if err != nil {
		return fmt.Errorf("create glamour renderer: %w", err)
	}

	out, err := renderer.Render(header + n.Body)
	if err != nil {
		return fmt.Errorf("render %q: %w", path, err)
	}
	if _, err := fmt.Fprint(w, out); err != nil {
		return fmt.Errorf("write rendered output: %w", err)
	}
	return nil
}
