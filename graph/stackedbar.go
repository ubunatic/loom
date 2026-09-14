// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"strings"
)

// defaultStackedBarGlyphs are the default per-segment fill glyphs used when
// StackedBarOptions.Glyphs is empty, cycling for bars with more segments
// than glyphs.
var defaultStackedBarGlyphs = []rune("█▓▒░")

// StackedBarOptions configures RenderStackedBar.
type StackedBarOptions struct {
	// Width is the total character width of the bar. Zero uses MaxWidth;
	// negative values clamp to 1.
	Width int
	// Glyphs supplies one fill glyph per segment, cycling if there are more
	// segments than glyphs. Empty uses defaultStackedBarGlyphs.
	Glyphs []rune
	// Left and Right wrap the bar. Empty uses the default "[" and "]"
	// unless NoWrapper is true.
	Left  string
	Right string
	// NoWrapper renders only the segment glyphs, useful for embedding in a
	// larger layout.
	NoWrapper bool
	// ANSI wraps segments in foreground/background ANSI sequences, matching
	// RenderBar's convention.
	ANSI bool
	// BackgroundANSI is the SGR background code applied across the whole
	// bar when ANSI is true. Empty uses DefaultBackgroundANSI. Passing
	// "none" suppresses the background code while allowing ForegroundANSI.
	BackgroundANSI string
	// ForegroundANSI supplies one SGR foreground code per segment, cycling
	// if there are more segments than codes. Nil leaves segments styled
	// with BackgroundANSI only (if any).
	ForegroundANSI []string
}

// RenderStackedBar renders values as a single-row proportional bar: width
// cells divided among values by magnitude using the largest-remainder
// method, so segment widths always sum to exactly width. NaN and negative
// values are treated as zero. All-zero (or empty) input renders width blank
// cells.
//
// This is a one-dimensional summary, not a 2D treemap: no per-segment
// labels fit in a single row. Pair it with a legend, or use RenderTreemap
// for a multi-row layout with boxes and labels.
func RenderStackedBar(values []float64, opts StackedBarOptions) string {
	width := optionWidth(opts.Width)

	glyphs := opts.Glyphs
	if len(glyphs) == 0 {
		glyphs = defaultStackedBarGlyphs
	}
	glyphs = oneCellGlyphs(glyphs, defaultStackedBarGlyphs)

	left, right := opts.Left, opts.Right
	if !opts.NoWrapper && left == "" && right == "" {
		left = "["
		right = "]"
	}

	widths := allocateCells(values, width)

	var glyphOut string
	if len(widths) == 0 {
		glyphOut = strings.Repeat(" ", width)
	} else {
		var bgCode string
		if opts.ANSI {
			bgCode = resolveBackgroundANSI(opts.BackgroundANSI)
		}
		var b strings.Builder
		for i, w := range widths {
			if w == 0 {
				continue
			}
			segment := strings.Repeat(string(glyphs[i%len(glyphs)]), w)
			if !opts.ANSI {
				b.WriteString(segment)
				continue
			}
			var fg string
			if len(opts.ForegroundANSI) > 0 {
				fg = opts.ForegroundANSI[i%len(opts.ForegroundANSI)]
			}
			code := bgCode
			if fg != "" {
				if code != "" {
					code += ";" + fg
				} else {
					code = fg
				}
			}
			if code == "" {
				b.WriteString(segment)
				continue
			}
			b.WriteString("\x1b[" + code + "m" + segment + "\x1b[0m")
		}
		glyphOut = b.String()
	}

	var b strings.Builder
	b.WriteString(left)
	b.WriteString(glyphOut)
	b.WriteString(right)
	return b.String()
}
