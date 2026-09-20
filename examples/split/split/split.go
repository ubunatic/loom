// Package split demonstrates independent scrolling and keyboard focus in a loom.Split.
package split

import (
	"fmt"

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

// Run runs the split-panes example. It ignores args; it exists so the split
// example matches the Run(args []string) error signature shared by the
// other examples for loom-demo/loom-bench registration.
func Run(_ []string) error {
	left := &scrollPane{View: loom.NewView(lines("left"))}
	right := &scrollPane{View: loom.NewView(lines("right"))}
	split := loom.NewSplit(left, right)
	split.Ratio = 0.4
	split.MinFirst = 12
	split.MinSecond = 12
	split.Divider = loom.DividerStyle{Glyph: "│", Style: loom.Style{Dim: true}}
	frame := &loom.Frame{
		Title: "Split panes", Status: "Tab / Shift-Tab: focus  •  [ / ]: ratio  •  arrows / PgUp / PgDn: scroll  •  q: quit",
		Boxes:   []loom.Box{{ID: "split", Title: "40 / 60", Width: 69, Height: 16, Dynamic: true, MinWidth: 26, Child: split}},
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
