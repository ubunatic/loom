// Command split demonstrates independent scrolling and keyboard focus in a Frame.
package main

import (
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom"
)

// scrollPane adds a focus cue to View without changing the library widget.
type scrollPane struct {
	*loom.View
	focused bool
}

func (p *scrollPane) Focused() bool         { return p.focused }
func (p *scrollPane) SetFocus(focused bool) { p.focused = focused }
func (p *scrollPane) Draw(c *loom.Canvas, r loom.Rect) {
	p.View.Draw(c, r)
	if r.W > 0 && r.H > 0 {
		marker := " "
		if p.focused {
			marker = "▶"
		}
		c.Write(r.X, r.Y, marker, loom.Style{})
	}
}

func lines(label string) []string {
	rows := make([]string, 40)
	for i := range rows {
		rows[i] = fmt.Sprintf("  %s line %02d", label, i+1)
	}
	return rows
}

func run() error {
	left := &scrollPane{View: loom.NewView(lines("left"))}
	right := &scrollPane{View: loom.NewView(lines("right"))}
	frame := &loom.Frame{
		Title: "Split panes", Status: "Tab / Shift-Tab: focus  •  arrows / PgUp / PgDn: scroll  •  q: quit",
		Gap: 1, Boxes: []loom.Box{
			{ID: "left", Title: "Left", Width: 34, Height: 16, Dynamic: true, MinWidth: 12, Child: left},
			{ID: "right", Title: "Right", Width: 34, Height: 16, Dynamic: true, MinWidth: 12, Child: right},
		},
		Actions: []loom.FrameAction{{ID: "quit", Action: "quit", Key: "q"}},
	}
	pane, err := loom.New(18)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	return pane.Run(frame)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
