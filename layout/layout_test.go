// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package layout

import (
	"reflect"
	"testing"
)

func TestPlanShrinksFromEndAndStretchesWithCaps(t *testing.T) {
	items := []Item{
		{Visible: true, Constraint: Constraint{Min: 2, Preferred: 4, Max: 6, HasMax: true}},
		{Visible: true, Constraint: Constraint{Min: 2, Preferred: 4, Max: 5, HasMax: true}},
	}
	got, err := Plan(7, 1, items)
	if err != nil {
		t.Fatal(err)
	}
	if want := []Allocation{{Offset: 0, Size: 4}, {Offset: 5, Size: 2}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("shrink=%v, want %v", got, want)
	}
	got, err = Plan(12, 1, items)
	if err != nil {
		t.Fatal(err)
	}
	if want := []Allocation{{Offset: 0, Size: 6}, {Offset: 7, Size: 5}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("stretch=%v, want %v", got, want)
	}
}

func TestPlanSkipsHiddenAndRejectsInsufficientSpace(t *testing.T) {
	items := []Item{
		{Visible: true, Constraint: Constraint{Min: 3}},
		{Visible: false, Constraint: Constraint{Min: 99}},
		{Visible: true, Constraint: Constraint{Min: 3}},
	}
	got, err := Plan(8, 1, items)
	if err != nil {
		t.Fatal(err)
	}
	if got[1] != (Allocation{}) || got[0].Offset != 0 || got[2].Offset != 5 {
		t.Fatalf("hidden allocation=%v", got)
	}
	if _, err := Plan(5, 1, items); err == nil {
		t.Fatal("expected minimum-size error")
	}
}

func TestPlanIsDeterministic(t *testing.T) {
	items := []Item{
		{Visible: true, Constraint: Constraint{Min: 1, Preferred: 1}},
		{Visible: true, Constraint: Constraint{Min: 1, Preferred: 1}},
		{Visible: true, Constraint: Constraint{Min: 1, Preferred: 1}},
	}
	a, err := Plan(8, 0, items)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Plan(8, 0, items)
	if err != nil || !reflect.DeepEqual(a, b) {
		t.Fatalf("plans differ: %v vs %v (err=%v)", a, b, err)
	}
}
