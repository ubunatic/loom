// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

// TestDecodeAltEnter pins that Alt+Enter has one stable name, alt-enter, in
// every encoding a terminal can send, and that the near misses keep their
// names (issue 398 M1).
func TestDecodeAltEnter(t *testing.T) {
	cases := []struct {
		name string
		seq  string
		want KeyEvent
	}{
		{"legacy_esc_cr", "\x1b\r", KeyEvent{Key: "alt-enter"}},
		{"modify_other_keys_alt_enter", "\x1b[27;3;13~", KeyEvent{Key: "alt-enter"}},
		{"kitty_alt_enter", "\x1b[13;3u", KeyEvent{Key: "alt-enter"}},
		{"near_miss_esc_lf", "\x1b\n", KeyEvent{Key: "alt-\n"}},
		{"modify_other_keys_shift_enter", "\x1b[27;2;13~", KeyEvent{Key: "shift-enter"}},
		{"modify_other_keys_ctrl_enter", "\x1b[27;5;13~", KeyEvent{Key: "ctrl-enter"}},
		{"modify_other_keys_bare_enter", "\x1b[27;1;13~", KeyEvent{}},
		{"modify_other_keys_alt_a", "\x1b[27;3;97~", KeyEvent{Key: "alt-a"}},
		{"modify_other_keys_ctrl_a", "\x1b[27;5;97~", KeyEvent{Key: "ctrl-a"}},
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
