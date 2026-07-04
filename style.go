// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "fmt"

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
	out += s.FG.fgSeq()
	out += s.BG.bgSeq()
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
