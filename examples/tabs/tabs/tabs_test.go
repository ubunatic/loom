// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package tabs

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestTabsRunHelp(t *testing.T) {
	if err := Run([]string{"--help"}); err != nil {
		t.Fatalf("Run(--help) unexpected error: %v", err)
	}
}

func TestTabsAppKeysAndDynamicManagement(t *testing.T) {
	app := newTabsApp()
	canvas := loom.NewCanvas(80, 20)
	app.Draw(canvas, canvas.Bounds())

	var b strings.Builder
	canvas.Flush(&b, 1)
	out := b.String()

	if !strings.Contains(out, "Log") || !strings.Contains(out, "Choices") || !strings.Contains(out, "Files") {
		t.Errorf("expected initial tab titles, got:\n%s", out)
	}
	if !strings.Contains(out, "1..3: select") {
		t.Errorf("expected key legend with 1..3, got:\n%s", out)
	}

	// Direct numeric jump to tab 2 (Choices, index 1)
	app.HandleKey(loom.KeyEvent{Key: "2"})
	if app.tabs.Focus() != 1 {
		t.Fatalf("expected focus=1 after key '2', got %d", app.tabs.Focus())
	}

	// Cycle tab with ctrl-t (to Files, index 2)
	app.HandleKey(loom.KeyEvent{Key: "ctrl-t"})
	if app.tabs.Focus() != 2 {
		t.Fatalf("expected focus=2 after ctrl-t, got %d", app.tabs.Focus())
	}

	// Cycle tab with right arrow (back to Log, index 0)
	app.HandleKey(loom.KeyEvent{Key: "right"})
	if app.tabs.Focus() != 0 {
		t.Fatalf("expected focus=0 after right, got %d", app.tabs.Focus())
	}

	// Add a dynamic tab with '+'
	app.HandleKey(loom.KeyEvent{Key: "+"})
	if len(app.tabs.Tabs) != 4 {
		t.Fatalf("expected 4 tabs after '+', got %d", len(app.tabs.Tabs))
	}
	if app.tabs.Focus() != 3 {
		t.Fatalf("expected focus on newly added tab (3), got %d", app.tabs.Focus())
	}
	if app.tabs.Tabs[3].Title != "Extra 1" {
		t.Fatalf("expected title 'Extra 1', got %q", app.tabs.Tabs[3].Title)
	}

	// Add another dynamic tab with 'a'
	app.HandleKey(loom.KeyEvent{Key: "a"})
	if len(app.tabs.Tabs) != 5 {
		t.Fatalf("expected 5 tabs after 'a', got %d", len(app.tabs.Tabs))
	}

	// Remove current tab with 'x'
	app.HandleKey(loom.KeyEvent{Key: "x"})
	if len(app.tabs.Tabs) != 4 {
		t.Fatalf("expected 4 tabs after 'x', got %d", len(app.tabs.Tabs))
	}

	// Remove current tab with 'd'
	app.HandleKey(loom.KeyEvent{Key: "d"})
	if len(app.tabs.Tabs) != 3 {
		t.Fatalf("expected 3 tabs after 'd', got %d", len(app.tabs.Tabs))
	}

	// Clean quit
	if !app.HandleKey(loom.KeyEvent{Key: "q"}) {
		t.Fatal("expected 'q' to trigger quit")
	}
	if !app.HandleKey(loom.KeyEvent{Key: "ctrl-q"}) {
		t.Fatal("expected 'ctrl-q' to trigger quit")
	}
	if !app.HandleKey(loom.KeyEvent{Key: "esc"}) {
		t.Fatal("expected 'esc' to trigger quit")
	}
}
