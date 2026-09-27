package loom_test

import (
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestCursorProximityHintAndBrighten(t *testing.T) {
	c := loom.NewCanvas(7, 5)
	c.SetCursorPosition(2, 1)
	if _, ok := c.CursorHintAt(6, 4); ok {
		t.Fatal("position outside configured radius should not have a hint")
	}
	hint, ok := c.CursorHintAt(1, 1)
	if !ok || hint.DX != -1 || hint.DY != 0 || hint.Distance != 1 {
		t.Fatalf("CursorHintAt(1,1) = %+v, %v; want dx=-1 dy=0 distance=1", hint, ok)
	}
	hint, ok = c.CursorHintAt(2, 1)
	if !ok || hint.DX != 0 || hint.DY != 0 || hint.Distance != 0 {
		t.Fatalf("cursor hint = %+v, %v; want zero delta at cursor", hint, ok)
	}
	c.PaintSurface(loom.Rect{X: 2, Y: 1, W: 2, H: 1}, loom.Style{BG: loom.ColorRGB(20, 40, 60)})
	c.ApplyCursorBrighten()
	if got := c.Get(2, 1).Style.BG; got != loom.ColorRGB(185, 191, 197) {
		t.Fatalf("cursor background = %+v; want brightest configured blend", got)
	}
	if got := c.Get(3, 1).Style.BG; got == loom.ColorRGB(20, 40, 60) {
		t.Fatal("neighbor background did not brighten")
	}
}

func TestCursorProximityEdgesMissingAndDisabled(t *testing.T) {
	c := loom.NewCanvas(3, 2)
	if _, ok := c.CursorHintAt(0, 0); ok {
		t.Fatal("new canvas unexpectedly has a cursor hint")
	}
	c.SetCursorPosition(0, 0)
	if _, ok := c.CursorHintAt(-1, 0); ok {
		t.Fatal("out-of-bounds position unexpectedly has a hint")
	}
	if _, ok := c.CursorHintAt(3, 0); ok {
		t.Fatal("out-of-bounds position unexpectedly has a hint")
	}
	c.ClearCursorPosition()
	if _, ok := c.CursorHintAt(0, 0); ok {
		t.Fatal("cleared cursor unexpectedly has a hint")
	}
	style := loom.Style{BG: loom.ColorRGB(20, 40, 60)}
	c.PaintSurface(loom.Rect{W: 1, H: 1}, style)
	c.ApplyCursorBrighten()
	if got := c.Get(0, 0); got.Style.BG != style.BG {
		t.Fatalf("background changed with no cursor: %+v", got.Style.BG)
	}
}

func TestCursorProximitySpec(t *testing.T) {
	if spec := loom.SpeccedCursorProximity; !spec.Enabled || spec.Radius != 3 || spec.Brightness != .7 {
		t.Fatalf("SpeccedCursorProximity = %+v", spec)
	}
}
