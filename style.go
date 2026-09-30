// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"os"
	"strings"
)

// ColorProfile is the color capability available at an ANSI output boundary.
type ColorProfile uint8

const (
	ColorProfileNone ColorProfile = iota
	ColorProfile16
	ColorProfile256
	ColorProfileTrueColor
)

// DetectColorProfile reports the profile selected by LOOMCOLOR, NO_COLOR,
// COLORTERM, and TERM. LOOMCOLOR accepts truecolor, 256, 16, or none.
func DetectColorProfile() ColorProfile {
	return DetectColorProfileFrom(os.Getenv("LOOMCOLOR"), os.Getenv("NO_COLOR"), os.Getenv("COLORTERM"), os.Getenv("TERM"))
}

// DetectColorProfileFrom applies the environment profile rules to supplied values.
func DetectColorProfileFrom(override, noColor, colorTerm, term string) ColorProfile {
	switch strings.ToLower(strings.TrimSpace(override)) {
	case "truecolor", "24bit":
		return ColorProfileTrueColor
	case "256", "256color":
		return ColorProfile256
	case "16", "basic":
		return ColorProfile16
	case "none", "off":
		return ColorProfileNone
	}
	if noColor != "" {
		return ColorProfileNone
	}
	if strings.EqualFold(colorTerm, "truecolor") || strings.EqualFold(colorTerm, "24bit") {
		return ColorProfileTrueColor
	}
	if strings.Contains(strings.ToLower(term), "256color") {
		return ColorProfile256
	}
	if strings.EqualFold(term, "dumb") {
		return ColorProfileNone
	}
	if term == "" {
		return ColorProfileTrueColor
	}
	return ColorProfile16
}

// Color is a terminal color: reset, 256-color index, or 24-bit RGB.
type Color struct {
	mode  colorMode
	index uint8
	r, g  uint8
	b     uint8
}

type colorMode uint8

const (
	colorReset colorMode = iota
	colorIndex
	colorRGB
)

// ColorReset is the default terminal foreground or background color.
func ColorReset() Color { return Color{} }

// ColorIndex returns a 256-color palette entry (0–255).
func ColorIndex(n uint8) Color { return Color{mode: colorIndex, index: n} }

// ColorRGB returns a 24-bit true-color value.
func ColorRGB(r, g, b uint8) Color { return Color{mode: colorRGB, r: r, g: g, b: b} }

// RGB returns an approximate 24-bit value for c and whether one could be
// resolved. colorReset has no fixed RGB value (terminal-defined) and returns
// ok=false. Indexed colors are approximated using the standard xterm 256
// palette (6x6x6 color cube and grayscale ramp; 0-15 use common ANSI RGB
// values).
func (c Color) RGB() (r, g, b uint8, ok bool) {
	switch c.mode {
	case colorRGB:
		return c.r, c.g, c.b, true
	case colorIndex:
		r, g, b = xterm256RGB(c.index)
		return r, g, b, true
	default:
		return 0, 0, 0, false
	}
}

var ansi16RGB = [16][3]uint8{
	{0, 0, 0}, {205, 0, 0}, {0, 205, 0}, {205, 205, 0},
	{0, 0, 238}, {205, 0, 205}, {0, 205, 205}, {229, 229, 229},
	{127, 127, 127}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
	{92, 92, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
}

func xterm256RGB(idx uint8) (r, g, b uint8) {
	switch {
	case idx < 16:
		rgb := ansi16RGB[idx]
		return rgb[0], rgb[1], rgb[2]
	case idx >= 232:
		gray := 8 + (int(idx)-232)*10
		return uint8(gray), uint8(gray), uint8(gray)
	default:
		n := int(idx) - 16
		levels := [6]uint8{0, 95, 135, 175, 215, 255}
		return levels[(n/36)%6], levels[(n/6)%6], levels[n%6]
	}
}

func (c Color) fgSeq() string {
	switch c.mode {
	case colorIndex:
		return fmt.Sprintf("\x1b[38;5;%dm", c.index)
	case colorRGB:
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", c.r, c.g, c.b)
	default:
		return "\x1b[39m" // default fg
	}
}

func (c Color) bgSeq() string {
	switch c.mode {
	case colorIndex:
		return fmt.Sprintf("\x1b[48;5;%dm", c.index)
	case colorRGB:
		return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", c.r, c.g, c.b)
	default:
		return "\x1b[49m" // default bg
	}
}

func (c Color) forProfile(profile ColorProfile) Color {
	if c.mode == colorReset || profile == ColorProfileTrueColor {
		return c
	}
	limit := 256
	if profile == ColorProfile16 {
		limit = 16
	} else if profile == ColorProfileNone {
		return ColorReset()
	}
	if c.mode == colorIndex && int(c.index) < limit {
		return c
	}
	r, g, b, ok := c.RGB()
	if !ok {
		return c
	}
	return nearestColor(r, g, b, limit)
}

func (c Color) sequence(profile ColorProfile, background bool) string {
	resolved := c.forProfile(profile)
	if profile == ColorProfile16 && resolved.mode == colorIndex {
		base := 30
		if background {
			base = 40
		}
		if resolved.index >= 8 {
			base += 60
			return fmt.Sprintf("\x1b[%dm", base+int(resolved.index)-8)
		}
		return fmt.Sprintf("\x1b[%dm", base+int(resolved.index))
	}
	if background {
		return resolved.bgSeq()
	}
	return resolved.fgSeq()
}

func nearestColor(r, g, b uint8, limit int) Color {
	best, bestDistance := 0, int(^uint(0)>>1)
	for i := 0; i < limit; i++ {
		cr, cg, cb := xterm256RGB(uint8(i))
		dr, dg, db := int(r)-int(cr), int(g)-int(cg), int(b)-int(cb)
		distance := dr*dr + dg*dg + db*db
		if distance < bestDistance {
			best, bestDistance = i, distance
		}
	}
	return ColorIndex(uint8(best))
}

// Style describes the visual appearance of a Canvas cell.
type Style struct {
	FG        Color
	BG        Color
	Bold      bool
	Underline bool
	Dim       bool
}

// ANSI returns the escape sequence that applies this style.
// Always starts with a full reset to avoid state bleed from previous cells.
func (s Style) ANSI() string {
	return s.ANSIFor(ColorProfileTrueColor)
}

// ANSIFor returns the style escape sequence downsampled for profile.
func (s Style) ANSIFor(profile ColorProfile) string {
	out := "\x1b[0m" // reset all attributes
	if s.Bold {
		out += "\x1b[1m"
	}
	if s.Underline {
		out += "\x1b[4m"
	}
	if s.Dim {
		out += "\x1b[2m"
	}
	out += s.FG.sequence(profile, false)
	out += s.BG.sequence(profile, true)
	return out
}

// Reset is a zero Style — default terminal colors, no attributes.
var Reset = Style{}

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
