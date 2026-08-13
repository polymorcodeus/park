package theme

import (
	"testing"
)

func TestCurrentGlyphsDefault(t *testing.T) {
	gs := CurrentGlyphs()
	if gs.ErrorBullet == "" {
		t.Error("default glyph set has empty error bullet")
	}
}

func TestCurrentGlyphsPlain(t *testing.T) {
	t.Setenv("PARK_PLAIN", "1")
	gs := CurrentGlyphs()
	if gs.ErrorBullet != ASCII.ErrorBullet {
		t.Errorf("PARK_PLAIN error bullet = %q, want %q", gs.ErrorBullet, ASCII.ErrorBullet)
	}
}
