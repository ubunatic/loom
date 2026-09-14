package loom

import "testing"

func TestScrollbarSpecGlyphsFitOneCell(t *testing.T) {
	if err := SpeccedDefaults.Scrollbar.validate(); err != nil {
		t.Fatal(err)
	}
	for _, glyph := range []string{"", " ", "AB", "界", "\x1b"} {
		defs := ScrollbarDefaults{ForegroundChar: glyph, BackgroundChar: "░"}
		if err := defs.validate(); err == nil {
			t.Fatalf("accepted invalid scrollbar glyph %q", glyph)
		}
	}
}
