package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestFormFocusValidationAndSubmit(t *testing.T) {
	name := loom.NewTextInput("")
	age := 7.0
	active := true
	form := loom.NewForm([]loom.FormField{
		{Label: "Name", Widget: name, Required: true},
		{Label: "Age", Widget: loom.NewNumberInput(&age, 0, 120)},
		{Label: "Active", Widget: loom.NewToggle(&active)},
	})
	form.Actions = []loom.FormAction{{Label: "Save"}, {Label: "Cancel", Cancel: true}}
	var submitted map[string]any
	form.OnSubmit = func(values map[string]any) { submitted = values }
	form.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if form.Validation[0] == "" || form.FocusIndex() != 0 {
		t.Fatal("invalid required field did not receive focus and an error")
	}
	if !strings.Contains(strings.Join(loom.Render(form, 50, 10), "\n"), "Name") {
		t.Fatal("form fields were not rendered")
	}
	name.SetValue("Ada")
	form.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if submitted == nil || submitted["Name"] != "Ada" || submitted["Age"] != 7.0 || submitted["Active"] != true {
		t.Fatalf("unexpected submitted values: %#v", submitted)
	}
}

func TestFormFocusTraversalAndCancel(t *testing.T) {
	first, second := loom.NewTextInput("a"), loom.NewTextInput("b")
	form := loom.NewForm([]loom.FormField{{Label: "First", Widget: first}, {Label: "Second", Widget: second}})
	form.Actions = []loom.FormAction{{Label: "Save"}, {Label: "Cancel", Cancel: true}}
	form.ConsumeKey(loom.KeyEvent{Key: "tab"})
	if form.FocusIndex() != 1 {
		t.Fatal("Tab did not focus second field")
	}
	form.ConsumeKey(loom.KeyEvent{Key: "shift-tab"})
	if form.FocusIndex() != 0 {
		t.Fatal("Shift-Tab did not focus first field")
	}
	called := false
	form.OnCancel = func() { called = true }
	form.ConsumeKey(loom.KeyEvent{Key: "esc"})
	if !called {
		t.Fatal("Esc did not invoke cancellation")
	}
}

func TestFormHoverPreservesEditorAndClickFocusesEditor(t *testing.T) {
	first, second := loom.NewTextInput("Ada"), loom.NewTextInput("Engineer")
	form := loom.NewForm([]loom.FormField{
		{Label: "Name", Widget: first, Help: "Display name"},
		{Label: "Role", Widget: second},
	})
	loom.Render(form, 50, 10)
	for _, action := range []loom.MouseAction{loom.MouseHover, loom.MouseDrag, loom.MouseRelease} {
		res := form.ConsumeMouse(loom.MouseEvent{Action: action, Button: loom.MouseLeft, X: 7, Y: 2})
		if res.Consumed || form.FocusIndex() != 0 {
			t.Fatalf("motion changed focus: action=%v result=%+v focus=%d", action, res, form.FocusIndex())
		}
	}
	if res := form.ConsumeKey(loom.KeyEvent{Text: "Z"}); !res.Consumed || first.Value() != "AdaZ" || second.Value() != "Engineer" {
		t.Fatalf("typing after hover: result=%+v first=%q second=%q", res, first.Value(), second.Value())
	}
	if res := form.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 7, Y: 2}); !res.Consumed || form.FocusIndex() != 1 {
		t.Fatalf("click did not focus editor: result=%+v focus=%d", res, form.FocusIndex())
	}
	if res := form.ConsumeKey(loom.KeyEvent{Text: "Y"}); !res.Consumed || second.Value() != "EngineerY" {
		t.Fatalf("typing after click: result=%+v value=%q", res, second.Value())
	}
}
