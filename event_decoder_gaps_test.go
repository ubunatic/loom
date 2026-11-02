// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

// TestDecodeKeyDecoderGaps pins the decoder cases from issue 404.
func TestDecodeKeyDecoderGaps(t *testing.T) {
	cases := []struct {
		name string
		seq  string
		want KeyEvent
	}{
		{"kitty_ctrl_shift_enter", "\x1b[13;6u", KeyEvent{Key: "ctrl-shift-enter"}},
		// Code 10 (LF) stays unmapped: no terminal sends it for Enter, and
		// naming it ctrl-enter would conflate it with Ctrl+J.
		{"modify_other_keys_ctrl_lf_unmapped", "\x1b[27;5;10~", KeyEvent{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DecodeKey([]byte(tc.seq)); got != tc.want {
				t.Errorf("DecodeKey(%q) = %+v, want %+v", tc.seq, got, tc.want)
			}
			got, n, ok := scanKey([]byte(tc.seq + "z"))
			if !ok || n != len(tc.seq) || got != tc.want {
				t.Errorf("scanKey(%q) = %+v, %d, %v; want %+v, %d", tc.seq, got, n, ok, tc.want, len(tc.seq))
			}
		})
	}
}

// TestScanKeyLoneEscThenCSI pins that a lone ESC at the end of a read is
// held back, and the full sequence decodes once the rest arrives (issue 404).
func TestScanKeyLoneEscThenCSI(t *testing.T) {
	if _, n, ok := scanKey([]byte{27}); ok || n != 0 {
		t.Fatalf("scanKey(ESC) = n=%d ok=%v, want n=0 ok=false", n, ok)
	}
	got, n, ok := scanKey([]byte("\x1b[A"))
	if !ok || n != 3 || got.Key != "up" {
		t.Fatalf("scanKey(ESC [ A) = %+v, %d, %v; want up, 3, true", got, n, ok)
	}
}
