// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// Popup overlays a fixed-size inner Widget centered over the pane.
// While Open, it captures all keyboard and mouse events before the
// background widget. Set Open=false to dismiss.
type Popup struct {
	Title     string
	Inner     Widget
	Open      bool
	Width     int // 0 = half the canvas width
	Height    int // 0 = half the canvas height
	Style     Style
	lastRect  Rect
	innerRect Rect
}

// ApplyTheme updates popup chrome and forwards the theme to its inner widget.
func (p *Popup) ApplyTheme(theme ThemeColors) {
	p.Style = theme.BoxStyle().Background
	if child, ok := p.Inner.(Themeable); ok {
		child.ApplyTheme(theme)
	}
}

// NewPopup creates a Popup wrapping inner with the given title.
func NewPopup(title string, inner Widget) *Popup {
	return &Popup{Title: title, Inner: inner, Open: true}
}

// Draw renders the popup centered in r, with a simple box border.
// If not Open, Draw does nothing.
func (p *Popup) Draw(c *Canvas, r Rect) {
	p.lastRect = r
	p.innerRect = Rect{}
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

	// Draw border and title using the new primitive (issue 037).
	// DrawBox uses Sharp style with display-width-aware title truncation.
	c.DrawBox(Rect{X: x, Y: y, W: pw, H: ph}, BoxBorderStyleSharp, p.Title, p.Style)

	// Inner content area (inside the border).
	if ph > 2 && pw > 2 {
		if f, ok := p.Inner.(Focusable); ok {
			f.SetFocus(p.Open)
		}
		p.innerRect = Rect{X: x + 1, Y: y + 1, W: pw - 2, H: ph - 2}
		p.Inner.Draw(c, p.innerRect)
	}
}

// ConsumeKey forwards to Inner while Open; Esc closes the popup.
func (p *Popup) ConsumeKey(e KeyEvent) (quit EventResult) {
	if !p.Open || p.Inner == nil {
		return Ignored()
	}
	if e.Key == "esc" {
		p.Open = false
		return Handled()
	}
	res := DispatchKeyEvent(p.Inner, e)
	return res
}

func (p *Popup) ConsumePaste(e PasteEvent) EventResult {
	if !p.Open || p.Inner == nil {
		return Ignored()
	}
	return DispatchPasteEvent(p.Inner, e)
}

// ConsumeMouse forwards to Inner while Open.
func (p *Popup) ConsumeMouse(e MouseEvent) (quit EventResult) {
	if !p.Open || p.Inner == nil {
		return Ignored()
	}
	x, y := e.X+p.lastRect.X, e.Y+p.lastRect.Y
	if !p.innerRect.Contains(x, y) {
		return Handled()
	}
	e.X, e.Y = x-p.innerRect.X, y-p.innerRect.Y
	res := DispatchMouseEvent(p.Inner, e)
	if !res.Consumed {
		return Handled()
	}
	return res
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
