// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"fmt"
	"math"
	"strings"

	"codeberg.org/ubunatic/loom/measure"
)

// DefaultBackgroundANSI is the SGR code RenderBar and RenderSparkline fall
// back to when ANSI is true and the caller leaves BackgroundANSI empty.
// It is an immutable constant ("100", bright-black) ensuring zero mutable
// global state.
const DefaultBackgroundANSI = "100"

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

// SparklinePresentation selects how a pair of time-series samples is drawn.
type SparklinePresentation uint8

const (
	// SparklineBlocks reduces each sample pair to its newer value and draws it
	// with the configured lower-block glyph scale.
	SparklineBlocks SparklinePresentation = iota
	// SparklineBraille draws the older sample in the left Braille dot column
	// and the newer sample in the right column. Each column is quantized to
	// four bottom-aligned dots (left: 7,3,2,1; right: 8,6,5,4).
	SparklineBraille
)

// SparklineOptions configures RenderSparkline. The zero value renders a
// no-ANSI, relative-scale block sparkline padded to MaxWidth cells.
type SparklineOptions struct {
	// Width is the exact number of terminal cells to render. Each cell
	// consumes two chronological samples; zero uses MaxWidth and negative
	// values clamp to one cell.
	Width int
	// Presentation selects block or Braille rendering. The zero value is
	// SparklineBlocks.
	Presentation SparklinePresentation
	// Min and Max define the scale when FixedRange is true.
	Min float64
	Max float64
	// FixedRange uses Min and Max instead of deriving a range from the
	// rendered values. Use this for percent or other absolute-scale charts.
	FixedRange bool
	// ANSI wraps the sparkline in a background ANSI sequence. Plain output is
	// the default so callers can opt in only for terminal contexts.
	ANSI bool
	// BackgroundANSI is the SGR code used when ANSI is true. Empty uses
	// DefaultBackgroundANSI ("100"). Passing "none" suppresses the background
	// code while allowing ForegroundANSI to be applied.
	BackgroundANSI string
	// ForegroundANSI returns the SGR foreground code for a rendered cell's
	// representative value. Nil leaves the chart monochrome. The callback is
	// deliberately spec-agnostic so this dependency-free package never needs
	// to know an application's palette.
	ForegroundANSI func(value float64) string
	// Glyphs supplies low-to-high sparkline frames. Empty uses the package
	// default (percentSparkChars).
	Glyphs []rune
	// NoPad disables automatic padding to Width cells. When true, empty
	// inputs return "" and short inputs render fewer cells.
	NoPad bool
	// PadRune specifies an explicit rune to pad short or empty histories.
	// If zero, the idle glyph appropriate for Presentation is used.
	PadRune rune
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

// RenderSparkline renders up to Width terminal cells from the newest 2*Width
// samples. Samples are oldest first. An odd-length window keeps its oldest
// singleton as the first cell (duplicated across Braille columns); all later
// cells contain adjacent pairs. Block cells deliberately use the newer value
// of a pair so the current/latest observation is always represented. Relative
// mode derives one range across the complete retained sample window.
//
// Unless opts.NoPad is true, RenderSparkline guarantees exact Width output
// even when values is empty or has fewer samples than 2*Width. Short histories
// are left-padded with idle glyphs.
func RenderSparkline(values []float64, opts SparklineOptions) string {
	width := optionWidth(opts.Width)
	if len(values) == 0 && opts.NoPad {
		return ""
	}
	resolution := width * 2
	if len(values) > resolution {
		values = values[len(values)-resolution:]
	}

	minimum, maximum, flat := sparkRange(values, opts)

	var spark []rune
	if len(values) > 0 {
		spark = make([]rune, 0, (len(values)+1)/2)
		start := 0
		if len(values)%2 != 0 {
			spark = append(spark, sparkCell(values[0], values[0], minimum, maximum, flat, opts))
			start = 1
		}
		for i := start; i < len(values); i += 2 {
			spark = append(spark, sparkCell(values[i], values[i+1], minimum, maximum, flat, opts))
		}
	}

	var padCount int
	if !opts.NoPad && len(spark) < width {
		padCount = width - len(spark)
		padRune := opts.PadRune
		if padRune == 0 {
			padRune = idleSparkGlyph(minimum, maximum, flat, opts)
		}
		padRune = oneCellGlyph(padRune, ' ')
		padded := make([]rune, width)
		for i := 0; i < padCount; i++ {
			padded[i] = padRune
		}
		copy(padded[padCount:], spark)
		spark = padded
	}

	out := string(spark)
	if !opts.ANSI {
		return out
	}
	code := resolveBackgroundANSI(opts.BackgroundANSI)
	if code == "" && opts.ForegroundANSI == nil {
		return out
	}
	if opts.ForegroundANSI == nil {
		return "\x1b[" + code + "m" + out + "\x1b[0m"
	}
	var styled strings.Builder
	for i, glyph := range spark {
		var value float64
		if i < padCount || len(values) == 0 {
			value = minimum
		} else {
			value = sparkCellValue(values, i-padCount, opts)
		}
		foreground := opts.ForegroundANSI(value)
		if code == "" && foreground == "" {
			styled.WriteRune(glyph)
			continue
		}
		styled.WriteString("\x1b[")
		if code != "" {
			styled.WriteString(code)
			if foreground != "" {
				styled.WriteByte(';')
				styled.WriteString(foreground)
			}
		} else {
			styled.WriteString(foreground)
		}
		styled.WriteString("m")
		styled.WriteRune(glyph)
		styled.WriteString("\x1b[0m")
	}
	return styled.String()
}

// idleSparkGlyph returns the appropriate zero/idle glyph for the given sparkline configuration.
func idleSparkGlyph(minimum, maximum float64, flat bool, opts SparklineOptions) rune {
	if opts.Presentation == SparklineBraille {
		if opts.FixedRange {
			return brailleGlyph(opts.Min, opts.Min, opts.Min, opts.Max, false)
		}
		return brailleGlyph(minimum, minimum, minimum, maximum, flat)
	}
	glyphs := opts.Glyphs
	if len(glyphs) == 0 {
		glyphs = percentSparkChars
	}
	return glyphs[0]
}

func resolveBackgroundANSI(code string) string {
	switch code {
	case "":
		return DefaultBackgroundANSI
	case "none", "-":
		return ""
	default:
		return code
	}
}

// sparkCellValue returns the value that owns a cell's foreground. Blocks use
// their newer sample; Braille uses the taller of its two columns so heat color
// follows the visually dominant bar in the cell.
func sparkCellValue(values []float64, cell int, opts SparklineOptions) float64 {
	index := cell * 2
	if len(values)%2 != 0 {
		if cell == 0 {
			return values[0]
		}
		index = cell*2 - 1
	}
	older, newer := values[index], values[index+1]
	if opts.Presentation != SparklineBraille || !isFinite(older) {
		return newer
	}
	if !isFinite(newer) || older > newer {
		return older
	}
	return newer
}

func sparkCell(older, newer, minimum, maximum float64, flat bool, opts SparklineOptions) rune {
	if opts.Presentation != SparklineBraille {
		return sparkGlyphWithGlyphs(newer, minimum, maximum, flat, opts.Glyphs)
	}
	return brailleGlyph(older, newer, minimum, maximum, flat)
}

// brailleGlyph maps each sample to a four-dot, bottom-aligned Braille column.
// The Unicode Braille bit positions are left 1,2,3,7 and right 4,5,6,8;
// filling from bottom to top makes zero U+2800 and the maximum U+28FF.
func brailleGlyph(older, newer, minimum, maximum float64, flat bool) rune {
	left := brailleColumnLevel(older, minimum, maximum, flat)
	right := brailleColumnLevel(newer, minimum, maximum, flat)
	var bits rune
	leftBits := [...]rune{0, 1 << 6, (1 << 6) | (1 << 2), (1 << 6) | (1 << 2) | (1 << 1), (1 << 6) | (1 << 2) | (1 << 1) | 1}
	rightBits := [...]rune{0, 1 << 7, (1 << 7) | (1 << 5), (1 << 7) | (1 << 5) | (1 << 4), (1 << 7) | (1 << 5) | (1 << 4) | (1 << 3)}
	bits = leftBits[left] | rightBits[right]
	return 0x2800 + bits
}

func brailleColumnLevel(value, minimum, maximum float64, flat bool) int {
	if flat {
		return 2
	}
	if math.IsNaN(value) || math.IsInf(value, -1) || value < minimum {
		value = minimum
	}
	if math.IsInf(value, 1) || value > maximum {
		value = maximum
	}
	// Match btop's visible baseline: normalized zero and invalid values use
	// the lowest dot rather than U+2800. The remaining rows divide the range
	// into (0, 25], (25, 50], (50, 75], and (75, 100].
	level := int(math.Ceil(((value - minimum) / (maximum - minimum)) * 4))
	if level < 1 {
		return 1
	}
	if level > 4 {
		return 4
	}
	return level
}

// RenderPercentSparkline renders values on a fixed 0-100 scale.
func RenderPercentSparkline(values []float64, opts SparklineOptions) string {
	opts.FixedRange = true
	opts.Min = 0
	opts.Max = 100
	return RenderSparkline(values, opts)
}

// FormatPercent formats a percent value with clamping and a trailing percent
// sign. It is useful for callers that want a label next to RenderBar output.
func FormatPercent(percent float64, precision int) string {
	if precision < 0 {
		precision = 0
	}
	if math.IsNaN(percent) || percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return fmt.Sprintf("%.*f%%", precision, percent)
}

func optionWidth(width int) int {
	switch {
	case width == 0:
		return MaxWidth
	case width < 0:
		return 1
	default:
		return width
	}
}

func oneCellGlyph(glyph, fallback rune) rune {
	if measure.RuneWidth(glyph) != 1 {
		return fallback
	}
	return glyph
}

func oneCellGlyphs(glyphs, fallback []rune) []rune {
	result := make([]rune, len(glyphs))
	for i, glyph := range glyphs {
		result[i] = oneCellGlyph(glyph, fallback[min(i, len(fallback)-1)])
	}
	return result
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

func normalizedPercent(value, minimum, maximum float64) float64 {
	if math.IsNaN(value) {
		value = minimum
	}
	if value < minimum {
		value = minimum
	}
	if value > maximum {
		value = maximum
	}
	return ((value - minimum) / (maximum - minimum)) * 100
}

func sparkRange(values []float64, opts SparklineOptions) (float64, float64, bool) {
	if opts.FixedRange {
		if isFinite(opts.Min) && isFinite(opts.Max) && opts.Max > opts.Min {
			return opts.Min, opts.Max, false
		}
		return 0, 0, true
	}

	var minimum, maximum float64
	haveFinite := false
	for _, value := range values {
		if !isFinite(value) {
			continue
		}
		if !haveFinite {
			minimum = value
			maximum = value
			haveFinite = true
			continue
		}
		if value < minimum {
			minimum = value
		}
		if value > maximum {
			maximum = value
		}
	}
	if !haveFinite || maximum == minimum {
		return 0, 0, true
	}
	return minimum, maximum, false
}

func sparkGlyphWithGlyphs(value, minimum, maximum float64, flat bool, glyphs []rune) rune {
	if len(glyphs) == 0 {
		glyphs = percentSparkChars
	}
	if flat {
		return glyphs[(len(glyphs)-1)/2]
	}
	if math.IsNaN(value) || value < minimum {
		value = minimum
	}
	if value > maximum {
		value = maximum
	}
	idx := int(((value - minimum) / (maximum - minimum)) * float64(len(glyphs)-1))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(glyphs) {
		idx = len(glyphs) - 1
	}
	return glyphs[idx]
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
