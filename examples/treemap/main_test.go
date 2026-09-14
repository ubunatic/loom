// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"testing"

	"codeberg.org/ubunatic/loom/graph"
)

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

// Row-clipping and terminal-safety behavior (auto-wrap suppression, per-row
// clipping, absolute positioning) now lives in loom.RawScreen/loom.ClipRow
// (see ../../rawscreen.go and ../../rawscreen_test.go) rather than being
// duplicated locally in this example -- run() and runWatch() just call
// loom.OpenRawScreen(os.Stdout).Draw(rows).

func TestParseThemeAcceptsOneThroughFour(t *testing.T) {
	if err := parseTheme(1, false); err != nil {
		t.Errorf("parseTheme(1, false) = %v, want nil", err)
	}
	for _, theme := range []int{2, 3, 4} {
		if err := parseTheme(theme, true); err != nil {
			t.Errorf("parseTheme(%d, true) = %v, want nil", theme, err)
		}
	}
}

func TestParseThemeRejectsUnknownValues(t *testing.T) {
	for _, theme := range []int{0, 5, -1} {
		if err := parseTheme(theme, true); err == nil {
			t.Errorf("parseTheme(%d, true) = nil, want an error", theme)
		}
	}
}

func TestParseThemeTwoThroughFourRequireANSI(t *testing.T) {
	for _, theme := range []int{2, 3, 4} {
		if err := parseTheme(theme, false); err == nil {
			t.Errorf("parseTheme(%d, false) = nil, want an error (no color to render edges with without --ansi)", theme)
		}
	}
}

func TestTreemapThemeMapsFlagValuesToGraphThemes(t *testing.T) {
	cases := map[int]graph.TreemapTheme{
		1: graph.TreemapThemeClassic,
		2: graph.TreemapThemeBlocks,
		3: graph.TreemapThemeNumbered,
		4: graph.TreemapThemeNumberedFilled,
	}
	for flag, want := range cases {
		if got := treemapTheme(flag); got != want {
			t.Errorf("treemapTheme(%d) = %v, want %v", flag, got, want)
		}
	}
}
