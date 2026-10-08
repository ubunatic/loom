// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"testing"
)

func TestDrawRuler(t *testing.T) {
	c := NewCanvas(20, 10)
	DrawRuler(c, c.Bounds())

	// Top left cell (x=0, y=0) should be space
	if cell := c.Get(0, 0); cell.Text != " " {
		t.Errorf("Get(0, 0) = %q, want space", cell.Text)
	}
	// x=10, y=0 should be "1"
	if cell := c.Get(10, 0); cell.Text != "1" {
		t.Errorf("Get(10, 0) = %q, want '1'", cell.Text)
	}
	// Check that reset background is used by PaintForeground
	cell := c.Get(5, 0)
	if cell.Style.BG != ColorReset() {
		t.Errorf("Get(5, 0).Style.BG = %v, want ColorReset()", cell.Style.BG)
	}
}

func TestPaneDebugToggleKey(t *testing.T) {
	p := &Pane{}
	if p.DebugMode {
		t.Errorf("initial DebugMode = true, want false")
	}

	p.dispatchKey(nil, KeyEvent{Key: "shift-f12"})
	if !p.DebugMode {
		t.Errorf("DebugMode = false after shift-f12, want true")
	}

	p.dispatchKey(nil, KeyEvent{Key: "shift-f12"})
	if p.DebugMode {
		t.Errorf("DebugMode = true after second shift-f12, want false")
	}
}
