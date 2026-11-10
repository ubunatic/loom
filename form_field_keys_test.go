package loom

import (
	"fmt"
	"testing"
)

func TestFormCapturesDownAndEnterInTextArea(t *testing.T) {
	tests := []struct {
		key      string
		field    any
		advance  bool
		expected string
	}{
		{"enter", NewTextArea("one\ntwo"), false, "one\n\ntwo"},
		{"down", NewTextArea("one\ntwo"), false, "caret on row 1"},
		{"pgdown", NewTextArea("one\ntwo"), false, "caret on row 1"},
		{"tab", NewTextArea("one\ntwo"), false, "focus 1"},
		{"esc", NewTextArea("one\ntwo"), false, "cancel"},

		{"enter", NewTextInput("single"), false, "submits=1"},
		{"enter", NewTextInput("single"), true, "focus 1"},
		{"down", NewTextInput("single"), false, "focus 1"},
		{"pgdown", NewTextInput("single"), false, "focus 1"},
		{"tab", NewTextInput("single"), false, "focus 1"},
		{"esc", NewTextInput("single"), false, "cancel"},

		{"enter", NewChoice([]Item{{Name: "A"}}), false, "submits=1"},
		{"down", NewChoice([]Item{{Name: "A"}}), false, "focus 1"},
		{"pgdown", NewChoice([]Item{{Name: "A"}}), false, "focus 1"},
		{"tab", NewChoice([]Item{{Name: "A"}}), false, "focus 1"},
		{"esc", NewChoice([]Item{{Name: "A"}}), false, "cancel"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%T_%s_adv%v", tt.field, tt.key, tt.advance)
		t.Run(name, func(t *testing.T) {
			submits := 0
			canceled := false
			form := NewForm([]FormField{{Label: "test", Widget: tt.field}, {Label: "dummy", Widget: NewTextInput("")}})
			form.AdvanceOnEnter = tt.advance
			form.OnSubmit = func(_ map[string]any) { submits++ }
			form.OnCancel = func() { canceled = true }
			form.SetFocus(true)

			if ta, ok := tt.field.(*TextArea); ok {
				ta.SetCaret(0, 0)
			}

			_ = form.ConsumeKey(KeyEvent{Key: tt.key})

			var actual string
			if canceled {
				actual = "cancel"
			} else if submits > 0 {
				actual = "submits=1"
			} else if form.focused == 1 {
				actual = "focus 1"
			} else {
				if ta, ok := tt.field.(*TextArea); ok {
					if tt.key == "down" || tt.key == "pgdown" {
						row, _ := ta.Caret()
						actual = fmt.Sprintf("caret on row %d", row)
					} else {
						actual = ta.Value()
					}
				} else if ti, ok := tt.field.(*TextInput); ok {
					actual = ti.Value()
				} else if ch, ok := tt.field.(*Choice); ok {
					actual = fmt.Sprintf("choice %v", ch)
				}
			}

			if actual != tt.expected {
				t.Errorf("got %q, want %q", actual, tt.expected)
			}
		})
	}
}
