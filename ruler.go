// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "strconv"

// DrawRuler frames r with a 1-cell ruler on the terminal's default background,
// so cells a snapshot leaves unpainted show through next to it.
func DrawRuler(c *Canvas, r Rect) {
	if c == nil || r.W < 1 || r.H < 1 {
		return
	}
	dim := Style{FG: ColorIndex(244), BG: ColorReset()}
	hot := Style{FG: ColorIndex(214), BG: ColorReset()}
	// PaintForeground keeps the reset background instead of inheriting the box's.
	put := func(x, y int, text string, style Style) {
		c.PaintForeground(x, y, Cell{Text: text, Style: style})
	}
	for x := 0; x < r.W; x++ {
		text, style := "·", dim
		switch {
		case x == 0 || x == r.W-1:
			text = " "
		case x%10 == 0:
			text, style = strconv.Itoa(x/10%10), hot
		case x%5 == 0:
			text = "┊"
		}
		put(r.X+x, r.Y, text, style)
		put(r.X+x, r.Y+r.H-1, text, style)
	}
	for y := 1; y < r.H-1; y++ {
		style := dim
		if y%5 == 0 {
			style = hot
		}
		put(r.X, r.Y+y, strconv.Itoa(y%10), style)
		put(r.X+r.W-1, r.Y+y, "│", dim)
	}
}
