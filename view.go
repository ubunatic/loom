// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "strings"

// View is a scrollable list of text lines rendered into a Rect.
// When content exceeds the visible area a "▐" scroll indicator appears on the
// right edge, proportional to the current scroll position.
type View struct {
	Lines  []string
	Scroll int   // first visible line index
	Style  Style // base style for all lines
	lastH  int   // height from last Draw; gates scroll in HandleKey
}

// NewView creates a View from a slice of pre-formatted lines.
func NewView(lines []string) *View { return &View{Lines: lines} }

// Draw renders visible lines into r. When scrollable, the rightmost column is
// reserved for the scroll indicator and content is truncated one column shorter.
func (v *View) Draw(c *Canvas, r Rect) {
	v.lastH = r.H
	total := len(v.Lines)
	scrollable := total > r.H

	// Clamp scroll.
	if v.Scroll < 0 {
		v.Scroll = 0
	}
	if scrollable && v.Scroll > total-r.H {
		v.Scroll = total - r.H
	}

	// Pre-compute indicator row (proportional to scroll position).
	indicatorRow := 0
	if scrollable && total > r.H {
		maxScroll := total - r.H
		ratio := float64(v.Scroll) / float64(maxScroll)
		indicatorRow = int(ratio * float64(r.H-1))
	}

	contentW := r.W
	if scrollable {
		contentW = r.W - 1 // reserve rightmost column for indicator
	}

	for row := 0; row < r.H; row++ {
		y := r.Y + row
		c.Fill(Rect{r.X, y, r.W, 1}, Cell{Text: " ", Style: v.Style})
		lineIdx := v.Scroll + row
		if lineIdx >= 0 && lineIdx < total {
			plain := stripANSI(v.Lines[lineIdx])
			if len([]rune(plain)) > contentW {
				plain = string([]rune(plain)[:contentW])
			}
			c.Write(r.X, y, plain, v.Style)
		}
		if scrollable && row == indicatorRow {
			c.Set(r.X+r.W-1, y, Cell{Text: "▐", Style: Style{Dim: true}})
		}
	}
}

// HandleKey supports up/down scroll.
func (v *View) HandleKey(e KeyEvent) (quit bool) {
	maxScroll := len(v.Lines) - v.lastH
	if maxScroll < 0 {
		maxScroll = 0
	}
	switch e.Key {
	case "up":
		if v.Scroll > 0 {
			v.Scroll--
		}
	case "down":
		if v.Scroll < maxScroll {
			v.Scroll++
		}
	case "esc", "ctrl-c":
		return true
	}
	return false
}

// HandleMouse supports scroll-wheel navigation.
func (v *View) HandleMouse(e MouseEvent) (quit bool) {
	maxScroll := len(v.Lines) - v.lastH
	if maxScroll < 0 {
		maxScroll = 0
	}
	switch e.Action {
	case MouseScrollUp:
		if v.Scroll > 0 {
			v.Scroll--
		}
	case MouseScrollDown:
		if v.Scroll < maxScroll {
			v.Scroll++
		}
	}
	return false
}

// stripANSI removes escape sequences from s, returning plain text.
func stripANSI(s string) string {
	var b strings.Builder
	esc := false
	csi := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			esc = true
		case esc && r == '[':
			csi = true
			esc = false
		case csi && r >= 0x40 && r <= 0x7e:
			csi = false
		case esc:
			esc = false
		case !csi:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ContentHeight estimates the required height for this view.
func (v *View) ContentHeight() int {
	return len(v.Lines)
}
