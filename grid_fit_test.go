// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"ubunatic.com/loom"
	"ubunatic.com/loom/measure"
)

type fitChild struct {
	mouse loom.MouseEvent
	keys  int
}

func (w *fitChild) Draw(c *loom.Canvas, r loom.Rect) {
	c.Fill(r, loom.Cell{Text: "x"})
}
func (w *fitChild) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if e.Key == "enter" {
		w.keys++
		return loom.Handled()
	}
	return loom.Ignored()
}
func (w *fitChild) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	w.mouse = e
	return loom.Handled()
}

type measuredFitChild struct {
	fitChild
	height, width int
}

func (w *measuredFitChild) Measure(width int) measure.Size {
	w.width = width
	return measure.Size{Height: w.height}
}

type wrappingFitChild struct{ fitChild }

func (*wrappingFitChild) Measure(width int) measure.Size {
	return measure.Size{Height: (10 + max(1, width) - 1) / max(1, width)}
}

func TestGridFitRows(t *testing.T) {
	for _, mode := range []loom.GridBorderMode{loom.GridBorderNone, loom.GridBorderInner, loom.GridBorderFull} {
		t.Run(string(mode), func(t *testing.T) {
			a := &measuredFitChild{height: 2}
			b := &measuredFitChild{height: 4}
			fallback := &fitChild{}
			zero := &measuredFitChild{height: -1}
			g := loom.NewGrid(2, a, b, fallback, zero)
			g.FitRows, g.BorderMode = true, mode
			c := loom.NewCanvas(30, 20)
			r := loom.Rect{X: 3, Y: 2, W: 20, H: 15}
			g.Draw(c, r)
			gap, inset := 0, 0
			if mode != loom.GridBorderNone {
				gap = 1
			}
			if mode == loom.GridBorderFull {
				inset = 1
			}
			for i, h := range []int{4, 4, 1, 1} {
				cr := g.ChildRect(i)
				wantY := r.Y + inset
				if i >= 2 {
					wantY += 4 + gap
				}
				if cr.H != h || cr.Y != wantY {
					t.Fatalf("child %d = %+v, want H=%d Y=%d", i, cr, h, wantY)
				}
			}
			if a.width != g.ChildRect(0).W || b.width != g.ChildRect(1).W {
				t.Fatal("measure did not receive cell width")
			}
			if mode != loom.GridBorderNone {
				separator := g.ChildRect(0)
				if g.ConsumeMouse(loom.MouseEvent{X: separator.X - r.X, Y: separator.Y - r.Y + separator.H, Action: loom.MousePress}).Consumed {
					t.Fatal("row separator routed to a child")
				}
			}
			cr := g.ChildRect(3)
			res := g.ConsumeMouse(loom.MouseEvent{X: cr.X - r.X + 2, Y: cr.Y - r.Y, Action: loom.MousePress})
			if !res.Consumed || g.Focus() != 3 || zero.mouse.X != 2 || zero.mouse.Y != 0 {
				t.Fatalf("mouse result=%+v focus=%d event=%+v", res, g.Focus(), zero.mouse)
			}
			for _, step := range []struct {
				key   string
				focus int
			}{{"up", 1}, {"down", 3}, {"down", 1}, {"shift-down", 3}} {
				if !g.ConsumeKey(loom.KeyEvent{Key: step.key}).Consumed || g.Focus() != step.focus {
					t.Fatalf("%s focus=%d, want %d", step.key, g.Focus(), step.focus)
				}
			}
			g.ConsumeKey(loom.KeyEvent{Key: "enter"})
			if zero.keys != 1 {
				t.Fatal("key did not reach focused child")
			}
			if g.ConsumeMouse(loom.MouseEvent{X: 2, Y: cr.Y - r.Y + 2, Action: loom.MousePress}).Consumed {
				t.Fatal("spare height routed to a child")
			}
			if c.Get(r.X, r.Y+r.H-2).Text == "x" {
				t.Fatal("fitted rows stretched into spare height")
			}
		})
	}
}

func TestGridFitRowsRemeasuresAtCellWidth(t *testing.T) {
	g := loom.NewGrid(1, &wrappingFitChild{})
	g.FitRows = true
	c := loom.NewCanvas(20, 10)
	for _, tt := range []struct{ width, height int }{{10, 1}, {3, 4}} {
		g.Draw(c, loom.Rect{W: tt.width, H: 10})
		if got := g.ChildRect(0).H; got != tt.height {
			t.Fatalf("width %d: row height=%d, want measured height %d", tt.width, got, tt.height)
		}
	}
}

func TestGridFitRowsFullBorderClipsMouse(t *testing.T) {
	g := loom.NewGrid(1, &measuredFitChild{height: 10})
	g.FitRows, g.BorderMode = true, loom.GridBorderFull
	c := loom.NewCanvas(10, 10)
	r := loom.Rect{X: 2, Y: 2, W: 6, H: 5}
	g.Draw(c, r)
	if got := c.Get(3, 6).Text; got != "─" {
		t.Fatalf("bottom border=%q", got)
	}
	if g.ConsumeMouse(loom.MouseEvent{X: 1, Y: 4, Action: loom.MousePress}).Consumed {
		t.Fatal("bottom border routed to overflowing child")
	}
}

func TestGridFitRowsClipsOverflow(t *testing.T) {
	a, b := &measuredFitChild{height: 4}, &measuredFitChild{height: 3}
	g := loom.NewGrid(1, a, b)
	g.FitRows = true
	c := loom.NewCanvas(12, 12)
	r := loom.Rect{X: 2, Y: 2, W: 6, H: 5}
	g.Draw(c, r)
	if got := g.ChildRect(1); got.Y != 6 || got.H != 3 {
		t.Fatalf("overflow rect=%+v", got)
	}
	if c.Get(2, 6).Text != "x" || c.Get(2, 7).Text == "x" {
		t.Fatal("overflow was not clipped to Grid area")
	}
	if g.ConsumeMouse(loom.MouseEvent{X: 0, Y: 5, Action: loom.MousePress}).Consumed {
		t.Fatal("mouse reached clipped content")
	}
}
