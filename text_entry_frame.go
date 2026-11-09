// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// TextEntryFrameStyle selects the decoration around a TextEntry.
type TextEntryFrameStyle int

const (
	// TextEntryFrameNone draws no decoration. It is the zero value.
	TextEntryFrameNone TextEntryFrameStyle = iota
	// TextEntryFrameBox draws a rounded box with the title and hint in the top
	// border and the footer in the bottom border (the taia.legacy look):
	//
	//	╭─ title ──────── hint ─╮
	//	│ ❯ text                │
	//	╰─────────────── footer ─╯
	TextEntryFrameBox
	// TextEntryFrameSimple draws a full-width rule above and below the text:
	//
	//	── title ─────── hint ──
	//	❯ text
	//	─────────────── footer ──
	TextEntryFrameSimple
)

var textEntryFrameNames = []string{"none", "box", "simple"}

// String returns "none", "box" or "simple".
func (s TextEntryFrameStyle) String() string {
	if s < 0 || int(s) >= len(textEntryFrameNames) {
		return "unknown"
	}
	return textEntryFrameNames[s]
}

// TextEntryFrame is the decoration of a TextEntry. Title and Hint sit in the
// top border (left and right), Footer in the bottom border (right). Empty
// texts leave plain border.
//
// BorderColor and LabelColor (for Title, Hint and Footer) accept "" or "dim"
// (dimmed default color, the default), "default" (the terminal default color),
// "#RRGGBB", a palette index "0".."255", or the name of a theme color such as
// "border_fg" or "status_fg" (the yaml names in ThemeColors, resolved against
// the theme given to ApplyTheme). An unknown name falls back to dim.
type TextEntryFrame struct {
	Style                   TextEntryFrameStyle
	Title, Hint, Footer     string
	BorderColor, LabelColor string
}

var textEntryRounded = sync.OnceValue(func() BoxBorder {
	b, err := getBoxBorderGlyphs(BoxBorderStyleRounded)
	if err != nil {
		panic("loom: " + err.Error())
	}
	return b
})

// SetFrame replaces the frame.
func (e *TextEntry) SetFrame(f TextEntryFrame) { e.frame = f }

// Frame returns the current frame.
func (e *TextEntry) Frame() TextEntryFrame { return e.frame }

// SetTitle sets the text in the top border, left.
func (e *TextEntry) SetTitle(s string) { e.frame.Title = s }

// SetHint sets the text in the top border, right.
func (e *TextEntry) SetHint(s string) { e.frame.Hint = s }

// SetFooter sets the text in the bottom border, right.
func (e *TextEntry) SetFooter(s string) { e.frame.Footer = s }

// SetHeights sets the row bounds (see TextEntryOptions). Zero takes the
// default; a negative max is unbounded.
func (e *TextEntry) SetHeights(height, minHeight, maxHeight int) {
	e.height, e.minHeight, e.maxHeight = height, minHeight, maxHeight
}

// Heights returns the row bounds as set, before defaults apply.
func (e *TextEntry) Heights() (height, minHeight, maxHeight int) {
	return e.height, e.minHeight, e.maxHeight
}

// SetSendButton shows or hides the send button; label "" takes the spec label.
func (e *TextEntry) SetSendButton(on bool, label string) {
	e.sendButton, e.sendLabel = on, label
}

func (e *TextEntry) buttonLabel() string {
	if e.sendLabel != "" {
		return e.sendLabel
	}
	return SpeccedDefaults.TextEntry.SendLabel
}

func (e *TextEntry) buttonWidth() int {
	if !e.sendButton {
		return 0
	}
	return StringWidth(e.buttonLabel())
}

// chromeSize is the columns and rows the frame takes.
func (e *TextEntry) chromeSize() (w, h int) {
	switch e.frame.Style {
	case TextEntryFrameBox:
		return 4, 2
	case TextEntryFrameSimple:
		return 0, 2
	}
	return 0, 0
}

// boundedRows clamps a text row count to the height bounds.
func (e *TextEntry) boundedRows(n int) int {
	d := SpeccedDefaults.TextEntry
	lo := e.minHeight
	if lo <= 0 {
		lo = d.MinHeight
	}
	preferred := e.height
	if preferred <= 0 {
		preferred = d.Height
	}
	n = max(max(n, preferred), lo)
	if hi := e.maxRows(); hi > 0 {
		n = min(n, max(hi, lo))
	}
	return n
}

// maxRows is the most text rows shown; 0 means unbounded.
func (e *TextEntry) maxRows() int {
	switch {
	case e.maxHeight > 0:
		return e.maxHeight
	case e.maxHeight < 0:
		return 0
	}
	d := SpeccedDefaults.TextEntry
	rows := e.appRows
	if rows <= 0 {
		if _, r, err := TerminalSize(); err == nil && r > 0 {
			rows = r
		} else {
			rows = d.FallbackTerminalRows
		}
	}
	return max(1, rows*d.MaxHeightPercent/100)
}

// frameStyles returns the border and label styles.
func (e *TextEntry) frameStyles() (border, label Style) {
	return e.frameColor(e.frame.BorderColor), e.frameColor(e.frame.LabelColor)
}

func (e *TextEntry) frameColor(name string) Style {
	dim := Style{Dim: true}
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "dim":
		return dim
	case "default":
		return Style{}
	}
	name = strings.TrimSpace(name)
	if len(name) == 7 && name[0] == '#' {
		rgb, err := strconv.ParseUint(name[1:], 16, 24)
		if err != nil {
			return dim
		}
		return Style{FG: ColorRGB(uint8(rgb>>16), uint8(rgb>>8), uint8(rgb))}
	}
	if n, err := strconv.ParseUint(name, 10, 8); err == nil {
		return Style{FG: ColorIndex(uint8(n))}
	}
	v := reflect.ValueOf(e.theme)
	for i := 0; i < v.NumField(); i++ {
		if v.Type().Field(i).Tag.Get("yaml") != name {
			continue
		}
		if tc, ok := v.Field(i).Interface().(ThemeColor); ok {
			return Style{FG: tc.Color()}
		}
	}
	return dim
}

// drawFrame paints the frame into r and returns the rect left for the content.
// A frame that does not fit is skipped and the content gets all of r.
func (e *TextEntry) drawFrame(c *Canvas, r Rect) Rect {
	cw, ch := e.chromeSize()
	if cw+ch == 0 || r.H < ch+1 || r.W < cw+3 {
		return r
	}
	border, label := e.frameStyles()
	lead := 2
	var topL, topR, botL, botR, side string
	if e.frame.Style == TextEntryFrameBox {
		b := textEntryRounded()
		lead = 1
		topL, topR, botL, botR, side = b.TopLeft, b.TopRight, b.BottomLeft, b.BottomRight, b.Vertical
	}
	fill := textEntryRounded().Horizontal
	e.drawRule(c, r.X, r.Y, r.W, topL, topR, fill, lead, e.frame.Title, e.frame.Hint, border, label)
	e.drawRule(c, r.X, r.Y+r.H-1, r.W, botL, botR, fill, lead, "", e.frame.Footer, border, label)
	if side != "" {
		for y := 1; y < r.H-1; y++ {
			c.Write(r.X, r.Y+y, side, border)
			c.Write(r.X+r.W-1, r.Y+y, side, border)
		}
	}
	return Rect{X: r.X + cw/2, Y: r.Y + ch/2, W: r.W - cw, H: r.H - ch}
}

// drawRule writes one border row: corners, a lead of fill, the left label, fill,
// the right label, a lead of fill. Labels that do not fit are dropped, the right
// one first, then cut.
func (e *TextEntry) drawRule(c *Canvas, x, y, w int, cl, cr, fill string, lead int, l, r string, border, label Style) {
	pad := func(s string) string {
		if s == "" {
			return ""
		}
		return " " + s + " "
	}
	l, r = pad(strings.ReplaceAll(l, "\n", " ")), pad(strings.ReplaceAll(r, "\n", " "))
	space := w - StringWidth(cl) - StringWidth(cr) - 2*lead - 1
	if StringWidth(l)+StringWidth(r) > space {
		r = ""
	}
	if StringWidth(l) > space {
		l = TruncateText(l, max(space, 0), "…")
	}
	col := x
	col += c.Write(col, y, cl, border)
	col += c.Write(col, y, strings.Repeat(fill, lead), border)
	col += c.Write(col, y, l, label)
	end := x + w - StringWidth(cr) - lead - StringWidth(r)
	for col < end {
		col += c.Write(col, y, fill, border)
	}
	col += c.Write(col, y, r, label)
	col += c.Write(col, y, strings.Repeat(fill, lead), border)
	c.Write(col, y, cr, border)
}
