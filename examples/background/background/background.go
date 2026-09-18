// Package background demonstrates the full-screen Astra star field.
package background

import (
	"fmt"

	"codeberg.org/ubunatic/loom"
)

type widget struct{ metrics *loom.RenderMetrics }

func (w widget) Draw(c *loom.Canvas, r loom.Rect) {
	c.Write(2, 1, "Loom Astra background", loom.Style{Bold: true, FG: loom.ColorRGB(240, 240, 240)})
	c.Write(2, 3, "Deterministic stars fade in and out behind the foreground.", loom.Style{FG: loom.ColorRGB(220, 220, 220)})
	c.Write(2, 5, fmt.Sprintf("Loom: %.1f FPS   Astra: %.1f FPS (target %.1f)   redraw: %s", w.metrics.LoomFPS, w.metrics.AstraFPS, w.metrics.AstraTargetFPS, w.metrics.RedrawTime), loom.Style{FG: loom.ColorRGB(180, 210, 180)})
}

func (widget) HandleKey(loom.KeyEvent) bool { return false }

func (widget) HandleMouse(loom.MouseEvent) bool { return false }

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
	return pane.Run(widget{metrics: metrics})
}
