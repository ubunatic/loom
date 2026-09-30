// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strings"

	"codeberg.org/ubunatic/loom/measure"
)

// FormField binds a label and optional validation to an existing input widget.
type FormField struct {
	Label    string
	Widget   any // commonly TextInput (which needs a focused Draw argument), NumberInput, Toggle, or Choice
	Help     string
	Validate func() error
	Required bool
}

// FormAction is an action in the form's footer. Cancel actions invoke OnCancel.
type FormAction struct {
	Label      string
	Cancel     bool
	OnActivate func()
}

// Form arranges input fields, validates them, and submits their current values.
type Form struct {
	Fields      []FormField
	Actions     []FormAction
	OnSubmit    func(map[string]any)
	OnCancel    func()
	Validation  map[int]string
	focused     int
	focusedSelf bool
}

// NewForm constructs a form with the first field focused.
func NewForm(fields []FormField) *Form {
	f := &Form{Fields: fields, Validation: make(map[int]string)}
	f.syncFocus()
	return f
}

// FocusIndex returns the index of the focused field, or the action index after the fields.
func (f *Form) FocusIndex() int {
	if f == nil {
		return 0
	}
	return f.focused
}
func (f *Form) Focused() bool { return f != nil && f.focusedSelf }
func (f *Form) SetFocus(v bool) {
	if f != nil {
		f.focusedSelf = v
	}
}
func (f *Form) focusCount() int { return len(f.Fields) + len(f.Actions) }
func (f *Form) syncFocus() {
	for i, field := range f.Fields {
		if w, ok := field.Widget.(Focusable); ok {
			w.SetFocus(i == f.focused)
		}
	}
}
func (f *Form) FocusNext() bool {
	if f == nil || f.focused+1 >= f.focusCount() {
		return false
	}
	f.focused++
	f.syncFocus()
	return true
}
func (f *Form) FocusPrevious() bool {
	if f == nil || f.focused <= 0 {
		return false
	}
	f.focused--
	f.syncFocus()
	return true
}

// Draw renders an error summary followed by aligned field labels, editors, help, and actions.
func (f *Form) Draw(c *Canvas, r Rect) {
	for y := 0; y < r.H; y++ {
		c.PaintSurface(Rect{r.X, r.Y + y, r.W, 1}, Style{})
	}
	y := r.Y
	if len(f.Validation) > 0 && y < r.Y+r.H {
		c.Write(r.X, y, "Please correct the following:", Style{Bold: true})
		y++
		for i := range f.Fields {
			if msg := f.Validation[i]; msg != "" && y < r.Y+r.H {
				c.Write(r.X, y, fmt.Sprintf("• %s: %s", f.Fields[i].Label, msg), Style{FG: ColorRGB(220, 70, 70)})
				y++
			}
		}
	}
	labelW := 0
	for _, field := range f.Fields {
		if w := measure.StringWidth(field.Label) + 2; w > labelW {
			labelW = w
		}
	}
	for i, field := range f.Fields {
		if y >= r.Y+r.H {
			return
		}
		label := field.Label
		if field.Required {
			label += " *"
		}
		c.Write(r.X, y, label, Style{Bold: i == f.focused})
		x := r.X + labelW
		h := 1
		if ch, ok := field.Widget.(ContentHeighter); ok {
			h = ch.ContentHeight()
			if h < 1 {
				h = 1
			}
		}
		fieldRect := Rect{x, y, max(0, r.X+r.W-x), min(h, r.Y+r.H-y)}
		switch w := field.Widget.(type) {
		case *TextInput:
			w.Draw(c, fieldRect, i == f.focused)
		case Widget:
			w.Draw(c, fieldRect)
		}
		y += h
		if field.Help != "" && y < r.Y+r.H {
			c.Write(r.X+labelW, y, field.Help, Style{Dim: true})
			y++
		}
	}
	if len(f.Actions) > 0 && y < r.Y+r.H {
		var parts []string
		for i, a := range f.Actions {
			mark := " "
			if f.focused == len(f.Fields)+i {
				mark = ">"
			}
			parts = append(parts, mark+" "+a.Label)
		}
		c.Write(r.X, y, strings.Join(parts, "   "), Style{Bold: true})
	}
}

func (f *Form) validate() int {
	f.Validation = make(map[int]string)
	first := -1
	for i, field := range f.Fields {
		var err error
		if field.Required && strings.TrimSpace(formValue(field.Widget)) == "" {
			err = fmt.Errorf("required")
		}
		if err == nil && field.Validate != nil {
			err = field.Validate()
		}
		if err != nil {
			f.Validation[i] = err.Error()
			if first < 0 {
				first = i
			}
		}
	}
	return first
}
func formValue(w any) string {
	switch v := w.(type) {
	case interface{ Value() string }:
		return v.Value()
	case interface{ String() string }:
		return v.String()
	case interface{ Value() *float64 }:
		if v.Value() == nil {
			return ""
		}
		return fmt.Sprint(*v.Value())
	case interface{ Value() *bool }:
		if v.Value() == nil {
			return ""
		}
		return fmt.Sprint(*v.Value())
	}
	return ""
}
func (f *Form) values() map[string]any {
	out := make(map[string]any, len(f.Fields))
	for _, field := range f.Fields {
		switch w := field.Widget.(type) {
		case *TextInput:
			out[field.Label] = w.Value()
		case *NumberInput:
			if w.Value != nil {
				out[field.Label] = *w.Value
			}
		case *Toggle:
			if w.Value != nil {
				out[field.Label] = *w.Value
			}
		case *Choice:
			if item, ok := w.Selected(); ok {
				out[field.Label] = item
			}
		}
	}
	return out
}
func (f *Form) activate() {
	if f.focused >= len(f.Fields) {
		if f.focused-len(f.Fields) >= len(f.Actions) {
			f.submit()
			return
		}
		a := f.Actions[f.focused-len(f.Fields)]
		if a.Cancel {
			if f.OnCancel != nil {
				f.OnCancel()
			}
		} else if a.OnActivate != nil {
			a.OnActivate()
		} else {
			f.submit()
		}
		return
	}
	if _, ok := f.Fields[f.focused].Widget.(*NumberInput); ok {
		w := f.Fields[f.focused].Widget.(*NumberInput)
		if !w.Editing() {
			w.HandleKey(KeyEvent{Key: "enter"})
			return
		}
		if !w.HandleKey(KeyEvent{Key: "enter"}) && w.Error() != "" {
			f.Validation[f.focused] = w.Error()
			return
		}
	}
	if w, ok := f.Fields[f.focused].Widget.(*Toggle); ok {
		w.HandleKey(KeyEvent{Key: "enter"})
		return
	}
	if w, ok := f.Fields[f.focused].Widget.(*Choice); ok {
		w.HandleKey(KeyEvent{Key: "enter"})
		return
	}
	if first := f.validate(); first >= 0 {
		f.focused = first
		f.syncFocus()
		return
	}
	f.submit()
}
func (f *Form) submit() {
	if f.OnSubmit != nil {
		f.OnSubmit(f.values())
	}
}

// HandleKey routes editing keys and form navigation.
func (f *Form) HandleKey(e KeyEvent) bool { return f.ConsumeKey(e).Quit }
func (f *Form) ConsumeKey(e KeyEvent) EventResult {
	if f == nil {
		return Ignored()
	}
	switch e.Key {
	case "tab":
		f.FocusNext()
		return Handled()
	case "shift-tab":
		f.FocusPrevious()
		return Handled()
	case "down", "pgdown", "pgdn", "pagedown":
		if f.FocusNext() {
			return Handled()
		}
	case "up", "pgup", "pageup":
		if f.FocusPrevious() {
			return Handled()
		}
	case "esc":
		if f.OnCancel != nil {
			f.OnCancel()
		}
		return Handled()
	case "enter":
		f.activate()
		return Handled()
	}
	if f.focused < len(f.Fields) {
		if w, ok := f.Fields[f.focused].Widget.(KeyConsumer); ok {
			quit, used := w.ConsumeKey(e)
			return EventResult{Quit: quit, Consumed: used}
		}
		switch w := f.Fields[f.focused].Widget.(type) {
		case *TextInput:
			return EventResult{Consumed: w.HandleKey(e)}
		case Widget:
			return EventResult{Quit: w.HandleKey(e), Consumed: true}
		}
	}
	return Ignored()
}
func (f *Form) HandleMouse(e MouseEvent) bool { return f.ConsumeMouse(e).Quit }
func (f *Form) ConsumeMouse(e MouseEvent) EventResult {
	if f == nil {
		return Ignored()
	}
	y := e.Y
	if len(f.Validation) > 0 {
		y-- // validation heading
		for i := range f.Fields {
			if f.Validation[i] != "" {
				y--
			}
		}
	}
	labelW := 0
	for _, field := range f.Fields {
		if width := StringWidth(field.Label) + 2; width > labelW {
			labelW = width
		}
	}
	for i, field := range f.Fields {
		h := 1
		if ch, ok := field.Widget.(ContentHeighter); ok {
			h = max(1, ch.ContentHeight())
		}
		if y >= 0 && y < h {
			f.focused = i
			f.syncFocus()
			e.X -= labelW
			e.Y = y
			if mc, ok := field.Widget.(MouseConsumer); ok {
				return mc.ConsumeMouse(e)
			}
			if w, ok := field.Widget.(Widget); ok {
				return EventResult{Quit: w.HandleMouse(e), Consumed: true}
			}
			return Handled()
		}
		y -= h
		if field.Help != "" {
			y--
		}
	}
	return Ignored()
}

var _ Widget = (*Form)(nil)
var _ EventConsumer = (*Form)(nil)
var _ MouseConsumer = (*Form)(nil)
var _ FocusContainer = (*Form)(nil)
