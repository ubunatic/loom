// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
)

func TestFrameVisibilitySequences(t *testing.T) {
	root, _, err := BuildWidget(strings.NewReader(shellFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	f := root.(*Frame)
	f.Title = "unchanged clock snapshot"
	for _, tc := range []struct {
		key    string
		hidden [2]bool
	}{
		{"u", [2]bool{true, false}}, {"l", [2]bool{true, true}},
		{"u", [2]bool{false, true}}, {"u", [2]bool{true, true}},
		{"l", [2]bool{true, false}}, {"u", [2]bool{false, false}},
	} {
		if f.HandleKey(KeyEvent{Text: tc.key}) {
			t.Fatal("toggle quit")
		}
		for i, want := range tc.hidden {
			if f.Boxes[i].Hidden != want {
				t.Fatalf("key %s: %+v", tc.key, f.Boxes)
			}
		}
		for _, width := range []int{64, 40, 64} {
			height := f.HeightForWidth(width)
			layout := f.Layout(width, height)
			for i, hidden := range tc.hidden {
				if hidden && layout[i] != (Rect{}) {
					t.Fatalf("hidden box retained space: %+v", layout)
				}
				if !hidden && layout[i].Y != 1 && (tc.hidden[0] || tc.hidden[1]) {
					t.Fatal("survivor did not reflow")
				}
			}
			if tc.hidden[0] && tc.hidden[1] && height != 2 {
				t.Fatal("all-hidden chrome height")
			}
			Render(f, width, height)
		}
		for i, name := range []string{"usage", "load"} {
			state := "on"
			if tc.hidden[i] {
				state = "off"
			}
			if !strings.Contains(f.StatusText(), name+":"+state) {
				t.Fatal(f.StatusText())
			}
		}
		if f.Title != "unchanged clock snapshot" {
			t.Fatal("state reset")
		}
	}
	if !f.HandleKey(KeyEvent{Text: "q"}) || !f.HandleKey(KeyEvent{Key: "ctrl-c"}) {
		t.Fatal("quit not wired")
	}
	if f.HandleKey(KeyEvent{Text: "z"}) {
		t.Fatal("unknown key handled")
	}
}

func TestFrameActionValidation(t *testing.T) {
	for _, tc := range []struct{ old, new string }{
		{"key: l", "key: u"}, {"target: load", "target: missing"},
		{"action: toggle", "action: launch"}, {"key: l", "key: unsupported"},
		{"id: toggle_load", "id: toggle_usage"}, {"hidden_hint: '[l]load:off'", "hidden_hint: ''"},
	} {
		if _, _, err := BuildWidget(strings.NewReader(strings.Replace(shellFixture(t), tc.old, tc.new, 1))); err == nil {
			t.Fatalf("accepted %s", tc.new)
		}
	}
	source := strings.Replace(shellFixture(t), "      - id: usage", "      - id: usage\n        hidden: true", 1)
	root, _, err := BuildWidget(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	f := root.(*Frame)
	if !f.Boxes[0].Hidden || strings.Contains(strings.Join(Render(f, 64, 9), "\n"), "All Usage") {
		t.Fatal("initial visibility ignored")
	}
	if !strings.Contains(f.StatusText(), "usage:off") {
		t.Fatal("hidden hint missing")
	}
}
