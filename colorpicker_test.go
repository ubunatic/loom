// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"testing"
)

func TestColorPickerRowsShareCursorAndApplyOnlyActiveRow(t *testing.T) {
	picker := NewColorPicker(ColorIndex(7), ColorIndex(4))
	if !picker.Focused() || picker.Target != ColorPickerFG || picker.Selected != 7 {
		t.Fatalf("new picker focused=%v target=%v selection=%d", picker.Focused(), picker.Target, picker.Selected)
	}
	changes := map[ColorPickerTarget]Color{}
	picker.OnChange = func(target ColorPickerTarget, color Color) { changes[target] = color }
	if got := picker.ConsumeKey(KeyEvent{Key: "right"}); got != Handled() || picker.Selected != 8 {
		t.Fatalf("FG Right result/selection=%+v/%d", got, picker.Selected)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "down"}); got != Handled() || picker.Target != ColorPickerBG || picker.Selected != 8 {
		t.Fatalf("Down result/target/selection=%+v/%v/%d", got, picker.Target, picker.Selected)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "left"}); got != Handled() || picker.Selected != 7 {
		t.Fatalf("BG Left result/selection=%+v/%d", got, picker.Selected)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "space"}); got != Handled() || picker.BG != picker.Palette[7] || picker.FG != ColorIndex(7) {
		t.Fatalf("BG Space result/colors=%+v/%v/%v", got, picker.FG, picker.BG)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "up"}); got != Handled() || picker.Target != ColorPickerFG || picker.Selected != 7 {
		t.Fatalf("Up result/target/selection=%+v/%v/%d", got, picker.Target, picker.Selected)
	}
	closed := 0
	picker.OnClose = func() { closed++ }
	if got := picker.ConsumeKey(KeyEvent{Key: "enter"}); got != DoneResult() || picker.FG != picker.Palette[7] || picker.BG != picker.Palette[7] || closed != 1 {
		t.Fatalf("Enter result/colors/close=%+v/%v/%v/%d", got, picker.FG, picker.BG, closed)
	}
	if changes[ColorPickerFG] != picker.Palette[7] || changes[ColorPickerBG] != picker.Palette[7] {
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
	if size.Width != 36 || size.Height != 7 {
		t.Fatalf("picker measure=%+v, want 36x7", size)
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
	if marker.Text != SpeccedDefaults.ColorPicker.Glyphs.EdgeUpper || marker.Style.FG != DefaultTheme().Selection.Color() {
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
		{Action: MousePress, Button: MouseLeft, X: 4, Y: 5},
	} {
		if got := picker.ConsumeMouse(mouse); got != Ignored() {
			t.Errorf("mouse %+v result=%+v, want ignored", mouse, got)
		}
	}
}

func TestColorPickerFaintToggleKeyboardMouseAndPreview(t *testing.T) {
	picker := NewColorPicker(ColorIndex(7), ColorIndex(4))
	picker.Selected = len(picker.Palette) - 1
	var changes []bool
	colorChanges := 0
	picker.OnDimChange = func(dim bool) { changes = append(changes, dim) }
	picker.OnChange = func(ColorPickerTarget, Color) { colorChanges++ }
	if got := picker.ConsumeKey(KeyEvent{Key: "right"}); got != Handled() || !picker.FaintSelected {
		t.Fatalf("Right from final FG swatch = %+v faint-selected=%v", got, picker.FaintSelected)
	}
	if !picker.PreviewStyle().Dim {
		t.Fatal("faint preview should include Dim")
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "left"}); got != Handled() || picker.FaintSelected || picker.Selected != len(picker.Palette)-1 {
		t.Fatalf("Left from faint = %+v faint-selected=%v selected=%d", got, picker.FaintSelected, picker.Selected)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "right"}); got != Handled() || !picker.FaintSelected {
		t.Fatalf("Right to faint = %+v faint-selected=%v", got, picker.FaintSelected)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "space"}); got != Handled() || !picker.Dim || len(changes) != 1 || !changes[0] {
		t.Fatalf("Space faint toggle result/dim/callbacks = %+v/%v/%v", got, picker.Dim, changes)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "down"}); got != Handled() || picker.Target != ColorPickerBG || picker.FaintSelected || picker.Selected != len(picker.Palette)-1 {
		t.Fatalf("Down from faint = %+v target=%v faint-selected=%v selected=%d", got, picker.Target, picker.FaintSelected, picker.Selected)
	}
	size := picker.Measure(0)
	picker.Draw(NewCanvas(size.Width, size.Height), Rect{W: size.Width, H: size.Height})
	defs := SpeccedDefaults.ColorPicker
	if got := picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: defs.LeadingPadding, Y: 4}); got != Handled() || picker.Dim || len(changes) != 2 || changes[1] {
		t.Fatalf("click faint toggle result/dim/callbacks = %+v/%v/%v", got, picker.Dim, changes)
	}
	if picker.FG != ColorIndex(7) || picker.BG != ColorIndex(4) {
		t.Fatalf("faint toggle changed colors: FG=%v BG=%v", picker.FG, picker.BG)
	}
	if colorChanges != 0 {
		t.Fatalf("faint toggles triggered %d color callbacks", colorChanges)
	}
	if got := picker.ConsumeKey(KeyEvent{Key: "right"}); got != Handled() || !picker.FaintSelected {
		t.Fatalf("Right to faint before Enter = %+v faint-selected=%v", got, picker.FaintSelected)
	}
	closed := 0
	picker.OnClose = func() { closed++ }
	if got := picker.ConsumeKey(KeyEvent{Key: "enter"}); got != DoneResult() || !picker.Dim || len(changes) != 3 || !changes[2] || closed != 1 {
		t.Fatalf("Enter faint toggle result/dim/callbacks/close = %+v/%v/%v/%d", got, picker.Dim, changes, closed)
	}
}

func TestColorPickerDrawUsesStaticSwatchesAndSlimEdges(t *testing.T) {
	picker := NewColorPicker(ColorIndex(11), ColorIndex(4))
	canvas := NewCanvas(40, 8)
	canvas.Write(20, 5, "Z", Style{FG: ColorIndex(2), BG: ColorIndex(3)})
	picker.Draw(canvas, Rect{W: 36, H: 7})
	panelBG := ColorIndex(uint8(SpeccedDefaults.RichTextEdit.ToolbarBG))
	for i, color := range picker.Palette {
		fg := canvas.Get(3+(i*2), 1)
		bg := canvas.Get(3+(i*2), 2)
		if fg.Style.FG != color || fg.Style.BG != panelBG {
			t.Errorf("FG swatch %d style=%+v", i, fg.Style)
		}
		if bg.Style.FG != picker.Style.FG || bg.Style.BG != color {
			t.Errorf("BG swatch %d style=%+v", i, bg.Style)
		}
	}
	if got := canvas.Get(1, 1); got.Style.FG != ColorReset() || got.Style.BG != panelBG {
		t.Errorf("FG reset cell style=%+v", got.Style)
	}
	if got := canvas.Get(1, 2); got.Style.FG != picker.Style.FG || got.Style.BG != panelBG {
		t.Errorf("BG reset cell style=%+v", got.Style)
	}
	for _, pos := range [][2]int{{0, 1}, {0, 2}, {0, 5}, {8, 5}} {
		if got := canvas.Get(pos[0], pos[1]); got.Style.BG != panelBG {
			t.Errorf("panel padding at (%d,%d) style=%+v, want panel BG %v", pos[0], pos[1], got.Style, panelBG)
		}
	}
	assertEdges := func(p *ColorPicker, frame *Canvas) {
		t.Helper()
		inactive := ColorIndex(uint8(SpeccedDefaults.ColorPicker.InactiveMarkerFG))
		editorBG := p.Style.BG
		fgIndex, fgReset := colorPickerIndex(p.Palette, p.FG)
		bgIndex, bgReset := colorPickerIndex(p.Palette, p.BG)
		fgX := colorPickerCellColumn(fgIndex, fgReset, len(p.Palette))
		bgX := colorPickerCellColumn(bgIndex, bgReset, len(p.Palette))
		fgCursorX, bgCursorX := -1, -1
		if p.Target == ColorPickerFG {
			fgCursorX = p.cursorMarkerColumn()
		} else {
			bgCursorX = p.cursorMarkerColumn()
		}
		for x := 0; x < 36; x++ {
			wantFG := panelBG
			if colorPickerMarkerAt(x, fgX, 2) {
				wantFG = inactive
			}
			if colorPickerMarkerAt(x, fgCursorX, 2) {
				wantFG = p.selection
			}
			if got := frame.Get(x, 0); got.Text != SpeccedDefaults.ColorPicker.Glyphs.EdgeLower || got.Style.FG != wantFG || got.Style.BG != editorBG {
				t.Errorf("lower edge cell (%d,0)=%+v, want FG %v BG %v", x, got, wantFG, editorBG)
			}
			bgApplied := colorPickerMarkerAt(x, bgX, 2)
			bgCursor := colorPickerMarkerAt(x, bgCursorX, 2)
			if x < SpeccedDefaults.ColorPicker.PreviewWidth && !bgApplied && !bgCursor {
				if got := frame.Get(x, 3); got.Text != " " || got.Style.BG != panelBG {
					t.Errorf("upper-edge gap cell (%d,3)=%+v, want panel BG %v", x, got, panelBG)
				}
				continue
			}
			wantFG = panelBG
			wantBG := editorBG
			if bgApplied {
				wantFG = inactive
				if x < SpeccedDefaults.ColorPicker.PreviewWidth {
					wantBG = panelBG
				}
			}
			if bgCursor {
				wantFG = p.selection
				if x < SpeccedDefaults.ColorPicker.PreviewWidth {
					wantBG = panelBG
				}
			}
			if got := frame.Get(x, 3); got.Text != SpeccedDefaults.ColorPicker.Glyphs.EdgeUpper || got.Style.FG != wantFG || got.Style.BG != wantBG {
				t.Errorf("upper edge cell (%d,3)=%+v, want FG %v BG %v", x, got, wantFG, wantBG)
			}
		}
		for x := 0; x < SpeccedDefaults.ColorPicker.PreviewWidth; x++ {
			if got := frame.Get(x, 6); got.Text != SpeccedDefaults.ColorPicker.Glyphs.EdgeUpper || got.Style.FG != panelBG || got.Style.BG != editorBG {
				t.Errorf("bottom edge cell (%d,6)=%+v, want FG %v BG %v", x, got, panelBG, editorBG)
			}
		}
	}
	assertEdges(picker, canvas)
	if got := canvas.Get(1, 5); got.Text != "P" || got.Style.FG != picker.FG || got.Style.BG != picker.BG {
		t.Fatalf("preview label cell=%+v", got)
	}
	if got := canvas.Get(9, 6); got.Text != " " {
		t.Fatalf("preview rug extends past column 8: %+v", got)
	}
	if got := canvas.Get(20, 5); got.Text != "Z" || got.Style.FG != ColorIndex(2) || got.Style.BG != ColorIndex(3) {
		t.Fatalf("preview rug cleared content outside its 9 columns: %+v", got)
	}
	markerPicker := NewColorPicker(ColorIndex(1), ColorIndex(15))
	markerCanvas := NewCanvas(36, 7)
	markerPicker.Draw(markerCanvas, Rect{W: 36, H: 7})
	assertEdges(markerPicker, markerCanvas)
	markerInsidePicker := NewColorPicker(ColorIndex(1), ColorIndex(0))
	markerInsideCanvas := NewCanvas(36, 7)
	markerInsidePicker.Draw(markerInsideCanvas, Rect{W: 36, H: 7})
	assertEdges(markerInsidePicker, markerInsideCanvas)
	if got := canvas.Get(35, 2); got.Text != " " || got.Style.BG != panelBG {
		t.Fatalf("right padding should use panel background: %+v", got)
	}
	if got := picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 35, Y: 2}); got != Ignored() {
		t.Fatalf("right padding click result=%+v, want ignored", got)
	}
	if picker.ConsumeKey(KeyEvent{Key: "down"}) != Handled() {
		t.Fatal("Down should activate the BG row")
	}
	activeBGCanvas := NewCanvas(36, 7)
	picker.Draw(activeBGCanvas, Rect{W: 36, H: 7})
	assertEdges(picker, activeBGCanvas)
	if markerPicker.ConsumeKey(KeyEvent{Key: "down"}) != Handled() {
		t.Fatal("Down should activate BG row for marker right of preview rug")
	}
	activeBGRightCanvas := NewCanvas(36, 7)
	markerPicker.Draw(activeBGRightCanvas, Rect{W: 36, H: 7})
	assertEdges(markerPicker, activeBGRightCanvas)
	if markerInsidePicker.ConsumeKey(KeyEvent{Key: "down"}) != Handled() {
		t.Fatal("Down should activate BG row for marker inside preview rug")
	}
	activeBGInsideCanvas := NewCanvas(36, 7)
	markerInsidePicker.Draw(activeBGInsideCanvas, Rect{W: 36, H: 7})
	assertEdges(markerInsidePicker, activeBGInsideCanvas)
	defs := SpeccedDefaults.ColorPicker
	old := defs.ShowPreview
	SpeccedDefaults.ColorPicker.ShowPreview = false
	defer func() { SpeccedDefaults.ColorPicker.ShowPreview = old }()
	if got := picker.Measure(0); got.Height != 5 {
		t.Fatalf("hidden preview measure=%+v", got)
	}
}

func TestColorPickerDrawStaysWithinSuppliedBounds(t *testing.T) {
	defs := SpeccedDefaults.ColorPicker
	SpeccedDefaults.ColorPicker.Glyphs.EdgeLower = "界"
	SpeccedDefaults.ColorPicker.Glyphs.EdgeUpper = "界"
	defer func() { SpeccedDefaults.ColorPicker = defs }()

	picker := NewColorPicker(ColorIndex(7), ColorIndex(4))
	rects := []Rect{
		{X: 4, Y: 2, W: 0, H: 3},
		{X: 4, Y: 2, W: 3, H: 0},
		{X: 4, Y: 2, W: 1, H: 1},
		{X: 4, Y: 2, W: 2, H: 2},
		{X: 4, Y: 2, W: 1, H: 6},
		{X: 4, Y: 2, W: 6, H: 1},
		{X: 4, Y: 2, W: 36, H: 7},
	}
	for _, rect := range rects {
		t.Run(fmt.Sprintf("%dx%d-offset", rect.W, rect.H), func(t *testing.T) {
			canvas := NewCanvas(48, 12)
			before := make([]Cell, canvas.Cols()*canvas.Rows())
			for y := 0; y < canvas.Rows(); y++ {
				for x := 0; x < canvas.Cols(); x++ {
					cell := Cell{Text: "s", Style: Style{FG: ColorIndex(uint8(x)), BG: ColorIndex(uint8(y))}, Claim: true}
					canvas.Set(x, y, cell)
					before[y*canvas.Cols()+x] = cell
				}
			}
			picker.Draw(canvas, rect)
			for y := 0; y < canvas.Rows(); y++ {
				for x := 0; x < canvas.Cols(); x++ {
					inside := x >= rect.X && x < rect.X+rect.W && y >= rect.Y && y < rect.Y+rect.H
					if !inside && canvas.Get(x, y) != before[y*canvas.Cols()+x] {
						t.Errorf("draw at %+v changed outside cell (%d,%d): got %+v, want %+v", rect, x, y, canvas.Get(x, y), before[y*canvas.Cols()+x])
					}
				}
			}
		})
	}

	picker.Draw(NewCanvas(48, 12), Rect{X: 4, Y: 2, W: 36, H: 7})
	picker.Draw(NewCanvas(48, 12), Rect{X: 4, Y: 2, W: 0, H: 6})
	if got := picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 1, Y: 1}); got != Ignored() {
		t.Fatalf("mouse after empty draw=%+v, want ignored", got)
	}
	picker.Draw(NewCanvas(48, 12), Rect{X: 4, Y: 2, W: 3, H: 1})
	if got := picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 2, Y: 1}); got != Ignored() {
		t.Fatalf("mouse outside offset tiny draw=%+v, want ignored", got)
	}
}

func TestColorPickerDrawPreservesWideGlyphsCrossingHorizontalBounds(t *testing.T) {
	picker := NewColorPicker(ColorIndex(7), ColorIndex(4))
	canvas := NewCanvas(12, 8)
	left := Cell{Text: "界", Style: Style{FG: ColorIndex(2), BG: ColorIndex(3)}}
	right := Cell{Text: "界", Style: Style{FG: ColorIndex(5), BG: ColorIndex(6)}}
	canvas.Set(2, 1, left)
	canvas.Set(3, 1, Cell{Continuation: true, Style: left.Style})
	canvas.Set(7, 2, right)
	canvas.Set(8, 2, Cell{Continuation: true, Style: right.Style})
	leftBefore := []Cell{canvas.Get(2, 1), canvas.Get(3, 1)}
	rightBefore := []Cell{canvas.Get(7, 2), canvas.Get(8, 2)}

	picker.Draw(canvas, Rect{X: 3, Y: 1, W: 5, H: 4})

	if got := []Cell{canvas.Get(2, 1), canvas.Get(3, 1)}; got[0] != leftBefore[0] || got[1] != leftBefore[1] {
		t.Fatalf("left crossing glyph changed: got %+v, want %+v", got, leftBefore)
	}
	if got := []Cell{canvas.Get(7, 2), canvas.Get(8, 2)}; got[0] != rightBefore[0] || got[1] != rightBefore[1] {
		t.Fatalf("right crossing glyph changed: got %+v, want %+v", got, rightBefore)
	}
}
