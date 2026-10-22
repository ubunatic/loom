// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestWidgetNavigationKeysAndMouse(t *testing.T) {
	// Tabs: key navigation and mouse selection
	tab1 := NewView([]string{"tab1"})
	tab2 := NewView([]string{"tab2"})
	tabs := NewTabs(Tab{Title: "One", Widget: tab1}, Tab{Title: "Two", Widget: tab2})
	tabs.Draw(NewCanvas(40, 10), Rect{W: 40, H: 10})
	if res := tabs.ConsumeKey(KeyEvent{Key: SpeccedDefaults.Tabs.Keys.Next}); !res.Consumed || tabs.Focus() != 1 {
		t.Fatalf("Tabs next key failed: focus=%d", tabs.Focus())
	}
	if res := tabs.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 1, Y: 0}); !res.Consumed || tabs.Focus() != 0 {
		t.Fatalf("Tabs mouse click failed: focus=%d", tabs.Focus())
	}

	// Split: ratio change via key and mouse drag
	split := NewSplit(tab1, tab2)
	split.Draw(NewCanvas(40, 10), Rect{W: 40, H: 10})
	initialRatio := split.Ratio
	split.ConsumeKey(KeyEvent{Key: SpeccedDefaults.Split.Bindings["grow"][0]})
	if split.Ratio <= initialRatio {
		t.Fatalf("Split grow key failed: ratio=%f", split.Ratio)
	}
	split.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 20, Y: 0})
	split.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 10, Y: 0})
	split.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 10, Y: 0})
	if split.Ratio >= initialRatio {
		t.Fatalf("Split mouse drag failed: ratio=%f", split.Ratio)
	}

	// View: line scroll via key and mouse wheel
	view := NewView([]string{"line1", "line2", "line3", "line4", "line5"})
	view.Draw(NewCanvas(20, 2), Rect{W: 20, H: 2})
	view.ConsumeKey(KeyEvent{Key: SpeccedDefaults.View.Bindings["down"][0]})
	if view.Scroll != 1 {
		t.Fatalf("View down key failed: scroll=%d", view.Scroll)
	}
	view.ConsumeMouse(MouseEvent{Action: MouseScrollDown, X: 5, Y: 1})
	if view.Scroll != 2 {
		t.Fatalf("View mouse scroll down failed: scroll=%d", view.Scroll)
	}

	// Choice: move via key and click via mouse
	choice := NewChoice([]Item{{Name: "alpha"}, {Name: "beta"}, {Name: "gamma"}})
	choice.Draw(NewCanvas(20, 5), Rect{W: 20, H: 5})
	choice.ConsumeKey(KeyEvent{Key: SpeccedDefaults.Choice.Bindings["down"][0]})
	if item, _ := choice.Selected(); item.Name != "beta" {
		t.Fatalf("Choice down key failed: selected=%v", item.Name)
	}
	choice.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 2, Y: 2})
	if item, _ := choice.Selected(); item.Name != "gamma" {
		t.Fatalf("Choice mouse click failed: selected=%v", item.Name)
	}
}
