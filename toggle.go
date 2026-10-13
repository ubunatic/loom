// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// Toggle presents and changes a boolean value.
type Toggle struct {
	Value *bool
	Label string
}

// NewToggle creates a toggle bound to value.
func NewToggle(value *bool) *Toggle { return &Toggle{Value: value} }

// String returns the current on/off mark.
func (t *Toggle) String() string {
	if t != nil && t.Value != nil && *t.Value {
		return SpeccedDefaults.Toggle.OnMark
	}
	return SpeccedDefaults.Toggle.OffMark
}

// Draw renders the current on/off mark.
func (t *Toggle) Draw(c *Canvas, r Rect) {
	c.PaintSurface(r, Style{})
	if t == nil {
		return
	}
	text := t.String()
	if t.Label != "" {
		text += " " + t.Label
	}
	c.Write(r.X, r.Y, text, Style{})
}

// ConsumeKey toggles the value on Enter or Space.
func (t *Toggle) ConsumeKey(e KeyEvent) (quit EventResult) {
	if t == nil || t.Value == nil {
		return Ignored()
	}
	if e.Key == "enter" || e.Key == "space" || e.Text == " " {
		*t.Value = !*t.Value
		return Handled()
	}
	return Ignored()
}

// ConsumeMouse toggles the value when clicked within its rendered mark.
func (t *Toggle) ConsumeMouse(e MouseEvent) EventResult {
	markLen := StringWidth(SpeccedDefaults.Toggle.OnMark)
	if t == nil || t.Value == nil || e.Action != MousePress || e.Button != MouseLeft || e.X < 0 || e.X >= markLen || e.Y != 0 {
		return Ignored()
	}
	*t.Value = !*t.Value
	return Handled()
}
