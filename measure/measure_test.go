// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package measure

import "testing"

func TestStringWidthUsesLoomCellPolicy(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
		want int
	}{
		{name: "plain", text: "CPU", want: 3},
		{name: "style", text: "\x1b[31mCPU\x1b[0m", want: 3},
		{name: "combining", text: "e\u0301", want: 1},
		{name: "wide", text: "界", want: 2},
		{name: "graph", text: "⣿", want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := StringWidth(tc.text); got != tc.want {
				t.Fatalf("StringWidth(%q)=%d, want %d", tc.text, got, tc.want)
			}
		})
	}
}

func TestLinesMeasuresVisibleBounds(t *testing.T) {
	got := Lines("CPU\n界界\n")
	want := Size{Width: 4, Height: 3}
	if got != want {
		t.Fatalf("Lines()=%+v, want %+v", got, want)
	}
}

func TestLinesKeepsControlStateAcrossNewlines(t *testing.T) {
	got := Lines("\x1b]title\nhidden\x07X\nY")
	want := Size{Width: 1, Height: 2}
	if got != want {
		t.Fatalf("Lines()=%+v, want %+v", got, want)
	}
}

func TestTruncatePreservesCellBudget(t *testing.T) {
	got := Truncate("12界34", 5, "…")
	if got != "12界…" {
		t.Fatalf("Truncate()=%q, want %q", got, "12界…")
	}
	if width := StringWidth(got); width > 5 {
		t.Fatalf("truncated width=%d exceeds budget", width)
	}
}

func TestTruncateLeftKeepsTail(t *testing.T) {
	got := TruncateLeft("abcdef", 4, "…")
	if got != "…def" {
		t.Fatalf("TruncateLeft()=%q, want %q", got, "…def")
	}
}

func TestFitAndPadUseTerminalCells(t *testing.T) {
	if got := Fit("中中", 3); got != "中" {
		t.Fatalf("Fit()=%q, want %q", got, "中")
	}
	if got := Pad("中", 3, false); got != "中 " {
		t.Fatalf("Pad(left)=%q, want %q", got, "中 ")
	}
	if got := Pad("中", 3, true); got != " 中" {
		t.Fatalf("Pad(right)=%q, want %q", got, " 中")
	}
	if got := StringWidth(Pad("中", 3, false)); got != 3 {
		t.Fatalf("padded width=%d, want 3", got)
	}
}

func TestTextSizeWrapsAtCellBoundaries(t *testing.T) {
	if got := TextSize("ab界d", 3); got != (Size{Width: 3, Height: 2}) {
		t.Fatalf("TextSize()=%+v, want width 3 height 2", got)
	}
	if got := TextSize("\n", 3); got != (Size{Height: 2}) {
		t.Fatalf("empty lines=%+v, want height 2", got)
	}
}

func TestMeasureTextAppliesInsetsAndConstraints(t *testing.T) {
	c := Constraints{
		Width:  Limit{Min: 6, Preferred: 8, Max: 10, HasMax: true},
		Height: Limit{Min: 4, Preferred: 0, Max: 5, HasMax: true},
	}
	got, err := MeasureText("abc", 0, Insets{Top: 1, Right: 1, Bottom: 1, Left: 1}, c)
	if err != nil {
		t.Fatal(err)
	}
	if got != (Size{Width: 8, Height: 4}) {
		t.Fatalf("MeasureText()=%+v, want 8x4", got)
	}
}

func TestConstraintsRejectInconsistentLimits(t *testing.T) {
	err := (Constraints{Width: Limit{Min: 4, Preferred: 2}}).Validate()
	if err == nil {
		t.Fatal("expected preferred-below-minimum error")
	}
}
