// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

// TestScanMouseDrainsFlood pins the fix for the multi-report read bug: 1003
// any-motion tracking floods several SGR reports into one read, and the Run
// loop drains them report-by-report via scanMouse. DecodeMouse alone rejects
// such a buffer (it expects exactly one report), which dropped all mouse input.
func TestScanMouseDrainsFlood(t *testing.T) {
	// Three hover reports concatenated, as foot delivers them in one read.
	raw := []byte("\x1b[<35;10;5M\x1b[<35;11;6M\x1b[<35;12;7M")

	// DecodeMouse on the whole buffer must fail — this is the original bug.
	if _, ok := DecodeMouse(raw); ok {
		t.Fatal("DecodeMouse should reject a multi-report buffer")
	}

	want := []MouseEvent{
		{Action: MouseHover, Button: MouseNone, X: 10, Y: 5},
		{Action: MouseHover, Button: MouseNone, X: 11, Y: 6},
		{Action: MouseHover, Button: MouseNone, X: 12, Y: 7},
	}
	for i, w := range want {
		ev, used, ok := scanMouse(raw)
		if !ok {
			t.Fatalf("report %d: scanMouse returned ok=false", i)
		}
		if ev != w {
			t.Fatalf("report %d: got %+v, want %+v", i, ev, w)
		}
		raw = raw[used:]
	}
	if len(raw) != 0 {
		t.Fatalf("buffer not fully drained: %q left", raw)
	}
	if _, _, ok := scanMouse(raw); ok {
		t.Fatal("scanMouse on empty buffer should return ok=false")
	}
}

// TestScanMousePartialReport guards a report split across reads: a buffer that
// starts a report but lacks the M/m terminator must report ok=false rather than
// mis-parse, so the loop can wait for the rest.
func TestScanMousePartialReport(t *testing.T) {
	if _, _, ok := scanMouse([]byte("\x1b[<35;10;5")); ok {
		t.Fatal("scanMouse should reject a report missing its terminator")
	}
}
