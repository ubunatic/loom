// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func Test361ColorPickerMarkerStatesAndPreview(t *testing.T) {
	// Cell 0 is reset; palette index n occupies cell n+1. -1 means no marker.
	for _, tc := range []struct {
		name     string
		fg, bg   Color
		keys     []string
		target   ColorPickerTarget
		fgMarker int
		bgMarker int
		preview  Style
		applied  Style
	}{
		{"initial", ColorIndex(4), ColorIndex(15), nil, ColorPickerFG, 5, 16, Style{FG: ColorIndex(4), BG: ColorIndex(15)}, Style{FG: ColorIndex(4), BG: ColorIndex(15)}},
		{"FG cursor moves", ColorIndex(4), ColorIndex(15), []string{"right"}, ColorPickerFG, 6, 16, Style{FG: ColorIndex(5), BG: ColorIndex(15)}, Style{FG: ColorIndex(4), BG: ColorIndex(15)}},
		{"Down without pick", ColorIndex(4), ColorIndex(15), []string{"right", "down"}, ColorPickerBG, 5, 6, Style{FG: ColorIndex(4), BG: ColorIndex(5)}, Style{FG: ColorIndex(4), BG: ColorIndex(15)}},
		{"Down after FG Space", ColorIndex(4), ColorIndex(15), []string{"right", "space", "right", "down"}, ColorPickerBG, 6, 7, Style{FG: ColorIndex(5), BG: ColorIndex(6)}, Style{FG: ColorIndex(5), BG: ColorIndex(15)}},
		{"Up after BG Space", ColorIndex(4), ColorIndex(15), []string{"down", "right", "space", "right", "up"}, ColorPickerFG, 7, 6, Style{FG: ColorIndex(6), BG: ColorIndex(5)}, Style{FG: ColorIndex(4), BG: ColorIndex(5)}},
		{"reset cursor Down", ColorIndex(4), ColorIndex(15), []string{"home", "down"}, ColorPickerBG, 5, 0, Style{FG: ColorIndex(4), BG: ColorReset()}, Style{FG: ColorIndex(4), BG: ColorIndex(15)}},
		{"applied reset inactive", ColorIndex(4), ColorIndex(15), []string{"home", "space", "down", "right"}, ColorPickerBG, 0, 1, Style{FG: ColorReset(), BG: ColorIndex(0)}, Style{FG: ColorReset(), BG: ColorIndex(15)}},
		{"BG outside palette", ColorIndex(4), ColorIndex(99), nil, ColorPickerFG, 5, -1, Style{FG: ColorIndex(4), BG: ColorIndex(99)}, Style{FG: ColorIndex(4), BG: ColorIndex(99)}},
		{"FG outside palette", ColorIndex(99), ColorIndex(4), []string{"down"}, ColorPickerBG, -1, 1, Style{FG: ColorIndex(99), BG: ColorIndex(0)}, Style{FG: ColorIndex(99), BG: ColorIndex(4)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NewColorPicker(tc.fg, tc.bg)
			for _, key := range tc.keys {
				p.ConsumeKey(KeyEvent{Key: key})
			}
			if p.Target != tc.target || p.FG != tc.applied.FG || p.BG != tc.applied.BG {
				t.Fatalf("target/applied=%v/%+v, want %v/%+v", p.Target, Style{FG: p.FG, BG: p.BG}, tc.target, tc.applied)
			}
			if got := p.PreviewStyle(); got != tc.preview {
				t.Fatalf("preview=%+v, want %+v", got, tc.preview)
			}
			canvas := NewCanvas(36, 6)
			p.Draw(canvas, canvas.Bounds())
			for row, cell := range []int{tc.fgMarker, tc.bgMarker} {
				y := row * 3
				markerColor := ColorIndex(uint8(SpeccedDefaults.ColorPicker.InactiveMarkerFG))
				if ColorPickerTarget(row) == tc.target {
					markerColor = DefaultTheme().Selection.Color()
				}
				for x := 0; x < 36; x++ {
					wantMarker := cell >= 0 && x >= 1+cell*2 && x < 3+cell*2
					if got := canvas.Get(x, y); (got.Style.FG == markerColor) != wantMarker {
						t.Errorf("row %d column %d=%+v, marker=%v", row, x, got, wantMarker)
					}
				}
			}
			if got := canvas.Get(1, 4); got.Text != "P" || got.Style != tc.preview {
				t.Errorf("rendered preview=%+v, want %+v", got, tc.preview)
			}
		})
	}
}

func Test361ColorPickerEnterAppliesOnlyActiveRow(t *testing.T) {
	for _, space := range []bool{false, true} {
		p := NewColorPicker(ColorIndex(7), ColorIndex(4))
		var changes []ColorPickerTarget
		p.OnChange = func(target ColorPickerTarget, _ Color) { changes = append(changes, target) }
		p.ConsumeKey(KeyEvent{Key: "right"})
		wantFG := ColorIndex(7)
		if space {
			p.ConsumeKey(KeyEvent{Key: "space"})
			wantFG = ColorIndex(8)
		}
		p.ConsumeKey(KeyEvent{Key: "right"}) // Merely passing over FG 9 never applies it.
		p.ConsumeKey(KeyEvent{Key: "down"})
		closed := false
		p.OnClose = func() {
			closed = true
			if p.FG != wantFG || p.BG != ColorIndex(9) {
				t.Errorf("OnClose colors FG=%v BG=%v", p.FG, p.BG)
			}
		}
		if result := p.ConsumeKey(KeyEvent{Key: "enter"}); result != DoneResult() || !closed {
			t.Fatalf("Enter result=%+v closed=%v", result, closed)
		}
		wantChanges := 1
		if space {
			wantChanges = 2
		}
		if len(changes) != wantChanges || changes[len(changes)-1] != ColorPickerBG {
			t.Fatalf("Space=%v changed channels=%v", space, changes)
		}
	}
}

func Test361ColorPickerMouseMovesSharedCursor(t *testing.T) {
	p := NewColorPicker(ColorIndex(7), ColorIndex(4))
	p.Draw(NewCanvas(36, 6), Rect{X: 10, Y: 5, W: 36, H: 6})
	// Mouse coordinates are child-local, including when drawn at an offset.
	p.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 5, Y: 2})
	if p.Target != ColorPickerBG || p.Selected != 1 || p.BG != ColorIndex(1) || p.FG != ColorIndex(7) {
		t.Fatalf("BG click state=%+v", p)
	}
	p.ConsumeKey(KeyEvent{Key: "up"})
	if p.Selected != 1 || p.ResetSelected || p.Target != ColorPickerFG {
		t.Fatalf("Up after click state=%+v", p)
	}
	p.ConsumeKey(KeyEvent{Key: "enter"})
	if p.FG != ColorIndex(1) || p.BG != ColorIndex(1) {
		t.Fatalf("Enter after click FG=%v BG=%v", p.FG, p.BG)
	}
}
