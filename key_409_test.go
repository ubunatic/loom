// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

// Issue 409: xterm modified cursor keys decode for every modifier value, so
// none is dropped or leaks its tail ("9D") as text. Foot was seen leaking
// "9D" for Left, which matches CSI 1;9D.
func TestDecodeModifiedCursorKeysAnyModifier(t *testing.T) {
	cases := []struct{ in, want string }{
		{"\x1b[D", "left"},
		{"\x1bOD", "left"},
		{"\x1b[H", "home"},
		{"\x1bOF", "end"},
		{"\x1b[1;1D", "left"},
		{"\x1b[1;1H", "home"},
		{"\x1b[1;2C", "shift-right"},
		{"\x1b[1;5D", "ctrl-left"},
		{"\x1b[1;5H", "ctrl-home"},
		{"\x1b[1;5F", "ctrl-end"},
		{"\x1b[1;8A", "ctrl-shift-alt-up"},
		{"\x1b[1;9D", "super-left"},
		{"\x1b[1;9C", "super-right"},
		{"\x1b[1;9A", "super-up"},
		{"\x1b[1;9B", "super-down"},
		{"\x1b[1;9H", "super-home"},
		{"\x1b[1;9F", "super-end"},
		{"\x1b[1;13D", "ctrl-super-left"},
		{"\x1b[1;65D", "left"},       // Caps Lock bit ignored
		{"\x1b[1;5:1D", "ctrl-left"}, // kitty event type
		{"\x1b[1;3R", "alt-f3"},
	}
	for _, tc := range cases {
		ke, n, ok := ScanKey([]byte(tc.in + "x"))
		if !ok || n != len(tc.in) {
			t.Errorf("ScanKey(%q) = n %d ok %v, want n %d", tc.in, n, ok, len(tc.in))
			continue
		}
		if ke.Key != tc.want || ke.Text != "" {
			t.Errorf("ScanKey(%q) = key %q text %q, want key %q", tc.in, ke.Key, ke.Text, tc.want)
		}
	}
}

// A cursor position report for row 1 must stay a report, not an F3 key.
func TestCursorReportRowOneIsNotAKey(t *testing.T) {
	for _, in := range []string{"\x1b[1;1R", "\x1b[1;80R", "\x1b[1;9R"} {
		ke, n, ok := ScanKey([]byte(in))
		if !ok || n != len(in) || ke.Key != "" || ke.Text != "" {
			t.Errorf("ScanKey(%q) = %+v n %d ok %v, want a report with no key", in, ke, n, ok)
		}
	}
}

func TestTextAreaLineAndWordNavigation(t *testing.T) {
	type step struct {
		key     string
		row, co int
	}
	ta := NewTextArea("foo bar_1  baz\nqux quux")
	// SetValue leaves the caret at the end of the text.
	steps := []step{
		{"ctrl-home", 0, 0},
		{"ctrl-right", 0, 3},
		{"ctrl-right", 0, 9},
		{"ctrl-right", 0, 14},
		{"ctrl-right", 1, 0},
		{"ctrl-right", 1, 3},
		{"ctrl-left", 1, 0},
		{"ctrl-left", 0, 14},
		{"ctrl-left", 0, 11},
		{"ctrl-a", 0, 0},
		{"ctrl-e", 0, 14},
		{"home", 0, 0},
		{"end", 0, 14},
		{"ctrl-end", 1, 8},
	}
	for i, s := range steps {
		if r := ta.ConsumeKey(KeyEvent{Key: s.key}); !r.Consumed {
			t.Fatalf("step %d %s not consumed", i, s.key)
		}
		if ta.row != s.row || ta.col != s.co {
			t.Fatalf("step %d %s: caret (%d,%d), want (%d,%d)", i, s.key, ta.row, ta.col, s.row, s.co)
		}
	}
}

func TestTextAreaCtrlUDeletesToLineStart(t *testing.T) {
	ta := NewTextArea("one\ntwo three")
	ta.ConsumeKey(KeyEvent{Key: "ctrl-end"})
	for range len("three") {
		ta.ConsumeKey(KeyEvent{Key: "left"})
	}
	ta.ConsumeKey(KeyEvent{Key: "ctrl-u"})
	if got := ta.Value(); got != "one\nthree" {
		t.Fatalf("Value() = %q, want %q", got, "one\nthree")
	}
	if ta.row != 1 || ta.col != 0 {
		t.Fatalf("caret (%d,%d), want (1,0)", ta.row, ta.col)
	}
}

// End to end: the foot byte form reaches TextArea as a cursor move, never
// as inserted text.
func TestTextAreaFootCursorBytesDoNotInsertText(t *testing.T) {
	ta := NewTextArea("abc")
	b := []byte("\x1b[1;1D\x1b[1;1H\x1b[1;1F\x1b[D")
	for len(b) > 0 {
		ke, n, ok := ScanKey(b)
		if !ok || n == 0 {
			t.Fatalf("ScanKey stuck on %q", b)
		}
		ta.ConsumeKey(ke)
		b = b[n:]
	}
	if got := ta.Value(); got != "abc" {
		t.Fatalf("Value() = %q, want %q (cursor bytes leaked as text)", got, "abc")
	}
	if ta.col != 2 {
		t.Fatalf("caret col %d, want 2", ta.col)
	}
}
