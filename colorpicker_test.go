// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestColorPickerRowsKeepIndependentCursorsAndApply(t *testing.T) {
	picker := NewColorPicker(ColorIndex(7), ColorIndex(4))
	if !picker.Focused() || picker.Target != ColorPickerFG || picker.Selected != 7 {
		t.Fatalf("new picker focused=%v target=%v selection=%d", picker.Focused(), picker.Target, picker.Selected)
	}
	changes := map[ColorPickerTarget]Color{}
	picker.OnChange = func(target ColorPickerTarget, color Color) { changes[target] = color }
	if got := picker.ConsumeKey(KeyEvent{Key: "right"}); got != Handled() || picker.Selected != 8 {
		t.Fatalf("FG Right result/selection=%+v/%d", got, picker.Selected)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "down"}); got != Handled() || picker.Target != ColorPickerBG || picker.Selected != 4 {
		t.Fatalf("Down result/target/selection=%+v/%v/%d", got, picker.Target, picker.Selected)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "left"}); got != Handled() || picker.Selected != 3 {
		t.Fatalf("BG Left result/selection=%+v/%d", got, picker.Selected)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "space"}); got != Handled() || picker.BG != picker.Palette[3] || picker.FG != ColorIndex(7) {
		t.Fatalf("BG Space result/colors=%+v/%v/%v", got, picker.FG, picker.BG)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "up"}); got != Handled() || picker.Target != ColorPickerFG || picker.Selected != 8 {
		t.Fatalf("Up result/target/selection=%+v/%v/%d", got, picker.Target, picker.Selected)
	}
	closed := 0
	picker.OnClose = func() { closed++ }
	if got := picker.ConsumeKey(KeyEvent{Key: "enter"}); got != DoneResult() || picker.FG != picker.Palette[8] || picker.BG != picker.Palette[3] || closed != 1 {
		t.Fatalf("Enter result/colors/close=%+v/%v/%v/%d", got, picker.FG, picker.BG, closed)
	}
	if changes[ColorPickerFG] != picker.Palette[8] || changes[ColorPickerBG] != picker.Palette[3] {
		t.Fatalf("Enter changes=%v", changes)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "tab"}); got != Ignored() {
		t.Fatalf("Tab result=%+v, want ignored for container traversal", got)
	}
	if got := picker.ConsumeKey(KeyEvent{Text: "b"}); got != Ignored() {
		t.Fatalf("b result=%+v, want ignored", got)
	}
}

func TestColorPickerResetAndMouseRows(t *testing.T) {
	picker := NewColorPicker(ColorIndex(7), ColorIndex(4))
	size := picker.Measure(0)
	if size.Width != 35 || size.Height != 6 {
		t.Fatalf("picker measure=%+v, want 35x6", size)
	}
	picker.Draw(NewCanvas(40, 8), Rect{W: size.Width, H: size.Height})
	var changed []ColorPickerTarget
	picker.OnChange = func(target ColorPickerTarget, _ Color) { changed = append(changed, target) }
	if got := picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 33, Y: 2}); got != Handled() || picker.Target != ColorPickerBG || picker.BG != picker.Palette[15] {
		t.Fatalf("BG palette 15 click result/target/BG=%+v/%v/%v", got, picker.Target, picker.BG)
	}
	canvas := NewCanvas(40, 8)
	picker.Draw(canvas, Rect{W: size.Width, H: size.Height})
	marker := canvas.Get(33, 3)
	if marker.Text != SpeccedDefaults.ColorPicker.Glyphs.EdgeUpper || marker.Style.FG != ColorIndex(uint8(SpeccedDefaults.ColorPicker.MarkerFG)) {
		t.Fatalf("BG marker at palette 15=%+v", marker)
	}
	if got := picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 1, Y: 1}); got != Handled() || picker.FG != ColorReset() || !picker.ResetSelected {
		t.Fatalf("FG reset click result/FG/reset=%+v/%v/%v", got, picker.FG, picker.ResetSelected)
	}
	if len(changed) != 2 || changed[0] != ColorPickerBG || changed[1] != ColorPickerFG {
		t.Fatalf("mouse change targets=%v", changed)
	}
	for _, mouse := range []MouseEvent{
		{Action: MousePress, Button: MouseRight, X: 4, Y: 1},
		{Action: MousePress, Button: MouseLeft, X: 4, Y: 0},
		{Action: MousePress, Button: MouseLeft, X: 4, Y: 4},
	} {
		if got := picker.ConsumeMouse(mouse); got != Ignored() {
			t.Errorf("mouse %+v result=%+v, want ignored", mouse, got)
		}
	}
}

func TestColorPickerDrawUsesStaticSwatchesAndSlimEdges(t *testing.T) {
	picker := NewColorPicker(ColorIndex(11), ColorIndex(4))
	canvas := NewCanvas(40, 8)
	canvas.Write(20, 5, "Z", Style{FG: ColorIndex(2), BG: ColorIndex(3)})
	picker.Draw(canvas, Rect{X: 2, Y: 1, W: 35, H: 6})
	for i, color := range picker.Palette {
		fg := canvas.Get(2+3+(i*2), 2)
		bg := canvas.Get(2+3+(i*2), 3)
		if fg.Style.FG != color || fg.Style.BG != picker.Style.BG {
			t.Errorf("FG swatch %d style=%+v", i, fg.Style)
		}
		if bg.Style.FG != picker.Style.FG || bg.Style.BG != color {
			t.Errorf("BG swatch %d style=%+v", i, bg.Style)
		}
	}
	if got := canvas.Get(3, 2); got.Style.FG != ColorReset() || got.Style.BG != picker.Style.BG {
		t.Errorf("FG reset cell style=%+v", got.Style)
	}
	if got := canvas.Get(3, 3); got.Style.FG != picker.Style.FG || got.Style.BG != picker.Style.BG {
		t.Errorf("BG reset cell style=%+v", got.Style)
	}
	for x := 0; x < 35; x++ {
		if got := canvas.Get(2+x, 1); got.Text != SpeccedDefaults.ColorPicker.Glyphs.EdgeLower {
			t.Errorf("lower edge column %d=%+v", x, got)
			break
		}
	}
	if got := canvas.Get(3, 5); got.Text != "P" || got.Style.FG != picker.FG || got.Style.BG != picker.BG {
		t.Fatalf("preview label cell=%+v", got)
	}
	if got := canvas.Get(11, 6); got.Text != " " {
		t.Fatalf("preview rug extends past column 8: %+v", got)
	}
	if got := canvas.Get(20, 5); got.Text != "Z" || got.Style.FG != ColorIndex(2) || got.Style.BG != ColorIndex(3) {
		t.Fatalf("preview rug cleared content outside its 9 columns: %+v", got)
	}
	defs := SpeccedDefaults.ColorPicker
	old := defs.ShowPreview
	SpeccedDefaults.ColorPicker.ShowPreview = false
	defer func() { SpeccedDefaults.ColorPicker.ShowPreview = old }()
	if got := picker.Measure(0); got.Height != 4 {
		t.Fatalf("hidden preview measure=%+v", got)
	}
}
