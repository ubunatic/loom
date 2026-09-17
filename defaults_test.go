package loom

import (
	"strings"
	"testing"
)

func TestScrollbarSpecGlyphsFitOneCell(t *testing.T) {
	if err := SpeccedDefaults.Scrollbar.validate(); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"foreground_char", "background_char"} {
		for _, glyph := range []string{"", " ", "AB", "界", "\x1b", "🖼️", "A\u200d"} {
			defs := ScrollbarDefaults{ForegroundChar: "▓", BackgroundChar: "░"}
			if field == "foreground_char" {
				defs.ForegroundChar = glyph
			} else {
				defs.BackgroundChar = glyph
			}
			if err := defs.validate(); err == nil || !strings.Contains(err.Error(), "scrollbar."+field) {
				t.Fatalf("accepted invalid scrollbar %s glyph %q: %v", field, glyph, err)
			}
		}
	}
}
