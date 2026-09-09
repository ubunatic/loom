// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

// percentSparkChars are the 8 sparkline glyphs used for a 0-100% value's
// height, low to high (▁ = idle/0%, █ = 100%). Unlike a relative
// (per-series min/max) sparkline, this uses an absolute scale: a flat history
// at 0% renders idle, not maxed out.
var percentSparkChars = []rune("▁▂▃▄▅▆▇█")

// PercentSparkline renders one glyph per value in pcts (each 0-100) on a
// rolling timeline, padded to exactly maxWidth characters with the idle
// glyph ('▁').
//
// If len(pcts) > maxWidth, only the most recent maxWidth values are rendered.
// If len(pcts) < maxWidth, the sparkline is left-padded with the idle glyph
// ('▁') so that the output glyph count always equals maxWidth.
//
// maxWidth < 1 clamps up to 1. The rendered glyphs are wrapped in a muted
// grey background ANSI sequence (SGR 100, bright-black) matching Harnez.
func PercentSparkline(pcts []float64, maxWidth int) string {
	if maxWidth < 1 {
		maxWidth = 1
	}
	if len(pcts) > maxWidth {
		pcts = pcts[len(pcts)-maxWidth:]
	}
	spark := make([]rune, maxWidth)
	padCount := maxWidth - len(pcts)
	for i := 0; i < padCount; i++ {
		spark[i] = percentSparkChars[0]
	}
	for i, p := range pcts {
		idx := int(p / 100 * float64(len(percentSparkChars)))
		if idx >= len(percentSparkChars) {
			idx = len(percentSparkChars) - 1
		}
		if idx < 0 {
			idx = 0
		}
		spark[padCount+i] = percentSparkChars[idx]
	}
	return "\x1b[" + DefaultBackgroundANSI + "m" + string(spark) + "\x1b[0m"
}
