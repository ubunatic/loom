// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
)

func TestChartRendersAxesAndSingleLineWithoutLegend(t *testing.T) {
	chart := &Chart{Series: []ChartSeries{{Name: "CPU", Values: []float64{0, 5, 10}}}}
	c := NewCanvas(24, 8)
	chart.Draw(c, Rect{W: 24, H: 8})
	joined := strings.Join(Render(chart, 24, 8), "\n")
	if !strings.Contains(joined, "0") || !strings.Contains(joined, "10") || !hasBrailleGlyph(joined) {
		t.Fatalf("chart should show scaled labels and plotted line: %q", joined)
	}
	if strings.Contains(joined, "Legend") {
		t.Fatalf("single series should not show legend: %q", joined)
	}
}

func hasBrailleGlyph(text string) bool {
	for _, r := range text {
		if r >= 0x2800 && r <= 0x28ff {
			return true
		}
	}
	return false
}

func TestChartGroupedBarsAndLegend(t *testing.T) {
	chart := &Chart{Mode: ChartGroupedBar, Series: []ChartSeries{
		{Name: "A", Values: []float64{2, 5}}, {Name: "B", Values: []float64{4, 3}},
	}}
	c := NewCanvas(30, 10)
	chart.Draw(c, Rect{W: 30, H: 10})
	joined := strings.Join(Render(chart, 30, 10), "\n")
	if !strings.Contains(joined, "A") || !strings.Contains(joined, "B") || !strings.Contains(joined, "█") {
		t.Fatalf("grouped chart missing legend or bars: %q", joined)
	}
}

func TestChartClipsAndHandlesConstantValues(t *testing.T) {
	chart := &Chart{Series: []ChartSeries{{Values: []float64{3, 3, 3}}}}
	c := NewCanvas(5, 2)
	chart.Draw(c, Rect{W: 5, H: 2})
	if c.Get(4, 1).Text == "" {
		t.Fatal("chart should safely draw within a small canvas")
	}
}

func TestChartLineUsesBrailleSubcellDots(t *testing.T) {
	chart := &Chart{Series: []ChartSeries{{Values: []float64{0, 5, 10}}}}
	rows := Render(chart, 32, 10)
	found := false
	for _, row := range rows {
		for _, r := range row {
			if r >= 0x2800 && r <= 0x28ff {
				found = true
				if r == '⠿' {
					t.Fatalf("line used dense braille cell %q", r)
				}
			}
		}
	}
	if !found {
		t.Fatal("line chart did not render braille subcell dots")
	}
}

func TestChartYAxisIsContinuousAndDataStartsAfterIt(t *testing.T) {
	chart := &Chart{Series: []ChartSeries{{Values: []float64{0, 10, 0, 10}}}}
	c := NewCanvas(32, 10)
	chart.Draw(c, Rect{W: 32, H: 10})
	axisX, plotY, plotH := 6, 0, 8
	for y := plotY; y < plotY+plotH; y++ {
		if got := c.Get(axisX, y).Text; got != "│" {
			t.Errorf("Y axis row %d = %q, want │", y, got)
		}
	}
	if got := c.Get(axisX+1, plotY).Text; got == "│" {
		t.Fatalf("data area starts with the Y axis glyph at x=%d", axisX+1)
	}
}

func TestChartUsesNiceOutwardYTicks(t *testing.T) {
	lo, hi, step := niceChartRange(9.25, 23.75)
	if lo != 5 || hi != 25 || step != 5 {
		t.Fatalf("niceChartRange(9.25, 23.75) = (%g, %g, %g), want (5, 25, 5)", lo, hi, step)
	}
	joined := strings.Join(Render(&Chart{Series: []ChartSeries{{Values: []float64{9.25, 23.75}}}}, 32, 10), "\n")
	if !strings.Contains(joined, "25") || !strings.Contains(joined, "5") {
		t.Fatalf("chart ticks omit outward-rounded bounds: %q", joined)
	}
}

func TestChartMeasureFitsPlotAxisAndLegend(t *testing.T) {
	one := &Chart{Series: []ChartSeries{{Values: []float64{1, 2}}}}
	if got := one.Measure(30); got.Width != 30 || got.Height != 7 {
		t.Fatalf("one series: %+v, want 30x7", got)
	}
	two := &Chart{Series: []ChartSeries{{Values: []float64{1}}, {Values: []float64{2}}}}
	if got := two.Measure(30).Height; got != 8 {
		t.Fatalf("two series height = %d, want 8 (legend line)", got)
	}
	if out := strings.Join(Render(one, 30, 7), "\n"); !strings.Contains(out, "└") || !hasBrailleGlyph(out) {
		t.Fatalf("chart at measured height drew no axis or line:\n%s", out)
	}
}
