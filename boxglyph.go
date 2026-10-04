package loom

// BoxArms identifies the connected sides of a light box-drawing glyph.
type BoxArms uint8

const (
	BoxArmUp BoxArms = 1 << iota
	BoxArmRight
	BoxArmDown
	BoxArmLeft
)

var boxGlyphs = [...]string{" ", "╵", "╶", "└", "╷", "│", "┌", "├", "╴", "┘", "─", "┴", "┐", "┤", "┬", "┼"}

// BoxGlyph returns the light box-drawing glyph represented by arms.
func BoxGlyph(arms BoxArms) string {
	return boxGlyphs[arms&0x0f]
}

// BoxGlyphArms returns the connected sides of a light box-drawing glyph.
func BoxGlyphArms(glyph string) BoxArms {
	switch glyph {
	case "╭":
		return BoxArmRight | BoxArmDown
	case "╮":
		return BoxArmLeft | BoxArmDown
	case "╰":
		return BoxArmRight | BoxArmUp
	case "╯":
		return BoxArmLeft | BoxArmUp
	}
	for arms, candidate := range boxGlyphs {
		if candidate == glyph {
			return BoxArms(arms)
		}
	}
	return 0
}
