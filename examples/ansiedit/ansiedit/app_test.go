// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansiedit

import (
	"path/filepath"
	"testing"

	"ubunatic.com/loom"
)

func TestAppDrawAndLayout(t *testing.T) {
	buf := NewBuffer(54, 22)
	buf.SetPath("testart.ansi")
	app := NewAnsiEditApp(buf)

	canvas := loom.NewCanvas(80, 24)
	app.Draw(canvas, loom.Rect{X: 0, Y: 0, W: 80, H: 24})

	// Verify top header
	row0 := canvas.Row(0)
	if row0 == "" {
		t.Fatalf("expected rendered row 0")
	}

	// Verify footer
	row23 := canvas.Row(23)
	if row23 == "" {
		t.Fatalf("expected rendered row 23")
	}

	// Verify side panel views draw without panic
	for _, mode := range []SidePanelMode{PanelInfo, PanelPalette, PanelKeys, PanelTheme} {
		app.panelMode = mode
		c := loom.NewCanvas(80, 24)
		app.Draw(c, loom.Rect{X: 0, Y: 0, W: 80, H: 24})
	}
}

func TestAppFocusAndPanelSwitching(t *testing.T) {
	buf := NewBuffer(54, 22)
	app := NewAnsiEditApp(buf)

	if app.activeFocus != FocusCanvas {
		t.Fatalf("expected default focus to be Canvas, got %v", app.activeFocus)
	}

	// Tab toggles focus
	app.ConsumeKey(loom.KeyEvent{Key: "tab"})
	if app.activeFocus != FocusSidePanel {
		t.Fatalf("expected focus SidePanel after Tab, got %v", app.activeFocus)
	}

	app.ConsumeKey(loom.KeyEvent{Key: "tab"})
	if app.activeFocus != FocusCanvas {
		t.Fatalf("expected focus Canvas after second Tab, got %v", app.activeFocus)
	}

	// Function keys switch side panel modes
	app.ConsumeKey(loom.KeyEvent{Key: "f1"})
	if app.panelMode != PanelInfo {
		t.Fatalf("expected PanelInfo after F1, got %v", app.panelMode)
	}

	app.ConsumeKey(loom.KeyEvent{Key: "f2"})
	if app.panelMode != PanelPalette {
		t.Fatalf("expected PanelPalette after F2, got %v", app.panelMode)
	}

	app.ConsumeKey(loom.KeyEvent{Key: "f8"})
	if app.panelMode != PanelKeys {
		t.Fatalf("expected PanelKeys after F8, got %v", app.panelMode)
	}

	app.ConsumeKey(loom.KeyEvent{Key: "f9"})
	if app.panelMode != PanelTheme {
		t.Fatalf("expected PanelTheme after F9, got %v", app.panelMode)
	}

	// F10 quits
	if app.Quit() {
		t.Fatalf("expected not quit yet")
	}
	app.ConsumeKey(loom.KeyEvent{Key: "f10"})
	if !app.Quit() {
		t.Fatalf("expected quit after F10")
	}
}

func TestPaletteNavigationAndColorApply(t *testing.T) {
	buf := NewBuffer(54, 22)
	app := NewAnsiEditApp(buf)

	app.panelMode = PanelPalette
	app.activeFocus = FocusSidePanel

	app.palRow = 0
	app.palCol = 0

	// Move down and right in palette
	app.ConsumeKey(loom.KeyEvent{Key: "down"})
	if app.palRow != 1 {
		t.Fatalf("expected palRow 1, got %d", app.palRow)
	}

	app.ConsumeKey(loom.KeyEvent{Key: "right"})
	if app.palCol != 1 {
		t.Fatalf("expected palCol 1, got %d", app.palCol)
	}

	targetColor := paletteGrid[1][1] // #94

	// Enter sets FG
	app.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if app.activeFG != loom.ColorIndex(targetColor) {
		t.Fatalf("expected activeFG %v, got %v", loom.ColorIndex(targetColor), app.activeFG)
	}

	// Shift-Enter sets BG
	app.palRow = 6
	app.palCol = 0 // #232
	bgTarget := paletteGrid[6][0]
	app.ConsumeKey(loom.KeyEvent{Key: "shift-enter"})
	if app.activeBG != loom.ColorIndex(bgTarget) {
		t.Fatalf("expected activeBG %v, got %v", loom.ColorIndex(bgTarget), app.activeBG)
	}
}

func TestCanvasTypingAndNavigation(t *testing.T) {
	buf := NewBuffer(54, 22)
	app := NewAnsiEditApp(buf)
	app.activeFocus = FocusCanvas

	// Type text
	app.ConsumeKey(loom.KeyEvent{Text: "Loom"})
	if app.cursorX != 4 || app.cursorY != 0 {
		t.Fatalf("expected cursor at (4, 0), got (%d, %d)", app.cursorX, app.cursorY)
	}
	if buf.Get(0, 0).Rune != 'L' || buf.Get(3, 0).Rune != 'm' {
		t.Fatalf("expected 'L' and 'm', got %c and %c", buf.Get(0, 0).Rune, buf.Get(3, 0).Rune)
	}

	// Move cursor left
	app.ConsumeKey(loom.KeyEvent{Key: "left"})
	if app.cursorX != 3 {
		t.Fatalf("expected cursor at 3, got %d", app.cursorX)
	}

	// Backspace
	app.ConsumeKey(loom.KeyEvent{Key: "backspace"})
	if app.cursorX != 2 {
		t.Fatalf("expected cursor at 2, got %d", app.cursorX)
	}

	// Insert mode toggle
	app.ConsumeKey(loom.KeyEvent{Key: "insert"})
	if app.editMode != ModeInsert {
		t.Fatalf("expected ModeInsert after insert key")
	}

	// Save with C-s
	tmp := t.TempDir()
	savePath := filepath.Join(tmp, "canvas_test.ansi")
	buf.SetPath(savePath)
	app.ConsumeKey(loom.KeyEvent{Key: "ctrl-s"})
	if buf.Modified() {
		t.Fatalf("expected buffer unmodified after save")
	}
}

func TestThemeSwitching(t *testing.T) {
	buf := NewBuffer(54, 22)
	app := NewAnsiEditApp(buf)
	app.panelMode = PanelTheme
	app.activeFocus = FocusSidePanel

	origTheme := app.currentTheme
	app.ConsumeKey(loom.KeyEvent{Key: "down"})
	if app.currentTheme == origTheme && len(app.themeNames) > 1 {
		t.Fatalf("expected theme to change on down arrow")
	}
}

func TestAnsiEditNavigationKeysNeverQuit(t *testing.T) {
	buf := NewBuffer(54, 22)
	app := NewAnsiEditApp(buf)

	navKeys := []loom.KeyEvent{
		{Key: "up"},
		{Key: "down"},
		{Key: "left"},
		{Key: "right"},
		{Key: "ctrl-up"},
		{Key: "ctrl-down"},
		{Key: "ctrl-left"},
		{Key: "ctrl-right"},
		{Key: "tab"},
		{Key: "backspace"},
		{Key: "delete"},
		{Key: "insert"},
		{Key: "enter"},
		{Key: "f1"},
		{Key: "f2"},
		{Key: "f8"},
		{Key: "f9"},
		{Key: "ctrl-s"},
		{Text: "a"},
		{Text: " "},
	}

	for _, k := range navKeys {
		// ConsumeKey must return Consumed: true, Quit: false
		res := app.ConsumeKey(k)
		if !res.Consumed || res.Quit {
			t.Errorf("ConsumeKey(%+v) = %+v, want Consumed:true, Quit:false", k, res)
		}
		// ConsumeKey must return false (meaning do not quit)
		if quit := app.ConsumeKey(k).Quit; quit {
			t.Errorf("ConsumeKey(%+v) = true (quit), want false", k)
		}
		if app.Quit() {
			t.Errorf("app.Quit() became true after key %+v", k)
		}
	}

	// Explicit quit keys
	quitKeys := []loom.KeyEvent{
		{Key: "f10"},
		{Key: "ctrl-q"},
	}

	for _, k := range quitKeys {
		app.quit = false
		res := app.ConsumeKey(k)
		if !res.Consumed || !res.Quit {
			t.Errorf("ConsumeKey(%+v) = %+v, want Consumed:true, Quit:true", k, res)
		}
		if !app.Quit() {
			t.Errorf("app.Quit() = false after quit key %+v, want true", k)
		}
		if quit := app.ConsumeKey(k).Quit; !quit {
			t.Errorf("ConsumeKey(%+v) = false, want true", k)
		}
	}
}
