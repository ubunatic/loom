// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestPaneClickTrackingIsDisabledOnTeardown(t *testing.T) {
	out, err := os.Create(filepath.Join(t.TempDir(), "mouse-sequences"))
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	p := &Pane{tty: out}
	p.EnableMouseClicks()
	p.disableMouse()
	p.disableMouse()
	if p.mouse || p.mouseMode != 0 {
		t.Fatal("mouse tracking remained enabled")
	}
	if _, err := out.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(out)
	if err != nil {
		t.Fatal(err)
	}
	want := "\x1b[?1000h\x1b[?1006h\x1b[?1000l\x1b[?1006l"
	if string(got) != want {
		t.Fatalf("mouse sequences = %q, want %q", got, want)
	}
}

func TestPaneOwnershipRejectsNestedPane(t *testing.T) {
	if !claimPaneOwnership() {
		t.Fatal("initial pane ownership claim failed")
	}
	defer releasePaneOwnership()
	if claimPaneOwnership() {
		t.Fatal("nested pane ownership claim succeeded")
	}
}

func TestPaneOwnershipCanBeReclaimedAfterRelease(t *testing.T) {
	if !claimPaneOwnership() {
		t.Fatal("pane ownership claim failed")
	}
	releasePaneOwnership()
	if !claimPaneOwnership() {
		t.Fatal("pane ownership was not released")
	}
	releasePaneOwnership()
}

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

func TestPaneHandleKeyFallback(t *testing.T) {
	p := &Pane{}

	// Default fallback exits on Esc, Ctrl-C, Ctrl-Q, Ctrl-D, and 'q'
	exitKeys := []KeyEvent{
		{Key: "esc"},
		{Key: "ctrl-c"},
		{Key: "ctrl-q"},
		{Key: "ctrl-d"},
		{Text: "q"},
	}

	for _, ke := range exitKeys {
		if !p.handleKeyFallback(ke) {
			t.Errorf("expected handleKeyFallback(%+v) = true", ke)
		}
	}

	// Non-exit keys do not quit
	nonExitKeys := []KeyEvent{
		{Key: "enter"},
		{Key: "up"},
		{Key: "down"},
		{Text: "a"},
		{Text: "x"},
	}

	for _, ke := range nonExitKeys {
		if p.handleKeyFallback(ke) {
			t.Errorf("expected handleKeyFallback(%+v) = false", ke)
		}
	}

	// When DisableDefaultQuit is set, no fallback exit occurs
	p.DisableDefaultQuit = true
	for _, ke := range exitKeys {
		if p.handleKeyFallback(ke) {
			t.Errorf("expected handleKeyFallback(%+v) = false when DisableDefaultQuit is true", ke)
		}
	}
}

func TestPaneHelpOverlayUsesRootCanvasAndCapturesInput(t *testing.T) {
	p := &Pane{}
	previous := paneHelpRequest
	defer func() { paneHelpRequest = previous }()
	paneHelpRequest = func(cmds []Cmd) {
		p.help = NewPopup("Help", newHelpWidget(cmds))
	}

	bar := newCmdBar()
	bar.active = true
	bar.query = "help"
	if bar.execute(bar.match()) != cmdNone {
		t.Fatal("help command unexpectedly changed navigation")
	}
	if p.help == nil {
		t.Fatal("help did not request the pane overlay")
	}

	canvas := NewCanvas(40, 12)
	canvas.Clear()
	p.help.Draw(canvas, canvas.Bounds())
	if got := canvas.Get(10, 3).Text; got != "┌" {
		t.Fatalf("overlay left border at %d,3 = %q, want popup corner", 10, got)
	}
	if p.handleHelpKey(KeyEvent{Text: "x"}) {
		// The event is consumed by the overlay; this branch documents that
		// underlying widgets must not see it.
	} else {
		t.Fatal("help key was not captured")
	}
	if p.help != nil {
		t.Fatal("help overlay remained open after dismissal key")
	}
}
