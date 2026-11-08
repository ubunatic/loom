// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

// ignoringTall is a 40-row child that ignores the mouse, so a middle press
// over it falls through to the Viewport and starts a pan.
type ignoringTall struct{ mouse []MouseEvent }

func (*ignoringTall) Draw(*Canvas, Rect) {}

func (*ignoringTall) ContentHeight() int { return 40 }

func (c *ignoringTall) ConsumeKey(KeyEvent) EventResult { return Ignored() }

func (c *ignoringTall) ConsumeMouse(e MouseEvent) EventResult {
	c.mouse = append(c.mouse, e)
	return Ignored()
}

// redraw repaints the fixture the way the pane does between input reads.
func (f captureFixture) redraw() {
	f.root.Draw(NewCanvas(12, 20), Rect{W: 12, H: 20})
}

// button routes one mouse event of the given button through the pane.
func (f captureFixture) button(action MouseAction, button MouseButton, x, y int) {
	f.pane.dispatchMouse(f.root, MouseEvent{Action: action, Button: button, X: x, Y: y})
}

// newPanFixture scrolls the 100-line View to offset 20 so a pan can move
// both up and down.
func newPanFixture(t *testing.T) captureFixture {
	t.Helper()
	f := newCaptureFixture(t)
	f.view.Scroll = 20
	f.redraw()
	return f
}

func TestPaneMiddlePanDragsViewContent(t *testing.T) {
	f := newPanFixture(t)
	f.button(MousePress, MouseMiddle, 3, 5)
	// Content follows the pointer one line per row: up moves the content up.
	f.button(MouseDrag, MouseMiddle, 3, 2)
	if f.view.Scroll != 23 {
		t.Fatalf("pan up 3 rows: scroll=%d, want 23", f.view.Scroll)
	}
	f.button(MouseDrag, MouseMiddle, 3, 8)
	if f.view.Scroll != 17 {
		t.Fatalf("pan down 3 rows from the press: scroll=%d, want 17", f.view.Scroll)
	}
	f.button(MouseRelease, MouseMiddle, 3, 8)
	if f.view.pan.active || len(f.probe.mouse) != 0 {
		t.Fatalf("release: active=%v probe=%d, want false 0", f.view.pan.active, len(f.probe.mouse))
	}
	f.button(MouseDrag, MouseLeft, 3, 2)
	if f.view.Scroll != 17 {
		t.Fatalf("drag after release: scroll=%d, want 17 unchanged", f.view.Scroll)
	}
}

func TestPaneMiddlePanLeavesViewAndReturns(t *testing.T) {
	f := newPanFixture(t)
	f.button(MousePress, MouseMiddle, 3, 5)
	// Below the View the pan keeps the pointer and never reaches the probe.
	f.button(MouseDrag, MouseMiddle, 3, 15)
	if f.view.Scroll != 10 {
		t.Fatalf("drag outside view: scroll=%d, want 10", f.view.Scroll)
	}
	f.button(MouseDrag, MouseMiddle, 3, 4)
	if f.view.Scroll != 21 {
		t.Fatalf("drag back inside: scroll=%d, want 21", f.view.Scroll)
	}
	f.button(MouseDrag, MouseMiddle, 3, 100)
	if f.view.Scroll != 0 {
		t.Fatalf("drag far past the top: scroll=%d, want clamped 0", f.view.Scroll)
	}
	f.button(MouseRelease, MouseMiddle, 3, 100)
	if len(f.probe.mouse) != 0 {
		t.Fatalf("pan reached the probe: %d events", len(f.probe.mouse))
	}
}

func TestPaneMiddlePanEscRestoresScroll(t *testing.T) {
	f := newPanFixture(t)
	f.button(MousePress, MouseMiddle, 3, 5)
	f.button(MouseDrag, MouseMiddle, 3, 0)
	if f.view.Scroll != 25 {
		t.Fatalf("pan before esc: scroll=%d, want 25", f.view.Scroll)
	}
	if f.pane.dispatchKey(f.root, KeyEvent{Key: "esc"}) {
		t.Fatal("esc during a middle pan requested quit")
	}
	if f.view.pan.active || f.view.Scroll != 20 || len(f.probe.keys) != 0 {
		t.Fatalf("esc: active=%v scroll=%d probeKeys=%d, want false 20 0",
			f.view.pan.active, f.view.Scroll, len(f.probe.keys))
	}
	f.button(MouseDrag, MouseMiddle, 3, 2)
	if f.view.Scroll != 20 {
		t.Fatalf("drag after esc: scroll=%d, want 20", f.view.Scroll)
	}
}

// TestViewMiddlePanEscKeyCancelsWithoutPane covers the View's own key path,
// for a host that routes keys to the View without a Pane.
func TestViewMiddlePanEscKeyCancelsWithoutPane(t *testing.T) {
	v := NewView(make([]string, 100))
	v.Draw(NewCanvas(12, 10), Rect{W: 12, H: 10})
	v.Scroll = 20
	v.Draw(NewCanvas(12, 10), Rect{W: 12, H: 10})
	v.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseMiddle, X: 3, Y: 5})
	v.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseMiddle, X: 3, Y: 0})
	if v.Scroll != 25 {
		t.Fatalf("pan: scroll=%d, want 25", v.Scroll)
	}
	v.ConsumeKey(KeyEvent{Key: "esc"})
	if v.Scroll != 20 || v.pan.active {
		t.Fatalf("esc: scroll=%d active=%v, want 20 false", v.Scroll, v.pan.active)
	}
}

func TestPaneLeftPressDoesNotPan(t *testing.T) {
	f := newPanFixture(t)
	f.button(MousePress, MouseLeft, 3, 5)
	f.button(MouseDrag, MouseLeft, 3, 2)
	if f.view.Scroll != 20 || f.view.pan.active {
		t.Fatalf("left drag: scroll=%d active=%v, want 20 false", f.view.Scroll, f.view.pan.active)
	}
}

func TestPaneMiddlePressOnViewScrollbarColumnDoesNotPan(t *testing.T) {
	f := newPanFixture(t)
	f.button(MousePress, MouseMiddle, 11, 5)
	f.button(MouseDrag, MouseMiddle, 11, 2)
	if f.view.Scroll != 20 || f.view.pan.active {
		t.Fatalf("middle press on scrollbar: scroll=%d active=%v, want 20 false", f.view.Scroll, f.view.pan.active)
	}
}

// TestPaneMiddlePanSurvivesRedraw: the pane redraws between input reads, and
// Stack.Draw re-applies focus to an unfocused View. Focus must not end a pan.
func TestPaneMiddlePanSurvivesRedraw(t *testing.T) {
	f := newPanFixture(t)
	f.button(MousePress, MouseMiddle, 3, 5)
	f.redraw()
	f.button(MouseDrag, MouseMiddle, 3, 2)
	f.redraw()
	f.button(MouseDrag, MouseMiddle, 3, 1)
	if f.view.Scroll != 24 || !f.view.pan.active {
		t.Fatalf("pan across redraws: scroll=%d active=%v, want 24 true", f.view.Scroll, f.view.pan.active)
	}
}

// TestPaneScrollbarDragSurvivesRedraw: the same redraw path must not cancel a
// scrollbar drag on an unfocused View (issue 449 capture).
func TestPaneScrollbarDragSurvivesRedraw(t *testing.T) {
	f := newCaptureFixture(t)
	f.grabThumb(t)
	f.redraw()
	f.button(MouseDrag, MouseLeft, 11, 4)
	if f.view.Scroll != 40 || !f.view.drag.active {
		t.Fatalf("thumb drag across redraw: scroll=%d active=%v, want 40 true", f.view.Scroll, f.view.drag.active)
	}
}

func TestViewFocusLossEndsPanAndRestores(t *testing.T) {
	f := newPanFixture(t)
	f.view.SetFocus(true)
	f.button(MousePress, MouseMiddle, 3, 5)
	f.button(MouseDrag, MouseMiddle, 3, 0)
	if !f.view.pan.active || f.view.Scroll != 25 {
		t.Fatalf("pan before focus loss: active=%v scroll=%d, want true 25", f.view.pan.active, f.view.Scroll)
	}
	f.view.SetFocus(false)
	if f.view.pan.active || f.view.Scroll != 20 {
		t.Fatalf("focus loss: active=%v scroll=%d, want false 20", f.view.pan.active, f.view.Scroll)
	}
}

// newViewportPanFixture stacks a 10-row Viewport over a probe. The 40-row
// child ignores the mouse, so a middle press in the content starts a pan.
func newViewportPanFixture(t *testing.T, child Widget) (*Pane, *Stack, *Viewport, *captureProbe) {
	t.Helper()
	probe := &captureProbe{}
	vp := NewViewport(child)
	vp.ScrollbarMode = ScrollbarAlways
	root := NewStack(Vertical, vp, probe)
	root.focus = 1
	root.Draw(NewCanvas(12, 20), Rect{W: 12, H: 20})
	vp.ScrollY = 20
	root.Draw(NewCanvas(12, 20), Rect{W: 12, H: 20})
	if vp.lastRect.H != 10 || vp.childH != 40 {
		t.Fatalf("viewport H=%d childH=%d, want 10 and 40", vp.lastRect.H, vp.childH)
	}
	return &Pane{}, root, vp, probe
}

func TestPaneMiddlePanDragsViewportContent(t *testing.T) {
	child := &ignoringTall{}
	p, root, vp, probe := newViewportPanFixture(t, child)
	press := func(action MouseAction, x, y int) {
		p.dispatchMouse(root, MouseEvent{Action: action, Button: MouseMiddle, X: x, Y: y})
	}
	press(MousePress, 3, 5)
	press(MouseDrag, 3, 2)
	if vp.ScrollY != 23 {
		t.Fatalf("pan up 3 rows: ScrollY=%d, want 23", vp.ScrollY)
	}
	press(MouseDrag, 3, 8)
	if vp.ScrollY != 17 {
		t.Fatalf("pan down 3 rows from the press: ScrollY=%d, want 17", vp.ScrollY)
	}
	press(MouseRelease, 3, 8)
	if vp.pan.active || len(probe.mouse) != 0 {
		t.Fatalf("release: active=%v probe=%d, want false 0", vp.pan.active, len(probe.mouse))
	}
	if len(child.mouse) != 1 || child.mouse[0].Y != 25 {
		t.Fatalf("child mouse=%v, want one press at child-local row 25", child.mouse)
	}
}

func TestPaneMiddlePanLeavesViewportAndReturns(t *testing.T) {
	p, root, vp, probe := newViewportPanFixture(t, &ignoringTall{})
	press := func(action MouseAction, x, y int) {
		p.dispatchMouse(root, MouseEvent{Action: action, Button: MouseMiddle, X: x, Y: y})
	}
	press(MousePress, 3, 5)
	press(MouseDrag, 3, 15)
	if vp.ScrollY != 10 {
		t.Fatalf("drag below viewport: ScrollY=%d, want 10", vp.ScrollY)
	}
	press(MouseDrag, 3, 4)
	if vp.ScrollY != 21 {
		t.Fatalf("drag back inside: ScrollY=%d, want 21", vp.ScrollY)
	}
	press(MouseRelease, 3, 4)
	if len(probe.mouse) != 0 {
		t.Fatalf("pan reached the probe: %d events", len(probe.mouse))
	}
}

func TestPaneMiddlePanEscRestoresViewportScroll(t *testing.T) {
	p, root, vp, probe := newViewportPanFixture(t, &ignoringTall{})
	p.dispatchMouse(root, MouseEvent{Action: MousePress, Button: MouseMiddle, X: 3, Y: 5})
	p.dispatchMouse(root, MouseEvent{Action: MouseDrag, Button: MouseMiddle, X: 3, Y: 0})
	if vp.ScrollY != 25 {
		t.Fatalf("pan before esc: ScrollY=%d, want 25", vp.ScrollY)
	}
	if p.dispatchKey(root, KeyEvent{Key: "esc"}) {
		t.Fatal("esc during a viewport pan requested quit")
	}
	if vp.pan.active || vp.ScrollY != 20 || len(probe.keys) != 0 {
		t.Fatalf("esc: active=%v ScrollY=%d probeKeys=%d, want false 20 0", vp.pan.active, vp.ScrollY, len(probe.keys))
	}
}

// TestPaneMiddlePressConsumedByChildIsNotPan: a child that consumes the middle
// press keeps it; the Viewport does not pan.
func TestPaneMiddlePressConsumedByChildIsNotPan(t *testing.T) {
	child := &tallProbe{}
	p, root, vp, _ := newViewportPanFixture(t, child)
	p.dispatchMouse(root, MouseEvent{Action: MousePress, Button: MouseMiddle, X: 3, Y: 5})
	p.dispatchMouse(root, MouseEvent{Action: MouseDrag, Button: MouseMiddle, X: 3, Y: 2})
	if vp.pan.active || vp.ScrollY != 20 {
		t.Fatalf("child consumed press: active=%v ScrollY=%d, want false 20", vp.pan.active, vp.ScrollY)
	}
	if len(child.mouse) != 2 || child.mouse[0].Y != 25 {
		t.Fatalf("child mouse=%v, want the press at child-local row 25 and the drag", child.mouse)
	}
}

func TestPaneMiddlePressOnViewportScrollbarDoesNotPan(t *testing.T) {
	p, root, vp, _ := newViewportPanFixture(t, &ignoringTall{})
	p.dispatchMouse(root, MouseEvent{Action: MousePress, Button: MouseMiddle, X: 11, Y: 5})
	p.dispatchMouse(root, MouseEvent{Action: MouseDrag, Button: MouseMiddle, X: 11, Y: 2})
	if vp.pan.active || vp.ScrollY != 20 {
		t.Fatalf("middle press on scrollbar: active=%v ScrollY=%d, want false 20", vp.pan.active, vp.ScrollY)
	}
}

func TestViewportFocusLossEndsPanAndRestores(t *testing.T) {
	p, root, vp, _ := newViewportPanFixture(t, &ignoringTall{})
	vp.SetFocus(true)
	p.dispatchMouse(root, MouseEvent{Action: MousePress, Button: MouseMiddle, X: 3, Y: 5})
	p.dispatchMouse(root, MouseEvent{Action: MouseDrag, Button: MouseMiddle, X: 3, Y: 0})
	if !vp.pan.active || vp.ScrollY != 25 {
		t.Fatalf("pan before focus loss: active=%v ScrollY=%d, want true 25", vp.pan.active, vp.ScrollY)
	}
	vp.SetFocus(false)
	if vp.pan.active || vp.ScrollY != 20 {
		t.Fatalf("focus loss: active=%v ScrollY=%d, want false 20", vp.pan.active, vp.ScrollY)
	}
}
