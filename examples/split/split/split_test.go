// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package split

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestSplitRunHelp(t *testing.T) {
	if err := Run([]string{"--help"}); err != nil {
		t.Fatalf("Run(--help) unexpected error: %v", err)
	}
}

func TestSplitAppLayoutAndFocusTraversal(t *testing.T) {
	app := newSplitApp()
	canvas := loom.NewCanvas(80, 20)
	app.Draw(canvas, canvas.Bounds())

	var b strings.Builder
	canvas.Flush(&b, 1)
	out := b.String()

	if !strings.Contains(out, "Nested Split") {
		t.Errorf("expected title to contain 'Nested Split', got:\n%s", out)
	}
	if !strings.Contains(out, "H: 40/60 • V: 50/50") {
		t.Errorf("expected status to contain dynamic ratios, got:\n%s", out)
	}

	// Initial focus on Left
	leftView := app.hSplit.First.(*loom.View)
	topRightView := app.vSplit.First.(*loom.View)
	bottomRightView := app.vSplit.Second.(*loom.View)

	if !leftView.Focused() {
		t.Fatal("expected left view to be focused initially")
	}
	if topRightView.Focused() || bottomRightView.Focused() {
		t.Fatal("expected right views not focused initially")
	}

	// Tab: moves focus to Top-Right
	app.HandleKey(loom.KeyEvent{Key: "tab"})
	if leftView.Focused() || !topRightView.Focused() || bottomRightView.Focused() {
		t.Fatal("expected top-right view focused after first tab")
	}

	// Tab: moves focus to Bottom-Right
	app.HandleKey(loom.KeyEvent{Key: "tab"})
	if leftView.Focused() || topRightView.Focused() || !bottomRightView.Focused() {
		t.Fatal("expected bottom-right view focused after second tab")
	}

	// Tab: cycles back to Left
	app.HandleKey(loom.KeyEvent{Key: "tab"})
	if !leftView.Focused() || topRightView.Focused() || bottomRightView.Focused() {
		t.Fatal("expected left view focused after cycling tab")
	}

	// Ratio key adjustments: "[" reduces ratio
	prevRatio := app.hSplit.Ratio
	app.HandleKey(loom.KeyEvent{Key: "["})
	if app.hSplit.Ratio >= prevRatio {
		t.Fatalf("expected ratio to decrease on '[', got %f (was %f)", app.hSplit.Ratio, prevRatio)
	}

	// Redraw reflects new ratio
	canvas.Clear()
	app.Draw(canvas, canvas.Bounds())
	b.Reset()
	canvas.Flush(&b, 1)
	out = b.String()
	if !strings.Contains(out, "H: 35/65") {
		t.Errorf("expected updated ratio H: 35/65, got:\n%s", out)
	}

	// Quit key
	if !app.HandleKey(loom.KeyEvent{Key: "q"}) {
		t.Fatal("expected 'q' to trigger quit")
	}
}
