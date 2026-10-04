package loom

import "testing"

func TestBoxGlyphAllArmMasks(t *testing.T) {
	want := []string{" ", "╵", "╶", "└", "╷", "│", "┌", "├", "╴", "┘", "─", "┴", "┐", "┤", "┬", "┼"}
	for mask, glyph := range want {
		if got := BoxGlyph(BoxArms(mask)); got != glyph {
			t.Errorf("BoxGlyph(%04b) = %q, want %q", mask, got, glyph)
		}
		if got := BoxGlyphArms(glyph); got != BoxArms(mask) {
			t.Errorf("BoxGlyphArms(%q) = %04b, want %04b", glyph, got, mask)
		}
	}
}
