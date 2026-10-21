// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestColorPickerModeAndKeyboardInput(t *testing.T) {
	picker := NewColorPicker(ColorIndex(7), ColorIndex(4))
	if !picker.Focused() || picker.Target != ColorPickerFG {
		t.Fatalf("new picker focused=%v target=%v", picker.Focused(), picker.Target)
	}
	fg, bg, selected := picker.FG, picker.BG, picker.Selected
	changed := 0
	picker.OnChange = func(ColorPickerTarget, Color) { changed++ }
	for _, key := range []KeyEvent{{Key: "down"}} {
		if got := picker.ConsumeKey(key); got != Handled() {
			t.Fatalf("BG mode key %+v result = %+v", key, got)
		}
		if picker.Target != ColorPickerBG || picker.FG != fg || picker.BG != bg || picker.Selected != selected || changed != 0 {
			t.Fatalf("BG mode switch changed picker: target=%v colors=%v/%v selected=%d callback=%d", picker.Target, picker.FG, picker.BG, picker.Selected, changed)
		}
	}
	for _, key := range []KeyEvent{{Key: "tab"}, {Text: "b"}} {
		picker.Target = ColorPickerFG
		if got := picker.ConsumeKey(key); got != Handled() || picker.Target != ColorPickerBG {
			t.Fatalf("toggle key %+v from FG result/target = %+v/%v", key, got, picker.Target)
		}
		if picker.FG != fg || picker.BG != bg || picker.Selected != selected || changed != 0 {
			t.Fatalf("toggle key %+v changed colors/selection/callbacks", key)
		}
		if got := picker.ConsumeKey(key); got != Handled() || picker.Target != ColorPickerFG {
			t.Fatalf("toggle key %+v twice result/target = %+v/%v", key, got, picker.Target)
		}
		if picker.FG != fg || picker.BG != bg || picker.Selected != selected || changed != 0 {
			t.Fatalf("toggle key %+v second press changed colors/selection/callbacks", key)
		}
		if got := picker.ConsumeKey(KeyEvent{Key: "up"}); got != Handled() || picker.Target != ColorPickerFG {
			t.Fatalf("Up result/target = %+v/%v", got, picker.Target)
		}
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "up"}); got != Handled() || picker.Target != ColorPickerFG {
		t.Fatalf("Up result/target = %+v/%v", got, picker.Target)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "right"}); got != Handled() || picker.Selected != selected+1 {
		t.Fatalf("Right result/selection = %+v/%d", got, picker.Selected)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "left"}); got != Handled() || picker.Selected != selected {
		t.Fatalf("Left result/selection = %+v/%d", got, picker.Selected)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "left"}); got != Handled() || picker.Selected != 0 || picker.Target != ColorPickerFG {
		t.Fatalf("left edge changed mode/selection: %+v selected=%d target=%v", got, picker.Selected, picker.Target)
	}
	for _, key := range []KeyEvent{{Text: "f"}, {Text: "x"}, {Key: "shift-space"}} {
		if got := picker.ConsumeKey(key); got != Ignored() {
			t.Errorf("removed key %+v result = %+v", key, got)
		}
	}
	if picker.FG != fg || picker.BG != bg || changed != 0 {
		t.Fatalf("removed keys changed colors/callbacks: %v/%v callback=%d", picker.FG, picker.BG, changed)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "space"}); got != Handled() || picker.FG != picker.Palette[0] || changed != 1 {
		t.Fatalf("Space pick result=%+v FG=%v callback=%d", got, picker.FG, changed)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "down"}); got != Handled() || picker.Target != ColorPickerBG {
		t.Fatalf("Down result/target = %+v/%v", got, picker.Target)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "space"}); got != Handled() || picker.BG != picker.Palette[0] || changed != 2 {
		t.Fatalf("BG Space pick result=%+v BG=%v callback=%d", got, picker.BG, changed)
	}
}

func TestColorPickerMouseModesAndBounds(t *testing.T) {
	picker := NewColorPicker(ColorIndex(7), ColorIndex(4))
	picker.Draw(NewCanvas(40, 5), Rect{X: 3, Y: 2, W: 32, H: 3})
	fg, bg, selected := picker.FG, picker.BG, picker.Selected
	changed := 0
	picker.OnChange = func(ColorPickerTarget, Color) { changed++ }
	for _, mouse := range []MouseEvent{
		{Action: MousePress, Button: MouseLeft, X: 32, Y: 0},
		{Action: MousePress, Button: MouseLeft, X: 2, Y: 1},
		{Action: MousePress, Button: MouseLeft, X: 2, Y: 2},
		{Action: MousePress, Button: MouseMiddle, X: 2, Y: 0},
		{Action: MousePress, Button: MouseLeft, X: 2, Y: 0, Shift: true},
	} {
		want := Ignored()
		if mouse.Button == MouseLeft && mouse.X == 2 && mouse.Y == 0 {
			want = Handled()
		}
		if got := picker.ConsumeMouse(mouse); got != want {
			t.Errorf("mouse %+v result = %+v, want %+v", mouse, got, want)
		}
	}
	if picker.FG != picker.Palette[1] || picker.BG != bg || picker.Selected != 1 || changed != 1 {
		t.Fatalf("left-click/shift result FG=%v BG=%v selected=%d callbacks=%d", picker.FG, picker.BG, picker.Selected, changed)
	}
	picker.Target, picker.Selected = ColorPickerFG, selected
	fg, bg = picker.FG, picker.BG
	for i, target := range []ColorPickerTarget{ColorPickerBG, ColorPickerFG} {
		if got := picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseRight, X: 7, Y: 0}); got != Handled() {
			t.Fatalf("right-click %d result = %+v", i+1, got)
		}
		if picker.Target != target || picker.Selected != selected || picker.FG != fg || picker.BG != bg || changed != 1 {
			t.Fatalf("right-click %d changed selection/colors: target=%v selected=%d FG=%v BG=%v callbacks=%d", i+1, picker.Target, picker.Selected, picker.FG, picker.BG, changed)
		}
	}
	picker.Target = ColorPickerBG
	if got := picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 6, Y: 0}); got != Handled() || picker.BG != picker.Palette[3] || changed != 2 {
		t.Fatalf("BG left-click result=%+v BG=%v callbacks=%d", got, picker.BG, changed)
	}
	if got := picker.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 4, Y: 0}); got != Ignored() {
		t.Fatalf("release result = %+v", got)
	}
}

func TestColorPickerDrawModesPreviewAndClose(t *testing.T) {
	picker := NewColorPicker(ColorIndex(11), ColorIndex(4))
	picker.Selected = 3
	canvas := NewCanvas(36, 5)
	picker.Draw(canvas, Rect{X: 2, Y: 1, W: 32, H: 3})
	for i := range picker.Palette {
		cell := canvas.Get(2+i*2, 1)
		if cell.Text != "A" {
			t.Errorf("FG swatch %d cell text = %q", i, cell.Text)
		}
		if cell.Style.FG != picker.Palette[i] || cell.Style.BG != picker.BG {
			t.Errorf("FG swatch %d style = %+v", i, cell.Style)
		}
		if cell.Style.Underline {
			t.Errorf("FG swatch %d unexpectedly underlined", i)
		}
	}
	for x := range 32 {
		cell := canvas.Get(2+x, 2)
		if x == picker.Selected*2 || x == picker.Selected*2+1 {
			if cell.Text != "▔" || cell.Style.FG != ColorIndex(uint8(SpeccedDefaults.ColorPicker.MarkerFG)) || cell.Style.BG != picker.Style.BG {
				t.Fatalf("FG marker cell = %+v", cell)
			}
		} else if cell.Text != " " {
			t.Fatalf("unexpected FG mode marker at %d: %+v", x, cell)
		}
	}
	preview := canvas.Get(2, 3)
	if preview.Text != "P" || preview.Style.FG != picker.FG || preview.Style.BG != picker.BG {
		t.Fatalf("preview cell = %+v", preview)
	}
	picker.Target = ColorPickerBG
	picker.Draw(canvas, Rect{X: 2, Y: 1, W: 32, H: 3})
	for i := range picker.Palette {
		cell := canvas.Get(2+i*2, 1)
		if cell.Style.FG != picker.FG || cell.Style.BG != picker.Palette[i] {
			t.Errorf("BG swatch %d style = %+v", i, cell.Style)
		}
		if cell.Style.Underline {
			t.Errorf("BG swatch %d unexpectedly underlined", i)
		}
	}
	if got := canvas.Get(8, 2); got.Text != "▔" || got.Style.FG != ColorIndex(uint8(SpeccedDefaults.ColorPicker.MarkerFG)) || got.Style.BG != picker.Style.BG {
		t.Fatalf("BG marker = %+v", got)
	}
	if got := canvas.Get(9, 2); got.Text != "▔" || got.Style.FG != ColorIndex(uint8(SpeccedDefaults.ColorPicker.MarkerFG)) || got.Style.BG != picker.Style.BG {
		t.Fatalf("BG marker second cell = %+v", got)
	}
	old := SpeccedDefaults.ColorPicker.ShowPreview
	SpeccedDefaults.ColorPicker.ShowPreview = false
	defer func() { SpeccedDefaults.ColorPicker.ShowPreview = old }()
	if size := picker.Measure(0); size.Width != 32 || size.Height != 2 {
		t.Fatalf("no-preview measure = %+v", size)
	}
	closed := 0
	picker.OnClose = func() { closed++ }
	for _, key := range []string{"enter", "esc", "ctrl-space"} {
		if got := picker.ConsumeKey(KeyEvent{Key: key}); got != DoneResult() {
			t.Errorf("close key %s result = %+v", key, got)
		}
	}
	if closed != 3 {
		t.Fatalf("close callbacks = %d", closed)
	}
}
