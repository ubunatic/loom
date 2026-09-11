// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"strings"

	"codeberg.org/ubunatic/loom/measure"
)

// BrailleSubCharacterGlyphs provides the partial-fill braille glyph for half-cell width.
// Left column (4 dots): ⡇ (U+2847). Full cell (8 dots): ⣿ (U+28FF).
var BrailleSubCharacterGlyphs = []rune{'⡇'}

// BracketedBarOptions configures RenderBracketedBar.
type BracketedBarOptions struct {
	// Width is the inner character width of the bar (excluding brackets).
	// Clamped to at least 1. Zero defaults to 32.
	Width int
	// Left and Right bracket strings. Default to "[" and "]".
	Left  string
	Right string
	// Fill and Empty runes for determinate progress. Default to '⣿' and ' '.
	Fill  rune
	Empty rune
	// Pattern overrides determinate fill with a fixed repeating string (e.g. ":").
	Pattern string
	// SubChar enables partial-cell braille boundary rendering.
	SubChar bool
	// SubCharacterGlyphs overrides the partial glyphs for sub-character rendering.
	SubCharacterGlyphs []rune
	// ANSI wraps the bar in styled escape sequences.
	ANSI           bool
	ForegroundANSI string
	BackgroundANSI string
}

// RenderBracketedBar renders a bracketed progress or pattern bar.
// When opts.Pattern is set, it fills the inner width with the repeated pattern.
// Otherwise, it renders value (clamped to [0, 100]) using the configured fill/empty glyphs.
func RenderBracketedBar(value float64, opts BracketedBarOptions) string {
	width := opts.Width
	if width <= 0 {
		width = 32
	}
	left := opts.Left
	if left == "" && opts.Right == "" {
		left = "["
		right := opts.Right
		if right == "" {
			right = "]"
		}
		opts.Right = right
	} else if left == "" {
		left = "["
	}
	right := opts.Right
	if right == "" {
		right = "]"
	}

	var inner string
	if opts.Pattern != "" {
		patRunes := []rune(opts.Pattern)
		if len(patRunes) == 0 {
			inner = strings.Repeat(" ", width)
		} else {
			var b strings.Builder
			for i := 0; i < width; i++ {
				b.WriteRune(patRunes[i%len(patRunes)])
			}
			inner = b.String()
		}
	} else {
		fill := opts.Fill
		if fill == 0 {
			fill = '⣿'
		}
		empty := opts.Empty
		if empty == 0 {
			empty = ' '
		}
		subGlyphs := opts.SubCharacterGlyphs
		if len(subGlyphs) == 0 && opts.SubChar {
			subGlyphs = BrailleSubCharacterGlyphs
		}

		barOpts := BarOptions{
			Width:              width,
			Fill:               fill,
			Empty:              empty,
			NoWrapper:          true,
			SubChar:            opts.SubChar,
			SubCharacterGlyphs: subGlyphs,
		}
		inner = RenderBar(value, barOpts)
	}

	glyphOut := inner
	if opts.ANSI {
		code := resolveBackgroundANSI(opts.BackgroundANSI)
		if code != "" {
			if opts.ForegroundANSI != "" {
				code += ";" + opts.ForegroundANSI
			}
			glyphOut = "\x1b[" + code + "m" + inner + "\x1b[0m"
		} else if opts.ForegroundANSI != "" {
			glyphOut = "\x1b[" + opts.ForegroundANSI + "m" + inner + "\x1b[0m"
		}
	}

	return left + glyphOut + right
}

// BracketedBarWidth returns the total visual column width of a bracketed bar.
func BracketedBarWidth(opts BracketedBarOptions) int {
	width := opts.Width
	if width <= 0 {
		width = 32
	}
	left := opts.Left
	if left == "" {
		left = "["
	}
	right := opts.Right
	if right == "" {
		right = "]"
	}
	return measure.StringWidth(left) + width + measure.StringWidth(right)
}
