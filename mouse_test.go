// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

// ── A1: DecodeMouse — SGR (1006) byte → MouseEvent ────────────────────────────

func TestDecodeMouseScroll(t *testing.T) {
	up, ok := loom.DecodeMouse([]byte("\x1b[<64;10;5M"))
	if !ok {
		t.Fatal("scroll-up report should decode")
	}
	if up.Action != loom.MouseScrollUp || up.Button != loom.MouseNone {
		t.Errorf("scroll up = {%v,%v}, want {ScrollUp,None}", up.Action, up.Button)
	}
	if up.X != 10 || up.Y != 5 {
		t.Errorf("scroll up coords = (%d,%d), want (10,5)", up.X, up.Y)
	}

	down, ok := loom.DecodeMouse([]byte("\x1b[<65;1;1M"))
	if !ok || down.Action != loom.MouseScrollDown {
		t.Errorf("scroll down = {%v,ok=%v}, want ScrollDown", down.Action, ok)
	}
}

func TestDecodeMousePressRelease(t *testing.T) {
	press, ok := loom.DecodeMouse([]byte("\x1b[<0;10;5M")) // trailing M = press
	if !ok {
		t.Fatal("press report should decode")
	}
	if press.Action != loom.MousePress || press.Button != loom.MouseLeft {
		t.Errorf("press = {%v,%v}, want {Press,Left}", press.Action, press.Button)
	}
	if press.X != 10 || press.Y != 5 {
		t.Errorf("press coords = (%d,%d), want (10,5)", press.X, press.Y)
	}

	release, ok := loom.DecodeMouse([]byte("\x1b[<0;10;5m")) // trailing m = release
	if !ok || release.Action != loom.MouseRelease || release.Button != loom.MouseLeft {
		t.Errorf("release = {%v,%v,ok=%v}, want {Release,Left}", release.Action, release.Button, ok)
	}
}

func TestDecodeMouseButtons(t *testing.T) {
	mid, ok := loom.DecodeMouse([]byte("\x1b[<1;2;3M"))
	if !ok || mid.Button != loom.MouseMiddle {
		t.Errorf("cb=1 button = %v, want Middle", mid.Button)
	}
	right, ok := loom.DecodeMouse([]byte("\x1b[<2;2;3M"))
	if !ok || right.Button != loom.MouseRight {
		t.Errorf("cb=2 button = %v, want Right", right.Button)
	}
}

func TestDecodeMouseHoverAndDrag(t *testing.T) {
	// motion bit (32) + button 3 (none) = 35 → hover.
	hover, ok := loom.DecodeMouse([]byte("\x1b[<35;7;8M"))
	if !ok {
		t.Fatal("hover report should decode")
	}
	if hover.Action != loom.MouseHover || hover.Button != loom.MouseNone {
		t.Errorf("hover = {%v,%v}, want {Hover,None}", hover.Action, hover.Button)
	}

	// motion bit (32) + button 0 (left) = 32 → drag with left button.
	drag, ok := loom.DecodeMouse([]byte("\x1b[<32;7;8M"))
	if !ok || drag.Action != loom.MouseDrag || drag.Button != loom.MouseLeft {
		t.Errorf("drag = {%v,%v,ok=%v}, want {Drag,Left}", drag.Action, drag.Button, ok)
	}
}

func TestDecodeMouseNotAReport(t *testing.T) {
	cases := [][]byte{
		[]byte("\x1b[A"),      // arrow key, not a mouse report
		[]byte("\x1b[<0;1;1"), // missing M/m terminator
		[]byte("hi"),          // plain text
		[]byte("\x1b[<0;1M"),  // only two ints — incomplete
		{27, '[', '<'},        // too short
	}
	for _, b := range cases {
		if _, ok := loom.DecodeMouse(b); ok {
			t.Errorf("DecodeMouse(%q) = ok, want not-ok", b)
		}
	}
}

// ── A2: HandleMouse per widget (synthetic events) ─────────────────────────────

// spyWidget records the mouse events forwarded to it.
type spyWidget struct {
	got  bool
	last loom.MouseEvent
	quit bool
}

func (s *spyWidget) Draw(*loom.Canvas, loom.Rect) {}
func (s *spyWidget) HandleKey(loom.KeyEvent) bool { return false }
func (s *spyWidget) HandleMouse(e loom.MouseEvent) bool {
	s.got = true
	s.last = e
	return s.quit
}

func TestViewHandleMouseScroll(t *testing.T) {
	v := loom.NewView([]string{"a", "b", "c", "d"})
	c := loom.NewCanvas(4, 2) // height 2 < 4 lines → scrollable, maxScroll=2
	v.Draw(c, c.Bounds())     // sets lastH so HandleMouse can clamp

	v.HandleMouse(loom.MouseEvent{Action: loom.MouseScrollDown})
	if v.Scroll != 1 {
		t.Errorf("Scroll = %d after wheel down, want 1", v.Scroll)
	}
	v.HandleMouse(loom.MouseEvent{Action: loom.MouseScrollDown})
	v.HandleMouse(loom.MouseEvent{Action: loom.MouseScrollDown}) // clamp at maxScroll=2
	if v.Scroll != 2 {
		t.Errorf("Scroll = %d after over-scroll, want 2 (clamped)", v.Scroll)
	}
	v.HandleMouse(loom.MouseEvent{Action: loom.MouseScrollUp})
	v.HandleMouse(loom.MouseEvent{Action: loom.MouseScrollUp})
	v.HandleMouse(loom.MouseEvent{Action: loom.MouseScrollUp}) // clamp at 0
	if v.Scroll != 0 {
		t.Errorf("Scroll = %d after over-scroll-up, want 0 (clamped)", v.Scroll)
	}
}

func TestChoiceHandleMouseClickSelects(t *testing.T) {
	items := []loom.Item{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	c := loom.NewChoice(items)

	// Left press at Y=1 selects the 2nd row (0-based) and quits (no OnSelect).
	quit := c.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, Y: 1})
	if !quit {
		t.Error("left click on a row should quit (done) when OnSelect is nil")
	}
	if c.FilteredSel() != 1 {
		t.Errorf("sel = %d after click Y=2, want 1", c.FilteredSel())
	}
}

func TestChoiceHandleMouseHover(t *testing.T) {
	items := []loom.Item{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	c := loom.NewChoice(items)

	quit := c.HandleMouse(loom.MouseEvent{Action: loom.MouseHover, Y: 2})
	if quit {
		t.Error("hover should not quit")
	}
	if c.FilteredSel() != 2 {
		t.Errorf("sel = %d after hover Y=2, want 2", c.FilteredSel())
	}

	// Out-of-range Y is a no-op (selection unchanged).
	c.HandleMouse(loom.MouseEvent{Action: loom.MouseHover, Y: 99})
	if c.FilteredSel() != 2 {
		t.Errorf("sel = %d after out-of-range hover, want 2 (unchanged)", c.FilteredSel())
	}
}

func TestChoiceHandleMouseOnSelectNoQuit(t *testing.T) {
	items := []loom.Item{{Name: "a"}, {Name: "b"}}
	var picked string
	c := loom.NewChoice(items)
	c.OnSelect = func(it loom.Item) { picked = it.Name }

	quit := c.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, Y: 0})
	if quit {
		t.Error("click with OnSelect set should not quit")
	}
	if picked != "a" {
		t.Errorf("OnSelect got %q, want a", picked)
	}
}

func TestChoiceInFrameHandleMouseUsesCanvasCoordinates(t *testing.T) {
	choice := loom.NewChoice([]loom.Item{{Name: "a"}, {Name: "b"}, {Name: "c"}})
	choice.SelectOnlyOnClick = true
	frame := &loom.Frame{Boxes: []loom.Box{{Child: choice, Width: 8, Height: 6}}}
	canvas := loom.NewCanvas(24, 14)
	frame.Draw(canvas, loom.Rect{X: 5, Y: 3, W: 16, H: 10})

	// Frame chrome is one row high and the box border is one cell high. The
	// second choice row is therefore at absolute canvas coordinate (7, 6).
	frame.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 7, Y: 6})
	if got := choice.FilteredSel(); got != 1 {
		t.Fatalf("selection after click = %d, want 1", got)
	}
}

func TestGridHandleMouseForwardsToFocusedChild(t *testing.T) {
	a, b := &spyWidget{}, &spyWidget{quit: true}
	g := loom.NewGrid(2, a, b)
	g.HandleKey(loom.KeyEvent{Key: "right"}) // focus → child 1 (b)

	ev := loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 4, Y: 0}
	quit := g.HandleMouse(ev)
	if !b.got || a.got {
		t.Errorf("mouse should reach focused child only: a.got=%v b.got=%v", a.got, b.got)
	}
	if b.last != ev {
		t.Errorf("forwarded event = %+v, want %+v", b.last, ev)
	}
	if !quit {
		t.Error("Grid should propagate child's quit=true")
	}
}

func TestStackHandleMouseForwardsToFocusedChild(t *testing.T) {
	a, b := &spyWidget{}, &spyWidget{}
	s := loom.NewStack(loom.Horizontal, a, b)
	s.HandleKey(loom.KeyEvent{Key: "tab"}) // focus → child 1 (b)

	s.HandleMouse(loom.MouseEvent{Action: loom.MouseHover})
	if !b.got || a.got {
		t.Errorf("stack mouse should reach focused child: a.got=%v b.got=%v", a.got, b.got)
	}
}

func TestPopupHandleMouseForwardsWhileOpen(t *testing.T) {
	inner := &spyWidget{}
	p := loom.NewPopup("t", inner)

	p.HandleMouse(loom.MouseEvent{Action: loom.MousePress})
	if !inner.got {
		t.Error("open popup should forward mouse to inner")
	}

	inner.got = false
	p.Open = false
	if p.HandleMouse(loom.MouseEvent{Action: loom.MousePress}) {
		t.Error("closed popup HandleMouse should return false")
	}
	if inner.got {
		t.Error("closed popup should not forward mouse to inner")
	}
}

func TestRouterHandleMouseForwardsToActiveView(t *testing.T) {
	// The router wires each choice's OnSelect to navigation, so a left click on
	// row 1 ("list") routes to the "list" view. Observing Current() proves the
	// mouse event reached the active view's Choice.
	yamlStr := `
app:
  root: main
  height: 6
views:
  - name: main
    grid: |
      +---+
      |L  |
      +---+
    elements:
      L:
        type: choice
        source: "static:list,other"
    on_key:
      "esc": exit
  - name: list
    grid: |
      +---+
      |V  |
      +---+
    elements:
      V:
        type: view
        static: "list view"
    on_key:
      "esc": back
`
	widget, _, err := loom.BuildWidget(strings.NewReader(yamlStr))
	if err != nil {
		t.Fatalf("BuildWidget failed: %v", err)
	}
	router, ok := widget.(*loom.Router)
	if !ok {
		t.Fatalf("expected *loom.Router, got %T", widget)
	}

	router.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, Y: 0})
	if router.Current() != "list" {
		t.Errorf("after click on row 1, current view = %q, want list", router.Current())
	}
}
