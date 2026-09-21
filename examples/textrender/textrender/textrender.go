// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package textrender demonstrates non-ASCII text rendering across loom widgets.
package textrender

import (
	"fmt"

	"codeberg.org/ubunatic/loom"
	"github.com/spf13/cobra"
)

// TestCase represents a labeled text sample with expected display width
type TestCase struct {
	Label     string
	Text      string
	WantWidth int
}

// Cases defines representative test samples covering various Unicode categories
var Cases = []TestCase{
	{Label: "ASCII", Text: "Hello", WantWidth: 5},
	{Label: "Accented Latin", Text: "Café", WantWidth: 4},
	{Label: "German umlauts (precomposed)", Text: "Müller", WantWidth: 6},
	{Label: "German umlauts (decomposed)", Text: "Müller", WantWidth: 6},
	{Label: "CJK", Text: "中文", WantWidth: 4},
	{Label: "Combining marks", Text: "e̊", WantWidth: 1},
	{Label: "Symbols", Text: "♠♣♥♦", WantWidth: 4},
	{Label: "Emoji", Text: "😀", WantWidth: 2},
	{Label: "ZWJ sequence", Text: "👨‍👩‍👧", WantWidth: 6},
	{Label: "Flag", Text: "🇩🇪", WantWidth: 4},
	{Label: "Mixed line", Text: "Test中文♠😀", WantWidth: 11},
}

type textRenderApp struct {
	tabs *loom.Tabs
}

func newTextRenderApp() *textRenderApp {
	borders := newBordersView()

	root := loom.NewTabs(
		loom.Tab{Title: "Borders", Widget: borders},
		loom.Tab{Title: "Buttons", Widget: loom.NewView([]string{"Not yet implemented"})},
		loom.Tab{Title: "Clipping", Widget: loom.NewView([]string{"Not yet implemented"})},
		loom.Tab{Title: "Scroll", Widget: loom.NewView([]string{"Not yet implemented"})},
	)

	return &textRenderApp{tabs: root}
}

// newBordersView creates a view that displays test cases with labels and widths
func newBordersView() loom.Widget {
	lines := []string{
		"Text samples with display widths:",
		"",
	}

	for _, tc := range Cases {
		lines = append(lines, fmt.Sprintf("  %s: %q (width: %d)", tc.Label, tc.Text, tc.WantWidth))
	}

	return loom.NewView(lines)
}

func (a *textRenderApp) Draw(c *loom.Canvas, r loom.Rect) {
	if r.H <= 0 || r.W <= 0 {
		return
	}
	footerH := 1
	tabsH := max(0, r.H-footerH)
	if tabsH > 0 {
		a.tabs.Draw(c, loom.Rect{X: r.X, Y: r.Y, W: r.W, H: tabsH})
	}
	footerY := r.Y + tabsH
	c.PaintSurface(loom.Rect{X: r.X, Y: footerY, W: r.W, H: 1}, loom.Style{Dim: true})
	legend := " ←/→: cycle tabs  •  q: quit"
	c.Write(r.X, footerY, legend, loom.Style{Dim: true})
}

func (a *textRenderApp) HandleKey(e loom.KeyEvent) bool {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	switch key {
	case "q", "ctrl-q", "ctrl-c", "esc":
		return true
	}
	return a.tabs.HandleKey(e)
}

func (a *textRenderApp) HandleMouse(e loom.MouseEvent) bool {
	return a.tabs.HandleMouse(e)
}

// PaneRequest declares the terminal requirements of the textrender widget.
func (a *textRenderApp) PaneRequest() loom.PaneRequest {
	return loom.PaneRequest{
		Mouse:      1000,
		Resizeable: true,
	}
}

// NewWidget builds the textrender example's root widget from command-line args,
// without creating or running a Pane.
func NewWidget(args []string) (loom.Widget, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("textrender: unexpected arguments: %v", args)
	}
	return newTextRenderApp(), nil
}

// Run runs the textrender example with cobra command support.
func Run(args []string) error {
	cmd := &cobra.Command{
		Use:           "textrender",
		Short:         "Demonstrate non-ASCII text rendering in loom widgets",
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
	app := newTextRenderApp()
	pane, err := loom.New(18)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	pane.EnableMouse()
	return pane.Run(app)
}
