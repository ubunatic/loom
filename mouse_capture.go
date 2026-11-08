// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// Pane-level mouse capture keeps a drag with the widget that started it, even
// when the pointer leaves that widget's bounds or the window and comes back.
//
// A widget claims the capture by reporting an active drag from
// mouseCaptureTarget. Containers forward that query to their visible children,
// the same way ModalMouseTarget walks the tree, and report the claimed widget's
// origin in their own coordinates. While a capture is active, Pane routes drag
// and release events straight to that widget. A press, a hover or Esc ends the
// capture; a release ends it on the widget itself.
//
// The contract is unexported, so only loom widgets claim captures.

// mouseCapturer is implemented by widgets that hold a drag capture and by
// containers that forward the query to their children. The returned origin is
// the claimed widget's top-left in the receiver's coordinates.
type mouseCapturer interface {
	mouseCaptureTarget() (Widget, Rect)
}

// mouseCaptureEnder is implemented by widgets that hold a capture. cancel
// restores the state from before the press, as Esc does; otherwise the
// capture ends where the pointer is.
type mouseCaptureEnder interface {
	endMouseCapture(cancel bool)
}

// mouseCaptureOf resolves the widget holding a drag capture below root. It
// returns nil when no widget claims one.
func mouseCaptureOf(root Widget) (Widget, Rect) {
	for root != nil {
		if capturer, ok := root.(mouseCapturer); ok {
			return capturer.mouseCaptureTarget()
		}
		wrapper, ok := root.(WidgetUnwrapper)
		if !ok {
			break
		}
		root = wrapper.Unwrap()
	}
	return nil, Rect{}
}

// dispatchCapturedMouse routes a drag or release to the widget holding a
// capture, whatever the pointer position. A press or hover ends the capture
// and returns handled=false, so the event takes ordinary routing.
func dispatchCapturedMouse(root Widget, e MouseEvent) (result EventResult, handled bool) {
	target, origin := mouseCaptureOf(root)
	if target == nil {
		return Ignored(), false
	}
	switch e.Action {
	case MouseDrag, MouseRelease:
		e.X -= origin.X
		e.Y -= origin.Y
		result = target.ConsumeMouse(e)
		result.Consumed = true
		return result, true
	case MousePress, MouseHover:
		endCapture(target, false)
	}
	return Ignored(), false
}

// cancelCapturedMouse ends an active capture for Esc. It reports whether a
// capture was active, in which case the key is consumed.
func cancelCapturedMouse(root Widget) bool {
	target, _ := mouseCaptureOf(root)
	if target == nil {
		return false
	}
	endCapture(target, true)
	return true
}

func endCapture(target Widget, cancel bool) {
	if ender, ok := target.(mouseCaptureEnder); ok {
		ender.endMouseCapture(cancel)
	}
}

// childMouseCapture forwards the capture query to child, whose allocation sits
// at allocation inside parent. It is nil when the child is not visible.
func childMouseCapture(child Widget, allocation, parent Rect) (Widget, Rect) {
	if child == nil || allocation.W <= 0 || allocation.H <= 0 || !rectsOverlap(allocation, parent) {
		return nil, Rect{}
	}
	target, origin := mouseCaptureOf(child)
	if target != nil {
		origin.X += allocation.X - parent.X
		origin.Y += allocation.Y - parent.Y
	}
	return target, origin
}

// mouseCaptureTarget reports the capture of the last drawn child that holds one.
func (s *Stack) mouseCaptureTarget() (Widget, Rect) {
	return childrenMouseCapture(s.Children, s.childRects, s.lastRect)
}

// mouseCaptureTarget reports the capture of the last visible child that holds one.
func (g *Grid) mouseCaptureTarget() (Widget, Rect) {
	return childrenMouseCapture(g.Children, g.childRects, g.lastRect)
}

func childrenMouseCapture(children []Widget, allocations []Rect, parent Rect) (Widget, Rect) {
	for i := min(len(children), len(allocations)) - 1; i >= 0; i-- {
		if target, origin := childMouseCapture(children[i], allocations[i], parent); target != nil {
			return target, origin
		}
	}
	return nil, Rect{}
}

// mouseCaptureTarget reports the capture of either split child.
func (s *Split) mouseCaptureTarget() (Widget, Rect) {
	if target, origin := childMouseCapture(s.Second, s.secondRect, s.lastRect); target != nil {
		return target, origin
	}
	return childMouseCapture(s.First, s.firstRect, s.lastRect)
}

// mouseCaptureTarget reports the capture of the active tab's panel.
func (t *Tabs) mouseCaptureTarget() (Widget, Rect) {
	if !t.drawn {
		return nil, Rect{}
	}
	return childMouseCapture(t.active(), t.childRect(t.lastRect), t.lastRect)
}

// mouseCaptureTarget reports a scrollbar drag, or the capture of the child
// translated through the current scroll position.
func (v *Viewport) mouseCaptureTarget() (Widget, Rect) {
	if v.dragging {
		return v, Rect{}
	}
	if v.content.W <= 0 || v.content.H <= 0 {
		return nil, Rect{}
	}
	target, origin := mouseCaptureOf(v.Child)
	if target != nil {
		origin.X -= v.ScrollX
		origin.Y -= v.ScrollY
	}
	return target, origin
}

// mouseCaptureTarget reports the capture of the child in the visible frame box.
func (f *Frame) mouseCaptureTarget() (Widget, Rect) {
	allocations := f.Layout(f.lastRect.W, f.lastRect.H)
	parent := Rect{W: f.lastRect.W, H: f.lastRect.H}
	for i := len(allocations) - 1; i >= 0; i-- {
		box, area := f.Boxes[i], allocations[i]
		if box.Hidden || area.W < 2 || area.H < 2 {
			continue
		}
		padding := max(0, box.Padding)
		inner := Rect{X: area.X + 1 + padding, Y: area.Y + 1 + padding,
			W: area.W - 2 - 2*padding, H: area.H - 2 - 2*padding}
		if target, origin := childMouseCapture(box.Child, inner, parent); target != nil {
			return target, origin
		}
	}
	return nil, Rect{}
}

// mouseCaptureTarget reports the capture of the last visible form field.
func (f *Form) mouseCaptureTarget() (Widget, Rect) {
	for i := min(len(f.Fields), len(f.fieldRects)) - 1; i >= 0; i-- {
		child, _ := f.Fields[i].Widget.(Widget)
		if target, origin := childMouseCapture(child, f.fieldRects[i], f.lastRect); target != nil {
			return target, origin
		}
	}
	return nil, Rect{}
}

// mouseCaptureTarget reports the capture of the visible child.
func (b *Box) mouseCaptureTarget() (Widget, Rect) {
	if b.Hidden || b.childRect.W <= 0 || b.childRect.H <= 0 {
		return nil, Rect{}
	}
	target, origin := mouseCaptureOf(b.Child)
	if target != nil {
		origin.X += b.childRect.X
		origin.Y += b.childRect.Y
	}
	return target, origin
}
