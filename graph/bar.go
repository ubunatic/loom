// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package graph provides dependency-free renderers for single-row,
// fixed-width terminal graphs: determinate fill bars and rolling sparklines.
//
// The rendering primitives are adapted from Harnez's internal/rograph package
// (commit 01e59b331c9d85699e55fbbb946c579e19901083).
package graph

import (
	"strings"
)

// MaxWidth is the default maximum render width, in characters/glyphs, for
// single-row graph bars and sparklines.
const MaxWidth = 10

// RenderProgressBar generates an ANSI/Unicode progress bar of the given
// character width: a "[filled empty]" bar using '█' for the filled portion
// and '░' for the empty portion, scaled to usedPercent (clamped to [0, 100]).
// The boundary character where fill transitions from filled to empty renders
// at eighth-block precision (▏▎▍▌▋▊▉█) rather than snapping to fully filled
// or empty.
//
// If width < 1, it clamps to 1.
func RenderProgressBar(usedPercent float64, width int) string {
	if width < 1 {
		width = 1
	}
	return RenderBar(usedPercent, BarOptions{Width: width, SubChar: true})
}

// BarOptions configures RenderBar. The zero value renders a 10-character
// 0-100 bar with the package's standard filled and empty glyphs.
type BarOptions struct {
	// Width is the number of graph glyphs inside the optional wrappers.
	// Zero uses MaxWidth; negative values clamp to 1.
	Width int
	// Min and Max define the numeric range. The zero value means 0-100.
	Min float64
	Max float64
	// Fill and Empty override the default terminal bar glyphs.
	Fill  rune
	Empty rune
	// Left and Right wrap the graph. Empty uses the default "[" and "]"
	// unless NoWrapper is true.
	Left  string
	Right string
	// NoWrapper renders only the graph glyphs, useful for stripe-style bars.
	NoWrapper bool
	// IncludePercent appends a formatted percent label after the bar.
	IncludePercent bool
	// PercentPrecision controls the optional percent label. Negative values
	// clamp to 0.
	PercentPrecision int
	// SubChar renders the single boundary character (where fill transitions
	// from filled to empty) at sub-character precision instead of snapping
	// it to fully filled or fully empty.
	SubChar bool
	// SubCharacterGlyphs supplies the ascending partial-fill glyphs. Empty
	// uses rograph's legacy eighth-block sequence; callers with a visual spec
	// should pass their resolved sequence.
	SubCharacterGlyphs []rune
	// ANSI wraps the rendered bar (including any percent label) in a
	// background ANSI sequence, matching RenderSparkline's convention.
	// Plain output is the default so callers can opt in only for terminal
	// contexts.
	ANSI bool
	// BackgroundANSI is the SGR code used when ANSI is true. Empty uses
	// DefaultBackgroundANSI ("100"). Passing "none" suppresses the background
	// code while allowing ForegroundANSI to be applied.
	BackgroundANSI string
	// ForegroundANSI is an optional SGR foreground code applied to the graph
	// glyphs only. Callers retain control of any nearby labels.
	ForegroundANSI string
}

// eighthBlockGlyphs are the horizontal eighth-block glyphs used for a
// sub-character fill boundary, 1/8 through 7/8 width.
var eighthBlockGlyphs = []rune("▏▎▍▌▋▊▉")

// subCharacterFill renders width default-glyph characters with the single
// boundary character rendered at sub-character precision rather than snapped
// to fully filled or fully empty. pct must already be clamped to [0, 100].
func subCharacterFill(pct float64, width int, fill, emptyRune rune, partial []rune) string {
	subdivisions := len(partial) + 1
	totalSubunits := int(float64(width*subdivisions) * (pct / 100))
	if totalSubunits < 0 {
		totalSubunits = 0
	}
	maxSubunits := width * subdivisions
	if totalSubunits > maxSubunits {
		totalSubunits = maxSubunits
	}

	fullChars := totalSubunits / subdivisions
	remainder := totalSubunits % subdivisions
	if fullChars >= width {
		fullChars = width
		remainder = 0
	}

	var b strings.Builder
	b.WriteString(strings.Repeat(string(fill), fullChars))
	if fullChars < width {
		if remainder == 0 {
			b.WriteRune(emptyRune)
		} else {
			b.WriteRune(partial[remainder-1])
		}
		b.WriteString(strings.Repeat(string(emptyRune), width-fullChars-1))
	}
	return b.String()
}

// RenderBar renders a single-value terminal bar or stripe. Values outside the
// configured range are clamped, NaN values render as the minimum, and inverted
// or empty ranges fall back to the zero-value 0-100 range.
func RenderBar(value float64, opts BarOptions) string {
	width := optionWidth(opts.Width)
	minimum, maximum := barRange(opts.Min, opts.Max)
	pct := normalizedPercent(value, minimum, maximum)

	fill := opts.Fill
	if fill == 0 {
		fill = '█'
	}
	fill = oneCellGlyph(fill, '█')
	empty := opts.Empty
	if empty == 0 {
		empty = '░'
	}
	empty = oneCellGlyph(empty, '░')
	left, right := opts.Left, opts.Right
	if !opts.NoWrapper && left == "" && right == "" {
		left = "["
		right = "]"
	}

	var glyphs string
	if opts.SubChar && (len(opts.SubCharacterGlyphs) > 0 || (fill == '█' && empty == '░')) {
		partial := opts.SubCharacterGlyphs
		if len(partial) == 0 {
			partial = eighthBlockGlyphs
		}
		partial = oneCellGlyphs(partial, eighthBlockGlyphs)
		emptyRune := empty
		if opts.ANSI {
			// The background wrap below already covers the whole glyph
			// run, so a flat space (no ink) reads as pure panel-bg here
			// instead of '░''s own stipple pattern layering a third tone
			// on top of it.
			emptyRune = ' '
		}
		glyphs = subCharacterFill(pct, width, fill, emptyRune, partial)
	} else {
		filledCount := int(float64(width) * (pct / 100))
		if filledCount < 0 {
			filledCount = 0
		}
		if filledCount > width {
			filledCount = width
		}
		glyphs = strings.Repeat(string(fill), filledCount) + strings.Repeat(string(empty), width-filledCount)
	}

	glyphOut := glyphs
	if opts.ANSI {
		code := resolveBackgroundANSI(opts.BackgroundANSI)
		if code != "" {
			if opts.ForegroundANSI != "" {
				code += ";" + opts.ForegroundANSI
			}
			glyphOut = "\x1b[" + code + "m" + glyphs + "\x1b[0m"
		} else if opts.ForegroundANSI != "" {
			glyphOut = "\x1b[" + opts.ForegroundANSI + "m" + glyphs + "\x1b[0m"
		}
	}

	var b strings.Builder
	b.WriteString(left)
	b.WriteString(glyphOut)
	b.WriteString(right)
	if opts.IncludePercent {
		b.WriteByte(' ')
		b.WriteString(FormatPercent(pct, opts.PercentPrecision))
	}

	return b.String()
}

func barRange(minimum, maximum float64) (float64, float64) {
	if minimum == 0 && maximum == 0 {
		return 0, 100
	}
	if !isFinite(minimum) || !isFinite(maximum) || maximum <= minimum {
		return 0, 100
	}
	return minimum, maximum
}
