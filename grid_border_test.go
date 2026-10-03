// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"ubunatic.com/loom"
)

type gridBorderChild struct {
	drawn    loom.Rect
	mouse    loom.MouseEvent
	gotMouse bool
}

func (w *gridBorderChild) Draw(c *loom.Canvas, r loom.Rect) {
	w.drawn = r
	if r.W > 0 && r.H > 0 {
		c.Write(r.X, r.Y, "X", loom.Reset)
	}
}
func (*gridBorderChild) ContentHeight() int                          { return 1 }
func (w *gridBorderChild) ConsumeKey(loom.KeyEvent) loom.EventResult { return loom.Ignored() }
func (w *gridBorderChild) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	w.mouse, w.gotMouse = e, true
	return loom.Handled()
}

func canvasRows(c *loom.Canvas) []string {
	rows := make([]string, c.Rows())
	for y := 0; y < c.Rows(); y++ {
		var b strings.Builder
		for x := 0; x < c.Cols(); x++ {
			b.WriteString(c.Get(x, y).Text)
		}
		rows[y] = b.String()
	}
	return rows
}

func TestGridBorderModesAndChildLayout(t *testing.T) {
	tests := []struct {
		name string
		mode loom.GridBorderMode
		want []string
	}{
		{name: "none", want: []string{"X   X    ", "         ", "X        ", "         ", "         "}},
		{name: "inner", mode: loom.GridBorderInner, want: []string{"X   │X   ", "    │    ", "────┼────", "X   │    ", "    │    "}},
		{name: "full", mode: loom.GridBorderFull, want: []string{"┌───┬───┐", "│X  │X  │", "├───┼───┤", "│X  │   │", "└───┴───┘"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := loom.NewGrid(2, &gridBorderChild{}, &gridBorderChild{}, &gridBorderChild{})
			children := []*gridBorderChild{grid.Children[0].(*gridBorderChild), grid.Children[1].(*gridBorderChild), grid.Children[2].(*gridBorderChild)}
			grid.BorderMode = tt.mode
			c := loom.NewCanvas(9, 5)
			grid.Draw(c, c.Bounds())
			got := canvasRows(c)
			for y := range tt.want {
				if got[y] != tt.want[y] {
					t.Errorf("row %d = %q, want %q", y, got[y], tt.want[y])
				}
			}
			first, second := grid.ChildRect(0), grid.ChildRect(1)
			for i, child := range children {
				if child.drawn != grid.ChildRect(i) {
					t.Errorf("child %d Draw rect = %+v, ChildRect = %+v", i, child.drawn, grid.ChildRect(i))
				}
			}
			if tt.mode == loom.GridBorderFull && (first != (loom.Rect{X: 1, Y: 1, W: 3, H: 1}) || second != (loom.Rect{X: 5, Y: 1, W: 3, H: 1})) {
				t.Errorf("full border child rects = %+v, %+v", first, second)
			}
		})
	}
}

func TestGridFullBorderNarrowAndMouseBounds(t *testing.T) {
	a, b := &gridBorderChild{}, &gridBorderChild{}
	g := loom.NewGrid(2, a, b)
	g.BorderMode = loom.GridBorderFull
	c := loom.NewCanvas(9, 5)
	r := loom.Rect{X: 1, Y: 1, W: 3, H: 2}
	g.Draw(c, r)
	if got := g.ChildRect(0); got.W != 0 || got.H != 0 {
		t.Errorf("narrow child rect = %+v, want zero content area", got)
	}
	if got := c.Get(1, 1).Text; got != "┌" {
		t.Fatalf("narrow grid corner = %q, want ┌", got)
	}
	for _, point := range []struct {
		x, y int
		want string
	}{{3, 1, "┐"}, {1, 2, "└"}, {3, 2, "┘"}} {
		if got := c.Get(point.x, point.y).Text; got != point.want {
			t.Errorf("narrow grid border (%d,%d) = %q, want %q", point.x, point.y, got, point.want)
		}
	}

	// Give the children content room and verify border cells don't route input.
	r = loom.Rect{X: 0, Y: 0, W: 9, H: 5}
	g.Draw(c, r)
	if got := g.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, X: 0, Y: 1}); got.Consumed {
		t.Fatal("outer border mouse press was consumed by a child")
	}
	g.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, X: 1, Y: 1})
	if !a.gotMouse || a.mouse.X != 0 || a.mouse.Y != 0 {
		t.Fatalf("child received mouse %+v, got=%v; want cell-local (0,0)", a.mouse, a.gotMouse)
	}
}

func TestGridBordersDoNotRouteMouseToSeparatorsOrEmptyCells(t *testing.T) {
	a, b, c := &gridBorderChild{}, &gridBorderChild{}, &gridBorderChild{}
	g := loom.NewGrid(2, a, b, c)
	g.BorderMode = loom.GridBorderFull
	g.Draw(loom.NewCanvas(14, 8), loom.Rect{X: 2, Y: 1, W: 9, H: 5})
	for _, point := range []struct{ x, y int }{{0, 1}, {4, 1}, {4, 2}, {5, 3}} {
		result := g.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, X: point.x, Y: point.y})
		if result.Consumed {
			t.Errorf("mouse at border/empty cell (%d,%d) reached a child", point.x, point.y)
		}
	}
	if result := g.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, X: 5, Y: 1}); !result.Consumed {
		t.Fatal("mouse in second child content was not consumed")
	}
	if !b.gotMouse || b.mouse.X != 0 || b.mouse.Y != 0 {
		t.Fatalf("second child mouse = %+v, got=%v; want cell-local (0,0)", b.mouse, b.gotMouse)
	}
}

func TestGridBorderContentHeight(t *testing.T) {
	for _, tc := range []struct {
		mode loom.GridBorderMode
		want int
	}{{loom.GridBorderNone, 2}, {loom.GridBorderInner, 3}, {loom.GridBorderFull, 5}} {
		grid := loom.NewGrid(1, &gridBorderChild{}, &gridBorderChild{})
		grid.BorderMode = tc.mode
		if got := grid.ContentHeight(); got != tc.want {
			t.Errorf("ContentHeight(%q) = %d, want %d", tc.mode, got, tc.want)
		}
	}
}

func TestGridFullBorderDistributesRemainder(t *testing.T) {
	a, b, c := &gridBorderChild{}, &gridBorderChild{}, &gridBorderChild{}
	g := loom.NewGrid(2, a, b, c)
	g.BorderMode = loom.GridBorderFull
	canvas := loom.NewCanvas(10, 6)
	g.Draw(canvas, loom.Rect{W: 10, H: 6})
	first, second, third := g.ChildRect(0), g.ChildRect(1), g.ChildRect(2)
	if first != (loom.Rect{X: 1, Y: 1, W: 4, H: 2}) {
		t.Errorf("first child rect = %+v, want content to absorb remainder", first)
	}
	if second != (loom.Rect{X: 6, Y: 1, W: 3, H: 2}) {
		t.Errorf("second child rect = %+v, want aligned inner border", second)
	}
	if third != (loom.Rect{X: 1, Y: 4, W: 4, H: 1}) {
		t.Errorf("third child rect = %+v, want final row to end at bottom border", third)
	}
	for i, child := range []*gridBorderChild{a, b, c} {
		if child.drawn != g.ChildRect(i) {
			t.Errorf("child %d drew at %+v, ChildRect = %+v", i, child.drawn, g.ChildRect(i))
		}
	}
	if canvas.Get(9, 0).Text != "┐" || canvas.Get(9, 5).Text != "┘" {
		t.Errorf("remainder borders at far edge = %q, %q; want ┐ and ┘", canvas.Get(9, 0).Text, canvas.Get(9, 5).Text)
	}
}
