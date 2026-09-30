// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestNumberInputStepsAndClamps(t *testing.T) {
	value := 4.0
	input := loom.NewNumberInput(&value, 0, 5)
	input.Step = .5
	input.HandleKey(loom.KeyEvent{Key: "right"})
	input.HandleKey(loom.KeyEvent{Key: "right"})
	if value != 5 {
		t.Fatalf("value = %g, want clamped value 5", value)
	}
	input.HandleKey(loom.KeyEvent{Key: "left"})
	if value != 4.5 {
		t.Fatalf("value = %g, want 4.5", value)
	}
}

func TestNumberInputInlineEditValidationAndCancel(t *testing.T) {
	value := 3.0
	input := loom.NewNumberInput(&value, -10, 10)
	input.HandleKey(loom.KeyEvent{Key: "enter"})
	for _, key := range []loom.KeyEvent{{Key: "home"}, {Text: "-"}, {Key: "delete"}, {Text: "2"}, {Text: "."}, {Text: "5"}, {Key: "enter"}} {
		input.HandleKey(key)
	}
	if value != -2.5 || input.Editing() {
		t.Fatalf("commit state = %g/editing=%v, want -2.5/false", value, input.Editing())
	}
	input.HandleKey(loom.KeyEvent{Key: "enter"})
	input.HandleKey(loom.KeyEvent{Key: "home"})
	for range "-2.5" {
		input.HandleKey(loom.KeyEvent{Key: "delete"})
	}
	input.HandleKey(loom.KeyEvent{Text: "9"})
	input.HandleKey(loom.KeyEvent{Text: "9"})
	input.HandleKey(loom.KeyEvent{Key: "enter"})
	if value != -2.5 || !input.Editing() || !strings.Contains(input.Error(), "above the maximum") {
		t.Fatalf("invalid edit state = %g/editing=%v/error=%q", value, input.Editing(), input.Error())
	}
	input.HandleKey(loom.KeyEvent{Key: "esc"})
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
	toggle.HandleKey(loom.KeyEvent{Key: "enter"})
	if !value {
		t.Fatal("Enter did not enable toggle")
	}
	canvas = loom.NewCanvas(8, 1)
	toggle.Draw(canvas, canvas.Bounds())
	if got := rowText(canvas, 0); !strings.Contains(got, "[✓]") {
		t.Fatalf("on toggle render = %q", got)
	}
}
