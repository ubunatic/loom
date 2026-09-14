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
