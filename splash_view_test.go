// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
)

func TestSplashViewGoldenInProgress(t *testing.T) {
	view := NewSplashView("harnez usage",
		ProviderPill{Symbol: "●", Name: "mic", State: ProviderDone},
		ProviderPill{Symbol: "✳", Name: "claude", State: ProviderFetching},
		ProviderPill{Symbol: "֍", Name: "codex", State: ProviderPending},
		ProviderPill{Symbol: "Λ", Name: "agy", State: ProviderPending},
	)
	view.SpinnerFrame = 1 // ⠙ in DefaultSpinnerFrames
	view.BracketWidth = 24
	view.Progress = 18.75 // [⣿⣿⣿⣿⡇                   ]
	view.StepText = "fetching claude..."

	c := NewCanvas(50, 10)
	view.Draw(c, c.Bounds())

	// Top line (row 1 because (10-8)/2 = 1)
	row1 := stripANSI(c.Row(1))
	if !strings.Contains(row1, "⠙  harnez usage") {
		t.Errorf("row1 does not contain title: %q", row1)
	}

	// Bar line (row 3)
	row3 := stripANSI(c.Row(3))
	if !strings.Contains(row3, "[⣿⣿⣿⣿⡇") {
		t.Errorf("row3 does not contain partial bar: %q", row3)
	}

	// Step text line (row 5)
	row5 := stripANSI(c.Row(5))
	if !strings.Contains(row5, "fetching claude...") {
		t.Errorf("row5 does not contain step text: %q", row5)
	}

	// Provider pills line (row 6)
	row6 := stripANSI(c.Row(6))
	if !strings.Contains(row6, "● mic  ✳ claude  ֍ codex  Λ agy") {
		t.Errorf("row6 does not contain pill cluster: %q", row6)
	}

	// Footer line (row 8)
	row8 := stripANSI(c.Row(8))
	if !strings.Contains(row8, "Esc to skip") {
		t.Errorf("row8 does not contain footer text: %q", row8)
	}
}

func TestSplashViewGoldenCompleted(t *testing.T) {
	view := NewSplashView("harnez usage",
		ProviderPill{Symbol: "●", Name: "mic", State: ProviderDone},
		ProviderPill{Symbol: "✳", Name: "claude", State: ProviderDone},
		ProviderPill{Symbol: "֍", Name: "codex", State: ProviderDone},
		ProviderPill{Symbol: "Λ", Name: "agy", State: ProviderDone},
	)
	view.BracketWidth = 32
	view.Pattern = ":"
	view.StepText = "agy done"

	c := NewCanvas(60, 12)
	view.Draw(c, c.Bounds())

	// Row (12-8)/2 = 2
	rowBar := stripANSI(c.Row(4))
	if !strings.Contains(rowBar, "["+strings.Repeat(":", 32)+"]") {
		t.Errorf("completed bar incorrect: %q", rowBar)
	}

	rowStep := stripANSI(c.Row(6))
	if !strings.Contains(rowStep, "agy done") {
		t.Errorf("completed step text incorrect: %q", rowStep)
	}
}

func TestSplashViewCenteringDimensions(t *testing.T) {
	view := NewSplashView("harnez usage")
	view.BracketWidth = 24

	geometries := []struct {
		cols, rows int
	}{
		{80, 24},
		{40, 20},
		{120, 40},
	}

	for _, g := range geometries {
		c := NewCanvas(g.cols, g.rows)
		view.Draw(c, c.Bounds())

		expectedOffsetY := (g.rows - 8) / 2
		// Check that the title row is written at expectedOffsetY
		titleRow := stripANSI(c.Row(expectedOffsetY))
		if !strings.Contains(titleRow, "harnez usage") {
			t.Errorf("geometry %dx%d: title not found at row %d (got: %q)", g.cols, g.rows, expectedOffsetY, titleRow)
		}
	}
}

func TestSplashViewKeyHandling(t *testing.T) {
	view := NewSplashView("harnez usage")

	// Direct key handling without controller
	if !view.HandleKey(KeyEvent{Key: "esc"}) {
		t.Errorf("HandleKey(esc) = false, want true")
	}
	if !view.HandleKey(KeyEvent{Text: "q"}) {
		t.Errorf("HandleKey(text 'q') = false, want true")
	}
	if !view.HandleKey(KeyEvent{Key: "enter"}) {
		t.Errorf("HandleKey(enter) = false, want true")
	}
	if view.HandleKey(KeyEvent{Text: "a"}) {
		t.Errorf("HandleKey(text 'a') = true, want false")
	}

	// Key handling with attached controller
	sc := NewSplashController(SplashConfig{})
	view.Controller = sc
	if !view.HandleKey(KeyEvent{Text: "q"}) {
		t.Errorf("Controller view HandleKey(q) = false, want true")
	}
	if !sc.Snapshot().Dismissed {
		t.Errorf("controller not dismissed after q key")
	}
}

