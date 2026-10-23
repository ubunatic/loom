// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "ubunatic.com/loom/ansi"

// ColorProfile is the color capability available at an ANSI output boundary.
type ColorProfile = ansi.ColorProfile

const (
	ColorProfileNone      = ansi.ColorProfileNone
	ColorProfile16        = ansi.ColorProfile16
	ColorProfile256       = ansi.ColorProfile256
	ColorProfileTrueColor = ansi.ColorProfileTrueColor
)

// Color is a terminal color: reset, 256-color index, or 24-bit RGB.
type Color = ansi.Color

// Style describes the visual appearance of a Canvas cell.
type Style = ansi.Style

// ColorReset is the default terminal foreground or background color.
func ColorReset() Color { return ansi.ColorReset() }

// ColorIndex returns a 256-color palette entry (0–255).
func ColorIndex(n uint8) Color { return ansi.ColorIndex(n) }

// ColorRGB returns a 24-bit true-color value.
func ColorRGB(r, g, b uint8) Color { return ansi.ColorRGB(r, g, b) }

// DetectColorProfile reports the profile selected by LOOMCOLOR, NO_COLOR,
// COLORTERM, and TERM. LOOMCOLOR accepts truecolor, 256, 16, or none.
func DetectColorProfile() ColorProfile { return ansi.DetectColorProfile() }

// DetectColorProfileFrom applies the environment profile rules to supplied values.
func DetectColorProfileFrom(override, noColor, colorTerm, term string) ColorProfile {
	return ansi.DetectColorProfileFrom(override, noColor, colorTerm, term)
}

// Reset is a zero Style — default terminal colors, no attributes.
var Reset = ansi.Reset

// Debug is a package-level flag. When enabled, containers draw dotted cell outlines.
var Debug bool

// DebugColor is the Color used for debug boundaries. Defaults to a medium grey (ColorIndex(242)).
var DebugColor = ColorIndex(242)

// drawDebugBorder draws a subtle dotted outline around r on canvas c.
func drawDebugBorder(c *Canvas, r Rect) {
	style := Style{Dim: true, FG: DebugColor}
	// Top & Bottom
	for x := r.X; x < r.X+r.W; x++ {
		if cur := c.Get(x, r.Y); !cur.Continuation && (cur.Text == " " || cur.Text == "") {
			c.Set(x, r.Y, Cell{Text: "⠂", Style: style})
		}
		if cur := c.Get(x, r.Y+r.H-1); !cur.Continuation && (cur.Text == " " || cur.Text == "") {
			c.Set(x, r.Y+r.H-1, Cell{Text: "⠂", Style: style})
		}
	}
	// Left & Right
	for y := r.Y; y < r.Y+r.H; y++ {
		if cur := c.Get(r.X, y); !cur.Continuation && (cur.Text == " " || cur.Text == "") {
			c.Set(r.X, y, Cell{Text: "⠆", Style: style})
		}
		if cur := c.Get(r.X+r.W-1, y); !cur.Continuation && (cur.Text == " " || cur.Text == "") {
			c.Set(r.X+r.W-1, y, Cell{Text: "⠆", Style: style})
		}
	}
}
