// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"sort"
	"strings"
)

// KeyHelp renders the labeled bindings in a KeyMap as one compact help line.
// The first configured key alias for each labeled action is shown.
type KeyHelp struct {
	KeyMap    *KeyMap
	Separator string
	Style     Style
}

// ApplyTheme sets the help text and surface colors.
func (h *KeyHelp) ApplyTheme(theme ThemeColors) {
	h.Style = Style{FG: theme.NormalFG.Color(), BG: theme.NormalBG.Color()}
}

// NewKeyHelp creates a compact key help widget for km.
func NewKeyHelp(km *KeyMap) *KeyHelp {
	return &KeyHelp{KeyMap: km}
}

// Text returns the compact, deterministic key help line.
func (h *KeyHelp) Text() string {
	if h == nil || h.KeyMap == nil {
		return ""
	}
	separator := h.Separator
	if separator == "" {
		separator = " · "
	}
	actions := make([]string, 0, len(h.KeyMap.labels))
	for action, label := range h.KeyMap.labels {
		if label != "" && len(h.KeyMap.actions[action]) != 0 {
			actions = append(actions, action)
		}
	}
	sort.Strings(actions)
	parts := make([]string, 0, len(actions))
	for _, action := range actions {
		parts = append(parts, KeyCap(h.KeyMap.actions[action][0])+" "+h.KeyMap.labels[action])
	}
	return joinKeyHelp(parts, separator)
}

// Draw renders key help clipped to the supplied rectangle.
func (h *KeyHelp) Draw(c *Canvas, r Rect) {
	if h == nil || c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	c.PaintDefaultSurface(r, h.Style)
	text := TruncateText(h.Text(), r.W, "…")
	if text != "" {
		c.WriteDefault(r.X, r.Y, text, h.Style)
	}
}

// ConsumeKey does not consume keyboard events.
func (h *KeyHelp) ConsumeKey(KeyEvent) EventResult { return Ignored() }

// ConsumeMouse does not consume mouse events.
func (h *KeyHelp) ConsumeMouse(MouseEvent) EventResult { return Ignored() }

func joinKeyHelp(parts []string, separator string) string {
	if len(parts) == 0 {
		return ""
	}
	text := parts[0]
	for _, part := range parts[1:] {
		text += separator + part
	}
	return text
}

// KeyHelpRow is one rendered row of a sectioned key help table.
type KeyHelpRow struct {
	Header bool   // section title row
	Key    string // padded keycap column (empty for headers and wrapped continuations)
	Text   string // description, or the section title for headers
}

// helpEntryCap returns the keycap text of an entry: Cap, else its keys joined by spaces.
func helpEntryCap(entry HelpEntry) (string, string) {
	label := entry.Label
	keys := entry.Keys
	if entry.Ref != "" {
		binding, refLabel, _ := SpeccedDefaults.RichTextEdit.helpRef(entry.Ref)
		keys = []string{binding}
		if label == "" {
			label = refLabel
		}
	}
	if entry.Cap != "" {
		return entry.Cap, label
	}
	caps := make([]string, 0, len(keys))
	for _, key := range keys {
		caps = append(caps, KeyCap(key))
	}
	return strings.Join(caps, " "), label
}

// KeyHelpSections lays out grouped help entries as a two-column table of the given
// width: section headers, an aligned keycap column and wrapped descriptions.
func KeyHelpSections(sections []HelpSection, width int) []KeyHelpRow {
	if width <= 0 {
		return nil
	}
	type pair struct{ key, label string }
	keyWidth := 0
	resolved := make([][]pair, len(sections))
	for i, section := range sections {
		for _, entry := range section.Entries {
			key, label := helpEntryCap(entry)
			resolved[i] = append(resolved[i], pair{key, label})
			keyWidth = max(keyWidth, StringWidth(key))
		}
	}
	keyWidth = min(keyWidth, max(1, width/2))
	const gap = 2
	descWidth := max(1, width-keyWidth-gap)
	var rows []KeyHelpRow
	for i, section := range sections {
		if len(rows) > 0 {
			rows = append(rows, KeyHelpRow{})
		}
		rows = append(rows, KeyHelpRow{Header: true, Text: TruncateText(section.Title, width, "…")})
		for _, p := range resolved[i] {
			key := TruncateText(p.key, keyWidth, "…")
			key += strings.Repeat(" ", keyWidth-StringWidth(key))
			for j, text := range wrapRichTextHelpLine(p.label, descWidth) {
				if j > 0 {
					key = strings.Repeat(" ", keyWidth)
				}
				rows = append(rows, KeyHelpRow{Key: key, Text: text})
			}
		}
	}
	return rows
}
