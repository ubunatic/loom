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

func TestMeasureTextRejectsUnsatisfiableParentWidth(t *testing.T) {
	_, err := MeasureText("a", 3, Insets{}, Constraints{Width: Limit{Min: 5}})
	if err == nil {
		t.Fatal("expected parent-width constraint error")
	}
}

func TestMalformedANSIIsSafeAndConsumesNoCells(t *testing.T) {
	for _, text := range []string{"\x1b[31", "\x1b]title", "\x1bPpayload"} {
		if got := StringWidth(text); got != 0 {
			t.Fatalf("StringWidth(%q)=%d, want 0", text, got)
		}
		if got := Truncate(text, 4, "…"); got != "" {
			t.Fatalf("Truncate(%q)=%q, want empty", text, got)
		}
	}
}

func TestMeasureOldNewParity(t *testing.T) {
	testCases := []string{
		"",
		"CPU",
		"Hello World",
		"\x1b[31mCPU\x1b[0m",
		"\x1b[38;5;196mFailed\x1b[0m - \x1b[1;34mDetails\x1b[0m",
		"\x1b[38;2;255;100;50mrgb text\x1b[0m",
		"e\u0301",
		"界",
		"⣿",
		"👍emoji",
		"CPU\n界界\n",
		"\x1b]title\nhidden\x07X\nY",
		"\x1b[31",
		"\x1b]title",
		"\x1bPpayload",
		"12界34",
		"ab界d",
	}

	for i, tc := range testCases {
		oldW := StringWidthOld(tc)
		newW := StringWidthNew(tc)
		if oldW != newW {
			t.Errorf("case %d (%q) StringWidth mismatch: Old=%d, New=%d", i, tc, oldW, newW)
		}

		oldC := ClustersOld(tc)
		newC := ClustersNew(tc)
		if len(oldC) != len(newC) {
			t.Errorf("case %d (%q) Clusters len mismatch: Old=%d (%v), New=%d (%v)", i, tc, len(oldC), oldC, len(newC), newC)
		} else {
			for j := range oldC {
				if oldC[j] != newC[j] {
					t.Errorf("case %d (%q) Cluster[%d] mismatch: Old=%q, New=%q", i, tc, j, oldC[j], newC[j])
				}
			}
		}

		oldP := plainTerminalTextOld(tc)
		newP := plainTerminalTextNew(tc)
		if oldP != newP {
			t.Errorf("case %d (%q) plainTerminalText mismatch: Old=%q, New=%q", i, tc, oldP, newP)
		}

		oldL := plainTerminalLinesOld(tc)
		newL := plainTerminalLinesNew(tc)
		if len(oldL) != len(newL) {
			t.Errorf("case %d (%q) plainTerminalLines len mismatch: Old=%d (%v), New=%d (%v)", i, tc, len(oldL), oldL, len(newL), newL)
		} else {
			for j := range oldL {
				if oldL[j] != newL[j] {
					t.Errorf("case %d (%q) plainTerminalLines[%d] mismatch: Old=%q, New=%q", i, tc, j, oldL[j], newL[j])
				}
			}
		}
	}
}

func TestFastMeasureEnvVarToggle(t *testing.T) {
	tc := "\x1b[31mERROR\x1b[0m: Failed"

	t.Setenv("LOOM_FAST_MEASURE", "0")
	if got := StringWidth(tc); got != 13 {
		t.Errorf("StringWidth with LOOM_FAST_MEASURE=0 got %d, want 13", got)
	}

	t.Setenv("LOOM_FAST_MEASURE", "1")
	if got := StringWidth(tc); got != 13 {
		t.Errorf("StringWidth with LOOM_FAST_MEASURE=1 got %d, want 13", got)
	}
}

// ── Benchmarks ──────────────────────────────────────────────────────────────

var benchTextPlain = "STATUS: 200 OK - Processing request completed successfully in 12ms"
var benchTextANSI = "\x1b[31mERROR\x1b[0m: " +
	"\x1b[38;5;196mFailed\x1b[0m - " +
	"\x1b[1;34mDetails\x1b[0m: " +
	"Some error message with lots of words that goes on for a bit"

func BenchmarkStringWidth_Old(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = StringWidthOld(benchTextANSI)
	}
}

func BenchmarkStringWidth_New(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = StringWidthNew(benchTextANSI)
	}
}

func BenchmarkStringWidth_Plain_Old(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = StringWidthOld(benchTextPlain)
	}
}

func BenchmarkStringWidth_Plain_New(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = StringWidthNew(benchTextPlain)
	}
}

func BenchmarkClusters_Old(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ClustersOld(benchTextANSI)
	}
}

func BenchmarkClusters_New(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ClustersNew(benchTextANSI)
	}
}

func BenchmarkPlainTerminalText_Old(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = plainTerminalTextOld(benchTextANSI)
	}
}

func BenchmarkPlainTerminalText_New(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = plainTerminalTextNew(benchTextANSI)
	}
}
