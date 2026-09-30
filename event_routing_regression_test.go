// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestChoiceLeavesContainerFocusKeysUnhandled(t *testing.T) {
	choice := NewChoice([]Item{{Name: "one"}, {Name: "two"}})
	for _, key := range []string{"tab", "shift-tab", "f9"} {
		if result := choice.ConsumeKey(KeyEvent{Key: key}); result != Ignored() {
			t.Errorf("%s = %+v, want ignored", key, result)
		}
	}
	if choice.Query() != "" || choice.FilteredSel() != 0 {
		t.Fatal("container keys changed the choice")
	}
}

func TestViewNavigationReportsConsumption(t *testing.T) {
	for _, key := range []string{"down", "pgdown", "pgdn", "pagedown"} {
		t.Run(key, func(t *testing.T) {
			view := NewView(make([]string, 30))
			view.Draw(NewCanvas(10, 4), Rect{W: 10, H: 4})
			if result := view.ConsumeKey(KeyEvent{Key: key}); result != Handled() {
				t.Fatalf("navigation result = %+v, want handled", result)
			}
			if view.Scroll == 0 {
				t.Fatal("navigation did not scroll")
			}
			for _, up := range []string{"pgup", "pageup"} {
				if result := view.ConsumeKey(KeyEvent{Key: up}); result != Handled() || view.Scroll != 0 {
					t.Fatalf("%s = %+v, scroll=%d, want handled at top", up, result, view.Scroll)
				}
			}
		})
	}
}

func TestViewScrollbarReportsConsumption(t *testing.T) {
	view := NewView(make([]string, 30))
	view.Draw(NewCanvas(14, 8), Rect{X: 2, Y: 3, W: 10, H: 4})
	for _, event := range []MouseEvent{
		{Action: MouseScrollDown},
		{Action: MousePress, Button: MouseLeft, X: 9, Y: 0},
		{Action: MouseDrag, Button: MouseLeft, X: 9, Y: 8},
		{Action: MouseRelease, Button: MouseLeft, X: 9, Y: 8},
	} {
		if result := view.ConsumeMouse(event); result != Handled() {
			t.Errorf("event %+v = %+v, want handled", event, result)
		}
	}
	if view.Scroll != 26 || view.drag.active {
		t.Fatalf("drag finished at scroll=%d active=%v, want bottom and released", view.Scroll, view.drag.active)
	}
	if result := view.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseRight}); result != Ignored() {
		t.Fatalf("right click = %+v, want ignored", result)
	}
}

func TestFrameChoiceFocusAndActionRouting(t *testing.T) {
	choice := NewChoice([]Item{{Name: "one"}, {Name: "two"}})
	view := NewView(make([]string, 30))
	frame := &Frame{
		Boxes:   []Box{{ID: "choice", Width: 18, Height: 8, Child: choice}, {ID: "view", Width: 18, Height: 8, Child: view}},
		Actions: []FrameAction{{Key: "q", Action: "quit"}},
	}
	frame.Draw(NewCanvas(40, 10), Rect{W: 40, H: 10})
	if result := frame.ConsumeKey(KeyEvent{Key: "tab"}); result != Handled() || frame.FocusedBox().ID != "view" {
		t.Fatalf("tab = %+v, focused=%s", result, frame.FocusedBox().ID)
	}
	if result := frame.ConsumeKey(KeyEvent{Key: "down"}); result != Handled() || view.Scroll != 1 {
		t.Fatalf("down = %+v, scroll=%d", result, view.Scroll)
	}
	frame.ConsumeKey(KeyEvent{Key: "shift-tab"})
	if !choice.Focused() || view.Focused() {
		t.Fatal("reverse traversal did not restore choice focus")
	}
	if result := frame.ConsumeKey(KeyEvent{Text: "q"}); result != QuitResult() || choice.Query() != "" {
		t.Fatalf("frame action = %+v, query=%q", result, choice.Query())
	}
}
