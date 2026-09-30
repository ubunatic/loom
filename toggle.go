// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// Toggle presents and changes a boolean value.
type Toggle struct{ Value *bool }

// NewToggle creates a toggle bound to value.
func NewToggle(value *bool) *Toggle { return &Toggle{Value: value} }

// String returns the current on/off mark.
func (t *Toggle) String() string {
	if t != nil && t.Value != nil && *t.Value {
		return "[✓]"
	}
	return "[ ]"
}

// Draw renders the current on/off mark.
func (t *Toggle) Draw(c *Canvas, r Rect) {
	c.PaintSurface(r, Style{})
	if t == nil {
		return
	}
	c.Write(r.X, r.Y, t.String(), Style{})
}

// HandleKey toggles the value on Enter or Space.
func (t *Toggle) HandleKey(e KeyEvent) (quit bool) {
	if t == nil || t.Value == nil {
		return false
	}
	if e.Key == "enter" || e.Key == "space" || e.Text == " " {
		*t.Value = !*t.Value
	}
	return false
}

// HandleMouse is a no-op.
func (*Toggle) HandleMouse(MouseEvent) bool { return false }
