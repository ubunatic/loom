// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom/graph"
)

func TestGaugeDrawsMetricValueWithinBounds(t *testing.T) {
	store := NewMetricStore()
	store.Publish("ram", 50, time.Unix(1, 0))
	gauge := Gauge{Store: store, Metric: "ram", Width: 4, Options: graph.BarOptions{SubChar: true}}
	canvas := NewCanvas(7, 2)
	gauge.Draw(canvas, Rect{X: 1, Y: 1, W: 6, H: 1})
	if got := strings.TrimSuffix(canvas.Row(1), "\x1b[0m"); got != " [██░░]" {
		t.Errorf("gauge row = %q, want %q", got, " [██░░]")
	}
	// A narrow allocation must not write into the neighboring column.
	narrow := NewCanvas(8, 1)
	gauge.Draw(narrow, Rect{X: 2, Y: 0, W: 3, H: 1})
	if got := strings.TrimSuffix(narrow.Row(0), "\x1b[0m"); got != "  [██   " {
		t.Errorf("narrow gauge row = %q, want content clipped to bounds: %q", got, "  [██   ")
	}
}

func TestSparklineDrawsBoundMetricSeries(t *testing.T) {
	store := NewMetricStore()
	store.Publish("cpu", 0, time.Unix(1, 0))
	store.Publish("cpu", 100, time.Unix(2, 0))
	sparkline := Sparkline{
		Store:  store,
		Series: "cpu",
		Width:  2,
		Min:    0,
		Max:    100,
	}
	canvas := NewCanvas(2, 1)
	sparkline.Draw(canvas, Rect{W: 2, H: 1})
	want := graph.RenderSparkline([]float64{0, 100}, graph.SparklineOptions{Width: 2, FixedRange: true, Min: 0, Max: 100})
	if got := strings.TrimSuffix(canvas.Row(0), "\x1b[0m"); got != want {
		t.Errorf("sparkline row = %q, want %q", got, want)
	}
}

func TestGraphWidgetsArePassiveAndHavePreferredSizes(t *testing.T) {
	gauge := Gauge{Width: 4}
	if gauge.ContentWidth() != 6 || gauge.ContentHeight() != 1 {
		t.Errorf("gauge preferred size = %dx%d, want 6x1", gauge.ContentWidth(), gauge.ContentHeight())
	}
	sparkline := Sparkline{Width: 4}
	if sparkline.ContentWidth() != 4 || sparkline.ContentHeight() != 1 {
		t.Errorf("sparkline preferred size = %dx%d, want 4x1", sparkline.ContentWidth(), sparkline.ContentHeight())
	}
	if gauge.HandleKey(KeyEvent{}) || gauge.HandleMouse(MouseEvent{}) || sparkline.HandleKey(KeyEvent{}) || sparkline.HandleMouse(MouseEvent{}) {
		t.Fatal("graph widgets consumed an input event")
	}
}
