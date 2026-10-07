// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"unicode"
)

var keyCapNames = map[string]string{
	"pgdown": "PgDn", "pgdn": "PgDn", "pagedown": "PgDn",
	"pgup": "PgUp", "pageup": "PgUp",
	"backspace": "Backspace", "esc": "Esc", "escape": "Esc",
}

// KeyCap renders a binding such as ctrl-alt-s as a keycap label such as ⌃⌥S.
// Modifier glyphs and their order come from the spec (keycaps.modifiers), so
// every hint bar, menu and help text shows the same caps.
func KeyCap(binding string) string {
	mods := SpeccedDefaults.KeyCaps.Modifiers
	rest := strings.TrimSpace(binding)
	if rest == "" {
		return ""
	}
	have := map[string]bool{}
	for {
		i := strings.IndexByte(rest, '-')
		if i <= 0 || i == len(rest)-1 {
			break
		}
		name := strings.ToLower(rest[:i])
		if name == "control" {
			name = "ctrl"
		}
		known := false
		for _, m := range mods {
			known = known || m.Name == name
		}
		if !known {
			break
		}
		have[name] = true
		rest = rest[i+1:]
	}
	var b strings.Builder
	for _, m := range mods {
		if have[m.Name] {
			b.WriteString(m.Glyph)
		}
	}
	b.WriteString(keyCapKey(rest))
	return b.String()
}

func keyCapKey(key string) string {
	if name, ok := keyCapNames[strings.ToLower(key)]; ok {
		return name
	}
	runes := []rune(key)
	if len(runes) == 1 {
		return strings.ToUpper(key)
	}
	if runes[0] == 'f' || runes[0] == 'F' {
		digits := true
		for _, r := range runes[1:] {
			digits = digits && unicode.IsDigit(r)
		}
		if digits {
			return strings.ToUpper(key)
		}
	}
	return strings.ToUpper(string(runes[:1])) + strings.ToLower(string(runes[1:]))
}

// capKey returns the entry's keycap: the explicit Key, else KeyCap(Binding).
func (e HintEntry) capKey() string {
	if e.Key != "" {
		return e.Key
	}
	return KeyCap(e.Binding)
}
