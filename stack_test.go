// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"reflect"
	"testing"

	"codeberg.org/ubunatic/loom/layout"
)

func TestMeasuredStackUsesConstraintsAndGap(t *testing.T) {
	s := NewStack(Horizontal, NewView([]string{"a"}), NewView([]string{"b"}))
	s.Measured = true
	s.Gap = 1
	s.Constraints = []layout.Constraint{
		{Min: 2, Preferred: 3, Max: 4, HasMax: true},
		{Min: 2, Preferred: 2, Max: 5, HasMax: true},
	}
	got := []Rect{s.childRect(Rect{W: 10, H: 4}, 0, 2), s.childRect(Rect{W: 10, H: 4}, 1, 2)}
	want := []Rect{{X: 0, W: 4, H: 4}, {X: 5, W: 5, H: 4}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("measured rects=%v, want %v", got, want)
	}
}

func TestStackDefaultAllocationRemainsEqualShare(t *testing.T) {
	s := NewStack(Horizontal, NewView([]string{"a"}), NewView([]string{"b"}))
	got := []Rect{s.childRect(Rect{W: 5, H: 4}, 0, 2), s.childRect(Rect{W: 5, H: 4}, 1, 2)}
	want := []Rect{{X: 0, W: 2, H: 4}, {X: 2, W: 3, H: 4}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("legacy rects=%v, want %v", got, want)
	}
}
