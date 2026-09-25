package syntax

import "unicode/utf8"

// ByteToRune converts a byte offset to a rune offset, clamping to source bounds.
func ByteToRune(source []byte, offset int) int {
	if offset < 0 {
		offset = 0
	}
	if offset > len(source) {
		offset = len(source)
	}
	count := 0
	for i := 0; i < offset; {
		_, size := utf8.DecodeRune(source[i:offset])
		if size == 1 && source[i] >= utf8.RuneSelf {
			break
		}
		i += size
		count++
	}
	return count
}

// RuneToByte converts a rune offset to a byte offset, clamping to source bounds.
func RuneToByte(source []byte, offset int) int {
	if offset <= 0 {
		return 0
	}
	for i, n := 0, 0; i < len(source); n++ {
		if n == offset {
			return i
		}
		_, size := utf8.DecodeRune(source[i:])
		i += size
	}
	return len(source)
}

// RuneToDisplayCol returns the terminal-cell column at a rune offset.
func RuneToDisplayCol(source []byte, offset int) int {
	if offset <= 0 {
		return 0
	}
	col := 0
	for i, n := 0, 0; i < len(source) && n < offset; n++ {
		r, size := utf8.DecodeRune(source[i:])
		i += size
		// Tabs advance to the next four-cell stop.
		if r == '\t' {
			col += 4 - col%4
		} else if displayWide(r) {
			col += 2
		} else if r >= 0x20 && r != 0x7f {
			col++
		}
	}
	return col
}

func displayWide(r rune) bool {
	return r >= 0x1100 && (r <= 0x115f || r == 0x2329 || r == 0x232a || (r >= 0x2e80 && r <= 0xa4cf) || (r >= 0xac00 && r <= 0xd7a3) || (r >= 0xf900 && r <= 0xfaff) || (r >= 0xfe10 && r <= 0xfe19) || (r >= 0xfe30 && r <= 0xfe6f) || (r >= 0xff00 && r <= 0xff60) || (r >= 0xffe0 && r <= 0xffe6) || (r >= 0x1f300 && r <= 0x1faff) || (r >= 0x20000 && r <= 0x3fffd))
}
