// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package vt

import "testing"

// queryStep feeds one input and checks the exact reply bytes it produces.
type queryStep struct {
	in   string
	want string
}

func runQuerySteps(t *testing.T, v *VT, steps []queryStep) {
	t.Helper()
	for _, s := range steps {
		v.Write([]byte(s.in)) //nolint:errcheck
		if got := string(v.DrainReplies()); got != s.want {
			t.Fatalf("after %q: reply = %q, want %q", s.in, got, s.want)
		}
	}
}

func TestVTAnswersDECRQM(t *testing.T) {
	v := NewVT(80, 24)
	runQuerySteps(t, v, []queryStep{
		// Modes the VT does not track are reported unknown (0), never guessed.
		{"\x1b[?2026$p", "\x1b[?2026;0$y"},
		{"\x1b[?9999$p", "\x1b[?9999;0$y"},
		// Tracked modes report their current state: 1 set, 2 reset.
		{"\x1b[?7$p", "\x1b[?7;1$y"}, // auto-wrap is on by default
		{"\x1b[?7l\x1b[?7$p", "\x1b[?7;2$y"},
		{"\x1b[?25$p", "\x1b[?25;1$y"}, // cursor visible by default
		{"\x1b[?25l\x1b[?25$p", "\x1b[?25;2$y"},
		{"\x1b[?1$p", "\x1b[?1;2$y"},
		{"\x1b[?1h\x1b[?1$p", "\x1b[?1;1$y"},
		{"\x1b[?2004h\x1b[?2004$p", "\x1b[?2004;1$y"},
		{"\x1b[?1000$p", "\x1b[?1000;2$y"},
		{"\x1b[?1000h\x1b[?1000$p", "\x1b[?1000;1$y"},
		// Mouse modes are mutually exclusive: 1003 is off while 1000 is on.
		{"\x1b[?1003$p", "\x1b[?1003;2$y"},
		{"\x1b[?1000l\x1b[?1049h\x1b[?1049$p", "\x1b[?1049;1$y"},
		{"\x1b[?1049l\x1b[?1049$p", "\x1b[?1049;2$y"},
	})
}

func TestVTAnswersKittyKeyboardQuery(t *testing.T) {
	v := NewVT(10, 5)
	runQuerySteps(t, v, []queryStep{
		{"\x1b[?u", "\x1b[?0u"},
		{"\x1b[>5u", ""},
		{"\x1b[?u", "\x1b[?5u"},
		{"\x1b[<u", ""},
		{"\x1b[?u", "\x1b[?0u"},
	})
}

// CSI ? u is a query, not SCORC (CSI u restore cursor): it must not restore a
// saved cursor.
func TestVTKittyQueryDoesNotRestoreCursor(t *testing.T) {
	v := NewVT(10, 5)
	v.Write([]byte("ab\x1b[s\x1b[3;5H\x1b[?u")) //nolint:errcheck
	if got := string(v.DrainReplies()); got != "\x1b[?0u" {
		t.Fatalf("reply = %q, want ESC[?0u", got)
	}
	if x, y := v.Cursor(); x != 4 || y != 2 {
		t.Fatalf("cursor after ESC[?u = (%d,%d), want (4,2): query restored saved cursor", x, y)
	}
}

func TestVTAnswersXTVersion(t *testing.T) {
	v := NewVT(80, 24)
	runQuerySteps(t, v, []queryStep{
		{"\x1b[>0q", "\x1bP>|loom-vt\x1b\\"},
		{"\x1b[>q", "\x1bP>|loom-vt\x1b\\"},
	})
}
