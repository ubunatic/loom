// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "codeberg.org/ubunatic/loom/measure"

// Viewport displays a scrollable window over a widget that can report its
// preferred content size through Measurer, ContentWidther, or ContentHeighter.
type Viewport struct {
	Child         Widget
	ScrollX       int
	ScrollY       int
	ScrollbarMode ScrollbarMode
	Scrollbar     ScrollbarStyle

	focused  bool
	lastRect Rect
	content  Rect
	childW   int
	childH   int
}

// NewViewport wraps child in a scrollable viewport.
func NewViewport(child Widget) *Viewport {
	return &Viewport{Child: child, ScrollbarMode: ScrollbarAuto, Scrollbar: DefaultScrollbarStyle()}
}

// Measure reports the viewport's natural content size.
func (v *Viewport) Measure(width int) measure.Size {
	w, h := v.childSize(width, 0, 0)
	return measure.Size{Width: w, Height: h}
}

func (v *Viewport) childSize(width, fallbackW, fallbackH int) (int, int) {
	w, h := 0, 0
	if v.Child != nil {
		if measured, ok := v.Child.(Measurer); ok {
			size := measured.Measure(width)
			w, h = size.Width, size.Height
		}
		if measured, ok := v.Child.(ContentWidther); ok {
			w = max(w, measured.ContentWidth())
		}
		if measured, ok := v.Child.(ContentHeighter); ok {
			h = max(h, measured.ContentHeight())
		}
	}
	return max(max(1, w), fallbackW), max(max(1, h), fallbackH)
}

// Draw renders the measured child offscreen, then blits the visible window.
func (v *Viewport) Draw(c *Canvas, r Rect) {
	v.lastRect = r
	if c == nil || r.W <= 0 || r.H <= 0 || v.Child == nil {
		return
	}
	w, h := v.childSize(0, r.W, r.H)
	showBar := scrollbarVisible(v.ScrollbarMode, h > r.H)
	contentW := r.W
	if showBar && contentW > 0 {
		contentW--
	}
	v.childW, v.childH = w, h
	v.content = Rect{X: r.X, Y: r.Y, W: contentW, H: r.H}
	v.ScrollX = min(max(0, v.ScrollX), max(0, w-contentW))
	v.ScrollY = min(max(0, v.ScrollY), max(0, h-r.H))
	offscreen := NewCanvas(w, h)
	offscreen.ColorProfile = c.ColorProfile
	v.Child.Draw(offscreen, offscreen.Bounds())
	window := offscreen.SubCanvas(Rect{X: v.ScrollX, Y: v.ScrollY, W: contentW, H: r.H})
	c.Blit(window, r.X, r.Y)
	if showBar && r.W > 0 {
		maxOffset := max(0, h-r.H)
		thumb := max(1, scrollbarThumbLength(r.H, h, r.H))
		start := scrollbarThumbStart(r.H, thumb, v.ScrollY, maxOffset)
		for row := 0; row < r.H; row++ {
			c.Set(r.X+r.W-1, r.Y+row, scrollbarCell(v.Scrollbar, row >= start && row < start+thumb))
		}
	}
}

// Focused reports whether the viewport owns focus.
func (v *Viewport) Focused() bool { return v.focused }

// SetFocus updates focus and forwards it to a focusable child.
func (v *Viewport) SetFocus(focused bool) {
	v.focused = focused
	if child, ok := v.Child.(Focusable); ok {
		child.SetFocus(focused)
	}
}

// HandleKey scrolls the viewport or forwards an unhandled key to its child.
func (v *Viewport) HandleKey(e KeyEvent) (quit bool) {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	page := max(1, v.lastRect.H)
	maxX, maxY := max(0, v.childW-v.content.W), max(0, v.childH-v.lastRect.H)
	switch key {
	case "up", "k":
		v.ScrollY = max(0, v.ScrollY-1)
	case "down", "j":
		v.ScrollY = min(maxY, v.ScrollY+1)
	case "left", "h":
		v.ScrollX = max(0, v.ScrollX-1)
	case "right", "l":
		v.ScrollX = min(maxX, v.ScrollX+1)
	case "pgdown", "pgdn", "pagedown", " ":
		v.ScrollY = min(maxY, v.ScrollY+page)
	case "pgup", "pageup":
		v.ScrollY = max(0, v.ScrollY-page)
	case "home":
		v.ScrollX, v.ScrollY = 0, 0
	case "end":
		v.ScrollX, v.ScrollY = maxX, maxY
	default:
		if consumer, ok := v.Child.(EventConsumer); ok {
			result := consumer.ConsumeKey(e)
			if result.Consumed {
				return result.Quit
			}
		}
		if consumer, ok := v.Child.(KeyConsumer); ok {
			if quit, consumed := consumer.ConsumeKey(e); consumed {
				return quit
			}
		}
		if v.Child != nil {
			return v.Child.HandleKey(e)
		}
	}
	return false
}

// HandleMouse scrolls with the wheel and translates pointer coordinates into
// the child's full content coordinate space.
func (v *Viewport) HandleMouse(e MouseEvent) (quit bool) {
	maxY := max(0, v.childH-v.lastRect.H)
	switch e.Action {
	case MouseScrollUp:
		v.ScrollY = max(0, v.ScrollY-1)
		return false
	case MouseScrollDown:
		v.ScrollY = min(maxY, v.ScrollY+1)
		return false
	}
	if v.Child == nil || e.X < 0 || e.Y < 0 || e.X >= v.content.W || e.Y >= v.lastRect.H {
		return false
	}
	e.X += v.ScrollX
	e.Y += v.ScrollY
	return v.Child.HandleMouse(e)
}
