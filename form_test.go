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
	form.HandleKey(loom.KeyEvent{Key: "enter"})
	if form.Validation[0] == "" || form.FocusIndex() != 0 {
		t.Fatal("invalid required field did not receive focus and an error")
	}
	if !strings.Contains(strings.Join(loom.Render(form, 50, 10), "\n"), "Name") {
		t.Fatal("form fields were not rendered")
	}
	name.SetValue("Ada")
	form.HandleKey(loom.KeyEvent{Key: "enter"})
	if submitted == nil || submitted["Name"] != "Ada" || submitted["Age"] != 7.0 || submitted["Active"] != true {
		t.Fatalf("unexpected submitted values: %#v", submitted)
	}
}

func TestFormFocusTraversalAndCancel(t *testing.T) {
	first, second := loom.NewTextInput("a"), loom.NewTextInput("b")
	form := loom.NewForm([]loom.FormField{{Label: "First", Widget: first}, {Label: "Second", Widget: second}})
	form.Actions = []loom.FormAction{{Label: "Save"}, {Label: "Cancel", Cancel: true}}
	form.HandleKey(loom.KeyEvent{Key: "tab"})
	if form.FocusIndex() != 1 {
		t.Fatal("Tab did not focus second field")
	}
	form.HandleKey(loom.KeyEvent{Key: "shift-tab"})
	if form.FocusIndex() != 0 {
		t.Fatal("Shift-Tab did not focus first field")
	}
	called := false
	form.OnCancel = func() { called = true }
	form.HandleKey(loom.KeyEvent{Key: "esc"})
	if !called {
		t.Fatal("Esc did not invoke cancellation")
	}
}
