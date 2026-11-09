// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func scrollbarEntry(lines int) (*TextEntry, *Canvas) {
	e := NewTextEntry(TextEntryOptions{MaxHeight: 4})
	text := ""
	for i := range lines {
		if i > 0 {
			text += "\n"
		}
		text += "x"
	}
	e.SetValue(text)
	c := NewCanvas(20, 4)
	e.Draw(c, Rect{W: 20, H: 4})
	return e, c
}

func TestTextEntryScrollbarHiddenWhenRowsFit(t *testing.T) {
	e, c := scrollbarEntry(4)
	if e.bar || e.layoutWidth != 20 {
		t.Fatalf("bar=%v text width=%d, want no bar and width 20", e.bar, e.layoutWidth)
	}
	if got := c.Get(19, 0).Text; got != " " {
		t.Fatalf("last column = %q, want blank", got)
	}
}

func TestTextEntryScrollbarShownWhenRowsOverflow(t *testing.T) {
	e, c := scrollbarEntry(10)
	if !e.bar || e.layoutWidth != 19 {
		t.Fatalf("bar=%v text width=%d, want bar and width 19", e.bar, e.layoutWidth)
	}
	// The caret is on the last row, so the thumb sits at the bottom of the track.
	if c.Get(19, 3).Text == c.Get(19, 0).Text && c.Get(19, 3).Style == c.Get(19, 0).Style {
		t.Fatal("thumb and track look the same in the bar column")
	}
	if row, _ := e.Cursor(); row != 3 {
		t.Fatalf("caret row = %d, want 3", row)
	}
}

func TestTextEntryScrollbarDragScrollsAndEscRestores(t *testing.T) {
	e, _ := scrollbarEntry(10)
	if e.scroll != 6 {
		t.Fatalf("scroll = %d, want 6 (caret at the end)", e.scroll)
	}
	// Press the thumb at the bottom row, drag to the top.
	e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 19, Y: 3})
	if w, _ := e.mouseCaptureTarget(); w != e {
		t.Fatal("a press on the thumb must claim the mouse capture")
	}
	e.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 19, Y: 0})
	if e.scroll != 0 {
		t.Fatalf("scroll after dragging to the top = %d, want 0", e.scroll)
	}
	e.endMouseCapture(true)
	if e.scroll != 6 {
		t.Fatalf("scroll after Esc = %d, want 6", e.scroll)
	}
	if w, _ := e.mouseCaptureTarget(); w != nil {
		t.Fatal("capture must end on cancel")
	}
	// A drag outside the bar keeps scrolling; release ends it.
	e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 19, Y: 3})
	e.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 40, Y: -5})
	e.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 40, Y: -5})
	if e.scroll != 0 || e.barDrag.active {
		t.Fatalf("scroll=%d active=%v, want 0 and released", e.scroll, e.barDrag.active)
	}
	// Scrolling by the bar does not move the caret or follow it back.
	if e.CursorOffset() != len(e.Value()) {
		t.Fatalf("caret offset = %d, want end of text", e.CursorOffset())
	}
	e.Draw(NewCanvas(20, 4), Rect{W: 20, H: 4})
	if e.scroll != 0 {
		t.Fatalf("Draw pulled the scroll back to %d", e.scroll)
	}
}

func TestTextEntryScrollbarTrackClickJumpsAndTextClickStillPlacesCaret(t *testing.T) {
	e, _ := scrollbarEntry(10)
	e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 19, Y: 0})
	e.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 19, Y: 0})
	if e.scroll != 0 {
		t.Fatalf("track click at the top: scroll = %d, want 0", e.scroll)
	}
	e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 0, Y: 1})
	if got, want := e.CursorOffset(), 2; got != want {
		t.Fatalf("caret offset after a text click = %d, want %d", got, want)
	}
}

func TestTextEntryMeasureMatchesDrawWrapWidthWithBar(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{MaxHeight: 2})
	e.SetValue("aaaa bbbb cccc dddd eeee ffff")
	if got := e.Measure(10).Height; got != 2 {
		t.Fatalf("Measure height = %d, want 2", got)
	}
	c := NewCanvas(10, 2)
	e.Draw(c, Rect{W: 10, H: 2})
	if !e.bar || e.layoutWidth != 9 {
		t.Fatalf("bar=%v width=%d, want bar and width 9", e.bar, e.layoutWidth)
	}
}

func TestTextEntryMiddleScrubScrollsAndEscRestores(t *testing.T) {
	e, _ := scrollbarEntry(10)
	e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseMiddle, X: 3, Y: 3})
	if w, _ := e.mouseCaptureTarget(); w != e {
		t.Fatal("a middle press on the text must claim the capture")
	}
	e.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseMiddle, X: 3, Y: 0})
	if e.scroll >= 6 {
		t.Fatalf("scroll after dragging up = %d, want less than 6", e.scroll)
	}
	e.endMouseCapture(true)
	if e.scroll != 6 {
		t.Fatalf("scroll after Esc = %d, want 6", e.scroll)
	}
	e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseMiddle, X: 3, Y: 3})
	e.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseMiddle, X: 3, Y: 0})
	e.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseMiddle, X: 3, Y: 0})
	if e.pan.active || e.scroll == 6 {
		t.Fatalf("pan active=%v scroll=%d, want released and scrolled", e.pan.active, e.scroll)
	}
}

func TestTextEntryMiddlePressIgnoredWithoutScrollbarOrOnBar(t *testing.T) {
	e, _ := scrollbarEntry(4)
	if r := e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseMiddle, X: 3, Y: 1}); r.Consumed {
		t.Fatal("no bar: a middle press must stay ignored")
	}
	e, _ = scrollbarEntry(10)
	if r := e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseMiddle, X: 19, Y: 1}); r.Consumed {
		t.Fatal("a middle press on the bar column must not pan")
	}
}

func TestTextEntrySetSubmitKeysSwapsEnter(t *testing.T) {
	var got string
	e := NewTextEntry(TextEntryOptions{OnSubmit: func(v string) EventResult { got = v; return Handled() }})
	e.SetValue("a")
	e.SetSubmitKeys([]string{"alt-enter"}, []string{"enter"})
	entryKey(e, "enter")
	if got != "" || e.Value() != "a\n" {
		t.Fatalf("enter: submitted %q value %q, want a line break", got, e.Value())
	}
	entryKey(e, "alt-enter")
	if got != "a\n" {
		t.Fatalf("alt-enter submitted %q, want %q", got, "a\n")
	}
}

func TestTextEntryCtrlEnterSubmitsInBothModes(t *testing.T) {
	var got []string
	e := NewTextEntry(TextEntryOptions{OnSubmit: func(v string) EventResult { got = append(got, v); return Handled() }})
	e.SetValue("a")
	entryKey(e, "ctrl-enter")
	e.SetSubmitKeys([]string{"alt-enter", "ctrl-enter"}, []string{"enter"})
	entryKey(e, "ctrl-enter")
	if len(got) != 2 {
		t.Fatalf("ctrl-enter submitted %d times, want 2 (default and swapped mode)", len(got))
	}
}
