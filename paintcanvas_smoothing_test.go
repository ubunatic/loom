package loom

import (
	"reflect"
	"testing"
)

func TestPaintCanvasSmoothingGeometry(t *testing.T) {
	stroke := func(t *testing.T, strength int, points [][2]int) map[[2]int]chartDots {
		t.Helper()
		p := &PaintCanvas{Smoothing: strength}
		p.Draw(NewCanvas(12, 12), Rect{W: 12, H: 12})
		for i, point := range points {
			action := MouseDrag
			if i == 0 {
				action = MousePress
			}
			p.ConsumeMouse(MouseEvent{Action: action, Button: MouseLeft, X: point[0], Y: point[1]})
		}
		end := points[len(points)-1]
		p.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: end[0], Y: end[1]})
		return p.dots
	}
	diagonal := [][2]int{{1, 1}, {2, 1}, {2, 2}, {3, 2}, {3, 3}, {4, 3}, {4, 4}, {5, 4}, {5, 5}}
	want := make(map[[2]int]chartDots)
	rasterChartLine(want, 2, 4, 10, 20, Style{})
	if got := stroke(t, 2, diagonal); !reflect.DeepEqual(got, want) {
		t.Fatalf("slow diagonal retained staircase jitter: %v, want %v", got, want)
	}
	raw := make(map[[2]int]chartDots)
	for i := 1; i < len(diagonal); i++ {
		a, b := diagonal[i-1], diagonal[i]
		rasterChartLine(raw, a[0]*2, a[1]*4, b[0]*2, b[1]*4, Style{})
	}
	if got := stroke(t, 0, diagonal); !reflect.DeepEqual(got, raw) {
		t.Fatal("option zero changed the existing rasterization")
	}
	// Even a one-cell-wide closed rectangle retains every corner and stays in bounds.
	rectangle := [][2]int{{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1}}
	if smooth, original := stroke(t, 2, rectangle), stroke(t, 0, rectangle); !reflect.DeepEqual(smooth, original) {
		t.Fatalf("smoothing rounded a small rectangle: %v, want %v", smooth, original)
	}
}
