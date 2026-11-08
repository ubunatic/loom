// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "ubunatic.com/loom/measure"

// Viewport displays a scrollable window over a widget that can report its
// preferred content size through Measurer, ContentWidther, or ContentHeighter.
type Viewport struct {
	Child         Widget
	ScrollX       int
	ScrollY       int
	ScrollbarMode ScrollbarMode
	Scrollbar     ScrollbarStyle
	Style         Style // base text and background colors

	focused    bool
	lastRect   Rect
	content    Rect
	childW     int
	childH     int
	dragging   bool
	dragStart  int // ScrollY at the scrollbar press, restored on cancel
	dragGrab   int // thumb row under the pointer at the press
	modalMouse ModalMouseCapture
	pan        contentPan
}

// NewViewport wraps child in a scrollable viewport.
func NewViewport(child Widget) *Viewport {
	return &Viewport{Child: child, ScrollbarMode: ScrollbarAuto, Scrollbar: DefaultScrollbarStyle()}
}

// ApplyTheme updates the viewport surface, scrollbar, and themeable child.
func (v *Viewport) ApplyTheme(theme ThemeColors) {
	v.Style = Style{FG: theme.NormalFG.Color(), BG: theme.NormalBG.Color()}
	v.Scrollbar = theme.ScrollbarStyle()
	// A default track or thumb colour takes the surface background. Themed
	// colours stay, so a distinct thumb shows (issue 450).
	if v.Scrollbar.Track.BG == ColorReset() {
		v.Scrollbar.Track.BG = v.Style.BG
	}
	if v.Scrollbar.Thumb.BG == ColorReset() {
		v.Scrollbar.Thumb.BG = v.Style.BG
	}
	if child, ok := UnwrapWidget(v.Child).(Themeable); ok {
		child.ApplyTheme(theme)
	}
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
	if c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	// Establish the default surface on the main canvas. On a blank canvas this
	// shows the theme NormalBG (ticket 222). When drawn inside a Grid cell the
	// parent surface (focusBG or ambient) wins and this is a no-op (ticket 243).
	c.PaintDefaultSurface(r, v.Style)
	if v.Child == nil {
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
	// Propagate the parent's ambient BG (which PaintDefaultSurface either set or
	// left as-is) into the offscreen so blitted cells inherit it rather than
	// overwriting the main canvas's surface with ColorReset.
	ambientBG := Style{BG: c.Get(r.X, r.Y).Style.BG}
	offscreen.PaintSurface(offscreen.Bounds(), ambientBG)
	v.Child.Draw(offscreen, offscreen.Bounds())
	window := offscreen.SubCanvas(Rect{X: v.ScrollX, Y: v.ScrollY, W: contentW, H: r.H})
	c.Blit(window, r.X, r.Y)
	if showBar && r.W > 0 {
		start, thumb := scrollbarUnitThumb(r.H, h, r.H, v.ScrollY, max(0, h-r.H))
		for row, cell := range scrollbarColumn(v.Scrollbar, c.ColorProfile, r.H, start, thumb) {
			c.Set(r.X+r.W-1, r.Y+row, cell)
		}
	}
}

// Focused reports whether the viewport owns focus.
func (v *Viewport) Focused() bool { return v.focused }

// SetFocus updates focus and forwards it to a focusable child.
func (v *Viewport) SetFocus(focused bool) {
	// As for View, only a real focus loss ends a capture; containers re-apply
	// focus on every draw.
	if !focused && v.focused {
		v.endMouseCapture(true)
	}
	v.focused = focused
	if child, ok := v.Child.(Focusable); ok {
		child.SetFocus(focused)
	}
}

// ConsumeKey scrolls the viewport or forwards an unhandled key to its child.
func (v *Viewport) ConsumeKey(e KeyEvent) EventResult {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	page := max(1, v.lastRect.H)
	maxX, maxY := max(0, v.childW-v.content.W), max(0, v.childH-v.lastRect.H)
	switch key {
	case "up", "k":
		v.ScrollY = max(0, v.ScrollY-1)
		return Handled()
	case "down", "j":
		v.ScrollY = min(maxY, v.ScrollY+1)
		return Handled()
	case "left", "h":
		v.ScrollX = max(0, v.ScrollX-1)
		return Handled()
	case "right", "l":
		v.ScrollX = min(maxX, v.ScrollX+1)
		return Handled()
	case "pgdown", "pgdn", "pagedown", "shift-pgdown", "shift-pgdn", "shift-pagedown", " ":
		v.ScrollY = min(maxY, v.ScrollY+page)
		return Handled()
	case "pgup", "pageup", "shift-pgup", "shift-pageup":
		v.ScrollY = max(0, v.ScrollY-page)
		return Handled()
	case "home":
		v.ScrollY = 0
		return Handled()
	case "end":
		v.ScrollY = maxY
		return Handled()
	default:
		if v.Child != nil {
			return v.Child.ConsumeKey(e)
		}
	}
	return Ignored()
}

// ConsumeMouse scrolls with the wheel and translates pointer coordinates into
// the child's full content coordinate space.
func (v *Viewport) ConsumeMouse(e MouseEvent) EventResult {
	if result, captured := v.modalMouse.Dispatch(v, e); captured {
		v.dragging = false
		return result
	}
	maxY := max(0, v.childH-v.lastRect.H)
	if e.Action == MousePress {
		v.pan.end()
	}
	if e.Action == MousePress && e.Button == MouseLeft && e.X == v.lastRect.W-1 && maxY > 0 {
		v.pressScrollbar(e.Y, maxY)
		return Handled()
	}
	if e.Action == MousePress && e.Button == MouseLeft {
		v.dragging = false
	}
	if e.Action == MouseRelease && v.dragging {
		v.dragTo(e.Y, maxY)
		v.dragging = false
		return Handled()
	}
	if e.Action == MouseDrag && v.dragging {
		v.dragTo(e.Y, maxY)
		return Handled()
	}
	if v.pan.active {
		switch e.Action {
		case MouseDrag:
			v.ScrollY = v.pan.offset(e.Y, maxY)
			return Handled()
		case MouseRelease:
			v.pan.end()
			return Handled()
		}
	}
	switch e.Action {
	case MouseScrollUp:
		v.ScrollY = max(0, v.ScrollY-1)
		return Handled()
	case MouseScrollDown:
		v.ScrollY = min(maxY, v.ScrollY+1)
		return Handled()
	}
	if v.Child == nil || e.X < 0 || e.Y < 0 || e.X >= v.content.W || e.Y >= v.lastRect.H {
		return Ignored()
	}
	// The child sees the press first. Only a press it does not consume starts
	// a middle-button pan.
	childEvent := e
	childEvent.X += v.ScrollX
	childEvent.Y += v.ScrollY
	result := v.Child.ConsumeMouse(childEvent)
	if !result.Consumed && !result.Quit && panStarts(e, false, v.content.W, v.lastRect.H, maxY) {
		v.pan.press(e.Y, v.ScrollY)
		return Handled()
	}
	return result
}

// endMouseCapture ends a scrollbar drag or a pan. cancel restores the offset from before the press.
func (v *Viewport) endMouseCapture(cancel bool) {
	if cancel && v.dragging {
		v.ScrollY = v.dragStart
	}
	if cancel && v.pan.active {
		v.ScrollY = v.pan.start
	}
	v.dragging = false
	v.pan.end()
}

// pressScrollbar starts a thumb drag. A press on the thumb keeps the thumb row
// under the pointer; a press on the track moves the thumb to the pointer and
// drags from there.
func (v *Viewport) pressScrollbar(row, maxY int) {
	track := v.lastRect.H
	start, thumb := scrollbarUnitThumb(track, v.childH, track, v.ScrollY, maxY)
	v.dragging = true
	v.dragStart = v.ScrollY
	v.dragGrab = scrollbarUnitPointer(row) - start
	if !scrollbarRowTouchesThumb(row, start, thumb) {
		// A track click centres the thumb on the pointer's row centre.
		v.dragGrab = thumb / 2
		v.dragTo(row, maxY)
	}
}

func (v *Viewport) dragTo(y, maxY int) {
	track := v.lastRect.H
	_, thumb := scrollbarUnitThumb(track, v.childH, track, v.ScrollY, maxY)
	v.ScrollY = scrollbarUnitOffset(track, thumb, scrollbarUnitPointer(y), v.dragGrab, maxY)
}
