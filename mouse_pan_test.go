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
	// The pointer row maps onto the whole range (10 rows, offsets 0..90, stops
	// at rows 1 and 8) and the press row (5) keeps offset 20: dragging up
	// scrolls toward the top, dragging down toward the end.
	f.button(MouseDrag, MouseMiddle, 3, 2)
	if f.view.Scroll != 5 {
		t.Fatalf("pan up to row 2: scroll=%d, want 5", f.view.Scroll)
	}
	f.button(MouseDrag, MouseMiddle, 3, 7)
	if f.view.Scroll != 67 {
		t.Fatalf("pan down to row 7: scroll=%d, want 67", f.view.Scroll)
	}
	f.button(MouseRelease, MouseMiddle, 3, 8)
	if f.view.pan.active || len(f.probe.mouse) != 0 {
		t.Fatalf("release: active=%v probe=%d, want false 0", f.view.pan.active, len(f.probe.mouse))
	}
	f.button(MouseDrag, MouseLeft, 3, 2)
	if f.view.Scroll != 67 {
		t.Fatalf("drag after release: scroll=%d, want 67 unchanged", f.view.Scroll)
	}
}

func TestPaneMiddlePanLeavesViewAndReturns(t *testing.T) {
	f := newPanFixture(t)
	f.button(MousePress, MouseMiddle, 3, 5)
	// Below the View the pan keeps the pointer and never reaches the probe.
	f.button(MouseDrag, MouseMiddle, 3, 15)
	if f.view.Scroll != 90 {
		t.Fatalf("drag below the view: scroll=%d, want the end (90)", f.view.Scroll)
	}
	f.button(MouseDrag, MouseMiddle, 3, 4)
	if f.view.Scroll != 15 {
		t.Fatalf("drag back inside: scroll=%d, want 15", f.view.Scroll)
	}
	f.button(MouseDrag, MouseMiddle, 3, -3)
	if f.view.Scroll != 0 {
		t.Fatalf("drag above the view: scroll=%d, want the top (0)", f.view.Scroll)
	}
	f.button(MouseRelease, MouseMiddle, 3, -3)
	if len(f.probe.mouse) != 0 {
		t.Fatalf("pan reached the probe: %d events", len(f.probe.mouse))
	}
}

func TestPaneMiddlePanEscRestoresScroll(t *testing.T) {
	f := newPanFixture(t)
	f.button(MousePress, MouseMiddle, 3, 5)
	f.button(MouseDrag, MouseMiddle, 3, 0)
	if f.view.Scroll != 0 {
		t.Fatalf("pan before esc: scroll=%d, want 0 (top row)", f.view.Scroll)
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
	if v.Scroll != 0 {
		t.Fatalf("pan: scroll=%d, want 0 (top row)", v.Scroll)
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
	if f.view.Scroll != 0 || !f.view.pan.active {
		t.Fatalf("pan across redraws: scroll=%d active=%v, want 0 true", f.view.Scroll, f.view.pan.active)
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
	if !f.view.pan.active || f.view.Scroll != 0 {
		t.Fatalf("pan before focus loss: active=%v scroll=%d, want true 0", f.view.pan.active, f.view.Scroll)
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
	if vp.ScrollY != 5 {
		t.Fatalf("pan up to row 2: ScrollY=%d, want 5", vp.ScrollY)
	}
	press(MouseDrag, 3, 7)
	if vp.ScrollY != 27 {
		t.Fatalf("pan down to row 7: ScrollY=%d, want 27", vp.ScrollY)
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
	if vp.ScrollY != 30 {
		t.Fatalf("drag below viewport: ScrollY=%d, want the end (30)", vp.ScrollY)
	}
	press(MouseDrag, 3, 4)
	if vp.ScrollY != 15 {
		t.Fatalf("drag back inside: ScrollY=%d, want 15", vp.ScrollY)
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
	if vp.ScrollY != 0 {
		t.Fatalf("pan before esc: ScrollY=%d, want 0 (top row)", vp.ScrollY)
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
	if !vp.pan.active || vp.ScrollY != 0 {
		t.Fatalf("pan before focus loss: active=%v ScrollY=%d, want true 0", vp.pan.active, vp.ScrollY)
	}
	vp.SetFocus(false)
	if vp.pan.active || vp.ScrollY != 20 {
		t.Fatalf("focus loss: active=%v ScrollY=%d, want false 20", vp.pan.active, vp.ScrollY)
	}
}

func TestMiddlePanOffsetMapsWholeRange(t *testing.T) {
	// 10 rows, offsets 0..90, stops at rows 1 and 8, pressed at row 5 holding offset 20.
	p := contentPan{start: 20, anchor: 5, rows: 10}
	for _, c := range []struct{ row, want int }{
		{-4, 0}, {0, 0}, {1, 0}, {5, 20}, {8, 90}, {9, 90}, {14, 90},
	} {
		if got := p.offset(c.row, 90); got != c.want {
			t.Errorf("row %d: offset=%d, want %d", c.row, got, c.want)
		}
	}
	// A press on the first row cannot go further up and a press on the last row
	// cannot go further down: no division by zero.
	if got := (contentPan{start: 7, anchor: 0, rows: 10}).offset(0, 90); got != 7 {
		t.Errorf("press on the top row: offset=%d, want 7", got)
	}
	if got := (contentPan{start: 90, anchor: 9, rows: 10}).offset(9, 90); got != 90 {
		t.Errorf("press on the last row: offset=%d, want 90", got)
	}
}

func TestMiddlePanStopsOneRowInside(t *testing.T) {
	// Without leaving the widget: row 1 is the top and row rows-2 the end.
	p := contentPan{start: 40, anchor: 5, rows: 12}
	if got := p.offset(1, 90); got != 0 {
		t.Errorf("row 1: offset=%d, want 0", got)
	}
	if got := p.offset(10, 90); got != 90 {
		t.Errorf("row 10 (rows-2): offset=%d, want 90", got)
	}
	// A view under 4 rows keeps the full-height map.
	small := contentPan{start: 5, anchor: 1, rows: 3}
	if got := small.offset(2, 10); got != 10 {
		t.Errorf("3-row view, last row: offset=%d, want 10", got)
	}
}
