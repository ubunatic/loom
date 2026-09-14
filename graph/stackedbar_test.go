// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"math"
	"strings"
	"testing"
	"unicode/utf8"

	"codeberg.org/ubunatic/loom/measure"
)

func TestRenderStackedBarExactWidth(t *testing.T) {
	cases := [][]float64{
		{1, 2, 3},
		{0, 0, 0},
		{},
		{100},
		{1, 1, 1, 1, 1, 1, 1},
		{-5, 10, 0},
	}
	for _, values := range cases {
		out := RenderStackedBar(values, StackedBarOptions{Width: 10})
		inner := strings.TrimSuffix(strings.TrimPrefix(out, "["), "]")
		if got := measure.StringWidth(inner); got != 10 {
			t.Errorf("RenderStackedBar(%v) inner width = %d, want 10 (out=%q)", values, got, out)
		}
	}
}

func TestRenderStackedBarProportions(t *testing.T) {
	// 1:1:2 over width 8 -> 2:2:4 cells.
	out := RenderStackedBar([]float64{1, 1, 2}, StackedBarOptions{Width: 8, Glyphs: []rune("ABC")})
	inner := strings.TrimSuffix(strings.TrimPrefix(out, "["), "]")
	want := "AABBCCCC"
	if inner != want {
		t.Errorf("RenderStackedBar proportions = %q, want %q", inner, want)
	}
}

func TestRenderStackedBarZeroSumRendersBlank(t *testing.T) {
	out := RenderStackedBar([]float64{0, 0}, StackedBarOptions{Width: 6, NoWrapper: true})
	if out != "      " {
		t.Errorf("RenderStackedBar zero-sum = %q, want 6 spaces", out)
	}
}

func TestRenderStackedBarEmptyInput(t *testing.T) {
	out := RenderStackedBar(nil, StackedBarOptions{Width: 5, NoWrapper: true})
	if out != "     " {
		t.Errorf("RenderStackedBar(nil) = %q, want 5 spaces", out)
	}
}

func TestRenderStackedBarNegativeAndNaNTreatedAsZero(t *testing.T) {
	out := RenderStackedBar([]float64{-1, 4, math.NaN()}, StackedBarOptions{Width: 4, Glyphs: []rune("ABC")})
	inner := strings.TrimSuffix(strings.TrimPrefix(out, "["), "]")
	if inner != "BBBB" {
		t.Errorf("RenderStackedBar negative/NaN handling = %q, want %q", inner, "BBBB")
	}
}

func TestRenderStackedBarNoWrapper(t *testing.T) {
	out := RenderStackedBar([]float64{1}, StackedBarOptions{Width: 3, NoWrapper: true})
	if strings.ContainsAny(out, "[]") {
		t.Errorf("RenderStackedBar NoWrapper leaked brackets: %q", out)
	}
}

func TestRenderStackedBarANSIPerSegmentForeground(t *testing.T) {
	out := RenderStackedBar([]float64{1, 1}, StackedBarOptions{
		Width:          4,
		Glyphs:         []rune("AA"),
		NoWrapper:      true,
		ANSI:           true,
		BackgroundANSI: "none",
		ForegroundANSI: []string{"31", "32"},
	})
	want := "\x1b[31mAA\x1b[0m\x1b[32mAA\x1b[0m"
	if out != want {
		t.Errorf("RenderStackedBar ANSI = %q, want %q", out, want)
	}
}

func TestRenderStackedBarNeverPanics(t *testing.T) {
	inputs := [][]float64{
		nil, {}, {0}, {math.NaN()}, {math.Inf(1)}, {math.Inf(-1)},
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	for _, values := range inputs {
		for _, width := range []int{-1, 0, 1, 3} {
			_ = RenderStackedBar(values, StackedBarOptions{Width: width})
		}
	}
}

func TestRenderStackedBarGlyphsAreSingleRune(t *testing.T) {
	out := RenderStackedBar([]float64{1}, StackedBarOptions{Width: 4, NoWrapper: true})
	if utf8.RuneCountInString(out) != 4 {
		t.Errorf("RenderStackedBar rune count = %d, want 4 (out=%q)", utf8.RuneCountInString(out), out)
	}
}
