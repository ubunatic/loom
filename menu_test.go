// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestMenuBarKeyboardAndAccelerators(t *testing.T) {
	called := 0
	bar := NewMenuBar(
		Menu{Title: "File", Mnemonic: 'F', Items: []MenuItem{{Label: "Save", Shortcut: "Ctrl+S", Action: func() { called++ }}}},
		Menu{Title: "Edit", Mnemonic: 'E', Items: []MenuItem{{Label: "Copy", Action: func() { called++ }}}},
	)
	if !bar.ConsumeKey(KeyEvent{Key: "f10"}).Consumed || !bar.Open {
		t.Fatal("F10 did not open the menu bar")
	}
	bar.ConsumeKey(KeyEvent{Key: "right"})
	if bar.ActiveMenu != 1 {
		t.Fatalf("right selected menu %d, want 1", bar.ActiveMenu)
	}
	bar.ConsumeKey(KeyEvent{Key: "esc"})
	if bar.Open || !bar.Focused() {
		t.Fatal("first Escape did not close the dropdown while retaining menu focus")
	}
	bar.ConsumeKey(KeyEvent{Key: "esc"})
	if bar.Focused() {
		t.Fatal("second Escape did not unfocus the menu bar")
	}
	if !bar.ConsumeKey(KeyEvent{Key: "ctrl-s"}).Consumed || called != 1 {
		t.Fatalf("shortcut called action %d times", called)
	}
}

func TestMenuBarSelectionToggleAndEscape(t *testing.T) {
	checked := false
	called := 0
	bar := NewMenuBar(Menu{Title: "View", Items: []MenuItem{
		{Label: "Status", Checked: &checked, Action: func() { called++ }},
		{Label: "Refresh", Action: func() { called++ }},
	}})
	bar.HandleKey(KeyEvent{Key: "f10"})
	bar.HandleKey(KeyEvent{Key: "enter"})
	if !checked || called != 1 {
		t.Fatalf("checked=%v called=%d", checked, called)
	}
	bar.HandleKey(KeyEvent{Key: "esc"})
	if bar.Open {
		t.Fatal("Escape left dropdown open")
	}
}

func TestMenuBarMouseOpenSelectAndOutsideDismiss(t *testing.T) {
	called := 0
	bar := NewMenuBar(Menu{Title: "File", Items: []MenuItem{{Label: "Open", Action: func() { called++ }}}})
	c := NewCanvas(30, 8)
	bar.Draw(c, Rect{W: 30, H: 8})
	bar.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 1, Y: 0})
	if !bar.Open {
		t.Fatal("clicking title did not open menu")
	}
	bar.Draw(c, Rect{W: 30, H: 8})
	bar.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 2, Y: 2})
	if called != 1 || bar.Open {
		t.Fatalf("called=%d open=%v after item click", called, bar.Open)
	}
	bar.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 1, Y: 0})
	bar.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 20, Y: 6})
	if bar.Open {
		t.Fatal("outside click left menu open")
	}
}

func TestMenuBarDrawsClippedDropdown(t *testing.T) {
	bar := NewMenuBar(Menu{Title: "File", Items: []MenuItem{{Label: "Open", Shortcut: "Ctrl+O"}, {Label: "Save"}}})
	bar.Open = true
	c := NewCanvas(14, 4)
	bar.Draw(c, c.Bounds())
	if got := c.Get(3, 2).Text; got != "O" {
		t.Fatalf("dropdown item starts with %q, want O", got)
	}
	if rows := Render(bar, 14, 4); len(rows) != 4 {
		t.Fatalf("render rows=%d", len(rows))
	}
}
