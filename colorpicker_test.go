// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestColorPickerTargetAndKeyboardInput(t *testing.T) {
	picker := NewColorPicker(ColorIndex(7), ColorIndex(0))
	if !picker.Focused() {
		t.Fatal("new picker should be focused")
	}
	changed := []ColorPickerTarget{}
	picker.OnChange = func(target ColorPickerTarget, _ Color) { changed = append(changed, target) }

	if got := picker.ConsumeKey(KeyEvent{Key: "space"}); got != Handled() {
		t.Fatalf("Space result = %+v", got)
	}
	if picker.FG != picker.Palette[picker.Selected] || picker.BG != ColorIndex(0) {
		t.Fatalf("Space colors = FG %v BG %v", picker.FG, picker.BG)
	}
	picker.ConsumeKey(KeyEvent{Key: "shift-space"})
	if picker.BG != picker.Palette[picker.Selected] || picker.FG != picker.BG {
		t.Fatalf("Shift+Space colors = FG %v BG %v", picker.FG, picker.BG)
	}
	picker.ConsumeKey(KeyEvent{Text: "f"})
	if picker.Target != ColorPickerFG || picker.FG != picker.Palette[picker.Selected] {
		t.Fatalf("f state = target %v FG %v", picker.Target, picker.FG)
	}
	picker.ConsumeKey(KeyEvent{Text: "b"})
	if picker.Target != ColorPickerBG || picker.BG != picker.Palette[picker.Selected] {
		t.Fatalf("b state = target %v BG %v", picker.Target, picker.BG)
	}
	if len(changed) != 4 || changed[0] != ColorPickerFG || changed[1] != ColorPickerBG || changed[2] != ColorPickerFG || changed[3] != ColorPickerBG {
		t.Fatalf("change targets = %v", changed)
	}
}

func TestColorPickerTargetToggleEdgeAndUnknownKeys(t *testing.T) {
	picker := NewColorPicker(ColorIndex(7), ColorIndex(0))
	if result := picker.ConsumeKey(KeyEvent{Text: "?"}); result != Ignored() {
		t.Fatalf("unknown key result = %+v", result)
	}
	if result := picker.ConsumeKey(KeyEvent{Key: "tab"}); result != Ignored() {
		t.Fatalf("Tab result = %+v", result)
	}
	if picker.Selected != 0 || picker.Target != ColorPickerFG {
		t.Fatalf("ignored input changed picker: selected %d target %v", picker.Selected, picker.Target)
	}
	picker.ConsumeKey(KeyEvent{Key: "left"})
	if picker.Target != ColorPickerBG || picker.Selected != 0 {
		t.Fatalf("left edge = target %v selected %d", picker.Target, picker.Selected)
	}
	picker.Selected = len(picker.Palette) - 1
	picker.ConsumeKey(KeyEvent{Key: "right"})
	if picker.Target != ColorPickerFG || picker.Selected != len(picker.Palette)-1 {
		t.Fatalf("right edge = target %v selected %d", picker.Target, picker.Selected)
	}
	picker.Selected = 0
	picker.ConsumeKey(KeyEvent{Text: "x"})
	if picker.Target != ColorPickerBG {
		t.Fatalf("x target = %v, want BG", picker.Target)
	}
	picker.ConsumeKey(KeyEvent{Key: "right"})
	if picker.Selected != 1 || picker.Target != ColorPickerBG {
		t.Fatalf("interior right = selected %d target %v", picker.Selected, picker.Target)
	}
}

func TestColorPickerMouseSelectionTargetAndBounds(t *testing.T) {
	picker := NewColorPicker(ColorIndex(7), ColorIndex(0))
	picker.Draw(NewCanvas(40, 4), Rect{X: 3, Y: 2, W: 30, H: 3})

	if got := picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 30, Y: 1}); got != Ignored() {
		t.Fatalf("outside swatch result = %+v", got)
	}
	if picker.FG != ColorIndex(7) {
		t.Fatal("outside click changed FG")
	}
	if got := picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 8, Y: 2}); got != Ignored() {
		t.Fatalf("preview click result = %+v", got)
	}
	if picker.FG != ColorIndex(7) || picker.BG != ColorIndex(0) {
		t.Fatal("preview click changed either color")
	}
	if got := picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 2, Y: 1, Shift: true}); got != Handled() {
		t.Fatalf("shift click result = %+v", got)
	}
	if picker.BG != picker.Palette[1] {
		t.Fatalf("Shift+click BG = %v", picker.BG)
	}
	picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseRight, X: 4, Y: 1})
	if picker.BG != picker.Palette[2] {
		t.Fatalf("right click BG = %v", picker.BG)
	}
	if got := picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseMiddle, X: 6, Y: 1}); got != Ignored() {
		t.Fatalf("middle click result = %+v", got)
	}
	picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 0, Y: 0})
	if picker.Target != ColorPickerFG {
		t.Fatalf("FG target toggle target = %v", picker.Target)
	}
	picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 5, Y: 0})
	if picker.Target != ColorPickerBG {
		t.Fatalf("target toggle click target = %v", picker.Target)
	}
	picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 6, Y: 1})
	if picker.BG != picker.Palette[3] {
		t.Fatalf("active target click BG = %v", picker.BG)
	}
	if got := picker.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 4, Y: 1}); got != Ignored() {
		t.Fatalf("release result = %+v", got)
	}
}

func TestColorPickerCloseAndPreview(t *testing.T) {
	picker := NewColorPicker(ColorIndex(7), ColorIndex(0))
	closed := 0
	picker.OnClose = func() { closed++ }
	if got := picker.ConsumeKey(KeyEvent{Key: "enter"}); got != DoneResult() || closed != 1 {
		t.Fatalf("Enter result/callback = %+v/%d", got, closed)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "esc"}); got != DoneResult() || closed != 2 {
		t.Fatalf("Esc result/callback = %+v/%d", got, closed)
	}
	canvas := NewCanvas(20, 4)
	picker.Draw(canvas, Rect{W: 20, H: 4})
	preview := picker.PreviewStyle()
	cell := canvas.Get(1, 2)
	if preview.FG != picker.FG || preview.BG != picker.BG || cell.Style.FG != picker.FG || cell.Style.BG != picker.BG {
		t.Fatalf("preview style=%+v cell=%+v colors FG=%v BG=%v", preview, cell, picker.FG, picker.BG)
	}
}
