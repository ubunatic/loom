// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// TextInput is a single-line text editor with a movable caret. It is the shared
// editing primitive behind Settings KindString rows and wizard value entry: it
// owns a rune buffer and a caret index, and renders an optional prompt prefix
// plus a dim placeholder when empty.
//
// TextInput does not decide when editing starts or ends — the host widget calls
// HandleKey for input keys and reads Value when it wants the result. HandleKey
// returns consumed=true for keys it acted on (printable text, backspace, delete,
// left/right/home/end) so the host can keep navigation keys for itself.
type TextInput struct {
	Prompt      string // drawn before the value, e.g. "title> "
	Placeholder string // dim hint shown when the buffer is empty

	runes []rune
	caret int // caret index in [0, len(runes)]
}

// NewTextInput creates a TextInput seeded with value.
func NewTextInput(value string) *TextInput {
	t := &TextInput{}
	t.SetValue(value)
	return t
}

// Value returns the current buffer as a string.
func (t *TextInput) Value() string { return string(t.runes) }

// SetValue replaces the buffer and places the caret at the end.
func (t *TextInput) SetValue(s string) {
	t.runes = []rune(s)
	t.caret = len(t.runes)
}

// Caret returns the current caret index (rune offset from the start).
func (t *TextInput) Caret() int { return t.caret }

// HandleKey applies an editing key. It returns consumed=true when the key was an
// editing action; the host should treat consumed=false keys (enter, esc, …) as
// its own.
func (t *TextInput) HandleKey(e KeyEvent) (consumed bool) {
	switch e.Key {
	case "left":
		if t.caret > 0 {
			t.caret--
		}
		return true
	case "right":
		if t.caret < len(t.runes) {
			t.caret++
		}
		return true
	case "home":
		t.caret = 0
		return true
	case "end":
		t.caret = len(t.runes)
		return true
	case "backspace":
		if t.caret > 0 {
			t.runes = append(t.runes[:t.caret-1], t.runes[t.caret:]...)
			t.caret--
		}
		return true
	case "delete":
		if t.caret < len(t.runes) {
			t.runes = append(t.runes[:t.caret], t.runes[t.caret+1:]...)
		}
		return true
	}
	if e.Text != "" {
		ins := []rune(e.Text)
		// Insert at the caret without aliasing the backing array.
		next := make([]rune, 0, len(t.runes)+len(ins))
		next = append(next, t.runes[:t.caret]...)
		next = append(next, ins...)
		next = append(next, t.runes[t.caret:]...)
		t.runes = next
		t.caret += len(ins)
		return true
	}
	return false
}

// Draw renders the prompt, the value (or the dim placeholder when empty) into r,
// and — when focused — sets the canvas cursor at the caret column. The host owns
// focus, so it passes focused so an unfocused field shows no cursor.
func (t *TextInput) Draw(c *Canvas, r Rect, focused bool) {
	c.Fill(Rect{r.X, r.Y, r.W, 1}, Cell{Text: " "})
	x := r.X
	if t.Prompt != "" {
		x += c.Write(x, r.Y, t.Prompt, Style{})
	}
	if len(t.runes) == 0 && t.Placeholder != "" {
		c.Write(x, r.Y, t.Placeholder, Style{Dim: true})
	} else {
		c.Write(x, r.Y, string(t.runes), Style{})
	}
	if focused {
		c.CursorX = x + StringWidth(string(t.runes[:t.caret]))
		c.CursorY = r.Y
	}
}
