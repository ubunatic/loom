package loom_test

import (
	"testing"
	"time"

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

func TestCursorBrightenTextAndIndexedBackgrounds(t *testing.T) {
	c := loom.NewCanvas(4, 2)
	c.SetCursorPosition(1, 0)
	c.Set(1, 0, loom.Cell{Text: "x", Style: loom.Style{FG: loom.ColorIndex(2), BG: loom.ColorRGB(10, 20, 30)}})
	c.Set(2, 0, loom.Cell{Text: " ", Style: loom.Style{BG: loom.ColorIndex(16)}})
	c.Set(3, 0, loom.Cell{Text: " ", Style: loom.Style{BG: loom.ColorReset()}})
	c.ApplyCursorBrighten()
	text := c.Get(1, 0)
	if text.Text != "x" || text.Style.FG != loom.ColorIndex(2) || text.Style.BG == loom.ColorRGB(10, 20, 30) {
		t.Fatalf("brightened text cell = %+v", text)
	}
	if got := c.Get(2, 0).Style.BG; got != loom.ColorRGB(134, 134, 134) {
		t.Fatalf("indexed background = %+v; want brightened RGB", got)
	}
	if got := c.Get(3, 0).Style.BG; got != loom.ColorReset() {
		t.Fatalf("terminal-default background changed: %+v", got)
	}
}

func TestCursorStarTrailFadesAndExpires(t *testing.T) {
	c := loom.NewCanvas(8, 3)
	now := time.Unix(100, 0)
	trail := []loom.CursorTrailPoint{
		{X: 1, Y: 1, At: now.Add(-200 * time.Millisecond)},
		{X: 2, Y: 1, At: now},
	}
	c.ApplyCursorStarTrail(trail, now)
	fresh, faded := c.Get(2, 1), c.Get(1, 1)
	if fresh.Text != "✦" || fresh.Style.FG != loom.ColorRGB(255, 255, 255) || fresh.Style.Dim {
		t.Fatalf("fresh star = %+v", fresh)
	}
	if faded.Text != "✦" || faded.Style.FG != loom.ColorRGB(128, 128, 128) || !faded.Style.Dim {
		t.Fatalf("fading star = %+v; want dim star", faded)
	}
	c.Set(4, 1, loom.Cell{Text: "x", Style: loom.Style{FG: loom.ColorIndex(2)}})
	c.ApplyCursorStarTrail([]loom.CursorTrailPoint{{X: 4, Y: 1, At: now}}, now)
	if got := c.Get(4, 1).Text; got != "x" {
		t.Fatalf("star trail replaced foreground text: %q", got)
	}
	c.Clear()
	c.ApplyCursorStarTrail(trail[:1], now.Add(loom.SpeccedCursorStarTrail.Lifetime))
	if got := c.Get(1, 1).Text; got == "✦" {
		t.Fatal("expired trail point was still rendered")
	}
}

func TestCursorStarTrailSpec(t *testing.T) {
	if spec := loom.SpeccedCursorStarTrail; !spec.Enabled || spec.Glyph != "✦" || spec.MaxPoints != 8 || spec.Lifetime != 400*time.Millisecond {
		t.Fatalf("SpeccedCursorStarTrail = %+v", spec)
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
