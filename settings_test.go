// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestSettingsNumberSteppingAndClamping(t *testing.T) {
	tests := []struct {
		name        string
		key         loom.KeyEvent
		start, want float64
	}{
		{name: "plus custom step", key: loom.KeyEvent{Text: "+"}, start: 4, want: 4.5},
		{name: "minus custom step", key: loom.KeyEvent{Text: "-"}, start: 4, want: 3.5},
		{name: "upper clamp", key: loom.KeyEvent{Text: "+"}, start: 5.5, want: 6},
		{name: "lower clamp", key: loom.KeyEvent{Text: "-"}, start: 0.5, want: 0},
		{name: "reversed bounds clamp to min", key: loom.KeyEvent{Text: "+"}, start: 4, want: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.start
			max := 6.0
			s := loom.NewSettings([]loom.Setting{{Kind: loom.KindNumber, Num: &value, Min: 0, Max: max, Step: .5}})
			if tt.name == "reversed bounds clamp to min" {
				s = loom.NewSettings([]loom.Setting{{Kind: loom.KindNumber, Num: &value, Min: 5, Max: -1, Step: .5}})
			}
			s.HandleKey(tt.key)
			if value != tt.want {
				t.Errorf("value = %g, want %g", value, tt.want)
			}
			if s.Editing() {
				t.Error("stepping entered edit mode")
			}
		})
	}

	value := 4.0
	s := loom.NewSettings([]loom.Setting{{Kind: loom.KindNumber, Num: &value, Min: 0, Max: 6}})
	s.HandleKey(loom.KeyEvent{Key: "left"})
	if value != 3 {
		t.Errorf("default step value = %g, want 3", value)
	}
	s.HandleKey(loom.KeyEvent{Key: "right"})
	if value != 4 {
		t.Errorf("right step value = %g, want 4", value)
	}
}

func TestSettingsNumberInlineEditCommitCancelAndNegative(t *testing.T) {
	value := 3.0
	s := loom.NewSettings([]loom.Setting{{Kind: loom.KindNumber, Num: &value, Min: -10, Max: 10}})
	s.HandleKey(loom.KeyEvent{Key: "enter"})
	if !s.Editing() {
		t.Fatal("Enter did not start number edit")
	}
	for _, key := range []loom.KeyEvent{{Key: "home"}, {Text: "-"}, {Key: "delete"}, {Text: "2"}, {Text: "."}, {Text: "5"}, {Key: "enter"}} {
		s.HandleKey(key)
	}
	if value != -2.5 || s.Editing() {
		t.Errorf("committed value/editing = %g/%v, want -2.5/false", value, s.Editing())
	}

	s.HandleKey(loom.KeyEvent{Key: "enter"})
	s.HandleKey(loom.KeyEvent{Text: "9"})
	s.HandleKey(loom.KeyEvent{Key: "esc"})
	if value != -2.5 || s.Editing() {
		t.Errorf("cancel value/editing = %g/%v, want -2.5/false", value, s.Editing())
	}

	positive := 1.0
	positiveSettings := loom.NewSettings([]loom.Setting{{Kind: loom.KindNumber, Num: &positive, Min: 0, Max: 10}})
	positiveSettings.HandleKey(loom.KeyEvent{Key: "enter"})
	positiveSettings.HandleKey(loom.KeyEvent{Key: "home"})
	positiveSettings.HandleKey(loom.KeyEvent{Text: "-"})
	positiveSettings.HandleKey(loom.KeyEvent{Key: "enter"})
	if positiveSettings.Editing() {
		t.Fatal("a minus sign is ignored when Min is nonnegative")
	}
}

func TestSettingsNumberValidationAndErrorRendering(t *testing.T) {
	value := 12.0
	s := loom.NewSettings([]loom.Setting{{Label: "size", Kind: loom.KindNumber, Num: &value, Min: 1, Max: 72}})
	s.HandleKey(loom.KeyEvent{Key: "enter"})
	s.HandleKey(loom.KeyEvent{Key: "home"})
	s.HandleKey(loom.KeyEvent{Key: "delete"})
	for _, r := range "99" {
		s.HandleKey(loom.KeyEvent{Text: string(r)})
	}
	s.HandleKey(loom.KeyEvent{Key: "enter"})
	if !s.Editing() || value != 12 {
		t.Fatalf("out-of-range commit state value/editing = %g/%v, want 12/true", value, s.Editing())
	}
	if got := s.ContentHeight(); got != 2 {
		t.Errorf("ContentHeight with error = %d, want 2", got)
	}
	c := loom.NewCanvas(50, 2)
	s.Draw(c, c.Bounds())
	if got := rowText(c, 1); !strings.Contains(got, "above the maximum 72") {
		t.Errorf("error row = %q, want maximum validation error", got)
	}

	// Replace invalid input and verify a valid commit clears the error line.
	s.HandleKey(loom.KeyEvent{Key: "home"})
	for range "99" {
		s.HandleKey(loom.KeyEvent{Key: "delete"})
	}
	s.HandleKey(loom.KeyEvent{Key: "home"})
	s.HandleKey(loom.KeyEvent{Key: "delete"})
	s.HandleKey(loom.KeyEvent{Text: "6"})
	s.HandleKey(loom.KeyEvent{Key: "enter"})
	if value != 6 || s.Editing() || s.ContentHeight() != 1 {
		t.Errorf("corrected commit value/editing/height = %g/%v/%d, want 6/false/1", value, s.Editing(), s.ContentHeight())
	}

	for _, input := range []string{"", "."} {
		v := 2.0
		widget := loom.NewSettings([]loom.Setting{{Kind: loom.KindNumber, Num: &v, Min: 0, Max: 10}})
		widget.HandleKey(loom.KeyEvent{Key: "enter"})
		widget.HandleKey(loom.KeyEvent{Key: "home"})
		widget.HandleKey(loom.KeyEvent{Key: "delete"})
		for _, r := range input {
			widget.HandleKey(loom.KeyEvent{Text: string(r)})
		}
		widget.HandleKey(loom.KeyEvent{Key: "enter"})
		if !widget.Editing() || widget.ContentHeight() != 2 {
			t.Errorf("input %q should remain in edit mode with error", input)
		}
		canvas := loom.NewCanvas(40, 2)
		widget.Draw(canvas, canvas.Bounds())
		if !strings.Contains(rowText(canvas, 1), "invalid number") {
			t.Errorf("input %q error row = %q, want invalid number", input, rowText(canvas, 1))
		}
	}
}

func TestSettingsNumberCustomFormattingAndPositiveKeys(t *testing.T) {
	value := 3.5
	s := loom.NewSettings([]loom.Setting{{Label: "scale", Kind: loom.KindNumber, Num: &value, Min: 0, Max: 10, Step: .25, Format: "%.2f"}})
	c := loom.NewCanvas(40, 1)
	s.Draw(c, c.Bounds())
	if row := rowText(c, 0); !strings.Contains(row, "◂ 3.50 ▸") {
		t.Errorf("formatted number row = %q, want ◂ 3.50 ▸", row)
	}
	s.HandleKey(loom.KeyEvent{Text: "+"})
	if value != 3.75 {
		t.Errorf("plus value = %g, want 3.75", value)
	}
	s.HandleKey(loom.KeyEvent{Text: "-"})
	if value != 3.5 {
		t.Errorf("minus value = %g, want 3.5", value)
	}
}
