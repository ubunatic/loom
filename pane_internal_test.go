// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

// TestWinchBounds covers the pane-placement math used when the terminal window
// is resized: height is clamped to leave the prompt line, and the top row is
// lifted (never below 1) when the pane would overflow the new bottom.
func TestWinchBounds(t *testing.T) {
	cases := []struct {
		name                     string
		startRow, rows, termRows int
		wantStartRow, wantRows   int
	}{
		{"fits unchanged", 5, 8, 24, 5, 8},
		{"height clamped then lifted to fit", 5, 30, 10, 2, 9},
		{"overflow lifts top row", 20, 6, 22, 17, 6},
		{"shrink clamps and lifts", 5, 8, 6, 2, 5},
		{"overflow with clamp lifts to fit", 3, 10, 5, 2, 4},
		{"single-row floor", 1, 2, 2, 1, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotStart, gotRows := winchBounds(c.startRow, c.rows, c.termRows)
			if gotStart != c.wantStartRow || gotRows != c.wantRows {
				t.Fatalf("winchBounds(%d,%d,%d) = (%d,%d), want (%d,%d)",
					c.startRow, c.rows, c.termRows, gotStart, gotRows, c.wantStartRow, c.wantRows)
			}
			// Invariants: stays on screen, at least one row, top row >= 1.
			if gotStart < 1 {
				t.Errorf("startRow %d < 1", gotStart)
			}
			if gotRows < 1 {
				t.Errorf("rows %d < 1", gotRows)
			}
			if gotStart+gotRows-1 > c.termRows {
				t.Errorf("bottom %d overflows termRows %d", gotStart+gotRows-1, c.termRows)
			}
		})
	}
}
