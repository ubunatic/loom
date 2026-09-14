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

func TestParseThemeAcceptsOneAndTwo(t *testing.T) {
	if err := parseTheme(1, false); err != nil {
		t.Errorf("parseTheme(1, false) = %v, want nil", err)
	}
	if err := parseTheme(2, true); err != nil {
		t.Errorf("parseTheme(2, true) = %v, want nil", err)
	}
}

func TestParseThemeRejectsUnknownValues(t *testing.T) {
	for _, theme := range []int{0, 3, -1} {
		if err := parseTheme(theme, true); err == nil {
			t.Errorf("parseTheme(%d, true) = nil, want an error", theme)
		}
	}
}

func TestParseThemeTwoRequiresANSI(t *testing.T) {
	if err := parseTheme(2, false); err == nil {
		t.Error("parseTheme(2, false) = nil, want an error (theme 2 has no color to render edges with without --ansi)")
	}
}

func TestTreemapThemeMapsFlagValuesToGraphThemes(t *testing.T) {
	cases := map[int]graph.TreemapTheme{
		1: graph.TreemapThemeClassic,
		2: graph.TreemapThemeNumbered,
	}
	for flag, want := range cases {
		if got := treemapTheme(flag); got != want {
			t.Errorf("treemapTheme(%d) = %v, want %v", flag, got, want)
		}
	}
}

func TestTreemapPaletteUsesDarkTextOnLightBackgrounds(t *testing.T) {
	opts := buildOptions(80, 20, 1, true, true)
	want := []struct{ bg, fg string }{
		{"41", "97"}, {"42", "30"}, {"43", "30"}, {"44", "97"},
		{"45", "97"}, {"46", "30"}, {"100", "97"}, {"47", "30"},
	}
	if len(opts.BackgroundANSI) != len(want) || len(opts.ForegroundANSI) != len(want) {
		t.Fatalf("palette lengths = %d backgrounds, %d foregrounds; want %d each", len(opts.BackgroundANSI), len(opts.ForegroundANSI), len(want))
	}
	for i, pair := range want {
		if opts.BackgroundANSI[i] != pair.bg || opts.ForegroundANSI[i] != pair.fg {
			t.Errorf("palette[%d] = background %q, foreground %q; want %q, %q", i, opts.BackgroundANSI[i], opts.ForegroundANSI[i], pair.bg, pair.fg)
		}
	}
}

func TestBuildProcTreeExcludesSelfAndDescendants(t *testing.T) {
	psOutput := "10 1 0.1 shell\n20 10 2.0 treemap\n21 20 9.0 ps\n30 10 3.0 firefox\n"
	all := buildProcTree(psOutput, 0)
	filtered := buildProcTree(psOutput, 20)
	names := func(root graph.TreemapNode) map[string]bool {
		seen := map[string]bool{}
		var visit func(graph.TreemapNode)
		visit = func(node graph.TreemapNode) {
			seen[node.Name] = true
			for _, child := range node.Children {
				visit(child)
			}
		}
		visit(root)
		return seen
	}
	if !names(all)["treemap"] || !names(all)["ps"] {
		t.Fatalf("unfiltered tree lost treemap or its ps child: %+v", all)
	}
	got := names(filtered)
	if got["treemap"] || got["ps"] {
		t.Errorf("excluded process or child remained: %+v", filtered)
	}
	if !got["shell"] || !got["firefox"] {
		t.Errorf("unrelated processes were lost: %+v", filtered)
	}
}

func TestIsTreemapGoRun(t *testing.T) {
	cases := []struct {
		command string
		want    bool
	}{
		{"/usr/local/go/bin/go run ./examples/treemap --exclude-self", true},
		{"go run ./examples/treemap/main.go --exclude-self", true},
		{"go run ./examples/monitor", false},
		{"/bin/zsh -c go run ./examples/treemap", false},
		{"/tmp/treemap --exclude-self", false},
	}
	for _, tc := range cases {
		if got := isTreemapGoRun(tc.command); got != tc.want {
			t.Errorf("isTreemapGoRun(%q) = %v, want %v", tc.command, got, tc.want)
		}
	}
}

func TestBuildProcTreeExcludesGoRunLauncherSubtree(t *testing.T) {
	psOutput := "10 1 0.1 shell\n20 10 250.0 go\n21 20 2.0 treemap\n22 21 9.0 ps\n30 10 3.0 firefox\n"
	root := buildProcTree(psOutput, 20)
	if len(root.Children) != 1 || root.Children[0].Name != "shell" {
		t.Fatalf("unexpected roots after launcher exclusion: %+v", root)
	}
	children := root.Children[0].Children
	if len(children) != 1 || children[0].Name != "firefox" {
		t.Errorf("launcher subtree should be gone, unrelated sibling retained: %+v", children)
	}
}
