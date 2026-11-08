// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

// captureProbe records the events routed to it. It never claims a capture.
type captureProbe struct {
	mouse []MouseEvent
	keys  []KeyEvent
}

func (p *captureProbe) Draw(*Canvas, Rect) {}

func (p *captureProbe) ConsumeKey(e KeyEvent) EventResult {
	p.keys = append(p.keys, e)
	return Ignored()
}

func (p *captureProbe) ConsumeMouse(e MouseEvent) EventResult {
	p.mouse = append(p.mouse, e)
	return Handled()
}

// captureFixture is a 12x20 vertical stack: a 100-line View on rows 0-9 and a
// non-claiming probe on rows 10-19. The probe is focused so Esc is not seen by
// the View through focus.
type captureFixture struct {
	pane  *Pane
	root  *Stack
	view  *View
	probe *captureProbe
}

func newCaptureFixture(t *testing.T) captureFixture {
	t.Helper()
	view := NewView(make([]string, 100))
	probe := &captureProbe{}
	root := NewStack(Vertical, view, probe)
	root.focus = 1
	root.Draw(NewCanvas(12, 20), Rect{W: 12, H: 20})
	if view.lastRect.H != 10 {
		t.Fatalf("view height=%d, want 10", view.lastRect.H)
	}
	return captureFixture{pane: &Pane{}, root: root, view: view, probe: probe}
}

func (f captureFixture) mouse(action MouseAction, x, y int) {
	f.pane.dispatchMouse(f.root, MouseEvent{Action: action, Button: MouseLeft, X: x, Y: y})
}

// grabThumb presses the scrollbar thumb at the top and drags to row 5, which
// scrolls to offset 50 of 90.
func (f captureFixture) grabThumb(t *testing.T) {
	t.Helper()
	f.mouse(MousePress, 11, 0)
	f.mouse(MouseDrag, 11, 5)
	if f.view.Scroll != 50 || !f.view.drag.active {
		t.Fatalf("thumb grab: scroll=%d active=%v, want 50 true", f.view.Scroll, f.view.drag.active)
	}
}

func TestPaneMouseCaptureDragLeavesViewAndReturns(t *testing.T) {
	f := newCaptureFixture(t)
	f.grabThumb(t)

	// Leaving the View's bounds: the drag stays with the View and clamps at
	// the track end instead of reaching the probe below it.
	f.mouse(MouseDrag, 5, 15)
	if f.view.Scroll != 90 {
		t.Fatalf("drag outside view: scroll=%d, want 90", f.view.Scroll)
	}
	// Returning inside the bounds continues the same drag.
	f.mouse(MouseDrag, 11, 4)
	if f.view.Scroll != 40 {
		t.Fatalf("drag back inside: scroll=%d, want 40", f.view.Scroll)
	}
	f.mouse(MouseRelease, 11, 4)
	if f.view.drag.active || len(f.probe.mouse) != 0 {
		t.Fatalf("release: active=%v probe events=%d, want false 0", f.view.drag.active, len(f.probe.mouse))
	}
}

func TestPaneMouseCaptureReleaseOutsideEndsDrag(t *testing.T) {
	f := newCaptureFixture(t)
	f.grabThumb(t)
	f.mouse(MouseDrag, 5, 15)
	f.mouse(MouseRelease, 5, 15)
	if f.view.drag.active {
		t.Fatal("release outside the view left the drag active")
	}
	if len(f.probe.mouse) != 0 {
		t.Fatalf("release outside the view reached the probe: %d events", len(f.probe.mouse))
	}
	// A later drag over the View must not scroll it: no stuck capture.
	f.mouse(MouseDrag, 11, 2)
	if f.view.Scroll != 90 {
		t.Fatalf("drag after release: scroll=%d, want 90 unchanged", f.view.Scroll)
	}
}

func TestPaneMouseCaptureHoverEndsDrag(t *testing.T) {
	f := newCaptureFixture(t)
	f.grabThumb(t)
	// foot reports hover, not drag, once the pointer leaves the window.
	f.mouse(MouseHover, 5, 15)
	if f.view.drag.active {
		t.Fatal("hover left the drag capture active")
	}
	if f.view.Scroll != 50 {
		t.Fatalf("hover moved the view: scroll=%d, want 50", f.view.Scroll)
	}
	// The hover itself takes ordinary routing and reaches the probe. A later
	// drag there follows normal routing too, so the probe gets it as well.
	f.mouse(MouseDrag, 5, 15)
	if f.view.Scroll != 50 || len(f.probe.mouse) != 2 || f.probe.mouse[0].Action != MouseHover {
		t.Fatalf("after hover: scroll=%d probe=%v, want 50 and [hover, drag]", f.view.Scroll, f.probe.mouse)
	}
}

func TestPaneMouseCaptureEscEndsDragAndRestoresScroll(t *testing.T) {
	f := newCaptureFixture(t)
	f.grabThumb(t)
	f.mouse(MouseDrag, 5, 15)
	if f.view.Scroll != 90 {
		t.Fatalf("drag outside view: scroll=%d, want 90", f.view.Scroll)
	}
	if f.pane.dispatchKey(f.root, KeyEvent{Key: "esc"}) {
		t.Fatal("esc during a capture requested quit")
	}
	if f.view.drag.active || f.view.Scroll != 0 {
		t.Fatalf("esc: active=%v scroll=%d, want false 0", f.view.drag.active, f.view.Scroll)
	}
	if len(f.probe.keys) != 0 {
		t.Fatalf("esc reached the focused probe during a capture: %d keys", len(f.probe.keys))
	}
	// After Esc the probe gets its keys and the View gets no drag.
	f.mouse(MouseDrag, 11, 4)
	if f.view.Scroll != 0 {
		t.Fatalf("drag after esc: scroll=%d, want 0", f.view.Scroll)
	}
}

func TestPaneEscWithoutCaptureReachesFocusedWidget(t *testing.T) {
	f := newCaptureFixture(t)
	if f.pane.dispatchKey(f.root, KeyEvent{Key: "esc"}) {
		t.Fatal("esc requested quit")
	}
	if len(f.probe.keys) != 1 || f.probe.keys[0].Key != "esc" {
		t.Fatalf("focused probe keys=%v, want one esc", f.probe.keys)
	}
}

func TestPaneMouseCaptureLeavesNonClaimingWidgetsUnchanged(t *testing.T) {
	left := &captureProbe{}
	right := &captureProbe{}
	root := NewStack(Vertical, left, right)
	root.Draw(NewCanvas(12, 20), Rect{W: 12, H: 20})
	p := &Pane{}

	p.dispatchMouse(root, MouseEvent{Action: MousePress, Button: MouseLeft, X: 3, Y: 2})
	p.dispatchMouse(root, MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 3, Y: 12})
	p.dispatchMouse(root, MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 3, Y: 12})

	// Hit-testing still decides: the press and the drag/release over each
	// probe's rows go where they went before the capture existed.
	if len(left.mouse) != 1 || left.mouse[0].Action != MousePress {
		t.Fatalf("left events=%v, want one press", left.mouse)
	}
	if len(right.mouse) != 2 || right.mouse[0].Action != MouseDrag || right.mouse[1].Action != MouseRelease {
		t.Fatalf("right events=%v, want drag and release", right.mouse)
	}
	if right.mouse[0].Y != 2 {
		t.Fatalf("right drag Y=%d, want child-local 2", right.mouse[0].Y)
	}
}

// tallProbe is a non-claiming child taller than its viewport.
type tallProbe struct{ captureProbe }

func (*tallProbe) ContentHeight() int { return 40 }

// newViewportCaptureFixture stacks a 10-row viewport over a probe, so drags
// outside the viewport would reach the probe without a pane-level capture.
func newViewportCaptureFixture(t *testing.T) (*Pane, *Stack, *Viewport, *captureProbe) {
	t.Helper()
	child := &tallProbe{}
	probe := &captureProbe{}
	vp := NewViewport(child)
	vp.ScrollbarMode = ScrollbarAlways
	root := NewStack(Vertical, vp, probe)
	root.focus = 1
	root.Draw(NewCanvas(12, 20), Rect{W: 12, H: 20})
	if vp.lastRect.H != 10 || vp.childH != 40 {
		t.Fatalf("viewport H=%d childH=%d, want 10 and 40", vp.lastRect.H, vp.childH)
	}
	return &Pane{}, root, vp, probe
}

func TestPaneMouseCaptureViewportScrollbarLeavesAndReturns(t *testing.T) {
	p, root, vp, probe := newViewportCaptureFixture(t)
	press := func(action MouseAction, x, y int) {
		p.dispatchMouse(root, MouseEvent{Action: action, Button: MouseLeft, X: x, Y: y})
	}
	press(MousePress, 11, 0)
	press(MouseDrag, 11, 5)
	// The 3-row thumb has 7 rows of travel; the thumb follows the pointer
	// (5 rows down = 5*30/7 rows of content). The old 16 came from mapping the
	// pointer straight onto the whole track, ignoring the thumb length.
	if vp.ScrollY != 21 {
		t.Fatalf("drag inside: ScrollY=%d, want 21", vp.ScrollY)
	}
	press(MouseDrag, 5, 25)
	if vp.ScrollY != 30 {
		t.Fatalf("drag outside: ScrollY=%d, want 30 (track end)", vp.ScrollY)
	}
	press(MouseRelease, 5, 25)
	if vp.dragging || len(probe.mouse) != 0 {
		t.Fatalf("release: dragging=%v probe=%d, want false 0", vp.dragging, len(probe.mouse))
	}
}

func TestPaneMouseCaptureViewportEscRestoresScroll(t *testing.T) {
	p, root, vp, probe := newViewportCaptureFixture(t)
	p.dispatchMouse(root, MouseEvent{Action: MousePress, Button: MouseLeft, X: 11, Y: 0})
	p.dispatchMouse(root, MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 5, Y: 25})
	if p.dispatchKey(root, KeyEvent{Key: "esc"}) {
		t.Fatal("esc during a viewport capture requested quit")
	}
	if vp.dragging || vp.ScrollY != 0 || len(probe.keys) != 0 {
		t.Fatalf("esc: dragging=%v ScrollY=%d probeKeys=%d, want false 0 0", vp.dragging, vp.ScrollY, len(probe.keys))
	}
}

// TestPaneMouseCaptureEndedOutsideDoesNotLeakIntoFrame covers a drag that the
// pane capture ends outside the frame. The release never reaches Frame, so
// Frame's own capture state must not route the next drag to the old box.
func TestPaneMouseCaptureEndedOutsideDoesNotLeakIntoFrame(t *testing.T) {
	view := NewView(make([]string, 100))
	probe := &captureProbe{}
	frame := &Frame{Boxes: []Box{{Child: view, Width: 15, Height: 4}, {Child: probe, Width: 15, Height: 4}}}
	frame.Draw(NewCanvas(30, 14), Rect{W: 30, H: 14})
	// Box areas are {0,1,15,4} and {15,1,15,4}; each child sits one cell in,
	// so the view's scrollbar column is x=13 and its top row is y=2 in frame
	// coordinates.
	if want := (Rect{X: 0, Y: 0, W: 13, H: 2}); view.lastRect != want {
		t.Fatalf("view rect=%+v, want %+v (child-local)", view.lastRect, want)
	}
	thumbX, thumbY := 13, 2
	p := &Pane{}
	press := func(action MouseAction, x, y int) {
		p.dispatchMouse(frame, MouseEvent{Action: action, Button: MouseLeft, X: x, Y: y})
	}
	press(MousePress, thumbX, thumbY)
	press(MouseDrag, thumbX, thumbY+1)
	press(MouseDrag, 5, 20) // outside the frame: the pane capture keeps the drag
	press(MouseRelease, 5, 20)
	if view.drag.active {
		t.Fatal("release outside left the view's drag active")
	}

	// A press on the frame border outside every box, then a drag over the probe.
	press(MousePress, 0, 0)
	press(MouseDrag, 20, 3)
	if len(probe.mouse) != 1 || probe.mouse[0].Action != MouseDrag {
		t.Fatalf("drag after a border press: probe events=%v, want one drag", probe.mouse)
	}
}

func TestViewportScrollbarPressOnThumbKeepsOffset(t *testing.T) {
	p, root, vp, _ := newViewportCaptureFixture(t)
	vp.ScrollY = 15 // thumb start 4 (15*7/30 rounded), rows 4..6
	p.dispatchMouse(root, MouseEvent{Action: MousePress, Button: MouseLeft, X: 11, Y: 6})
	if vp.ScrollY != 15 {
		t.Fatalf("press on thumb moved ScrollY to %d, want 15", vp.ScrollY)
	}
}

func TestViewportScrollbarTrackPressCentersThumb(t *testing.T) {
	p, root, vp, _ := newViewportCaptureFixture(t)
	p.dispatchMouse(root, MouseEvent{Action: MousePress, Button: MouseLeft, X: 11, Y: 8})
	// thumb centered on row 8: start 7 (clamped to travel 7) -> max offset.
	if vp.ScrollY != 30 || !vp.dragging {
		t.Fatalf("track press: ScrollY=%d dragging=%v, want 30 true", vp.ScrollY, vp.dragging)
	}
}

func TestViewportScrollbarThumbTravelReachesEnd(t *testing.T) {
	p, root, vp, _ := newViewportCaptureFixture(t)
	p.dispatchMouse(root, MouseEvent{Action: MousePress, Button: MouseLeft, X: 11, Y: 1}) // middle of rows 0..2
	p.dispatchMouse(root, MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 11, Y: 8})  // 7 rows of travel
	if vp.ScrollY != 30 {
		t.Fatalf("after 7 rows of travel ScrollY=%d, want the end (30)", vp.ScrollY)
	}
}
