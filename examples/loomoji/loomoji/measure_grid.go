// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomoji

import (
	"fmt"
	"strconv"
	"strings"

	"codeberg.org/ubunatic/loom/measure"
)

// RenderEmojiGrid renders a formatted ascii grid of emojis grouped by VTE visual width.
func RenderEmojiGrid(allGlyphs []string, widthFilter string) (string, error) {
	const cols = 10
	glyphsByWidth := make(map[int][]string)
	for _, g := range allGlyphs {
		w := measure.StringWidth(g)
		glyphsByWidth[w] = append(glyphsByWidth[w], g)
	}

	widths := []int{2, 3, 4, 1}
	if widthFilter != "" {
		w, err := strconv.Atoi(widthFilter)
		if err != nil || w < 1 || w > 4 {
			return "", fmt.Errorf("loomoji debug: -W/--width must be 1, 2, 3, or 4")
		}
		widths = []int{w}
	}

	var sections []string
	for _, w := range widths {
		if list, ok := glyphsByWidth[w]; ok && len(list) > 0 {
			sections = append(sections, FormatEmojiGrid(list, w, cols))
		}
	}
	return strings.Join(sections, "\n"), nil
}

// FormatEmojiGrid renders a single width section grid.
func FormatEmojiGrid(glyphs []string, width int, cols int) string {
	if len(glyphs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("VTE Width %d:\n", width))

	// Column header: 3 spaces matching row prefix "a |"
	b.WriteString("   ")
	for c := 1; c <= cols; c++ {
		switch width {
		case 1:
			if c < 10 {
				b.WriteString(fmt.Sprintf("%d ", c))
			} else {
				b.WriteString(fmt.Sprintf("%d", c))
			}
		case 2:
			b.WriteString(fmt.Sprintf(" %-2d", c))
		case 3:
			b.WriteString(fmt.Sprintf(" %-3d", c))
		case 4:
			b.WriteString(fmt.Sprintf(" %-4d", c))
		default:
			b.WriteString(fmt.Sprintf(" %-*d", width, c))
		}
	}
	b.WriteString("\n")

	// Grid rows
	for r := 0; r*cols < len(glyphs); r++ {
		b.WriteString(fmt.Sprintf("%-2s|", rowLabel(r)))
		for c := 0; c < cols; c++ {
			idx := r*cols + c
			if idx < len(glyphs) {
				glyph := glyphs[idx]
				mode := EvaluateVTEMode(glyph)
				rendered := ApplyRenderMode(glyph, mode)
				b.WriteString(rendered)
				b.WriteString("|")
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func rowLabel(index int) string {
	if index < 26 {
		return string(rune('a' + index))
	}
	first := (index / 26) - 1
	second := index % 26
	return string([]rune{rune('a' + first), rune('a' + second)})
}
