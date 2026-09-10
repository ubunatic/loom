// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"unicode/utf8"

	"codeberg.org/ubunatic/loom/measure"
)

func stripAnsi(s string) string {
	s = strings.TrimPrefix(s, "\x1b[100m")
	s = strings.TrimPrefix(s, "\x1b[44m")
	s = strings.TrimPrefix(s, "\x1b[40m")
	s = strings.TrimSuffix(s, "\x1b[0m")
	return s
}

// ── RenderProgressBar Tests ─────────────────────────────────────────────────

func TestRenderProgressBar(t *testing.T) {
	tests := []struct {
		name string
		pct  float64
		w    int
		want string
	}{
		{"normal width 0%", 0, 10, "[░░░░░░░░░░]"},
		{"normal width 50%", 50, 10, "[█████░░░░░]"},
		{"normal width 100%", 100, 10, "[██████████]"},
		{"clamp over 100%", 120, 10, "[██████████]"},
		{"clamp under 0%", -10, 10, "[░░░░░░░░░░]"},
		{"shrink to 1 char, filled", 100, 1, "[█]"},
		{"shrink to 1 char, empty", 0, 1, "[░]"},
		{"width 0 clamps to 1, sub-char boundary", 50, 0, "[▌]"},
		{"negative width clamps to 1, sub-char boundary", 50, -5, "[▌]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderProgressBar(tt.pct, tt.w)
			if got != tt.want {
				t.Errorf("RenderProgressBar(%v, %d) = %q, want %q", tt.pct, tt.w, got, tt.want)
			}
		})
	}
}

func TestRenderProgressBarEighthBlockBoundary(t *testing.T) {
	tests := []struct {
		name string
		pct  float64
		want string
	}{
		{"0%", 0, "[░░░░]"},
		{"12.5% boundary at half of char 0", 12.5, "[▌░░░]"},
		{"24% boundary at seven-eighths of char 0", 24, "[▉░░░]"},
		{"26% char 0 full, boundary empty at char 1", 26, "[█░░░]"},
		{"49% char 0 full, boundary at seven-eighths of char 1", 49, "[█▉░░]"},
		{"51% chars 0-1 full, boundary empty at char 2", 51, "[██░░]"},
		{"100% fully filled, no boundary character", 100, "[████]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderProgressBar(tt.pct, 4)
			if got != tt.want {
				t.Errorf("RenderProgressBar(%v, 4) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}

func TestRenderProgressBarNeverPanics(t *testing.T) {
	for _, w := range []int{-100, -1, 0, 1, 2, 100} {
		for _, pct := range []float64{-1000, -1, 0, 50, 100, 1000} {
			_ = RenderProgressBar(pct, w)
		}
	}
}

// ── RenderBar Options Tests ─────────────────────────────────────────────────

func TestRenderBarOptions(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		opts  BarOptions
		want  string
	}{
		{"zero options use default bar", 50, BarOptions{}, "[█████░░░░░]"},
		{"width clamps up from negative", 100, BarOptions{Width: -4}, "[█]"},
		{"custom range supports negatives", 0, BarOptions{Width: 4, Min: -10, Max: 10}, "[██░░]"},
		{"value clamps above range", 20, BarOptions{Width: 4, Min: -10, Max: 10}, "[████]"},
		{"value clamps below range", -20, BarOptions{Width: 4, Min: -10, Max: 10}, "[░░░░]"},
		{"nan renders as minimum", math.NaN(), BarOptions{Width: 4}, "[░░░░]"},
		{"infinity clamps above range", math.Inf(1), BarOptions{Width: 4}, "[████]"},
		{"invalid range falls back to percent scale", 50, BarOptions{Width: 4, Min: 10, Max: 10}, "[██░░]"},
		{"stripe omits wrappers", 50, BarOptions{Width: 4, NoWrapper: true}, "██░░"},
		{"custom glyphs and wrappers", 50, BarOptions{Width: 4, Fill: '#', Empty: '-', Left: "{", Right: "}"}, "{##--}"},
		{"percent label defaults to whole percent", 12.5, BarOptions{Width: 4, IncludePercent: true}, "[░░░░] 12%"},
		{"percent label honors precision", 12.5, BarOptions{Width: 4, IncludePercent: true, PercentPrecision: 1}, "[░░░░] 12.5%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderBar(tt.value, tt.opts)
			if got != tt.want {
				t.Errorf("RenderBar(%v, %+v) = %q, want %q", tt.value, tt.opts, got, tt.want)
			}
		})
	}
}

func TestRenderBarBrailleTwoLevelBoundary(t *testing.T) {
	opts := func(width int) BarOptions {
		return BarOptions{
			Width:              width,
			SubChar:            true,
			Fill:               '⣿',
			Empty:              '⠀',
			SubCharacterGlyphs: []rune{'⡇'},
		}
	}
	tests := []struct {
		name string
		pct  float64
		want string
	}{
		{"0% is fully empty", 0, "[⠀⠀⠀⠀]"},
		{"half-boundary at 12.5%", 12.5, "[⡇⠀⠀⠀]"},
		{"one full cell at 25%", 25, "[⣿⠀⠀⠀]"},
		{"87.5% is three full cells plus a half cell", 87.5, "[⣿⣿⣿⡇]"},
		{"100% is fully filled", 100, "[⣿⣿⣿⣿]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderBar(tt.pct, opts(4))
			if got != tt.want {
				t.Errorf("RenderBar(%v, braille width 4) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}

func TestRenderBarANSIBackground(t *testing.T) {
	got := RenderBar(50, BarOptions{Width: 4, ANSI: true})
	want := "[" + "\x1b[100m" + "██░░" + "\x1b[0m" + "]"
	if got != want {
		t.Errorf("RenderBar ANSI default = %q, want %q", got, want)
	}

	got = RenderBar(50, BarOptions{Width: 4, ANSI: true, BackgroundANSI: "44"})
	want = "[" + "\x1b[44m" + "██░░" + "\x1b[0m" + "]"
	if got != want {
		t.Errorf("RenderBar ANSI custom code = %q, want %q", got, want)
	}

	// "none" suppresses background SGR escape
	got = RenderBar(50, BarOptions{Width: 4, ANSI: true, BackgroundANSI: "none"})
	want = "[██░░]"
	if got != want {
		t.Errorf("RenderBar ANSI none = %q, want %q", got, want)
	}
}

// ── PercentSparkline Tests ──────────────────────────────────────────────────

func TestPercentSparkline(t *testing.T) {
	tests := []struct {
		name     string
		pcts     []float64
		maxWidth int
		want     string
	}{
		{"single value padded to 10", []float64{0}, 10, "▁▁▁▁▁▁▁▁▁▁"},
		{"single full value padded to 10", []float64{100}, 10, "▁▁▁▁▁▁▁▁▁█"},
		{"fewer values than maxWidth padded", []float64{0, 50, 100}, 10, "▁▁▁▁▁▁▁▁▅█"},
		{"exact maxWidth length", []float64{0, 25, 50, 75, 100}, 5, "▁▃▅▇█"},
		{"more values than maxWidth uses most recent", []float64{0, 25, 50, 75, 100}, 2, "▇█"},
		{"maxWidth 0 clamps to 1, keeps most recent", []float64{0, 100}, 0, "█"},
		{"negative maxWidth clamps to 1", []float64{0, 100}, -3, "█"},
		{"nil slice pads full width", nil, 4, "▁▁▁▁"},
		{"empty slice pads full width", []float64{}, 3, "▁▁▁"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := PercentSparkline(tt.pcts, tt.maxWidth)
			got := stripAnsi(raw)
			if got != tt.want {
				t.Errorf("PercentSparkline(%v, %d) = %q, want %q", tt.pcts, tt.maxWidth, got, tt.want)
			}
			expectedWidth := tt.maxWidth
			if expectedWidth < 1 {
				expectedWidth = 1
			}
			if runeCount := utf8.RuneCountInString(got); runeCount != expectedWidth {
				t.Errorf("PercentSparkline width = %d, want %d", runeCount, expectedWidth)
			}
		})
	}
}

// ── RenderSparkline Tests ───────────────────────────────────────────────────

func TestRenderSparklineExactWidthPadding(t *testing.T) {
	// Assert that regardless of slice length (empty, 1 sample, short, full),
	// output rune count strictly equals Width unless NoPad is explicitly requested.
	tests := []struct {
		name         string
		values       []float64
		presentation SparklinePresentation
		width        int
		padRune      rune
		wantRunes    int
	}{
		{"blocks empty width 10", nil, SparklineBlocks, 10, 0, 10},
		{"blocks 1 sample width 10", []float64{50}, SparklineBlocks, 10, 0, 10},
		{"blocks 3 samples width 10", []float64{10, 20, 30}, SparklineBlocks, 10, 0, 10},
		{"blocks custom space pad", nil, SparklineBlocks, 6, ' ', 6},
		{"braille empty width 10", nil, SparklineBraille, 10, 0, 10},
		{"braille 1 sample width 10", []float64{50}, SparklineBraille, 10, 0, 10},
		{"braille 2 samples width 4", []float64{0, 100}, SparklineBraille, 4, 0, 4},
		{"braille 7 samples width 6", []float64{1, 2, 3, 4, 5, 6, 7}, SparklineBraille, 6, 0, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := SparklineOptions{
				Presentation: tt.presentation,
				Width:        tt.width,
				FixedRange:   true,
				Min:          0,
				Max:          100,
				PadRune:      tt.padRune,
			}
			got := RenderSparkline(tt.values, opts)
			runes := utf8.RuneCountInString(got)
			if runes != tt.wantRunes {
				t.Errorf("RenderSparkline() rune count = %d, want %d (output: %q)", runes, tt.wantRunes, got)
			}
		})
	}
}

func TestGraphNormalizesWideCustomGlyphsToOneCell(t *testing.T) {
	bar := RenderBar(50, BarOptions{Width: 4, Fill: '界', Empty: '界'})
	if got := measure.StringWidth(bar); got != 6 {
		t.Fatalf("bar width=%d, want wrapper plus four cells", got)
	}
	spark := RenderSparkline([]float64{50}, SparklineOptions{Width: 3, PadRune: '界'})
	if got := measure.StringWidth(spark); got != 3 {
		t.Fatalf("spark width=%d, want 3", got)
	}
}

func TestRenderSparklineNoPad(t *testing.T) {
	opts := SparklineOptions{Width: 10, NoPad: true}
	if got := RenderSparkline(nil, opts); got != "" {
		t.Errorf("RenderSparkline(nil, NoPad) = %q, want empty", got)
	}
	if got := RenderSparkline([]float64{0, 50, 100}, opts); utf8.RuneCountInString(got) != 2 {
		t.Errorf("RenderSparkline(3 samples, NoPad) rune count = %d, want 2", utf8.RuneCountInString(got))
	}
}

func TestRenderBrailleSparkline(t *testing.T) {
	fixed := SparklineOptions{Presentation: SparklineBraille, FixedRange: true, Min: 0, Max: 100, NoPad: true}
	tests := []struct {
		name string
		in   []float64
		opts SparklineOptions
		want string
	}{
		{"pair orientation", []float64{0, 100, 100, 0}, fixed, "⣸⣇"},
		{"odd leading singleton is duplicated", []float64{10, 20, 30}, fixed, "⣀⣠"},
		{"latest doubled window", []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100}, SparklineOptions{Presentation: SparklineBraille, FixedRange: true, Min: 0, Max: 100, Width: 10, NoPad: true}, "⣀⣀⣀⣀⣀⣀⣀⣀⣀⣸"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RenderSparkline(tt.in, tt.opts); got != tt.want {
				t.Fatalf("RenderSparkline() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderBrailleSparklineCellForegroundUsesTallerColumn(t *testing.T) {
	got := RenderPercentSparkline([]float64{5, 90}, SparklineOptions{
		Presentation:   SparklineBraille,
		Width:          1,
		ANSI:           true,
		BackgroundANSI: "40",
		ForegroundANSI: func(value float64) string {
			if value > 75 {
				return "31"
			}
			return "34"
		},
	})
	if want := "\x1b[40;31m⣸\x1b[0m"; got != want {
		t.Fatalf("colored Braille = %q, want %q", got, want)
	}
}

func TestBrailleColumnLevelFixedRangeBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  int
	}{
		{"zero uses visible baseline", 0, 1},
		{"low positive keeps baseline", 0.01, 1},
		{"twenty percent is visible", 20, 1},
		{"first band endpoint", 25, 1},
		{"second band begins", 25.01, 2},
		{"second band endpoint", 50, 2},
		{"third band begins", 50.01, 3},
		{"third band endpoint", 75, 3},
		{"fourth band begins", 75.01, 4},
		{"maximum fills all dots", 100, 4},
		{"below range clamps to baseline", -1, 1},
		{"above range clamps to full", 101, 4},
		{"nan normalizes to baseline", math.NaN(), 1},
		{"negative infinity normalizes to baseline", math.Inf(-1), 1},
		{"positive infinity clamps to full", math.Inf(1), 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := brailleColumnLevel(tt.value, 0, 100, false); got != tt.want {
				t.Errorf("brailleColumnLevel(%v) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func ExampleRenderSparkline() {
	fmt.Println(RenderSparkline([]float64{-2, -1, 0, 1, 2}, SparklineOptions{NoPad: true}))
	// Output:
	// ▁▄█
}
