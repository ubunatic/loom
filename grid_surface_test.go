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
			gotFocusCellBG := c.Get(cr0.X+cr0.W-1, cr0.Y+cr0.H-1).Style.BG
			if gotFocusCellBG != focusBG {
				t.Errorf("focused cell (%s) background = %v, want FocusBG %v", w.name, gotFocusCellBG, focusBG)
			}

			// In cell 1 (unfocused), empty/unselected/background cells should take ambientBG.
			gotUnfocusCellBG := c.Get(cr1.X+cr1.W-1, cr1.Y+cr1.H-1).Style.BG
			if gotUnfocusCellBG != ambientBG {
				t.Errorf("unfocused cell (%s) background = %v, want ambient %v", w.name, gotUnfocusCellBG, ambientBG)
			}

			// Detailed check across all non-selection cells:
			// For KeyHelp, Viewport, Tree unselected rows, etc., unselected cells must carry focusBG in cr0 and ambientBG in cr1.
			switch w.name {
			case "KeyHelp":
				for x := cr0.X; x < cr0.X+cr0.W; x++ {
					if bg := c.Get(x, cr0.Y).Style.BG; bg != focusBG {
						t.Errorf("KeyHelp focused cell at (%d,%d) BG = %v, want FocusBG %v", x, cr0.Y, bg, focusBG)
					}
					if bg := c.Get(x+cr1.X-cr0.X, cr1.Y).Style.BG; bg != ambientBG {
						t.Errorf("KeyHelp unfocused cell at (%d,%d) BG = %v, want ambientBG %v", x+cr1.X-cr0.X, cr1.Y, bg, ambientBG)
					}
				}
			case "Viewport":
				for y := cr0.Y; y < cr0.Y+cr0.H; y++ {
					for x := cr0.X; x < cr0.X+cr0.W; x++ {
						if bg := c.Get(x, y).Style.BG; bg != focusBG {
							t.Errorf("Viewport focused cell at (%d,%d) BG = %v, want FocusBG %v", x, y, bg, focusBG)
						}
					}
				}
				for y := cr1.Y; y < cr1.Y+cr1.H; y++ {
					for x := cr1.X; x < cr1.X+cr1.W; x++ {
						if bg := c.Get(x, y).Style.BG; bg != ambientBG {
							t.Errorf("Viewport unfocused cell at (%d,%d) BG = %v, want ambientBG %v", x, y, bg, ambientBG)
						}
					}
				}
			case "Tree":
				// Row 0 is selected in focused tree (if focused) or row 0 is unselected in unfocused tree.
				// In unfocused tree (cr1), all rows are unselected and must have ambientBG.
				for y := cr1.Y; y < cr1.Y+cr1.H; y++ {
					for x := cr1.X; x < cr1.X+cr1.W; x++ {
						if bg := c.Get(x, y).Style.BG; bg != ambientBG {
							t.Errorf("Tree unfocused cell at (%d,%d) BG = %v, want ambientBG %v", x, y, bg, ambientBG)
						}
					}
				}
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

// TestGridCellNormalRowsInheritCellBG covers issue 243 M3: ProgressBar fill,
// and unselected Choice and Table rows, carry the Grid cell BG.
func TestGridCellNormalRowsInheritCellBG(t *testing.T) {
	theme := loom.Theme("mc-dark")
	ambient := loom.ColorIndex(42)
	cases := []struct {
		name      string
		make      func() loom.Widget
		firstRow  int // first row that is not selected or highlighted
		skipFirst bool
	}{
		{"ProgressBar", func() loom.Widget {
			b := loom.NewProgressBar()
			b.Options.Width = 10
			b.ShowPercent = true
			b.Set(50)
			return b
		}, 0, false},
		{"Choice", func() loom.Widget {
			return loom.NewChoice([]loom.Item{{Name: "One"}, {Name: "Two"}, {Name: "Three"}})
		}, 1, true},
		{"Table", func() loom.Widget {
			return loom.NewTable([]loom.Column{{Header: "Col", Width: 8}},
				[]loom.Row{{Cells: []string{"A"}}, {Cells: []string{"B"}}, {Cells: []string{"C"}}})
		}, 2, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, b := tc.make(), tc.make()
			a.(loom.Themeable).ApplyTheme(theme)
			b.(loom.Themeable).ApplyTheme(theme)
			grid := loom.NewGrid(2, a, b)
			grid.ApplyTheme(theme)
			c := loom.NewCanvas(40, 6)
			c.PaintSurface(c.Bounds(), loom.Style{BG: ambient})
			grid.Draw(c, c.Bounds())
			for i, want := range []loom.Color{grid.FocusBG, ambient} {
				cr := grid.ChildRect(i)
				for y := cr.Y + tc.firstRow; y < cr.Y+cr.H; y++ {
					if tc.skipFirst && y == cr.Y+cr.H-1 {
						continue // prompt row has its own style
					}
					for x := cr.X; x < cr.X+cr.W; x++ {
						if bg := c.Get(x, y).Style.BG; bg != want {
							t.Fatalf("cell %d (%d,%d) BG = %v, want %v", i, x, y, bg, want)
						}
					}
				}
			}
		})
	}
}
