// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"
	"time"
	"unicode/utf8"

	"ubunatic.com/loom"
)

func TestGridCellBackgroundInheritance(t *testing.T) {
	theme := loom.Theme("mc-dark")

	mkBar := func() loom.Widget {
		b := loom.NewProgressBar()
		b.Options.Width = 10
		b.ShowPercent = true
		b.Set(50)
		return b
	}
	mkChoice := func() loom.Widget {
		return loom.NewChoice([]loom.Item{{Name: "Item 1"}, {Name: "Item 2"}})
	}
	mkKeyHelp := func() loom.Widget {
		return loom.NewKeyHelp(loom.NewKeyMapWithLabels(map[string][]string{"q": {"q"}}, map[string]string{"q": "Quit"}))
	}
	mkTable := func() loom.Widget {
		return loom.NewTable([]loom.Column{{Header: "Col", Width: 8}}, []loom.Row{{Cells: []string{"Val"}}})
	}
	mkTree := func() loom.Widget {
		return loom.NewTree([]*loom.TreeNode{{Label: "Node"}})
	}
	mkViewport := func() loom.Widget {
		return loom.NewViewport(loom.NewView([]string{"Line 1", "Line 2"}))
	}

	widgets := []struct {
		name string
		make func() loom.Widget
	}{
		{"ProgressBar", mkBar},
		{"Choice", mkChoice},
		{"KeyHelp", mkKeyHelp},
		{"Table", mkTable},
		{"Tree", mkTree},
		{"Viewport", mkViewport},
	}

	for _, w := range widgets {
		t.Run(w.name, func(t *testing.T) {
			childFocused := w.make()
			childUnfocused := w.make()
			if th, ok := childFocused.(loom.Themeable); ok {
				th.ApplyTheme(theme)
			}
			if th, ok := childUnfocused.(loom.Themeable); ok {
				th.ApplyTheme(theme)
			}

			// Grid with 2 columns: cell 0 (focused) and cell 1 (unfocused).
			grid := loom.NewGrid(2, childFocused, childUnfocused)
			grid.ApplyTheme(theme)

			focusBG := grid.FocusBG
			ambientBG := loom.ColorIndex(42) // ambient surface under grid

			c := loom.NewCanvas(40, 10)
			c.PaintSurface(c.Bounds(), loom.Style{BG: ambientBG})
			grid.Draw(c, c.Bounds())

			cr0 := grid.ChildRect(0)
			cr1 := grid.ChildRect(1)

			// In cell 0 (focused), empty/unselected/background cells should take focusBG.
			// Specifically, the corner or padding cells of the widget should have focusBG.
			gotFocusCellBG := c.Get(cr0.X+cr0.W-1, cr0.Y+cr0.H-1).Style.BG
			if gotFocusCellBG != focusBG {
				t.Errorf("focused cell (%s) background = %v, want FocusBG %v", w.name, gotFocusCellBG, focusBG)
			}

			// In cell 1 (unfocused), empty/unselected/background cells should take ambientBG.
			gotUnfocusCellBG := c.Get(cr1.X+cr1.W-1, cr1.Y+cr1.H-1).Style.BG
			if gotUnfocusCellBG != ambientBG {
				t.Errorf("unfocused cell (%s) background = %v, want ambient %v", w.name, gotUnfocusCellBG, ambientBG)
			}
		})
	}
}

func TestStandaloneWidgetsRetainThemeNormalBG(t *testing.T) {
	theme := loom.Theme("mc-dark")

	widgets := []struct {
		name   string
		widget loom.Widget
	}{
		{"KeyHelp", loom.NewKeyHelp(loom.NewKeyMapWithLabels(map[string][]string{"q": {"q"}}, map[string]string{"q": "Quit"}))},
		{"Viewport", loom.NewViewport(loom.NewView([]string{"Line 1", "Line 2"}))},
		{"Choice", loom.NewChoice([]loom.Item{{Name: "Item 1"}})},
		{"Table", loom.NewTable([]loom.Column{{Header: "Col", Width: 8}}, []loom.Row{{Cells: []string{"Val"}}})},
		{"Tree", loom.NewTree([]*loom.TreeNode{{Label: "Node"}})},
	}

	for _, w := range widgets {
		t.Run(w.name, func(t *testing.T) {
			if th, ok := w.widget.(loom.Themeable); ok {
				th.ApplyTheme(theme)
			}
			c := loom.NewCanvas(20, 5)
			r := loom.Rect{X: 1, Y: 1, W: 18, H: 3}
			w.widget.Draw(c, r)

			// Standalone on blank canvas (BG == ColorReset), the widget's default surface must show theme.NormalBG
			gotBG := c.Get(r.X+r.W-1, r.Y+r.H-1).Style.BG
			if gotBG != theme.NormalBG.Color() {
				t.Errorf("standalone %s background = %v, want theme.NormalBG %v", w.name, gotBG, theme.NormalBG.Color())
			}
			if c.Get(0, 0).Style.BG != loom.ColorReset() {
				t.Errorf("standalone %s surface escaped widget bounds", w.name)
			}
		})
	}
}

func TestGridFocusedCellKeepsAstraDecoration(t *testing.T) {
	theme := loom.Theme("mc-dark")
	child := loom.NewChoice([]loom.Item{{Name: "One"}})
	child.ApplyTheme(theme)

	grid := loom.NewGrid(1, child)
	grid.ApplyTheme(theme)

	c := loom.NewCanvas(30, 10)
	grid.Draw(c, c.Bounds())

	bg := loom.NewAstraBackground()
	c.ComposeBackground(bg, c.Bounds(), time.Now())

	cr := grid.ChildRect(0)
	foundStar := false
	for y := cr.Y; y < cr.Y+cr.H; y++ {
		for x := cr.X; x < cr.X+cr.W; x++ {
			cell := c.Get(x, y)
			// Astra paints Braille Pattern glyphs (U+2800–U+28FF) into eligible
			// (unclaimed) cells. Only check those — prompt characters and other
			// foreground content are not astra stars.
			r, _ := utf8.DecodeRuneInString(cell.Text)
			if r < 0x2800 || r > 0x28FF {
				continue
			}
			foundStar = true
			if cell.Style.BG != grid.FocusBG {
				t.Errorf("star cell (%d,%d) has BG %v, want FocusBG %v", x, y, cell.Style.BG, grid.FocusBG)
			}
		}
	}
	if !foundStar {
		t.Errorf("no astra decoration rendered inside focused grid cell")
	}
}
