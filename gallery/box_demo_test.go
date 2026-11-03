// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package gallery

import (
	"testing"

	"ubunatic.com/loom"
)

// TestBoxDemoTitleSegmentStyles checks each title run of the Box demo cell by
// cell: hotkey letter, key badge, plain title text, and dim metadata.
func TestBoxDemoTitleSegmentStyles(t *testing.T) {
	widget, err := New("Box")
	if err != nil {
		t.Fatal(err)
	}
	const cols, rows = 60, 5
	canvas := loom.NewCanvas(cols, rows)
	widget.Draw(canvas, loom.Rect{W: cols, H: rows})

	hot := loom.Style{FG: loom.ColorIndex(1), Bold: true}
	text := loom.Style{FG: loom.ColorIndex(7)}
	meta := loom.Style{FG: loom.ColorIndex(8), Dim: true}
	// Title row: " menu [L] Load (@host · source)" starting at column 1.
	want := []struct {
		x    int
		text string
		st   loom.Style
	}{
		{1, " ", loom.Style{}},
		{2, "m", hot},
		{3, "e", text},
		{6, " ", text},
		{7, "[", hot},
		{8, "L", hot},
		{9, "]", hot},
		{11, "L", text},
		{16, "(", meta},
		{17, "@", meta},
	}
	for _, w := range want {
		got := canvas.Get(w.x, 0)
		if got.Text != w.text || got.Style != w.st {
			t.Errorf("x=%d: got %q %+v, want %q %+v", w.x, got.Text, got.Style, w.text, w.st)
		}
	}
}
