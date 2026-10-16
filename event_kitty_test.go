// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"testing"
)

// TestDecodeCSIuSpecialKeys covers the keys kitty flag 1 (Pane.KittyKeyboard)
// turns into CSI-u: every modifier combination keeps the legacy base name.
func TestDecodeCSIuSpecialKeys(t *testing.T) {
	keys := []struct {
		code int
		name string
	}{{13, "enter"}, {27, "esc"}, {127, "backspace"}, {9, "tab"}}
	mods := []struct {
		param  int
		prefix string
	}{
		{1, ""},
		{2, "shift-"},
		{3, "alt-"},
		{4, "shift-alt-"},
		{5, "ctrl-"},
		{6, "ctrl-shift-"},
		{7, "ctrl-alt-"},
		{8, "ctrl-shift-alt-"},
	}
	for _, k := range keys {
		t.Run(k.name+" bare", func(t *testing.T) {
			if got := DecodeKey([]byte(fmt.Sprintf("\x1b[%du", k.code))); got != (KeyEvent{Key: k.name}) {
				t.Errorf("got %+v, want %s", got, k.name)
			}
		})
		for _, m := range mods {
			want := KeyEvent{Key: m.prefix + k.name}
			for _, seq := range []string{
				fmt.Sprintf("\x1b[%d;%du", k.code, m.param),
				fmt.Sprintf("\x1b[%d;%d:1u", k.code, m.param),  // event type: press
				fmt.Sprintf("\x1b[%d;%du", k.code, m.param+64), // caps lock is ignored
			} {
				t.Run(fmt.Sprintf("%q", seq), func(t *testing.T) {
					if got := DecodeKey([]byte(seq)); got != want {
						t.Errorf("DecodeKey(%q) = %+v, want %+v", seq, got, want)
					}
					got, n, ok := scanKey([]byte(seq + "z"))
					if !ok || n != len(seq) || got != want {
						t.Errorf("scanKey(%q) = %+v, %d, %v; want %+v, %d", seq, got, n, ok, want, len(seq))
					}
				})
			}
		}
	}
}

// TestDecodeCSIuPrintable keeps legacy names for keys the protocol re-encodes:
// unmodified printable code points are text, modified ones "<mods>-<rune>".
func TestDecodeCSIuPrintable(t *testing.T) {
	for seq, want := range map[string]KeyEvent{
		"\x1b[105;5u":   {Key: "ctrl-i"},
		"\x1b[106;5u":   {Key: "ctrl-j"},
		"\x1b[109;5u":   {Key: "ctrl-m"},
		"\x1b[99;5u":    {Key: "ctrl-c"},
		"\x1b[97;3u":    {Key: "alt-a"},
		"\x1b[49;3u":    {Key: "alt-1"},
		"\x1b[47;5u":    {Key: "ctrl-/"},
		"\x1b[228;3u":   {Key: "alt-ä"},
		"\x1b[97u":      {Text: "a"},
		"\x1b[97:65;2u": {Key: "shift-a"},
		"\x1b[57441u":   {Key: "key-57441"}, // other private-use key (left shift) is named, not dropped
		"\x1b[57358;5u": {Key: "ctrl-key-57358"},
		"\x1b[57376u":   {Key: "f13"},
		"\x1b[57398;2u": {Key: "shift-f35"},
		"\x1b[97;9u":    {Key: "super-a"},
		"\x1b[97;17u":   {Key: "hyper-a"},
		"\x1b[97;33u":   {Key: "meta-a"},
		"\x1b[97;13u":   {Key: "ctrl-super-a"},
		"\x1b[13;9u":    {Key: "super-enter"},
		"\x1b[97;65u":   {Text: "a"}, // caps lock alone stays text
	} {
		if got := DecodeKey([]byte(seq)); got != want {
			t.Errorf("DecodeKey(%q) = %+v, want %+v", seq, got, want)
		}
	}
}

// TestDecodeLegacyKeysUnchanged pins the legacy byte forms the CSI-u decoder
// must not disturb: without the protocol these are what terminals send.
func TestDecodeLegacyKeysUnchanged(t *testing.T) {
	for seq, want := range map[string]string{
		"\r":     "enter",
		"\n":     "enter",
		"\t":     "tab",
		"\x7f":   "backspace",
		"\x08":   "backspace",
		"\x1b":   "esc",
		"\x1b[Z": "shift-tab",
		"\x1b\r": "alt-\r",
		"\x1ba":  "alt-a",
		"\x01":   "ctrl-a",
	} {
		if got := DecodeKey([]byte(seq)).Name(); got != want {
			t.Errorf("DecodeKey(%q) = %q, want %q", seq, got, want)
		}
	}
}

// TestDecodeCSIuKeypad maps kitty keypad codes (57399-57427) to the legacy
// names and characters the same keys produce without the protocol.
func TestDecodeCSIuKeypad(t *testing.T) {
	want := map[int]KeyEvent{
		57399: {Text: "0"}, 57400: {Text: "1"}, 57401: {Text: "2"}, 57402: {Text: "3"},
		57403: {Text: "4"}, 57404: {Text: "5"}, 57405: {Text: "6"}, 57406: {Text: "7"},
		57407: {Text: "8"}, 57408: {Text: "9"}, 57409: {Text: "."}, 57410: {Text: "/"},
		57411: {Text: "*"}, 57412: {Text: "-"}, 57413: {Text: "+"}, 57414: {Key: "enter"},
		57415: {Text: "="}, 57416: {Text: ","}, 57417: {Key: "left"}, 57418: {Key: "right"},
		57419: {Key: "up"}, 57420: {Key: "down"}, 57421: {Key: "pgup"}, 57422: {Key: "pgdown"},
		57423: {Key: "home"}, 57424: {Key: "end"}, 57425: {Key: "insert"}, 57426: {Key: "delete"},
		57427: {Key: "begin"},
	}
	for code := 57399; code <= 57427; code++ {
		w := want[code]
		if got := DecodeKey([]byte(fmt.Sprintf("\x1b[%du", code))); got != w {
			t.Errorf("code %d = %+v, want %+v", code, got, w)
		}
		name := w.Key
		if name == "" {
			name = w.Text
		}
		mod := KeyEvent{Key: "ctrl-shift-" + name}
		seq := fmt.Sprintf("\x1b[%d;6u", code)
		if got := DecodeKey([]byte(seq)); got != mod {
			t.Errorf("DecodeKey(%q) = %+v, want %+v", seq, got, mod)
		}
		if got, n, ok := scanKey([]byte(seq)); !ok || n != len(seq) || got != mod {
			t.Errorf("scanKey(%q) = %+v, %d, %v", seq, got, n, ok)
		}
	}
}
