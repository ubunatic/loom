// Package background demonstrates the full-screen Astra star field.
package background

import (
	"fmt"

	"codeberg.org/ubunatic/loom"
)

type widget struct {
	metrics *loom.RenderMetrics
	frame   *loom.Frame
}

func (w widget) Draw(c *loom.Canvas, r loom.Rect) {
	c.Write(2, 1, "Loom Astra background", loom.Style{Bold: true, FG: loom.ColorRGB(240, 240, 240)})
	c.Write(2, 3, "Deterministic stars fade in and out behind the foreground.", loom.Style{FG: loom.ColorRGB(220, 220, 220)})
	c.Write(2, 5, fmt.Sprintf("Loom: %.1f FPS   Astra: %.1f FPS (target %.1f)   redraw: %s", w.metrics.LoomFPS, w.metrics.AstraFPS, w.metrics.AstraTargetFPS, w.metrics.RedrawTime), loom.Style{FG: loom.ColorRGB(180, 210, 180)})
	if w.frame != nil && r.H > 7 {
		w.frame.Draw(c, loom.Rect{X: 1, Y: 7, W: r.W - 2, H: r.H - 7})
	}
}

func (w widget) HandleKey(k loom.KeyEvent) bool {
	return w.frame != nil && w.frame.HandleKey(k)
}

func (w widget) HandleMouse(k loom.MouseEvent) bool {
	return w.frame != nil && w.frame.HandleMouse(k)
}

func demoFrame() *loom.Frame {
	left := loom.NewView([]string{
		"▸ projects/",
		"  documents/",
		"  downloads/",
		"  .gitignore                 1.2 KiB",
		"  AGENTS.md                  4.8 KiB",
		"  background.go              5.1 KiB",
		"  canvas.go                  7.6 KiB",
		"  frame.go                   18 KiB",
		"  pane.go                    23 KiB",
		"  README.md                  3.4 KiB",
		"  spec/",
	})
	right := loom.NewView([]string{
		"Selected: frame.go",
		"",
		"Type:       Go source",
		"Size:       18.0 KiB",
		"Modified:   just now",
		"Mode:       -rw-r--r--",
		"",
		"The star field is composed after the foreground",
		"and skips claimed cells, borders, and the cursor.",
		"",
		"This pane intentionally leaves open surface so",
		"the animation remains visible during interaction.",
	})
	return &loom.Frame{
		Title: "Composition test area", Status: "Tab: focus  •  arrows: scroll  •  q: quit",
		Gap: 1, Breakpoint: 70,
		Boxes: []loom.Box{
			{ID: "left", Title: "Left pane", Dynamic: true, MinWidth: 18, Height: 12, Child: left},
			{ID: "right", Title: "Right pane", Dynamic: true, MinWidth: 24, Height: 12, Child: right},
		},
		Actions: []loom.FrameAction{{ID: "quit", Action: "quit", Key: "q"}},
	}
}

// Run starts the interactive image-background example.
func Run(_ []string) error {
	// A deliberately oversized request makes Pane reserve all available
	// terminal rows; New clamps it to the current screen height.
	pane, err := loom.New(1000)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.MaxCols = 0 // use the full terminal width for the image background
	pane.Background = loom.NewAstraBackground()
	metrics := &loom.RenderMetrics{}
	pane.Metrics = metrics
	return pane.Run(widget{metrics: metrics, frame: demoFrame()})
}
