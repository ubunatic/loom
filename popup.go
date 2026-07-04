// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "strings"

// Popup overlays a fixed-size inner Widget centered over the pane.
// While Open, it captures all keyboard and mouse events before the
// background widget. Set Open=false to dismiss.
type Popup struct {
	Title  string
	Inner  Widget
	Open   bool
	Width  int // 0 = half the canvas width
	Height int // 0 = half the canvas height
	Style  Style
}

// NewPopup creates a Popup wrapping inner with the given title.
func NewPopup(title string, inner Widget) *Popup {
	return &Popup{Title: title, Inner: inner, Open: true}
}

// Draw renders the popup centered in r, with a simple box border.
// If not Open, Draw does nothing.
func (p *Popup) Draw(c *Canvas, r Rect) {
	if !p.Open || p.Inner == nil {
		return
	}
	pw, ph := p.dims(r)
	x := r.X + (r.W-pw)/2
	y := r.Y + (r.H-ph)/2
	if x < r.X {
		x = r.X
	}
	if y < r.Y {
		y = r.Y
	}

	// Fill background.
	c.Fill(Rect{x, y, pw, ph}, Cell{Text: " ", Style: p.Style})

	// Top border.
	title := " " + p.Title + " "
	if len(title) > pw-2 {
		title = title[:pw-2]
	}
	top := "┌" + title + strings.Repeat("─", pw-2-len([]rune(title))) + "┐"
	c.Write(x, y, top, p.Style)

	// Side borders.
	for row := 1; row < ph-1; row++ {
		c.Write(x, y+row, "│", p.Style)
		c.Write(x+pw-1, y+row, "│", p.Style)
	}

	// Bottom border.
	bottom := "└" + strings.Repeat("─", pw-2) + "┘"
	c.Write(x, y+ph-1, bottom, p.Style)

	// Inner content area (inside the border).
	if ph > 2 && pw > 2 {
		if f, ok := p.Inner.(Focusable); ok {
			f.SetFocus(p.Open)
		}
		p.Inner.Draw(c, Rect{X: x + 1, Y: y + 1, W: pw - 2, H: ph - 2})
	}
}

// HandleKey forwards to Inner while Open; Esc closes the popup.
func (p *Popup) HandleKey(e KeyEvent) (quit bool) {
	if !p.Open {
		return false
	}
	if e.Key == "esc" {
		p.Open = false
		return false
	}
	return p.Inner.HandleKey(e)
}

// HandleMouse forwards to Inner while Open.
func (p *Popup) HandleMouse(e MouseEvent) (quit bool) {
	if !p.Open {
		return false
	}
	return p.Inner.HandleMouse(e)
}

func (p *Popup) dims(r Rect) (w, h int) {
	w = p.Width
	h = p.Height
	if w <= 0 {
		w = r.W / 2
	}
	if h <= 0 {
		h = r.H / 2
	}
	if w < 4 {
		w = 4
	}
	if h < 3 {
		h = 3
	}
	if w > r.W {
		w = r.W
	}
	if h > r.H {
		h = r.H
	}
	return w, h
}

// ContentHeight estimates the required height for this popup.
func (p *Popup) ContentHeight() int {
	if p.Inner == nil {
		return 0
	}
	ch := 1
	if chWidget, ok := p.Inner.(ContentHeighter); ok {
		ch = chWidget.ContentHeight()
	}
	return ch + 2 // borders
}
