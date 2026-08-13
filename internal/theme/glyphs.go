package theme

import "os"

// Glyphs holds the terminal glyphs used in styled CLI output and the TUI.
// Two sets are provided: NerdFont for terminals with a Nerd Font installed,
// and ASCII for plain terminals. The active set is selected by the
// PARK_PLAIN environment variable; when non-empty, ASCII glyphs are used.
type Glyphs struct {
	ErrorBullet string
	SubmitLeft  string
	SubmitRight string
}

var (
	// NerdFont uses private-use-area glyphs. These render correctly only when
	// the terminal font includes Nerd Font symbols.
	NerdFont = Glyphs{
		ErrorBullet: "\U000f0bf7",
		SubmitLeft:  "\U000f013d ",
		SubmitRight: " \U000f013e",
	}

	// ASCII uses plain characters so output is readable on any terminal.
	ASCII = Glyphs{
		ErrorBullet: "!",
		SubmitLeft:  "[ ",
		SubmitRight: " ]",
	}
)

// CurrentGlyphs returns the glyph set selected by the environment.
func CurrentGlyphs() Glyphs {
	if os.Getenv("PARK_PLAIN") != "" {
		return ASCII
	}
	return NerdFont
}
