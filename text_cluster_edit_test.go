package loom

import "testing"

// Issue 432: caret moves, backspace and delete act on whole grapheme
// clusters, not single runes. Skin-tone modifiers are left out until the
// measure scanner joins them (issue 433).
var clusterEditCases = []struct{ name, in string }{
	{"combining acute", "xé"},
	{"zwj family", "x👨‍👩‍👧"},
	{"flag", "x🇩🇪"},
	{"vs16 heart", "x❤️"},
	{"cjk", "x界"},
}

func TestTextInputClusterEdit(t *testing.T) {
	for _, tc := range clusterEditCases {
		t.Run(tc.name, func(t *testing.T) {
			ti := NewTextInput(tc.in)
			ti.ConsumeKey(KeyEvent{Key: "end"})
			ti.ConsumeKey(KeyEvent{Key: "backspace"})
			if got := ti.Value(); got != "x" {
				t.Errorf("backspace: got %q, want %q", got, "x")
			}

			ti = NewTextInput(tc.in)
			ti.ConsumeKey(KeyEvent{Key: "home"})
			ti.ConsumeKey(KeyEvent{Key: "right"})
			ti.ConsumeKey(KeyEvent{Key: "delete"})
			if got := ti.Value(); got != "x" {
				t.Errorf("right then delete: got %q, want %q", got, "x")
			}

			ti = NewTextInput(tc.in)
			ti.ConsumeKey(KeyEvent{Key: "end"})
			ti.ConsumeKey(KeyEvent{Key: "left"})
			ti.ConsumeKey(KeyEvent{Key: "left"})
			if ti.caret != 0 {
				t.Errorf("end, left, left: caret %d, want 0", ti.caret)
			}
		})
	}
}

func TestTextAreaClusterEdit(t *testing.T) {
	for _, tc := range clusterEditCases {
		t.Run(tc.name, func(t *testing.T) {
			end := len([]rune(tc.in))
			ta := NewTextArea(tc.in)
			ta.SetCaret(0, end)
			ta.ConsumeKey(KeyEvent{Key: "backspace"})
			if got := ta.Value(); got != "x" {
				t.Errorf("backspace: got %q, want %q", got, "x")
			}

			ta = NewTextArea(tc.in)
			ta.SetCaret(0, 1)
			ta.ConsumeKey(KeyEvent{Key: "delete"})
			if got := ta.Value(); got != "x" {
				t.Errorf("delete: got %q, want %q", got, "x")
			}

			ta = NewTextArea(tc.in)
			ta.SetCaret(0, end)
			ta.moveLeft()
			if ta.col != 1 {
				t.Errorf("moveLeft from end: col %d, want 1", ta.col)
			}
			ta.moveRight()
			if ta.col != end {
				t.Errorf("moveRight: col %d, want %d", ta.col, end)
			}
		})
	}
}
