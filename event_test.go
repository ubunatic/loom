// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestDecodeKey(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected KeyEvent
	}{
		{"empty", nil, KeyEvent{}},
		{"ctrl-c", []byte{3}, KeyEvent{Key: "ctrl-c"}},
		{"ctrl-d", []byte{4}, KeyEvent{Key: "ctrl-d"}},
		{"ctrl-b", []byte{2}, KeyEvent{Key: "ctrl-b"}},
		{"ctrl-f", []byte{6}, KeyEvent{Key: "ctrl-f"}},
		{"ctrl-u", []byte{21}, KeyEvent{Key: "ctrl-u"}},
		{"shift-tab", []byte{27, '[', 'Z'}, KeyEvent{Key: "shift-tab"}},
		{"enter", []byte{13}, KeyEvent{Key: "enter"}},
		{"esc", []byte{27}, KeyEvent{Key: "esc"}},
		{"up", []byte{27, '[', 'A'}, KeyEvent{Key: "up"}},
		{"down", []byte{27, '[', 'B'}, KeyEvent{Key: "down"}},
		{"left", []byte{27, '[', 'D'}, KeyEvent{Key: "left"}},
		{"right", []byte{27, '[', 'C'}, KeyEvent{Key: "right"}},
		{"shift-up xterm", []byte{27, '[', '1', ';', '2', 'A'}, KeyEvent{Key: "shift-up"}},
		{"shift-down xterm", []byte{27, '[', '1', ';', '2', 'B'}, KeyEvent{Key: "shift-down"}},
		{"ctrl-up xterm", []byte{27, '[', '1', ';', '5', 'A'}, KeyEvent{Key: "ctrl-up"}},
		{"ctrl-down xterm", []byte{27, '[', '1', ';', '5', 'B'}, KeyEvent{Key: "ctrl-down"}},
		{"alt-up xterm", []byte{27, '[', '1', ';', '3', 'A'}, KeyEvent{Key: "alt-up"}},
		{"alt-down xterm", []byte{27, '[', '1', ';', '3', 'B'}, KeyEvent{Key: "alt-down"}},
		{"shift-up rxvt", []byte{27, '[', 'a'}, KeyEvent{Key: "shift-up"}},
		{"shift-down rxvt", []byte{27, '[', 'b'}, KeyEvent{Key: "shift-down"}},
		{"f1 SS3", []byte{27, 'O', 'P'}, KeyEvent{Key: "f1"}},
		{"f2 SS3", []byte{27, 'O', 'Q'}, KeyEvent{Key: "f2"}},
		{"f3 SS3", []byte{27, 'O', 'R'}, KeyEvent{Key: "f3"}},
		{"f4 SS3", []byte{27, 'O', 'S'}, KeyEvent{Key: "f4"}},
		{"f5 tilde", []byte{27, '[', '1', '5', '~'}, KeyEvent{Key: "f5"}},
		{"f10 tilde", []byte{27, '[', '2', '1', '~'}, KeyEvent{Key: "f10"}},
		{"f12 tilde", []byte{27, '[', '2', '4', '~'}, KeyEvent{Key: "f12"}},
		{"home", []byte{27, '[', '1', '~'}, KeyEvent{Key: "home"}},
		{"delete", []byte{27, '[', '3', '~'}, KeyEvent{Key: "delete"}},
		{"end", []byte{27, '[', '4', '~'}, KeyEvent{Key: "end"}},
		{"pgup", []byte{27, '[', '5', '~'}, KeyEvent{Key: "pgup"}},
		{"pgdown", []byte{27, '[', '6', '~'}, KeyEvent{Key: "pgdown"}},
		{"printable t", []byte{'t'}, KeyEvent{Text: "t"}},
		{"printable ?", []byte{'?'}, KeyEvent{Text: "?"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecodeKey(tt.input)
			if got != tt.expected {
				t.Errorf("DecodeKey(%v) = %+v, want %+v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestKeyEventName(t *testing.T) {
	tests := []struct {
		name  string
		event KeyEvent
		want  string
	}{
		{"up", KeyEvent{Key: "up"}, "up"},
		{"pgup", KeyEvent{Key: "pgup"}, "pgup"},
		{"esc", KeyEvent{Key: "esc"}, "esc"},
		{"j", KeyEvent{Text: "j"}, "j"},
		{"q", KeyEvent{Text: "q"}, "q"},
		{"question mark", KeyEvent{Text: "?"}, "?"},
		{"key takes precedence", KeyEvent{Key: "up", Text: "j"}, "up"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.Name(); got != tt.want {
				t.Errorf("Name() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKeyEventIs(t *testing.T) {
	if !((KeyEvent{Key: "up"}).Is("down", "up", "left")) {
		t.Error("Is() = false, want true for a matching candidate")
	}
	if (KeyEvent{Text: "j"}).Is("q", "?") {
		t.Error("Is() = true, want false when no candidate matches")
	}
}

func TestKeyEventRune(t *testing.T) {
	tests := []struct {
		name  string
		event KeyEvent
		want  rune
	}{
		{"printable", KeyEvent{Text: "j"}, 'j'},
		{"unicode printable", KeyEvent{Text: "🙂"}, '🙂'},
		{"empty", KeyEvent{}, 0},
		{"non-printable", KeyEvent{Text: "\n"}, 0},
		{"special key", KeyEvent{Key: "up"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.Rune(); got != tt.want {
				t.Errorf("Rune() = %q, want %q", got, tt.want)
			}
		})
	}
}
