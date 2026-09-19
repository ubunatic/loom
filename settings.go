// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "fmt"

// SettingKind identifies how a Setting is rendered and edited.
type SettingKind int

const (
	KindBool   SettingKind = iota // toggle: true / false
	KindString                    // editable text value
	KindChoice                    // cycle through a fixed list of options
)

// Setting is a single configurable option.
// Set exactly one of Bool, Str, or Options+Index depending on Kind.
type Setting struct {
	Label   string
	Kind    SettingKind
	Bool    *bool    // KindBool: pointer to the governed value
	Str     *string  // KindString: pointer to the governed value
	Options []string // KindChoice: available values
	Index   *int     // KindChoice: pointer to the current index
}

// value returns the human-readable current value for the setting.
func (s Setting) value() string {
	switch s.Kind {
	case KindBool:
		if s.Bool != nil && *s.Bool {
			return "[✓]"
		}
		return "[ ]"
	case KindString:
		if s.Str != nil {
			return *s.Str
		}
		return ""
	case KindChoice:
		if s.Index != nil && len(s.Options) > 0 {
			i := *s.Index
			if i < 0 || i >= len(s.Options) {
				i = 0
			}
			out := ""
			for j, opt := range s.Options {
				if j > 0 {
					out += "  "
				}
				if j == i {
					out += "● " + opt
				} else {
					out += "○ " + opt
				}
			}
			return out
		}
	}
	return ""
}

// activate toggles or advances the setting's value.
func (s Setting) activate() {
	switch s.Kind {
	case KindBool:
		if s.Bool != nil {
			*s.Bool = !*s.Bool
		}
	case KindChoice:
		if s.Index != nil && len(s.Options) > 0 {
			*s.Index = (*s.Index + 1) % len(s.Options)
		}
	}
}

// prev moves a KindChoice setting one step backward.
func (s Setting) prev() {
	if s.Kind == KindChoice && s.Index != nil && len(s.Options) > 0 {
		n := len(s.Options)
		*s.Index = (*s.Index + n - 1) % n
	}
}

// Settings is a navigable list of configurable options.
// ↑ ↓ move between settings; Enter / → toggle or advance;
// ← steps a KindChoice backward. Esc returns quit=true.
//
// A KindString row enters inline edit mode on Enter/→: printable keys append,
// Backspace deletes, Enter commits, Esc cancels (restoring the prior value
// without quitting the widget). Navigation is locked while editing.
type Settings struct {
	Items    []Setting
	Prompt   string // header line; empty = no header
	sel      int
	editing  bool       // a KindString row is in inline edit mode
	editOrig string     // value captured at edit start, restored on cancel
	editor   *TextInput // active line editor while editing a KindString row
}

// NewSettings creates a Settings widget.
func NewSettings(items []Setting) *Settings {
	return &Settings{Items: items}
}

// Draw renders each setting as a labeled row with its current value.
// The selected row is bold; unselected rows are normal.
func (s *Settings) Draw(c *Canvas, r Rect) {
	start := 0
	if s.Prompt != "" {
		c.Write(r.X, r.Y, s.Prompt, Style{Dim: true})
		start = 1
	}
	for i, item := range s.Items {
		y := r.Y + start + i
		if y >= r.Y+r.H {
			break
		}
		c.PaintSurface(Rect{r.X, y, r.W, 1}, Style{})
		sel := i == s.sel
		labelStyle := Style{}
		valueStyle := Style{Dim: true}
		marker := "  "
		if sel {
			labelStyle = Style{Bold: true}
			valueStyle = Style{}
			marker = "▶ "
		}
		label := fmt.Sprintf("%-16s", item.Label)
		n := c.Write(r.X, y, marker+label, labelStyle)
		editing := sel && s.editing && item.Kind == KindString
		if editing {
			valueStyle = Style{} // never dim the value being edited
		}
		if editing && s.editor != nil {
			// Draw the live editor (with caret) instead of the static value.
			s.editor.Draw(c, Rect{r.X + n, y, r.W - n, 1}, true)
		} else {
			c.Write(r.X+n, y, item.value(), valueStyle)
		}
	}
}

// HandleKey navigates settings and edits values.
func (s *Settings) HandleKey(e KeyEvent) (quit bool) {
	n := len(s.Items)
	if n == 0 {
		if e.Key == "esc" || e.Key == "ctrl-c" {
			return true
		}
		return false
	}
	if s.editing {
		return s.handleEditKey(e)
	}
	switch e.Key {
	case "esc", "ctrl-c":
		return true
	case "up":
		if s.sel > 0 {
			s.sel--
		} else {
			s.sel = n - 1
		}
	case "down":
		if s.sel < n-1 {
			s.sel++
		} else {
			s.sel = 0
		}
	case "enter", "right":
		if s.Items[s.sel].Kind == KindString {
			s.beginEdit()
		} else {
			s.Items[s.sel].activate()
		}
	case "left":
		s.Items[s.sel].prev()
	}
	return false
}

// beginEdit puts the selected KindString row into edit mode, capturing the
// current value so Esc can restore it. No-op for other kinds or a nil pointer.
func (s *Settings) beginEdit() {
	item := s.Items[s.sel]
	if item.Kind != KindString || item.Str == nil {
		return
	}
	s.editOrig = *item.Str
	s.editor = NewTextInput(*item.Str)
	s.editing = true
}

// handleEditKey drives the line editor for the in-edit KindString value. Enter
// commits the editor's value into Str, Esc cancels (restoring the captured
// value); neither quits the widget. All other editing keys go to the editor.
func (s *Settings) handleEditKey(e KeyEvent) (quit bool) {
	item := s.Items[s.sel]
	switch e.Key {
	case "enter":
		if item.Str != nil && s.editor != nil {
			*item.Str = s.editor.Value()
		}
		s.endEdit()
	case "esc", "ctrl-c":
		if item.Str != nil {
			*item.Str = s.editOrig
		}
		s.endEdit()
	default:
		// Keep the governed value live so callers see edits as they happen;
		// editOrig still restores it on cancel.
		if s.editor != nil {
			s.editor.HandleKey(e)
			if item.Str != nil {
				*item.Str = s.editor.Value()
			}
		}
	}
	return false
}

// endEdit leaves inline edit mode and releases the editor.
func (s *Settings) endEdit() {
	s.editing = false
	s.editor = nil
}

// Editing reports whether a KindString row is currently being edited.
func (s *Settings) Editing() bool { return s.editing }

// HandleMouse is a no-op for now.
func (s *Settings) HandleMouse(MouseEvent) (quit bool) { return false }

// Sel returns the index of the currently selected setting.
func (s *Settings) Sel() int { return s.sel }

// ContentHeight estimates the required height for this settings list.
func (s *Settings) ContentHeight() int {
	h := len(s.Items)
	if s.Prompt != "" {
		h++
	}
	return h
}
