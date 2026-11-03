package background

import (
	"testing"

	"ubunatic.com/loom"
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
	bg1 := loom.NewAstraBackground()
	bg2, err := catBackground()
	if err != nil {
		t.Fatal(err)
	}
	pane.Background = bg1
	w := &widget{pane: pane, frame: demoFrame(), themes: []string{"plain", "mc"}, bgs: []loom.Background{bg1, bg2}}
	if w.ConsumeKey(loom.KeyEvent{Key: "m"}).Quit || !pane.ReduceMotion {
		t.Fatal("m should toggle reduced motion")
	}
	if w.ConsumeKey(loom.KeyEvent{Key: "b"}).Quit || w.bgIdx != 1 || pane.Background != bg2 {
		t.Fatal("b should toggle the background")
	}
	if w.ConsumeKey(loom.KeyEvent{Key: "T"}).Quit || w.theme != 1 {
		t.Fatal("T should cycle the theme")
	}
	if !w.ConsumeKey(loom.KeyEvent{Key: "q"}).Quit {
		t.Fatal("background widget should delegate quit to its frame")
	}
}
