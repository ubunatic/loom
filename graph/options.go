// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"fmt"
	"math"

	"ubunatic.com/loom/measure"
)

// DefaultBackgroundANSI is the default SGR background code used when ANSI is
// true and the caller leaves BackgroundANSI empty.
const DefaultBackgroundANSI = "100"

func init() {
	if graphDefaults.DefaultBackgroundANSI != DefaultBackgroundANSI {
		panic("graph: spec default_background_ansi differs from exported constant")
	}
	if graphDefaults.Bar.MaxWidth != MaxWidth {
		panic("graph: spec graph.bar.max_width differs from exported constant")
	}
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
		return graphDefaults.Bar.MaxWidth
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

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
