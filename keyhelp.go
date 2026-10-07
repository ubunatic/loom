// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "sort"

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
