// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package textrender demonstrates non-ASCII text rendering across loom widgets.
package textrender

import (
	"fmt"
	"strings"

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
	{Label: "German umlauts (decomposed)", Text: "Mu\u0308ller", WantWidth: 6},
	{Label: "CJK", Text: "中文", WantWidth: 4},
	{Label: "Combining marks", Text: "e\u030a", WantWidth: 1},
	{Label: "Symbols", Text: "♠♣♥♦", WantWidth: 4},
	{Label: "Emoji", Text: "😀", WantWidth: 2},
	{Label: "ZWJ sequence", Text: "👨\u200d👩\u200d👧", WantWidth: 2},
	{Label: "Flag", Text: "🇩🇪", WantWidth: 2},
	{Label: "Mixed line", Text: "Test中文♠😀", WantWidth: 11},
}

// knownDivergences records library behavior that differs from modern terminal
// cluster widths; WantWidth remains the terminal expectation.
var knownDivergences = map[string]int{
	"ZWJ sequence": 6,
	"Flag":         4,
}

type textRenderApp struct {
	tabs *loom.Tabs
}

type staticView struct{ draw func(*loom.Canvas, loom.Rect) }

func (v staticView) Draw(c *loom.Canvas, r loom.Rect) { v.draw(c, r) }
func (v staticView) HandleKey(loom.KeyEvent) bool     { return false }
func (v staticView) HandleMouse(loom.MouseEvent) bool { return false }

func codepoints(s string) string {
	parts := make([]string, 0, len([]rune(s)))
	for _, r := range s {
		parts = append(parts, fmt.Sprintf("U+%04X", r))
	}
	return strings.Join(parts, " ")
}

func sampleLabel(tc TestCase) string {
	return fmt.Sprintf("%s: %s [%s]", tc.Label, tc.Text, codepoints(tc.Text))
}

func newTextRenderApp() *textRenderApp {
	borders := newBordersView()
	buttons := newButtonsView()
	clipping := newClippingView()
	scroll := newScrollView()

	root := loom.NewTabs(
		loom.Tab{Title: "Borders", Widget: borders},
		loom.Tab{Title: "Buttons", Widget: buttons},
		loom.Tab{Title: "Clipping", Widget: clipping},
		loom.Tab{Title: "Scroll", Widget: scroll},
	)

	return &textRenderApp{tabs: root}
}

// newBordersView creates real titled boxes with samples as titles and bodies.
func newBordersView() loom.Widget {
	return staticView{draw: func(c *loom.Canvas, r loom.Rect) {
		c.Write(r.X, r.Y, "DrawBox titles and bodies (narrow boxes truncate titles)", loom.Style{Bold: true})
		for i, tc := range Cases {
			if r.Y+1+i*2+1 >= r.Y+r.H {
				break
			}
			y := r.Y + 1 + i*2
			w := r.W - 2
			if i%2 == 1 {
				w = 12
			}
			if w < 4 {
				w = 4
			}
			c.DrawBox(loom.Rect{X: r.X, Y: y, W: w, H: 2}, loom.BoxBorderStyleSharp, tc.Text, loom.Reset)
			c.Write(r.X+1, y+1, loom.TruncateText(sampleLabel(tc), w-2, ""), loom.Reset)
		}
	}}
}

type activatableView struct {
	choice *loom.Choice
	status string
}

func (v *activatableView) Draw(c *loom.Canvas, r loom.Rect) {
	c.Write(r.X, r.Y, "Activatable samples (Enter activates selection)", loom.Style{Bold: true})
	v.choice.Draw(c, loom.Rect{X: r.X, Y: r.Y + 1, W: r.W, H: r.H - 2})
	c.Write(r.X, r.Y+r.H-1, loom.TruncateText(v.status, r.W, ""), loom.Style{Dim: true})
}
func (v *activatableView) HandleKey(e loom.KeyEvent) bool     { return v.choice.HandleKey(e) }
func (v *activatableView) HandleMouse(e loom.MouseEvent) bool { return v.choice.HandleMouse(e) }

// newButtonsView uses Choice as the closest existing activatable widget.
func newButtonsView() loom.Widget {
	items := make([]loom.Item, len(Cases))
	for i, tc := range Cases {
		items[i] = loom.Item{Name: tc.Text, Desc: fmt.Sprintf("%s [%s]", tc.Label, codepoints(tc.Text))}
	}
	v := &activatableView{}
	v.choice = loom.NewChoice(items)
	v.choice.OnSelect = func(item loom.Item) { v.status = "Activated: " + item.Name }
	return v
}

// newClippingView renders each sample into real cell-limited areas.
func newClippingView() loom.Widget {
	return staticView{draw: func(c *loom.Canvas, r loom.Rect) {
		c.Write(r.X, r.Y, "Cluster-safe clipping: widths 8 / 6 / 4 / 2", loom.Style{Bold: true})
		for i, tc := range Cases {
			y := r.Y + 1 + i
			if y >= r.Y+r.H {
				break
			}
			line := fmt.Sprintf("%-20s %s", tc.Label, loom.TruncateText(tc.Text, 8, "…"))
			for _, w := range []int{6, 4, 2} {
				line += " | " + loom.TruncateText(tc.Text, w, "…")
			}
			c.Write(r.X, y, loom.TruncateText(line, r.W, ""), loom.Reset)
		}
	}}
}

// newScrollView creates a scrollable list of test cases
func newScrollView() loom.Widget {
	items := make([]loom.Item, len(Cases))
	for i, tc := range Cases {
		items[i] = loom.Item{Name: tc.Label, Desc: fmt.Sprintf("%s [%s]", tc.Text, codepoints(tc.Text))}
	}
	return loom.NewChoice(items)
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
