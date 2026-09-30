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
	if !strings.Contains(joined, "0") || !strings.Contains(joined, "10") || !strings.Contains(joined, "⠿") {
		t.Fatalf("chart should show scaled labels and plotted line: %q", joined)
	}
	if strings.Contains(joined, "Legend") {
		t.Fatalf("single series should not show legend: %q", joined)
	}
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
