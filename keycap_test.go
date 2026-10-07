// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestKeyCap(t *testing.T) {
	for binding, want := range map[string]string{
		"ctrl-alt-s":          "⌃⌥S",
		"alt-ctrl-s":          "⌃⌥S",
		"ctrl-shift-z":        "⇧⌃Z",
		"shift-tab":           "⇧Tab",
		"ctrl-shift-alt-left": "⇧⌃⌥Left",
		"ctrl-s":              "⌃S",
		"f10":                 "F10",
		"f1":                  "F1",
		"pgdown":              "PgDn",
		"ctrl-space":          "⌃Space",
		"ctrl--":              "⌃-",
		"q":                   "Q",
		"":                    "",
	} {
		if got := KeyCap(binding); got != want {
			t.Errorf("KeyCap(%q) = %q, want %q", binding, got, want)
		}
	}
}

func TestMenuShortcutGlyphsMapBackToBindings(t *testing.T) {
	for shortcut, want := range map[string]string{"⌃⌥S": "ctrl-alt-s", "⇧⌃Z": "ctrl-shift-z", "⌃S": "ctrl-s", "^S": "ctrl-s", "F10": "f10"} {
		if got := menuKeyName(shortcut); got != want {
			t.Errorf("menuKeyName(%q) = %q, want %q", shortcut, got, want)
		}
	}
}

func TestSaveAsMenuHasNoShortcut(t *testing.T) {
	edit := NewRichTextEdit(nil)
	item := edit.ensureFileBar().menu.Menus[0].Items[3]
	if item.Shortcut != "" || item.Label != SpeccedDefaults.RichTextEdit.HotkeySaveAsLabel+"…" {
		t.Fatalf("Save as menu item = %+v", item)
	}
}
