// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package measure exposes Loom's terminal-cell measurement policy to layout
// planners and applications without introducing a second width contract.
package measure

import (
	"strings"

	"codeberg.org/ubunatic/loom"
)

// Size describes the visible terminal-cell dimensions of text lines.
type Size struct {
	Width  int
	Height int
}

// StringWidth returns the visible terminal-cell width of text according to
// Loom's rendering policy.
func StringWidth(text string) int {
	return loom.StringWidth(text)
}

// Lines returns the visible dimensions of a multi-line string. An empty
// string is one empty line.
func Lines(text string) Size {
	lines := strings.Split(text, "\n")
	result := Size{Height: len(lines)}
	for _, line := range lines {
		if width := StringWidth(line); width > result.Width {
			result.Width = width
		}
	}
	return result
}

// Truncate fits text to a terminal-cell budget while preserving Loom's
// cluster and terminal-control handling.
func Truncate(text string, width int, marker string) string {
	return loom.TruncateText(text, width, marker)
}

// TruncateLeft fits text to a terminal-cell budget while keeping its trailing
// content and preserving Loom's cluster and terminal-control handling.
func TruncateLeft(text string, width int, marker string) string {
	return loom.TruncateTextLeft(text, width, marker)
}
