// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"testing"
)

func TestAnsiEditorBasics(t *testing.T) {
	buf := NewAnsiBuffer(20, 10)
	editor := NewAnsiEditor(buf)

	if !editor.Focused() {
		t.Fatalf("expected focused by default")
	}
	editor.SetFocus(false)
	if editor.Focused() {
		t.Fatalf("expected unfocused after SetFocus(false)")
	}

	x, y := editor.Cursor()
	if x != 0 || y != 0 {
		t.Fatalf("expected cursor at (0,0), got (%d,%d)", x, y)
	}

	editor.SetCursor(5, 7)
	x, y = editor.Cursor()
	if x != 5 || y != 7 {
		t.Fatalf("expected cursor at (5,7), got (%d,%d)", x, y)
	}

	// Clamping
	editor.SetCursor(100, 100)
	x, y = editor.Cursor()
	if x != 19 || y != 9 {
		t.Fatalf("expected cursor clamped to (19,9), got (%d,%d)", x, y)
	}

	editor.SetCursor(-5, -5)
	x, y = editor.Cursor()
	if x != 0 || y != 0 {
		t.Fatalf("expected cursor clamped to (0,0), got (%d,%d)", x, y)
	}
}

func TestAnsiEditorDraw(t *testing.T) {
	buf := NewAnsiBuffer(10, 5)
	buf.Put(0, 0, 'A', ColorIndex(1), ColorReset(), false, false, false, false)
	buf.Put(1, 0, 'B', ColorIndex(2), ColorReset(), false, false, false, false)

	editor := NewAnsiEditor(buf)
	editor.SetFocus(true)
	editor.SetCursor(0, 0)

	c := NewCanvas(10, 5)
	editor.Draw(c, Rect{X: 0, Y: 0, W: 10, H: 5})

	// Cursor at (0,0) should be inverted / styled on canvas
	cell00 := c.Get(0, 0)
	if cell00.Text != "A" || cell00.Style.BG != ColorIndex(15) || cell00.Style.FG != ColorIndex(0) {
		t.Fatalf("unexpected cursor cell at (0,0): %+v", cell00)
	}

	cell10 := c.Get(1, 0)
	if cell10.Text != "B" || cell10.Style.FG != ColorIndex(2) {
		t.Fatalf("unexpected cell at (1,0): %+v", cell10)
	}
}

func TestAnsiEditorKeyNavigationAndTyping(t *testing.T) {
	buf := NewAnsiBuffer(20, 10)
	editor := NewAnsiEditor(buf)
	editor.ActiveFG = ColorIndex(36)

	// Type "Hello"
	res := editor.ConsumeKey(KeyEvent{Text: "Hello"})
	if !res.Consumed || res.Quit {
		t.Fatalf("expected consumed without quit, got %+v", res)
	}
	if x, y := editor.Cursor(); x != 5 || y != 0 {
		t.Fatalf("expected cursor at (5,0), got (%d,%d)", x, y)
	}
	if buf.Get(0, 0).Rune != 'H' || buf.Get(4, 0).Rune != 'o' {
		t.Fatalf("unexpected buffer content: %c..%c", buf.Get(0, 0).Rune, buf.Get(4, 0).Rune)
	}

	// Arrow navigation: left 2 -> (3, 0)
	editor.ConsumeKey(KeyEvent{Key: "left"})
	editor.ConsumeKey(KeyEvent{Key: "left"})
	if x, y := editor.Cursor(); x != 3 || y != 0 {
		t.Fatalf("expected cursor at (3,0), got (%d,%d)", x, y)
	}

	// Down -> (3, 1)
	editor.ConsumeKey(KeyEvent{Key: "down"})
	if x, y := editor.Cursor(); x != 3 || y != 1 {
		t.Fatalf("expected cursor at (3,1), got (%d,%d)", x, y)
	}

	// Up -> (3, 0)
	editor.ConsumeKey(KeyEvent{Key: "up"})
	if x, y := editor.Cursor(); x != 3 || y != 0 {
		t.Fatalf("expected cursor at (3,0), got (%d,%d)", x, y)
	}

	// Right -> (4, 0)
	editor.ConsumeKey(KeyEvent{Key: "right"})
	if x, y := editor.Cursor(); x != 4 || y != 0 {
		t.Fatalf("expected cursor at (4,0), got (%d,%d)", x, y)
	}

	// Home -> (0, 0)
	editor.ConsumeKey(KeyEvent{Key: "home"})
	if x, y := editor.Cursor(); x != 0 || y != 0 {
		t.Fatalf("expected cursor at (0,0), got (%d,%d)", x, y)
	}

	// End -> (19, 0)
	editor.ConsumeKey(KeyEvent{Key: "end"})
	if x, y := editor.Cursor(); x != 19 || y != 0 {
		t.Fatalf("expected cursor at (19,0), got (%d,%d)", x, y)
	}

	// Enter -> (0, 1)
	editor.ConsumeKey(KeyEvent{Key: "enter"})
	if x, y := editor.Cursor(); x != 0 || y != 1 {
		t.Fatalf("expected cursor at (0,1), got (%d,%d)", x, y)
	}
}

func TestAnsiEditorEditingOperations(t *testing.T) {
	buf := NewAnsiBuffer(10, 5)
	editor := NewAnsiEditor(buf)

	editor.ConsumeKey(KeyEvent{Text: "ABC"}) // cursor at (3, 0)
	editor.ConsumeKey(KeyEvent{Key: "backspace"})
	if x, _ := editor.Cursor(); x != 2 {
		t.Fatalf("expected cursor at 2, got %d", x)
	}
	if buf.Get(2, 0).Rune != ' ' {
		t.Fatalf("expected backspaced cell to be blank")
	}

	// Toggle insert mode
	if editor.EditMode != AnsiModeOvertype {
		t.Fatalf("expected overtype mode default")
	}
	editor.ConsumeKey(KeyEvent{Key: "insert"})
	if editor.EditMode != AnsiModeInsert {
		t.Fatalf("expected insert mode after toggle")
	}

	// Copy and Paste
	editor.SetCursor(0, 0) // on 'A'
	editor.ConsumeKey(KeyEvent{Key: "ctrl-c"})
	editor.SetCursor(5, 2)
	editor.ConsumeKey(KeyEvent{Key: "ctrl-v"})
	if buf.Get(5, 2).Rune != 'A' {
		t.Fatalf("expected 'A' pasted at (5,2), got %c", buf.Get(5, 2).Rune)
	}
}

func TestAnsiEditorMouseHandling(t *testing.T) {
	buf := NewAnsiBuffer(20, 10)
	editor := NewAnsiEditor(buf)

	c := NewCanvas(20, 10)
	editor.Draw(c, Rect{X: 2, Y: 1, W: 15, H: 8})

	// Click at canvas coordinate (5, 4) -> should position cursor at (5-2, 4-1) = (3, 3)
	res := editor.ConsumeMouse(MouseEvent{
		Action: MousePress,
		Button: MouseLeft,
		X:      5,
		Y:      4,
	})
	if !res.Consumed {
		t.Fatalf("expected mouse press consumed")
	}
	if x, y := editor.Cursor(); x != 3 || y != 3 {
		t.Fatalf("expected cursor at (3,3), got (%d,%d)", x, y)
	}

	// Mouse scroll down
	editor.ConsumeMouse(MouseEvent{Action: MouseScrollDown})
	if editor.ScrollY != 1 {
		t.Fatalf("expected ScrollY=1, got %d", editor.ScrollY)
	}

	// Mouse scroll up
	editor.ConsumeMouse(MouseEvent{Action: MouseScrollUp})
	if editor.ScrollY != 0 {
		t.Fatalf("expected ScrollY=0, got %d", editor.ScrollY)
	}
}

func TestAnsiEditorInterfaces(t *testing.T) {
	editor := NewAnsiEditor(nil)

	var _ Widget = editor
	var _ EventConsumer = editor
	var _ MouseConsumer = editor
	var _ Focusable = editor
	var _ PaneRequester = editor

	req := editor.PaneRequest()
	if req.Mouse != 1003 || !req.Resizeable {
		t.Fatalf("unexpected pane request: %+v", req)
	}
}
