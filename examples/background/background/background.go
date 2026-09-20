// Package background demonstrates the full-screen Astra star field.
package background

import (
	"fmt"

	"codeberg.org/ubunatic/loom"
)

type widget struct {
	metrics *loom.RenderMetrics
	frame   *loom.Frame
	pane    *loom.Pane
	themes  []string
	theme   int
}

func (w widget) Draw(c *loom.Canvas, r loom.Rect) {
	c.Write(2, 1, "Loom Astra background", loom.Style{Bold: true, FG: loom.ColorRGB(240, 240, 240)})
	c.Write(2, 3, "Deterministic stars fade in and out behind the foreground.", loom.Style{FG: loom.ColorRGB(220, 220, 220)})
	c.Write(2, 5, fmt.Sprintf("Loom: %.1f FPS   Astra: %.1f FPS (target %.1f)   redraw: %s", w.metrics.LoomFPS, w.metrics.AstraFPS, w.metrics.AstraTargetFPS, w.metrics.RedrawTime), loom.Style{FG: loom.ColorRGB(180, 210, 180)})
	if w.frame != nil && r.H > 8 {
		// Keep one quiet row below the frame status line so it does not sit
		// against the terminal's bottom edge.
		w.frame.Draw(c, loom.Rect{X: 1, Y: 7, W: r.W - 2, H: r.H - 8})
	}
}

func (w *widget) HandleKey(k loom.KeyEvent) bool {
	key := k.Key
	if key == "" {
		key = k.Text
	}
	switch key {
	case "m", "M":
		if w.pane != nil {
			w.pane.ReduceMotion = !w.pane.ReduceMotion
		}
		return false
	case "t", "T":
		if len(w.themes) > 0 {
			w.theme = (w.theme + 1) % len(w.themes)
			applyTheme(w.frame, loom.Theme(w.themes[w.theme]))
		}
		return false
	}
	return w.frame != nil && w.frame.HandleKey(k)
}

func (w *widget) HandleMouse(k loom.MouseEvent) bool {
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
		"and passes through transparent blank cells only.",
		"Text, borders, colored surfaces, and the cursor stay protected.",
		"",
		"This pane intentionally leaves open surface so",
		"the animation remains visible during interaction.",
	})
	panelTheme := loom.SpeccedThemes["mc-dark"]
	panelStyle := loom.Style{FG: panelTheme.NormalFG.Color(), BG: panelTheme.NormalBG.Color()}
	left.Style = panelStyle
	right.Style = panelStyle
	left.FocusStyle = loom.Style{FG: panelTheme.SelectedFG.Color(), BG: panelTheme.SelectedBG.Color()}
	right.FocusStyle = left.FocusStyle
	split := loom.NewSplit(left, right)
	split.Ratio = 0.45
	split.MinFirst, split.MinSecond = 18, 24
	return &loom.Frame{
		Title: "Composition test area", Status: "Tab: focus  •  arrows: scroll  •  q: quit",
		Gap: 1, Breakpoint: 70,
		Boxes:   []loom.Box{{ID: "panels", Title: "Interactive split (drag divider)", Dynamic: true, FillHeight: true, MinWidth: 42, Height: 12, Child: split}},
		Actions: []loom.FrameAction{{ID: "quit", Action: "quit", Key: "q"}},
	}
}

func applyTheme(frame *loom.Frame, theme loom.ThemeColors) {
	frame.Style = theme.FrameStyle()
	for i := range frame.Boxes {
		frame.Boxes[i].Style = theme.BoxStyle()
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
	themes := make([]string, 0, len(loom.SpeccedThemes))
	for name := range loom.SpeccedThemes {
		themes = append(themes, name)
	}
	return pane.Run(&widget{metrics: metrics, frame: demoFrame(), pane: pane, themes: themes})
}
