// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"testing"

	"codeberg.org/ubunatic/loom/layout"
)

type mockWidget struct {
	w, h        int
	drawnRect   Rect
	handledKey  string
	handledMove bool
}

func (m *mockWidget) Draw(c *Canvas, r Rect) {
	m.drawnRect = r
	c.Fill(r, Cell{Text: "X"})
}

func (m *mockWidget) HandleKey(e KeyEvent) bool {
	m.handledKey = e.Key
	return e.Key == "q"
}

func (m *mockWidget) HandleMouse(e MouseEvent) bool {
	m.handledMove = true
	return false
}

func (m *mockWidget) ContentWidth() int  { return m.w }
func (m *mockWidget) ContentHeight() int { return m.h }

func TestCenterDrawPlacement(t *testing.T) {
	child := &mockWidget{w: 20, h: 5}
	center := NewCenter(child)

	c := NewCanvas(80, 25)
	center.Draw(c, c.Bounds())

	// 80 - 20 = 60 -> offset X = 30
	// 25 - 5 = 20 -> offset Y = 10
	want := Rect{X: 30, Y: 10, W: 20, H: 5}
	if child.drawnRect != want {
		t.Fatalf("drawnRect = %+v, want %+v", child.drawnRect, want)
	}
}

func TestAlignPlacement(t *testing.T) {
	child := &mockWidget{w: 10, h: 4}
	alignEnd := NewAlign(child, layout.AlignEnd, layout.AlignEnd)

	c := NewCanvas(50, 20)
	alignEnd.Draw(c, c.Bounds())

	// 50 - 10 = 40 -> offset X = 40
	// 20 - 4 = 16 -> offset Y = 16
	want := Rect{X: 40, Y: 16, W: 10, H: 4}
	if child.drawnRect != want {
		t.Fatalf("drawnRect = %+v, want %+v", child.drawnRect, want)
	}
}

func TestAlignOverflowClamping(t *testing.T) {
	child := &mockWidget{w: 100, h: 50}
	center := NewCenter(child)

	c := NewCanvas(40, 10)
	center.Draw(c, c.Bounds())

	// Container smaller than content: clamps to container dimensions at 0,0
	want := Rect{X: 0, Y: 0, W: 40, H: 10}
	if child.drawnRect != want {
		t.Fatalf("drawnRect = %+v, want %+v", child.drawnRect, want)
	}
}

func TestAlignDelegation(t *testing.T) {
	child := &mockWidget{w: 10, h: 2}
	center := NewCenter(child)

	if !center.HandleKey(KeyEvent{Key: "q"}) {
		t.Errorf("expected quit from child")
	}
	if child.handledKey != "q" {
		t.Errorf("handledKey = %q, want 'q'", child.handledKey)
	}

	center.HandleMouse(MouseEvent{Action: MouseHover})
	if !child.handledMove {
		t.Errorf("expected mouse event delegated to child")
	}

	if center.ContentWidth() != 10 {
		t.Errorf("ContentWidth = %d, want 10", center.ContentWidth())
	}
	if center.ContentHeight() != 2 {
		t.Errorf("ContentHeight = %d, want 2", center.ContentHeight())
	}
}
