// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
)

// ── A7: helpWidget (unexported → white-box in-package test) ────────────────────

func TestHelpWidgetDraw(t *testing.T) {
	hw := newHelpWidget([]Cmd{
		{Name: "help", Title: "show help"},
		{Name: "back", Title: "go back"},
	})
	c := NewCanvas(40, hw.ContentHeight())
	hw.Draw(c, c.Bounds())

	if !rowContains(c, 0, ":help") {
		t.Error("help row should list :help command")
	}
	// Last row is the close hint.
	if !rowContains(c, c.Rows()-1, "press any key") {
		t.Error("help footer should prompt to close")
	}
}

func TestHelpWidgetScrollStaysOpen(t *testing.T) {
	hw := newHelpWidget([]Cmd{
		{Name: "a"}, {Name: "b"}, {Name: "c"},
	})
	if hw.HandleKey(KeyEvent{Key: "down"}) {
		t.Error("down should scroll, not close")
	}
	if hw.scroll != 1 {
		t.Errorf("scroll = %d after down, want 1", hw.scroll)
	}
	if hw.HandleKey(KeyEvent{Key: "up"}) {
		t.Error("up should scroll, not close")
	}
	if hw.scroll != 0 {
		t.Errorf("scroll = %d after up, want 0", hw.scroll)
	}
	// up at top clamps (no underflow) and stays open.
	if hw.HandleKey(KeyEvent{Key: "up"}); hw.scroll != 0 {
		t.Errorf("scroll = %d after clamp, want 0", hw.scroll)
	}
}

func TestHelpWidgetClosesOnAnyOtherKey(t *testing.T) {
	for _, e := range []KeyEvent{{Key: "enter"}, {Key: "esc"}, {Text: "x"}} {
		hw := newHelpWidget([]Cmd{{Name: "a"}})
		if !hw.HandleKey(e) {
			t.Errorf("key %+v should close the help widget", e)
		}
	}
}

func TestHelpWidgetDrawTruncatesToContentWidth(t *testing.T) {
	hw := newHelpWidget([]Cmd{{Name: "長い", Title: "a very long description that must stay inside the modal"}})
	c := NewCanvas(12, hw.ContentHeight())
	hw.Draw(c, c.Bounds())

	for y := 0; y < c.Rows(); y++ {
		if got := StringWidth(c.Row(y)); got > c.Cols() {
			t.Errorf("row %d width = %d, want <= %d: %q", y, got, c.Cols(), c.Row(y))
		}
	}
}

// rowContains reports whether row y of canvas c contains sub (ANSI stripped).
func rowContains(c *Canvas, y int, sub string) bool {
	return strings.Contains(stripANSI(c.Row(y)), sub)
}
