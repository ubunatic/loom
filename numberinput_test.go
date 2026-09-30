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
