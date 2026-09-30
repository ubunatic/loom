// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "strings"

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
	Mask        rune   // optional display rune repeated once per value rune; zero shows the value

	runes []rune
	caret int // caret index in [0, len(runes)]
	view  int // first rune shown when the value exceeds the available width
}

// NewTextInput creates a TextInput seeded with value.
func NewTextInput(value string) *TextInput {
	t := &TextInput{}
	t.SetValue(value)
	return t
}

// Value returns the current buffer as a string.
func (t *TextInput) Value() string { return string(t.runes) }

// ConsumePaste inserts single-line paste text at the caret.
func (t *TextInput) ConsumePaste(event PasteEvent) EventResult {
	text := strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ").Replace(event.Text)
	if text != "" {
		t.HandleKey(KeyEvent{Text: text})
	}
	return Handled()
}

// SetValue replaces the buffer and places the caret at the end.
func (t *TextInput) SetValue(s string) {
	t.runes = []rune(s)
	t.caret = len(t.runes)
	t.view = 0
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
		t.keepCaretInView()
		return true
	case "right":
		if t.caret < len(t.runes) {
			t.caret++
		}
		t.keepCaretInView()
		return true
	case "home":
		t.caret = 0
		t.view = 0
		return true
	case "end":
		t.caret = len(t.runes)
		return true
	case "backspace":
		if t.caret > 0 {
			t.runes = append(t.runes[:t.caret-1], t.runes[t.caret:]...)
			t.caret--
		}
		t.keepCaretInView()
		return true
	case "delete":
		if t.caret < len(t.runes) {
			t.runes = append(t.runes[:t.caret], t.runes[t.caret+1:]...)
		}
		t.keepCaretInView()
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
		t.keepCaretInView()
		return true
	}
	return false
}

func (t *TextInput) keepCaretInView() {
	if t.caret < t.view {
		t.view = t.caret
	}
}

// Draw renders the prompt, the value (or the dim placeholder when empty) into r,
// and — when focused — sets the canvas cursor at the caret column. The host owns
// focus, so it passes focused so an unfocused field shows no cursor.
func (t *TextInput) Draw(c *Canvas, r Rect, focused bool) {
	c.PaintSurface(Rect{r.X, r.Y, r.W, 1}, Style{})
	x := r.X
	if t.Prompt != "" {
		x += c.Write(x, r.Y, t.Prompt, Style{})
	}
	if len(t.runes) == 0 && t.Placeholder != "" {
		c.Write(x, r.Y, t.Placeholder, Style{Dim: true})
	} else {
		available := max(0, r.X+r.W-x)
		displayed := t.displayRunes()
		start := min(t.view, len(t.runes))
		// Move the window until the caret fits, accounting for one-column markers.
		for start < t.caret {
			reserve := 1 // leave a cell for the caret after the displayed text
			if start > 0 {
				reserve++ // left clipping marker
			}
			if displayWidth(displayed[start:t.caret]) <= max(0, available-reserve) {
				break
			}
			start++
		}
		left := start > 0
		used := 0
		if left && available > 0 {
			c.Write(x, r.Y, "‹", Style{})
			x++
			used++
		}
		end := start
		for end < len(displayed) {
			w := RuneWidth(displayed[end])
			remaining := available - used
			if end+1 < len(displayed) {
				remaining-- // reserve the right clipping marker
			} else if focused && t.caret == len(displayed) {
				remaining-- // keep the end caret inside the field
			}
			if w > remaining {
				break
			}
			used += w
			end++
		}
		c.Write(x, r.Y, string(displayed[start:end]), Style{})
		if end < len(displayed) && available-used > 0 {
			c.Write(x+used, r.Y, "›", Style{})
		}
		if focused {
			caret := min(t.caret, end)
			c.CursorX = x + displayWidth(displayed[start:caret])
			c.CursorY = r.Y
		}
		return
	}
	if focused {
		if t.Mask != 0 {
			c.CursorX = x + StringWidth(strings.Repeat(string(t.Mask), t.caret))
		} else {
			c.CursorX = x + StringWidth(string(t.runes[:t.caret]))
		}
		c.CursorY = r.Y
	}
}

func (t *TextInput) displayRunes() []rune {
	if t.Mask == 0 {
		return t.runes
	}
	return []rune(strings.Repeat(string(t.Mask), len(t.runes)))
}

func displayWidth(runes []rune) int {
	return StringWidth(string(runes))
}
