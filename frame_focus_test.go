package loom

import "testing"

type focusProbe struct {
	focused bool
	keys    []KeyEvent
	quit    bool
}

func (*focusProbe) Draw(*Canvas, Rect)          {}
func (*focusProbe) HandleMouse(MouseEvent) bool { return false }
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
