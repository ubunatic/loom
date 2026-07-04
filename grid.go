// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// Grid lays out widgets in a fixed number of columns.
// Row count is inferred from len(Children) and Cols.
// Navigation: Left/Right move within a row; Up/Down move between rows.
// OnSelect is called when Enter is pressed on a cell; if nil, Enter
// delegates to the focused child widget (which may return quit=true).
type Grid struct {
	Cols     int
	Children []Widget
	FocusBG  Color        // background color of the focused cell; zero = default
	OnSelect func(i int)  // called on Enter if non-nil; prevents quit propagation

	focus int // flat index of the focused child
}

// NewGrid creates a Grid with cols columns.
func NewGrid(cols int, children ...Widget) *Grid {
	if cols < 1 {
		cols = 1
	}
	return &Grid{Cols: cols, Children: children, FocusBG: ColorIndex(238)}
}

// Focus returns the index of the currently focused child.
func (g *Grid) Focus() int { return g.focus }

// Draw renders all children into a uniform grid within r.
// The focused cell receives a FocusBG background highlight before its child draws.
func (g *Grid) Draw(c *Canvas, r Rect) {
	n := len(g.Children)
	if n == 0 || g.Cols == 0 {
		return
	}
	rows := (n + g.Cols - 1) / g.Cols
	cellW := r.W / g.Cols
	cellH := r.H / rows
	if cellW < 1 {
		cellW = 1
	}
	if cellH < 1 {
		cellH = 1
	}
	for i, child := range g.Children {
		if f, ok := child.(Focusable); ok {
			f.SetFocus(i == g.focus)
		}
		cr := Rect{
			X: r.X + (i%g.Cols)*cellW,
			Y: r.Y + (i/g.Cols)*cellH,
			W: cellW,
			H: cellH,
		}
		if i == g.focus {
			c.Fill(cr, Cell{Text: " ", Style: Style{BG: g.FocusBG}})
		}
		child.Draw(c, cr)
		if Debug {
			drawDebugBorder(c, cr)
		}
	}
}

// HandleKey moves focus with arrow keys; Enter calls OnSelect or delegates to child.
func (g *Grid) HandleKey(e KeyEvent) (quit bool) {
	n := len(g.Children)
	if n == 0 {
		return false
	}
	switch e.Key {
	case "left":
		if g.focus > 0 {
			g.focus--
		} else {
			g.focus = n - 1
		}
		return false
	case "right":
		if g.focus < n-1 {
			g.focus++
		} else {
			g.focus = 0
		}
		return false
	case "up":
		if g.focus >= g.Cols {
			g.focus -= g.Cols
		} else {
			last := g.focus + ((n-1)/g.Cols)*g.Cols
			if last >= n {
				last -= g.Cols
			}
			g.focus = last
		}
		return false
	case "down":
		next := g.focus + g.Cols
		if next < n {
			g.focus = next
		} else {
			g.focus = g.focus % g.Cols
		}
		return false
	case "enter":
		if g.OnSelect != nil {
			g.OnSelect(g.focus)
			return false
		}
	}
	return g.Children[g.focus].HandleKey(e)
}

// HandleMouse forwards to the focused child.
func (g *Grid) HandleMouse(e MouseEvent) (quit bool) {
	if len(g.Children) == 0 {
		return false
	}
	return g.Children[g.focus].HandleMouse(e)
}

// ContentHeight estimates the required height for this grid layout.
func (g *Grid) ContentHeight() int {
	n := len(g.Children)
	if n == 0 || g.Cols == 0 {
		return 0
	}
	rows := (n + g.Cols - 1) / g.Cols
	maxH := 0
	for _, child := range g.Children {
		ch := 1
		if chWidget, ok := child.(ContentHeighter); ok {
			ch = chWidget.ContentHeight()
		}
		if ch > maxH {
			maxH = ch
		}
	}
	if maxH == 0 {
		maxH = 1
	}
	return rows * maxH
}
