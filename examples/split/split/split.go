// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package split demonstrates nested split layout, divider dragging,
// and keyboard focus traversal in loom.Split.
package split

import (
	"fmt"
	"math"

	"codeberg.org/ubunatic/loom"
	"github.com/spf13/cobra"
)

func lines(label string) []string {
	rows := make([]string, 40)
	for i := range rows {
		rows[i] = fmt.Sprintf("%s line %02d", label, i+1)
	}
	return rows
}

type splitApp struct {
	frame  *loom.Frame
	hSplit *loom.Split
	vSplit *loom.Split
}

func newSplitApp() *splitApp {
	left := loom.NewView(lines("Left"))
	left.Style = loom.Style{Dim: true}
	left.FocusStyle = loom.Style{Bold: true}

	topRight := loom.NewView(lines("Top-Right"))
	topRight.Style = loom.Style{Dim: true}
	topRight.FocusStyle = loom.Style{Bold: true}

	bottomRight := loom.NewView(lines("Bottom-Right"))
	bottomRight.Style = loom.Style{Dim: true}
	bottomRight.FocusStyle = loom.Style{Bold: true}

	vSplit := loom.NewSplit(topRight, bottomRight)
	vSplit.Orientation = loom.Vertical
	vSplit.Ratio = 0.5
	vSplit.MinFirst = 3
	vSplit.MinSecond = 3
	vSplit.Divider = loom.DividerStyle{Glyph: "─", Style: loom.Style{Dim: true}}

	hSplit := loom.NewSplit(left, vSplit)
	hSplit.Orientation = loom.Horizontal
	hSplit.Ratio = 0.4
	hSplit.MinFirst = 12
	hSplit.MinSecond = 16
	hSplit.Divider = loom.DividerStyle{Glyph: "│", Style: loom.Style{Dim: true}}

	frame := &loom.Frame{
		Title: "Split panes",
		Boxes: []loom.Box{
			{ID: "split", Title: "Nested Split", Width: 69, Height: 16, Dynamic: true, MinWidth: 26, Child: hSplit},
		},
		Actions: []loom.FrameAction{
			{ID: "quit", Action: "quit", Key: "q"},
		},
	}

	return &splitApp{
		frame:  frame,
		hSplit: hSplit,
		vSplit: vSplit,
	}
}

func (a *splitApp) Draw(c *loom.Canvas, r loom.Rect) {
	hRatio := int(math.Round(a.hSplit.Ratio * 100))
	vRatio := int(math.Round(a.vSplit.Ratio * 100))
	a.frame.Boxes[0].Title = fmt.Sprintf("Nested Split (H: %02d/%02d • V: %02d/%02d)", hRatio, 100-hRatio, vRatio, 100-vRatio)
	a.frame.Status = fmt.Sprintf("Ratio: H %d%% V %d%%  •  Tab: focus  •  [ / ]: ratio  •  drag divider  •  q: quit", hRatio, vRatio)
	a.frame.Draw(c, r)
}

func (a *splitApp) HandleKey(e loom.KeyEvent) bool {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	if key == "/" {
		if a.hSplit.Ratio < 0.5 {
			a.hSplit.SetRatio(0.5)
		} else {
			a.hSplit.SetRatio(0.4)
		}
		return false
	}
	return a.frame.HandleKey(e)
}

func (a *splitApp) HandleMouse(e loom.MouseEvent) bool {
	return a.frame.HandleMouse(e)
}

// PaneRequest declares the terminal requirements of the split widget.
func (a *splitApp) PaneRequest() loom.PaneRequest {
	return loom.PaneRequest{
		Mouse:      1003, // SGR mouse tracking with motion events
		Resizeable: true,
		OwnsQuit:   true,
	}
}

// NewWidget builds the split example's root widget from command-line args,
// without creating or running a Pane.
func NewWidget(args []string) (loom.Widget, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("split: unexpected arguments: %v", args)
	}
	return newSplitApp(), nil
}

// Run runs the split example with cobra command support.
func Run(args []string) error {
	cmd := &cobra.Command{
		Use:           "split",
		Short:         "Demonstrate nested split layout, divider dragging, and focus traversal",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run()
		},
	}
	cmd.SetArgs(args)
	return cmd.Execute()
}

func run() error {
	app := newSplitApp()
	pane, err := loom.New(18)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	pane.DisableDefaultQuit = true
	pane.EnableMouse()
	return pane.Run(app)
}
