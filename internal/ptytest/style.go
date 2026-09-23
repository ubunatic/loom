// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ptytest

import (
	"fmt"
	"strings"
)

// ColorType specifies the color encoding mode.
type ColorType uint8

const (
	ColorDefault ColorType = iota
	ColorIndexed
	ColorDirect
)

// Color represents a terminal foreground or background color: default, 256-color
// palette index (0-255), or 24-bit truecolor RGB.
type Color struct {
	Type    ColorType
	Index   uint8
	R, G, B uint8
}

// ColorReset returns the default terminal color.
func ColorReset() Color { return Color{Type: ColorDefault} }

// ColorIndex returns a 256-color palette entry (0-15 standard/bright ANSI, 16-255 xterm).
func ColorIndex(n uint8) Color { return Color{Type: ColorIndexed, Index: n} }

// ColorRGB returns a 24-bit truecolor RGB value.
func ColorRGB(r, g, b uint8) Color { return Color{Type: ColorDirect, R: r, G: g, B: b} }

// RGB returns the approximate 24-bit RGB value for c and whether one could be
// resolved. ColorDefault returns ok=false. Indexed colors are resolved using
// the standard xterm 256 palette.
func (c Color) RGB() (r, g, b uint8, ok bool) {
	switch c.Type {
	case ColorDirect:
		return c.R, c.G, c.B, true
	case ColorIndexed:
		r, g, b = xterm256RGB(c.Index)
		return r, g, b, true
	default:
		return 0, 0, 0, false
	}
}

func (c Color) String() string {
	switch c.Type {
	case ColorDefault:
		return "default"
	case ColorIndexed:
		return fmt.Sprintf("index(%d)", c.Index)
	case ColorDirect:
		return fmt.Sprintf("rgb(%d,%d,%d)", c.R, c.G, c.B)
	default:
		return "unknown"
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

// Style describes the visual appearance (colors and SGR text attributes) of a VT cell.
type Style struct {
	FG        Color
	BG        Color
	Bold      bool
	Dim       bool
	Italic    bool
	Underline bool
	Blink     bool
	Reverse   bool
	Hidden    bool
	Strike    bool
}

// Effective returns the rendered foreground and background colors, swapping
// them when Reverse is active.
func (s Style) Effective() (fg, bg Color) {
	if s.Reverse {
		return s.BG, s.FG
	}
	return s.FG, s.BG
}

func (s Style) String() string {
	var parts []string
	if s.FG.Type != ColorDefault {
		parts = append(parts, "fg="+s.FG.String())
	}
	if s.BG.Type != ColorDefault {
		parts = append(parts, "bg="+s.BG.String())
	}
	if s.Bold {
		parts = append(parts, "bold")
	}
	if s.Dim {
		parts = append(parts, "dim")
	}
	if s.Italic {
		parts = append(parts, "italic")
	}
	if s.Underline {
		parts = append(parts, "underline")
	}
	if s.Blink {
		parts = append(parts, "blink")
	}
	if s.Reverse {
		parts = append(parts, "reverse")
	}
	if s.Hidden {
		parts = append(parts, "hidden")
	}
	if s.Strike {
		parts = append(parts, "strike")
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ",")
}

// Cell represents a single terminal cell containing a rune and its visual style.
// Wide characters occupy two consecutive cells: the left cell stores the rune,
// while the right cell stores Rune=0 with the same Style.
type Cell struct {
	Rune  rune
	Style Style
}

func (c Cell) String() string {
	if c.Rune == 0 {
		return fmt.Sprintf("[cont %s]", c.Style)
	}
	return fmt.Sprintf("[%c %s]", c.Rune, c.Style)
}
