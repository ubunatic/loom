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

func TestBoxGlyphArmsIncludesRoundedCorners(t *testing.T) {
	want := map[string]BoxArms{
		"╭": BoxArmRight | BoxArmDown,
		"╮": BoxArmLeft | BoxArmDown,
		"╰": BoxArmRight | BoxArmUp,
		"╯": BoxArmLeft | BoxArmUp,
	}
	for glyph, arms := range want {
		if got := BoxGlyphArms(glyph); got != arms {
			t.Errorf("BoxGlyphArms(%q) = %04b, want %04b", glyph, got, arms)
		}
	}
	if got := boxGlyphForStroke("╭", BoxArmRight|BoxArmDown); got != "╭" {
		t.Errorf("unchanged rounded corner = %q, want ╭", got)
	}
	if got := boxGlyphForStroke("╭", BoxArmRight|BoxArmDown|BoxArmLeft); got != "┬" {
		t.Errorf("rounded corner with added arm = %q, want sharp ┬", got)
	}
}
