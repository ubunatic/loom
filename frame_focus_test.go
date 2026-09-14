package loom

import "testing"

type focusProbe struct {
	focused bool
	keys    []KeyEvent
	mice    []MouseEvent
	quit    bool
}

func (*focusProbe) Draw(*Canvas, Rect) {}
func (p *focusProbe) HandleMouse(e MouseEvent) bool {
	p.mice = append(p.mice, e)
	return p.quit
}
func (p *focusProbe) HandleKey(k KeyEvent) bool { p.keys = append(p.keys, k); return p.quit }
func (p *focusProbe) Focused() bool             { return p.focused }
func (p *focusProbe) SetFocus(focused bool)     { p.focused = focused }

func TestFrameFocusRouting(t *testing.T) {
	a, b, c := &focusProbe{}, &focusProbe{}, &focusProbe{}
	f := &Frame{Boxes: []Box{
		{ID: "a", Child: a}, {ID: "b", Child: b, Hidden: true}, {ID: "c", Child: c},
	}, Actions: []FrameAction{{Key: "x", Action: "toggle", Target: "c"}}}
	if f.FocusedBox().ID != "a" || !a.Focused() || b.Focused() || c.Focused() {
		t.Fatal("initial focus must select first visible child")
	}
	f.HandleKey(KeyEvent{Key: "tab"})
	if f.FocusedBox().ID != "c" || a.Focused() || !c.Focused() {
		t.Fatal("tab did not skip hidden child")
	}
	c.quit = true
	if !f.HandleKey(KeyEvent{Key: "down"}) || len(c.keys) != 1 || len(a.keys) != 0 || len(b.keys) != 0 {
		t.Fatal("key/quit was not isolated to focused child")
	}
	if f.HandleKey(KeyEvent{Text: "x"}) || !f.Boxes[2].Hidden || f.FocusedBox().ID != "a" || c.Focused() {
		t.Fatal("toggle did not move focus from hidden child")
	}
	if len(c.keys) != 1 {
		t.Fatal("frame action reached child")
	}
	f.HandleKey(KeyEvent{Key: "shift-tab"})
	if f.FocusedBox().ID != "a" {
		t.Fatal("reverse cycling of single visible child")
	}
	f.Boxes[2].Hidden = false
	f.HandleKey(KeyEvent{Key: "shift-tab"})
	if f.FocusedBox().ID != "c" {
		t.Fatal("reverse wrap did not select last visible child")
	}
	f.Boxes[0].Hidden, f.Boxes[2].Hidden = true, true
	if f.FocusedBox() != nil || a.Focused() || b.Focused() || c.Focused() {
		t.Fatal("all-hidden frame kept focus")
	}
}

func TestFrameCustomFocusKeysAndBoxForwarding(t *testing.T) {
	a, b := &focusProbe{}, &focusProbe{}
	f := &Frame{FocusNextKey: "n", FocusPrevKey: "p", Boxes: []Box{{ID: "a", Child: a}, {ID: "b", Child: b}}}
	f.HandleKey(KeyEvent{Text: "n"})
	if f.FocusedBox().ID != "b" {
		t.Fatal("custom forward key")
	}
	f.HandleKey(KeyEvent{Text: "p"})
	if f.FocusedBox().ID != "a" {
		t.Fatal("custom backward key")
	}
	a.quit = true
	if !f.Boxes[0].HandleKey(KeyEvent{Text: "q"}) || len(a.keys) != 1 {
		t.Fatal("standalone box did not forward child input and quit")
	}
	if (&Box{}).HandleKey(KeyEvent{}) {
		t.Fatal("empty box quit")
	}
}

func TestFrameMouseRoutesToVisibleChildAndFocusesClick(t *testing.T) {
	left, right := &focusProbe{}, &focusProbe{}
	f := &Frame{Gap: 1, Boxes: []Box{
		{ID: "left", Width: 20, Height: 10, Child: left},
		{ID: "right", Width: 20, Height: 10, Child: right},
	}}
	f.Draw(NewCanvas(50, 12), Rect{W: 50, H: 12})
	rects := f.Layout(50, 12)
	click := MouseEvent{Action: MousePress, Button: MouseLeft, X: rects[1].X + 2, Y: rects[1].Y + 2}
	if f.HandleMouse(click) || f.FocusedBox().ID != "right" || len(left.mice) != 0 || len(right.mice) != 1 {
		t.Fatalf("right click not isolated: focus=%s left=%d right=%d", f.FocusedBox().ID, len(left.mice), len(right.mice))
	}
	if right.mice[0].X != 1 || right.mice[0].Y != 1 {
		t.Fatalf("child coordinates = %+v, want 1,1", right.mice[0])
	}
	wheel := MouseEvent{Action: MouseScrollDown, X: rects[0].X + 2, Y: rects[0].Y + 2}
	f.HandleMouse(wheel)
	if len(left.mice) != 1 || f.FocusedBox().ID != "right" {
		t.Fatal("wheel should reach hovered child without changing focus")
	}
	f.Boxes[0].Hidden = true
	f.HandleMouse(wheel)
	if len(left.mice) != 1 {
		t.Fatal("hidden child received mouse event")
	}
}
