// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"math"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom/graph"
)

func TestProgressBarSetInvalidatesOnlyOnVisibleChange(t *testing.T) {
	bar := NewProgressBar()
	bar.Options.Width = 10 // 20 half-cell steps: one step per 5%
	calls := 0
	bar.SetInvalidate(func() { calls++ })

	for v := 0.0; v < 5; v += 0.1 {
		bar.Set(v)
	}
	if calls != 0 {
		t.Fatalf("sub-step updates invalidated %d times, want 0", calls)
	}
	bar.Set(5)
	bar.Set(5.5)
	if calls != 1 {
		t.Fatalf("one visible step invalidated %d times, want 1", calls)
	}
	for v := 0.0; v <= 100; v += 0.01 {
		bar.Set(v)
	}
	// One redraw for dropping back to 0, then one per step up to 100.
	if calls != 21 {
		t.Fatalf("full sweep invalidated %d times, want 21", calls)
	}
}

func TestProgressBarClampsAndIgnoresNaN(t *testing.T) {
	bar := NewProgressBar()
	bar.Options.Width = 4
	for _, v := range []float64{-50, math.NaN(), math.Inf(-1)} {
		bar.Set(v)
		if got := renderProgressBar(bar, 6); got != "[    ]" {
			t.Fatalf("Set(%v) drew %q, want empty bar", v, got)
		}
	}
	bar.Set(math.Inf(1))
	if got := renderProgressBar(bar, 6); got != "[⣿⣿⣿⣿]" {
		t.Fatalf("Set(+Inf) drew %q, want full bar", got)
	}
}

func TestProgressBarDoneShowsPattern(t *testing.T) {
	bar := NewProgressBar()
	bar.Options.Width = 4
	calls := 0
	bar.SetInvalidate(func() { calls++ })
	bar.Set(100)
	bar.Done()
	bar.Done()
	if calls != 2 {
		t.Fatalf("Set+Done+Done invalidated %d times, want 2", calls)
	}
	if got := renderProgressBar(bar, 6); got != "[::::]" {
		t.Fatalf("done bar drew %q, want [::::]", got)
	}
	bar.Reset()
	if bar.IsDone() || renderProgressBar(bar, 6) != "[    ]" {
		t.Fatal("Reset did not restore an empty bar")
	}
}

func TestProgressBarClipsToRect(t *testing.T) {
	bar := NewProgressBar()
	bar.Options.Width = 10
	bar.Set(100)
	c := NewCanvas(20, 1)
	bar.Draw(c, Rect{X: 0, Y: 0, W: 5, H: 1})
	if got := strings.TrimRight(plainRow(c), " "); got != "[⣿⣿⣿⣿" {
		t.Fatalf("clipped bar drew %q", got)
	}
}

func TestProgressBarIndeterminateLabelsAndStyles(t *testing.T) {
	bar := NewProgressBar()
	bar.Options.Width = 6
	bar.Total = 24
	bar.ShowPercent, bar.ShowCount, bar.Unit = true, true, " files"
	bar.Indeterminate = true
	bar.Set(12)
	if got := bar.TickInterval(); got <= 0 {
		t.Fatal("indeterminate bar has no tick interval")
	}
	before := renderProgressBar(bar, 40)
	bar.Tick(time.Time{})
	after := renderProgressBar(bar, 40)
	if before == after {
		t.Fatal("tick did not move the marquee")
	}
	if !strings.Contains(after, "50%") || !strings.Contains(after, "12/24 files") {
		t.Fatalf("labels missing from %q", after)
	}
	bar.StyleFill = Style{FG: ColorIndex(2)}
	bar.StyleEmpty = Style{FG: ColorIndex(4)}
	bar.Draw(NewCanvas(40, 1), Rect{W: 40, H: 1})
}

func TestBracketedBarStepMatchesRender(t *testing.T) {
	opts := graph.BracketedBarOptions{Width: 7, SubChar: true}
	last, lastBar := -1, ""
	for v := 0.0; v <= 100; v += 0.37 {
		step, bar := graph.BracketedBarStep(v, opts), graph.RenderBracketedBar(v, opts)
		if (step == last) != (bar == lastBar) {
			t.Fatalf("v=%v: step change %v but bar change %v", v, step != last, bar != lastBar)
		}
		last, lastBar = step, bar
	}
}

func renderProgressBar(bar *ProgressBar, w int) string {
	c := NewCanvas(w, 1)
	bar.Draw(c, Rect{X: 0, Y: 0, W: w, H: 1})
	return plainRow(c)
}

func plainRow(c *Canvas) string {
	return strings.TrimSuffix(c.Row(0), "\x1b[0m")
}
