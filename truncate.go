// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "codeberg.org/ubunatic/loom/measure"

// TruncateText fits text to a terminal-cell budget, adding the caller's marker
// only when truncation is needed. It follows Canvas's supported text policy:
// terminal instructions are stripped, combining marks stay with their base,
// and wide glyphs are never split. Pass styles separately through Style.
// A marker wider than the budget is itself clipped, without recursive markers.
func TruncateText(text string, width int, marker string) string {
	if width <= 0 {
		return ""
	}
	return measure.Truncate(text, width, marker)
}

// TruncateTextLeft fits text to a terminal-cell budget by truncating leading clusters,
// keeping trailing characters and prepending the caller's marker when truncated.
// It follows the same supported text policy as TruncateText.
func TruncateTextLeft(text string, width int, marker string) string {
	if width <= 0 {
		return ""
	}
	return measure.TruncateLeft(text, width, marker)
}
