// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
)

// SettingKind identifies how a Setting is rendered and edited.
type SettingKind int

const (
	KindBool   SettingKind = iota // toggle: true / false
	KindString                    // editable text value
	KindChoice                    // cycle through a fixed list of options
	KindNumber                    // bounded numeric value
)

// Setting is a single configurable option.
// Set the fields used by Kind: Bool, Str, Options+Index, or Num+Min+Max+Step.
type Setting struct {
	Label   string
	Kind    SettingKind
	Bool    *bool    // KindBool: pointer to the governed value
	Str     *string  // KindString: pointer to the governed value
	Options []string // KindChoice: available values
	Index   *int     // KindChoice: pointer to the current index
	Num     *float64 // KindNumber: pointer to the governed value
	Min     float64  // KindNumber: inclusive lower bound
	Max     float64  // KindNumber: inclusive upper bound
	Step    float64  // KindNumber: step amount; zero defaults to 1
	Format  string   // KindNumber: fmt format; empty defaults to %g
}

// value returns the human-readable current value for the setting.
func (s Setting) value() string {
	switch s.Kind {
	case KindBool:
		return NewToggle(s.Bool).String()
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
	case KindNumber:
		return s.numberInput().String()
	}
	return ""
}

func (s Setting) numberInput() *NumberInput {
	return &NumberInput{Value: s.Num, Min: s.Min, Max: s.Max, Step: s.Step, Format: s.Format}
}

// activate toggles or advances the setting's value.
func (s Setting) activate() {
	switch s.Kind {
	case KindBool:
		NewToggle(s.Bool).HandleKey(KeyEvent{Key: "enter"})
	case KindChoice:
		if s.Index != nil && len(s.Options) > 0 {
			*s.Index = (*s.Index + 1) % len(s.Options)
		}
	case KindNumber:
		s.stepNumber(1)
	}
}

// prev moves a KindChoice setting one step backward.
func (s Setting) prev() {
	if s.Kind == KindNumber {
		s.stepNumber(-1)
	} else if s.Kind == KindChoice && s.Index != nil && len(s.Options) > 0 {
		n := len(s.Options)
		*s.Index = (*s.Index + n - 1) % n
	}
}

func (s Setting) stepNumber(direction float64) {
	s.numberInput().StepBy(direction)
}

// Settings is a navigable list of configurable options.
// ↑ ↓ move between settings; Enter / → toggle, advance or edit;
// ← steps a KindChoice or KindNumber backward. Esc returns quit=true.
//
// A KindString row enters inline edit mode on Enter/→: printable keys append,
// Backspace deletes, Enter commits, Esc cancels (restoring the prior value
// without quitting the widget). Navigation is locked while editing.
type Settings struct {
	Items       []Setting
	Prompt      string // header line; empty = no header
	sel         int
	editing     bool       // a KindString or KindNumber row is in inline edit mode
	editOrig    string     // value captured at edit start, restored on cancel
	editor      *TextInput // active line editor while editing a KindString row
	numberInput *NumberInput
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
		editing := sel && s.editing && (item.Kind == KindString || item.Kind == KindNumber)
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
	if s.numberInput != nil && s.numberInput.Error() != "" && len(s.Items) > 0 {
		y := r.Y + start + len(s.Items)
		if y < r.Y+r.H {
			c.Write(r.X, y, s.numberInput.Error(), Style{Dim: true})
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
		if s.Items[s.sel].Kind == KindNumber && e.Key == "right" {
			s.Items[s.sel].stepNumber(1)
		} else if s.Items[s.sel].Kind == KindString {
			s.beginEdit()
		} else if s.Items[s.sel].Kind == KindNumber {
			s.beginNumberEdit()
		} else {
			s.Items[s.sel].activate()
		}
	case "left":
		if s.Items[s.sel].Kind == KindNumber {
			s.Items[s.sel].stepNumber(-1)
		} else {
			s.Items[s.sel].prev()
		}
	case "-", "minus":
		if s.Items[s.sel].Kind == KindNumber {
			s.Items[s.sel].stepNumber(-1)
		}
	case "+", "plus":
		if s.Items[s.sel].Kind == KindNumber {
			s.Items[s.sel].stepNumber(1)
		}
	default:
		if s.Items[s.sel].Kind == KindNumber && (e.Text == "+" || e.Text == "-") {
			direction := 1.0
			if e.Text == "-" {
				direction = -1
			}
			s.Items[s.sel].stepNumber(direction)
		}
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

func (s *Settings) beginNumberEdit() {
	item := s.Items[s.sel]
	if item.Kind != KindNumber || item.Num == nil {
		return
	}
	s.numberInput = item.numberInput()
	s.numberInput.beginEdit()
	s.editor = s.numberInput.editor
	s.editing = true
}

// handleEditKey drives the line editor for the in-edit KindString value. Enter
// commits the editor's value into Str, Esc cancels (restoring the captured
// value); neither quits the widget. All other editing keys go to the editor.
func (s *Settings) handleEditKey(e KeyEvent) (quit bool) {
	item := s.Items[s.sel]
	if item.Kind == KindNumber {
		if s.numberInput != nil {
			s.numberInput.HandleKey(e)
		}
		s.editor = nil
		if s.numberInput != nil {
			s.editor = s.numberInput.editor
		}
		if !s.numberInput.Editing() {
			s.endEdit()
		}
		return false
	}
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
	s.numberInput = nil
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
	if s.numberInput != nil && s.numberInput.Error() != "" {
		h++
	}
	if s.Prompt != "" {
		h++
	}
	return h
}
