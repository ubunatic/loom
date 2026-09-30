// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/measure"
)

type viewportChild struct {
	keys   []loom.KeyEvent
	mouse  loom.MouseEvent
	width  int
	height int
}

func (w *viewportChild) Draw(c *loom.Canvas, r loom.Rect) {
	for y := 0; y < r.H; y++ {
		for x := 0; x < r.W; x++ {
			c.Set(r.X+x, r.Y+y, loom.Cell{Text: string(rune('A' + y))})
		}
	}
}
func (w *viewportChild) HandleKey(e loom.KeyEvent) bool     { w.keys = append(w.keys, e); return false }
func (w *viewportChild) HandleMouse(e loom.MouseEvent) bool { w.mouse = e; return false }
func (w *viewportChild) Measure(int) measure.Size {
	return measure.Size{Width: w.width, Height: w.height}
}

func TestViewportDrawScrollAndClamp(t *testing.T) {
	child := &viewportChild{width: 4, height: 6}
	viewport := loom.NewViewport(child)
	viewport.ScrollbarMode = loom.ScrollbarNever
	canvas := loom.NewCanvas(4, 3)
	viewport.Draw(canvas, canvas.Bounds())
	if got := canvas.Get(0, 0).Text; got != "A" {
		t.Fatalf("top cell = %q, want A", got)
	}
	viewport.HandleKey(loom.KeyEvent{Key: "pgdown"})
	viewport.Draw(canvas, canvas.Bounds())
	if viewport.ScrollY != 3 {
		t.Fatalf("ScrollY = %d, want 3", viewport.ScrollY)
	}
	if got := canvas.Get(0, 0).Text; got != "D" {
		t.Fatalf("scrolled cell = %q, want D", got)
	}
	viewport.HandleKey(loom.KeyEvent{Key: "end"})
	viewport.HandleKey(loom.KeyEvent{Key: "down"})
	if viewport.ScrollY != 3 {
		t.Fatalf("clamped ScrollY = %d, want 3", viewport.ScrollY)
	}
	viewport.HandleKey(loom.KeyEvent{Key: "home"})
	if viewport.ScrollY != 0 {
		t.Fatalf("home ScrollY = %d, want 0", viewport.ScrollY)
	}
}

func TestViewportTranslatesMouseAndScrollWheel(t *testing.T) {
	child := &viewportChild{width: 8, height: 8}
	viewport := loom.NewViewport(child)
	viewport.ScrollbarMode = loom.ScrollbarNever
	canvas := loom.NewCanvas(4, 3)
	viewport.Draw(canvas, loom.Rect{X: 2, Y: 1, W: 4, H: 3})
	viewport.HandleKey(loom.KeyEvent{Key: "down"})
	viewport.HandleKey(loom.KeyEvent{Key: "right"})
	viewport.HandleMouse(loom.MouseEvent{Action: loom.MousePress, X: 1, Y: 2})
	if child.mouse.X != 2 || child.mouse.Y != 3 {
		t.Fatalf("child mouse = (%d,%d), want (2,3)", child.mouse.X, child.mouse.Y)
	}
	viewport.HandleMouse(loom.MouseEvent{Action: loom.MouseScrollDown})
	if viewport.ScrollY != 2 {
		t.Fatalf("wheel ScrollY = %d, want 2", viewport.ScrollY)
	}
}

func TestViewportKeyAliases(t *testing.T) {
	for _, key := range []string{"pgdown", "pgdn", "pagedown"} {
		viewport := loom.NewViewport(&viewportChild{width: 2, height: 8})
		viewport.Draw(loom.NewCanvas(2, 2), loom.Rect{W: 2, H: 2})
		viewport.HandleKey(loom.KeyEvent{Key: key})
		if viewport.ScrollY != 2 {
			t.Errorf("%s ScrollY = %d, want 2", key, viewport.ScrollY)
		}
	}
	for _, key := range []string{"pgup", "pageup"} {
		viewport := loom.NewViewport(&viewportChild{width: 2, height: 8})
		viewport.ScrollY = 4
		viewport.Draw(loom.NewCanvas(2, 2), loom.Rect{W: 2, H: 2})
		viewport.HandleKey(loom.KeyEvent{Key: key})
		if viewport.ScrollY != 2 {
			t.Errorf("%s ScrollY = %d, want 2", key, viewport.ScrollY)
		}
	}
}
