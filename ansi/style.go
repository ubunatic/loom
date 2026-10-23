// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package ansi provides ANSI colors, styles, and SGR parsing.
package ansi

import (
	"os"
	"strconv"
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

var palette = func() (table [256][3]uint8) {
	for i := range table {
		r, g, b := computeXterm256RGB(uint8(i))
		table[i] = [3]uint8{r, g, b}
	}
	return
}()

var indexed16 = func() (table [256]Color) {
	for i, rgb := range palette {
		table[i] = nearestColor(rgb[0], rgb[1], rgb[2], 16)
	}
	return
}()

func xterm256RGB(idx uint8) (r, g, b uint8) {
	rgb := palette[idx]
	return rgb[0], rgb[1], rgb[2]
}

func computeXterm256RGB(idx uint8) (r, g, b uint8) {
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
	if c.mode == colorIndex && profile == ColorProfile16 {
		return indexed16[c.index]
	}
	r, g, b, ok := c.RGB()
	if !ok {
		return c
	}
	return nearestColor(r, g, b, limit)
}

// Index returns the palette index and whether c is an indexed color.
func (c Color) Index() (uint8, bool) { return c.index, c.mode == colorIndex }

// AppendFG appends the foreground SGR sequence for profile.
// It allocates only when b has insufficient capacity.
func (c Color) AppendFG(b []byte, profile ColorProfile) []byte {
	return c.appendSequence(b, profile, false)
}

// AppendBG appends the background SGR sequence for profile.
// It allocates only when b has insufficient capacity.
func (c Color) AppendBG(b []byte, profile ColorProfile) []byte {
	return c.appendSequence(b, profile, true)
}

func (c Color) appendSequence(b []byte, profile ColorProfile, background bool) []byte {
	c = c.forProfile(profile)
	base := 30
	if background {
		base = 40
	}
	b = append(b, '\x1b', '[')
	if c.mode == colorReset {
		b = strconv.AppendInt(b, int64(base+9), 10)
	} else if profile == ColorProfile16 {
		n := int(c.index)
		if n >= 8 {
			base += 60
			n -= 8
		}
		b = strconv.AppendInt(b, int64(base+n), 10)
	} else {
		b = strconv.AppendInt(b, int64(base+8), 10)
		if c.mode == colorIndex {
			b = append(b, ';', '5', ';')
			b = strconv.AppendInt(b, int64(c.index), 10)
		} else {
			b = append(b, ';', '2', ';')
			b = strconv.AppendInt(b, int64(c.r), 10)
			b = append(b, ';')
			b = strconv.AppendInt(b, int64(c.g), 10)
			b = append(b, ';')
			b = strconv.AppendInt(b, int64(c.b), 10)
		}
	}
	return append(b, 'm')
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
	Italic    bool
	Strike    bool
	Invert    bool
}

// ANSI returns the escape sequence that applies this style.
// Always starts with a full reset to avoid state bleed from previous cells.
func (s Style) ANSI() string {
	return s.ANSIFor(ColorProfileTrueColor)
}

// ANSIFor returns the style escape sequence downsampled for profile.
func (s Style) ANSIFor(profile ColorProfile) string {
	var b [80]byte
	return string(s.AppendANSIFor(b[:0], profile))
}

// AppendANSI appends this style using true color, starting with a full reset.
// It allocates only when b has insufficient capacity.
func (s Style) AppendANSI(b []byte) []byte {
	return s.AppendANSIFor(b, ColorProfileTrueColor)
}

// AppendANSIFor appends this style downsampled for profile, starting with a full reset.
// It allocates only when b has insufficient capacity.
func (s Style) AppendANSIFor(b []byte, profile ColorProfile) []byte {
	b = append(b, "\x1b[0m"...)
	for _, attr := range [...]struct {
		on   bool
		code byte
	}{
		{s.Bold, '1'}, {s.Dim, '2'}, {s.Italic, '3'},
		{s.Underline, '4'}, {s.Invert, '7'}, {s.Strike, '9'},
	} {
		if attr.on {
			b = append(b, '\x1b', '[', attr.code, 'm')
		}
	}
	b = s.FG.AppendFG(b, profile)
	return s.BG.AppendBG(b, profile)
}

// Reset is a zero Style — default terminal colors, no attributes.
var Reset = Style{}
