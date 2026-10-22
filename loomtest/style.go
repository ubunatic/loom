// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomtest

import "ubunatic.com/loom/vt"

// ColorType specifies the color encoding mode.
type ColorType = vt.ColorType

const (
	ColorDefault = vt.ColorDefault
	ColorIndexed = vt.ColorIndexed
	ColorDirect  = vt.ColorDirect
)

// Color represents a terminal foreground or background color: default, 256-color
// palette index (0-255), or 24-bit truecolor RGB.
type Color = vt.Color

// ColorReset returns the default terminal color.
func ColorReset() Color { return vt.ColorReset() }

// ColorIndex returns a 256-color palette entry (0-15 standard/bright ANSI, 16-255 xterm).
func ColorIndex(n uint8) Color { return vt.ColorIndex(n) }

// ColorRGB returns a 24-bit truecolor RGB value.
func ColorRGB(r, g, b uint8) Color { return vt.ColorRGB(r, g, b) }

// Style describes the visual appearance (colors and SGR text attributes) of a VT cell.
type Style = vt.Style

// Cell represents a single terminal cell containing a rune and its visual style.
type Cell = vt.Cell
