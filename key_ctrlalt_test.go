// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

// Canary for issue 295: every wire form of Ctrl+Alt+S decodes to ctrl-alt-s.
func TestDecodeKeyCtrlAltS(t *testing.T) {
	for name, seq := range map[string][]byte{
		"legacy ESC ^S":   {27, 0x13},
		"CSI-u":           []byte("\x1b[115;7u"),
		"modifyOtherKeys": []byte("\x1b[27;7;115~"),
	} {
		if got := DecodeKey(seq).Key; got != "ctrl-alt-s" {
			t.Errorf("%s %q = %q, want ctrl-alt-s", name, seq, got)
		}
	}
	for _, seq := range [][]byte{{27, 8}, {27, 9}, {27, 13}} {
		if got := DecodeKey(seq).Key; got == "ctrl-alt-h" || got == "ctrl-alt-i" || got == "ctrl-alt-m" {
			t.Errorf("%q = %q, want historical alt- name", seq, got)
		}
	}
}

// A split CSI-u or modifyOtherKeys sequence must be framed as one key, not
// leak its parameters into the text stream.
func TestScanKeyFramesModifiedLetterSequences(t *testing.T) {
	for _, seq := range []string{"\x1b[115;7u", "\x1b[27;7;115~"} {
		key, used, ok := scanKey([]byte(seq + "x"))
		if !ok || used != len(seq) || key.Key != "ctrl-alt-s" {
			t.Errorf("scanKey(%q) = %q, %d, %v", seq, key.Key, used, ok)
		}
		if _, _, ok := scanKey([]byte(seq[:len(seq)-2])); ok {
			t.Errorf("scanKey(%q) accepted an unfinished prefix", seq[:len(seq)-2])
		}
	}
}
