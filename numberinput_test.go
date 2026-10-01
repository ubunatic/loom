// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestNumberInputRightAlignedFixedField(t *testing.T) {
	value := 5.0
	input := loom.NewNumberInput(&value, -100, 100)
	input.Format, input.FixedWidth = "%.2f", 11
	input.Align = loom.AlignRight
	for _, test := range []struct {
		value float64
		want  string
	}{
		{-100, "◂ -100.00 ▸"}, {5, "◂    5.00 ▸"}, {100, "◂  100.00 ▸"},
	} {
		value = test.value
		c := loom.NewCanvas(14, 2)
		input.Draw(c, loom.Rect{X: 2, Y: 1, W: 11, H: 1})
		var text strings.Builder
		for x := 2; x < 13; x++ {
			text.WriteString(c.Get(x, 1).Text)
		}
		if got := text.String(); got != test.want {
			t.Fatalf("value %g: got %q, want %q", value, got, test.want)
		}
	}
}

func TestNumberInputStepsAndClamps(t *testing.T) {
	value := 4.0
	input := loom.NewNumberInput(&value, 0, 5)
	input.Step = .5
	input.ConsumeKey(loom.KeyEvent{Key: "right"})
	input.ConsumeKey(loom.KeyEvent{Key: "right"})
	if value != 5 {
		t.Fatalf("value = %g, want clamped value 5", value)
	}
	input.ConsumeKey(loom.KeyEvent{Key: "left"})
	if value != 4.5 {
		t.Fatalf("value = %g, want 4.5", value)
	}
}

func TestNumberInputInlineEditValidationAndCancel(t *testing.T) {
	value := 3.0
	input := loom.NewNumberInput(&value, -10, 10)
	input.ConsumeKey(loom.KeyEvent{Key: "enter"})
	for _, key := range []loom.KeyEvent{{Key: "home"}, {Text: "-"}, {Key: "delete"}, {Text: "2"}, {Text: "."}, {Text: "5"}, {Key: "enter"}} {
		input.ConsumeKey(key)
	}
	if value != -2.5 || input.Editing() {
		t.Fatalf("commit state = %g/editing=%v, want -2.5/false", value, input.Editing())
	}
	input.ConsumeKey(loom.KeyEvent{Key: "enter"})
	input.ConsumeKey(loom.KeyEvent{Key: "home"})
	for range "-2.5" {
		input.ConsumeKey(loom.KeyEvent{Key: "delete"})
	}
	input.ConsumeKey(loom.KeyEvent{Text: "9"})
	input.ConsumeKey(loom.KeyEvent{Text: "9"})
	input.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if value != -2.5 || !input.Editing() || !strings.Contains(input.Error(), "above the maximum") {
		t.Fatalf("invalid edit state = %g/editing=%v/error=%q", value, input.Editing(), input.Error())
	}
	input.ConsumeKey(loom.KeyEvent{Key: "esc"})
	if value != -2.5 || input.Editing() {
		t.Fatalf("cancel state = %g/editing=%v, want -2.5/false", value, input.Editing())
	}
}

func TestNumberInputConsumeMouseScroll(t *testing.T) {
	value := 4.0
	input := loom.NewNumberInput(&value, 0, 5)
	input.Step = 0.5

	res := input.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollUp})
	if !res.Consumed || value != 4.5 {
		t.Fatalf("scroll up: consumed=%v, value=%g, want 4.5", res.Consumed, value)
	}
	input.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollUp})
	if value != 5.0 {
		t.Fatalf("scroll up clamp: value=%g, want 5.0", value)
	}
	input.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollUp})
	if value != 5.0 {
		t.Fatalf("scroll up beyond max: value=%g, want 5.0", value)
	}
	res = input.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollDown})
	if !res.Consumed || value != 4.5 {
		t.Fatalf("scroll down: consumed=%v, value=%g, want 4.5", res.Consumed, value)
	}
	for range 10 {
		input.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollDown})
	}
	if value != 0.0 {
		t.Fatalf("scroll down clamp: value=%g, want 0.0", value)
	}
}

func TestNumberInputConsumeMouseClick(t *testing.T) {
	value := 5.0
	input := loom.NewNumberInput(&value, -100, 100)
	input.Step = 0.5
	input.Format, input.FixedWidth, input.Align = "%.2f", 11, loom.AlignRight

	// Rendered text is "◂    5.00 ▸" (width 11).
	// Left stepper is at X=0, right stepper is at X=10.
	res := input.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 10, Y: 0})
	if !res.Consumed || value != 5.5 {
		t.Fatalf("click right stepper: consumed=%v, value=%g, want 5.5", res.Consumed, value)
	}
	res = input.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 0, Y: 0})
	if !res.Consumed || value != 5.0 {
		t.Fatalf("click left stepper: consumed=%v, value=%g, want 5.0", res.Consumed, value)
	}

	// Clicks outside bounds or wrong buttons are ignored.
	if res := input.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 11, Y: 0}); res.Consumed {
		t.Fatalf("click X=11 out of bounds should be ignored: %#v", res)
	}
	if res := input.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: -1, Y: 0}); res.Consumed {
		t.Fatalf("click X=-1 out of bounds should be ignored: %#v", res)
	}
	if res := input.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 0, Y: 1}); res.Consumed {
		t.Fatalf("click Y=1 out of bounds should be ignored: %#v", res)
	}
	if res := input.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseRight, X: 0, Y: 0}); res.Consumed {
		t.Fatalf("right click should be ignored: %#v", res)
	}
	if res := input.ConsumeMouse(loom.MouseEvent{Action: loom.MouseHover, X: 0, Y: 0}); res.Consumed {
		t.Fatalf("hover should be ignored: %#v", res)
	}
}

func TestNumberInputConsumeMouseWhileEditingIgnored(t *testing.T) {
	value := 3.0
	input := loom.NewNumberInput(&value, 0, 10)
	input.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if !input.Editing() {
		t.Fatal("expected input to be editing")
	}
	if res := input.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollUp}); res.Consumed {
		t.Fatalf("scroll while editing should be ignored: %#v", res)
	}
	if res := input.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 0, Y: 0}); res.Consumed {
		t.Fatalf("click while editing should be ignored: %#v", res)
	}
	if value != 3.0 {
		t.Fatalf("value changed while editing: %g", value)
	}
}

func TestNumberInputConsumeMouseNilOrUnset(t *testing.T) {
	var nilInput *loom.NumberInput
	if res := nilInput.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollUp}); res.Consumed {
		t.Fatalf("nil NumberInput should return Ignored: %#v", res)
	}
	emptyInput := &loom.NumberInput{}
	if res := emptyInput.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollUp}); res.Consumed {
		t.Fatalf("NumberInput with nil Value should return Ignored: %#v", res)
	}
}

func TestToggleDrawAndActivate(t *testing.T) {
	value := false
	toggle := loom.NewToggle(&value)
	canvas := loom.NewCanvas(8, 1)
	toggle.Draw(canvas, canvas.Bounds())
	if got := rowText(canvas, 0); !strings.Contains(got, "[ ]") {
		t.Fatalf("off toggle render = %q", got)
	}
	toggle.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if !value {
		t.Fatal("Enter did not enable toggle")
	}
	canvas = loom.NewCanvas(8, 1)
	toggle.Draw(canvas, canvas.Bounds())
	if got := rowText(canvas, 0); !strings.Contains(got, "[✓]") {
		t.Fatalf("on toggle render = %q", got)
	}
}
