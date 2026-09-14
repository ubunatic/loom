// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import "testing"

// TestClampDimensionsNeverExceedsTerminal is a regression test: an explicit
// --width/--height wider/taller than the real terminal used to be honored
// as-is, so RenderTreemap's grid overflowed the terminal's column count and
// the terminal's own auto-wrap scrambled every subsequent row (confirmed
// via a real PTY repro at 80x24 with --width 150). Every combination here
// must come back within the terminal bounds.
func TestClampDimensionsNeverExceedsTerminal(t *testing.T) {
	cols, rows := 80, 24
	for _, width := range []int{-1, 0, 1, 79, 80, 81, 150, 1000} {
		for _, height := range []int{-1, 0, 1, 22, 23, 24, 25, 100} {
			w, h, _, _ := clampDimensions(width, height, cols, rows)
			if w > cols {
				t.Errorf("clampDimensions(width=%d, height=%d, %dx%d): w=%d exceeds terminal width %d", width, height, cols, rows, w, cols)
			}
			if h > rows-1 {
				t.Errorf("clampDimensions(width=%d, height=%d, %dx%d): h=%d exceeds usable terminal height %d", width, height, cols, rows, h, rows-1)
			}
			if w < 1 || h < 1 {
				t.Errorf("clampDimensions(width=%d, height=%d, %dx%d) = (%d, %d), want both >= 1", width, height, cols, rows, w, h)
			}
		}
	}
}

func TestClampDimensionsReportsWhenItClamps(t *testing.T) {
	cols, rows := 80, 24

	if _, _, widthClamped, _ := clampDimensions(150, 10, cols, rows); !widthClamped {
		t.Error("expected widthClamped=true for --width 150 on an 80-column terminal")
	}
	if _, _, widthClamped, _ := clampDimensions(80, 10, cols, rows); widthClamped {
		t.Error("expected widthClamped=false for --width exactly matching terminal width")
	}
	if _, _, widthClamped, _ := clampDimensions(0, 10, cols, rows); widthClamped {
		t.Error("expected widthClamped=false for an unset (auto) width")
	}

	if _, _, _, heightClamped := clampDimensions(40, 100, cols, rows); !heightClamped {
		t.Error("expected heightClamped=true for --height 100 on a 24-row terminal")
	}
	if _, _, _, heightClamped := clampDimensions(40, rows-1, cols, rows); heightClamped {
		t.Error("expected heightClamped=false for --height exactly at the usable max")
	}
}

func TestClampDimensionsDefaultsLeaveRoomForPrompt(t *testing.T) {
	w, h, widthClamped, heightClamped := clampDimensions(0, 0, 80, 24)
	if w != 80 {
		t.Errorf("default width = %d, want 80 (full terminal width)", w)
	}
	if h != 22 {
		t.Errorf("default height = %d, want 22 (rows-2, leaving room for the shell prompt)", h)
	}
	if widthClamped || heightClamped {
		t.Errorf("auto (0,0) dimensions should never report as clamped, got widthClamped=%v heightClamped=%v", widthClamped, heightClamped)
	}
}
