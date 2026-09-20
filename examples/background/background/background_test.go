package background

import (
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestDemoFrameInitializesSplit(t *testing.T) {
	frame := demoFrame()
	if frame == nil || len(frame.Boxes) != 1 {
		t.Fatalf("demoFrame = %#v, want one split panel", frame)
	}
	if frame.Boxes[0].Child == nil {
		t.Fatal("demoFrame has no panel child")
	}
}

func TestWidgetDraw(t *testing.T) {
	w := &widget{metrics: &loom.RenderMetrics{}, frame: demoFrame()}
	c := loom.NewCanvas(100, 30)
	w.Draw(c, c.Bounds())
	if c.Get(2, 1).Text == "" {
		t.Fatal("Draw did not paint the heading")
	}
}

func TestWidgetKeyHandling(t *testing.T) {
	pane := &loom.Pane{}
	w := &widget{pane: pane, frame: demoFrame(), themes: []string{"plain", "mc"}}
	if w.HandleKey(loom.KeyEvent{Key: "m"}) || !pane.ReduceMotion {
		t.Fatal("m should toggle reduced motion")
	}
	if w.HandleKey(loom.KeyEvent{Key: "T"}) || w.theme != 1 {
		t.Fatal("T should cycle the theme")
	}
	if !w.HandleKey(loom.KeyEvent{Key: "q"}) {
		t.Fatal("background widget should delegate quit to its frame")
	}
}
