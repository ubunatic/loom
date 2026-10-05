// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
)

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
	bar.ConsumeKey(KeyEvent{Key: "f10"})
	bar.ConsumeKey(KeyEvent{Key: "enter"})
	if !checked || called != 1 {
		t.Fatalf("checked=%v called=%d", checked, called)
	}
	bar.ConsumeKey(KeyEvent{Key: "esc"})
	if bar.Open {
		t.Fatal("Escape left dropdown open")
	}
}

func TestMenuBarMouseOpenSelectAndOutsideDismiss(t *testing.T) {
	called := 0
	bar := NewMenuBar(Menu{Title: "File", Items: []MenuItem{{Label: "Open", Action: func() { called++ }}}})
	c := NewCanvas(30, 8)
	bar.Draw(c, Rect{W: 30, H: 8})
	bar.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 1, Y: 0})
	if !bar.Open {
		t.Fatal("clicking title did not open menu")
	}
	bar.Draw(c, Rect{W: 30, H: 8})
	bar.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 2, Y: 2})
	if called != 1 || bar.Open {
		t.Fatalf("called=%d open=%v after item click", called, bar.Open)
	}
	bar.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 1, Y: 0})
	bar.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 20, Y: 6})
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

func TestMenuBarBottomPlacementOpensDropdownUpward(t *testing.T) {
	bar := NewMenuBar(Menu{Title: "File", Items: []MenuItem{{Label: "Save"}, {Label: "Save as"}}})
	bar.Bottom = true
	bar.ConsumeKey(KeyEvent{Key: "f10"})
	canvas := NewCanvas(24, 8)
	bar.Draw(canvas, canvas.Bounds())
	if bar.barRect.Y != 7 {
		t.Fatalf("bottom bar row = %d, want 7", bar.barRect.Y)
	}
	if bar.titleRects[0].Y != 7 || !strings.Contains(canvas.Row(7), "File") {
		t.Fatalf("bottom menu title is not on row 7: rect=%+v row=%q", bar.titleRects[0], canvas.Row(7))
	}
	if bar.menuRect.Y+bar.menuRect.H != bar.barRect.Y {
		t.Fatalf("dropdown rect %+v does not open upward from bar %+v", bar.menuRect, bar.barRect)
	}
	if !bar.menuRect.Contains(bar.itemRects[0].X, bar.itemRects[0].Y) || bar.itemRects[0].Y >= bar.barRect.Y {
		t.Fatalf("first dropdown item is not above bottom bar: %+v", bar.itemRects[0])
	}
}

func TestMenuBarNestedSubmenuKeyboardAndDismissal(t *testing.T) {
	called := 0
	bar := NewMenuBar(Menu{Title: "File", Items: []MenuItem{
		{Label: "Open", Submenu: []MenuItem{{Label: "Recent", Submenu: []MenuItem{{Label: "Project", Action: func() { called++ }}}}}},
	}})
	bar.ConsumeKey(KeyEvent{Key: "f10"})
	bar.ConsumeKey(KeyEvent{Key: "right"})
	if len(bar.submenus) != 1 {
		t.Fatalf("right opened %d submenu levels, want 1", len(bar.submenus))
	}
	bar.ConsumeKey(KeyEvent{Key: "right"})
	if len(bar.submenus) != 2 {
		t.Fatalf("nested right opened %d levels, want 2", len(bar.submenus))
	}
	bar.ConsumeKey(KeyEvent{Key: "enter"})
	if called != 1 || bar.Open {
		t.Fatalf("called=%d open=%v after leaf activation", called, bar.Open)
	}
	bar.ConsumeKey(KeyEvent{Key: "f10"})
	bar.ConsumeKey(KeyEvent{Key: "right"})
	bar.ConsumeKey(KeyEvent{Key: "esc"})
	if len(bar.submenus) != 0 || !bar.Open {
		t.Fatalf("Escape should close one level; levels=%d open=%v", len(bar.submenus), bar.Open)
	}
}

func TestMenuBarNestedSubmenuFlipsAndRendersMarker(t *testing.T) {
	bar := NewMenuBar(Menu{Title: "File"}, Menu{Title: "Edit"}, Menu{Title: "Help", Items: []MenuItem{{Label: "More", Submenu: []MenuItem{{Label: "Leaf"}}}}})
	bar.ActiveMenu = 2
	bar.Open = true
	c := NewCanvas(20, 8)
	bar.Draw(c, c.Bounds())
	bar.ConsumeKey(KeyEvent{Key: "right"})
	bar.Draw(c, c.Bounds())
	if len(bar.submenus) != 1 {
		t.Fatal("submenu did not open")
	}
	level := bar.submenus[0]
	if level.rect.X+level.rect.W > bar.menuRect.X {
		t.Fatalf("submenu overflows screen: %+v", level.rect)
	}
	if got := c.Get(bar.itemRects[0].X+bar.itemRects[0].W-1, bar.itemRects[0].Y).Text; got != "›" {
		t.Fatalf("submenu marker=%q, want ›", got)
	}
}

func TestMenuBarNestedSubmenuMouseHoverAndClick(t *testing.T) {
	called := 0
	bar := NewMenuBar(Menu{Title: "File", Items: []MenuItem{{Label: "More", Submenu: []MenuItem{{Label: "Leaf", Action: func() { called++ }}}}}})
	c := NewCanvas(24, 8)
	bar.Draw(c, c.Bounds())
	bar.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 1, Y: 0})
	bar.Draw(c, c.Bounds())
	bar.ConsumeMouse(MouseEvent{Action: MouseHover, X: bar.itemRects[0].X, Y: bar.itemRects[0].Y})
	if len(bar.submenus) != 1 {
		t.Fatal("hover did not immediately open submenu")
	}
	bar.Draw(c, c.Bounds())
	leaf := bar.submenus[0].itemRects[0]
	bar.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: leaf.X, Y: leaf.Y})
	if called != 1 || bar.Open {
		t.Fatalf("called=%d open=%v after submenu click", called, bar.Open)
	}
}
