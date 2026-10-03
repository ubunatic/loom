// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "time"

// Grid lays out widgets in a fixed number of columns.
// Row count is inferred from len(Children) and Cols.
// Navigation: Left/Right move within a row; Up/Down move between rows.
// OnSelect is called when Enter is pressed on a cell; if nil, Enter
// delegates to the focused child widget (which may return quit=true).
type Grid struct {
	Cols     int
	Children []Widget
	FocusBG  Color       // background color of the focused cell; zero = default
	OnSelect func(i int) // called on Enter if non-nil; prevents quit propagation

	focus      int // flat index of the focused child
	childRects []Rect
	lastRect   Rect
}

// NewGrid creates a Grid with cols columns.
func NewGrid(cols int, children ...Widget) *Grid {
	if cols < 1 {
		cols = 1
	}
	return &Grid{Cols: cols, Children: children, FocusBG: Theme("plain").FocusBGColor()}
}

func (g *Grid) TickInterval() (shortest time.Duration) {
	for _, child := range g.Children {
		if t, ok := child.(Ticker); ok && t.TickInterval() > 0 && (shortest == 0 || t.TickInterval() < shortest) {
			shortest = t.TickInterval()
		}
	}
	return shortest
}

func (g *Grid) Tick(now time.Time) {
	for _, child := range g.Children {
		if t, ok := child.(Ticker); ok && t.TickInterval() > 0 {
			t.Tick(now)
		}
	}
}

// PaneRequest merges the terminal requirements of all children.
func (g *Grid) PaneRequest() (request PaneRequest) {
	first := true
	for _, child := range g.Children {
		if requester, ok := child.(PaneRequester); ok {
			childRequest := requester.PaneRequest()
			if first {
				request, first = childRequest, false
				continue
			}
			mergePaneRequest(&request, childRequest)
		}
	}
	return request
}

// Focus returns the index of the currently focused child.
func (g *Grid) Focus() int { return g.focus }

// ChildRect returns the drawn bounds of child i from the most recent Draw.
func (g *Grid) ChildRect(i int) Rect {
	if i < 0 || i >= len(g.childRects) {
		return Rect{}
	}
	return g.childRects[i]
}

func (g *Grid) setFocus(index int) {
	n := len(g.Children)
	if n == 0 {
		g.focus = 0
		return
	}
	if index < 0 {
		index = 0
	} else if index >= n {
		index = n - 1
	}
	if g.focus == index {
		return
	}
	old := g.focus
	g.focus = index
	if old >= 0 && old < n {
		if f, ok := g.Children[old].(Focusable); ok {
			f.SetFocus(false)
		}
	}
	if index >= 0 && index < n {
		if f, ok := g.Children[index].(Focusable); ok {
			f.SetFocus(true)
		}
	}
}

// Draw renders all children into a uniform grid within r.
// The focused cell receives a FocusBG background highlight before its child draws.
func (g *Grid) Draw(c *Canvas, r Rect) {
	g.lastRect = r
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
		if len(g.childRects) != n {
			g.childRects = make([]Rect, n)
		}
		g.childRects[i] = cr
		if i == g.focus {
			c.PaintSurface(cr, Style{BG: g.FocusBG})
		}
		child.Draw(c, cr)
		if Debug {
			drawDebugBorder(c, cr)
		}
	}
}

// ConsumeKey moves focus with arrow keys; Enter calls OnSelect or delegates to child.
func (g *Grid) ConsumeKey(e KeyEvent) EventResult {
	n := len(g.Children)
	if n == 0 {
		return Ignored()
	}
	switch e.Key {
	case "shift-left":
		e.Key = "left"
	case "shift-right":
		e.Key = "right"
	case "shift-up":
		e.Key = "up"
	case "shift-down":
		e.Key = "down"
	default:
		if g.focus >= 0 && g.focus < n {
			if res := g.Children[g.focus].ConsumeKey(e); res.Consumed {
				return res
			}
		}
	}
	switch e.Key {
	case "left":
		if g.focus > 0 {
			g.setFocus(g.focus - 1)
		} else {
			g.setFocus(n - 1)
		}
		return Handled()
	case "right":
		if g.focus < n-1 {
			g.setFocus(g.focus + 1)
		} else {
			g.setFocus(0)
		}
		return Handled()
	case "up":
		if g.focus >= g.Cols {
			g.setFocus(g.focus - g.Cols)
		} else {
			last := g.focus + ((n-1)/g.Cols)*g.Cols
			if last >= n {
				last -= g.Cols
			}
			g.setFocus(last)
		}
		return Handled()
	case "down":
		next := g.focus + g.Cols
		if next < n {
			g.setFocus(next)
		} else {
			g.setFocus(g.focus % g.Cols)
		}
		return Handled()
	case "enter":
		if g.OnSelect != nil {
			g.OnSelect(g.focus)
			return Handled()
		}
	}
	return Ignored()
}

func (g *Grid) ConsumePaste(e PasteEvent) EventResult {
	if g.focus < 0 || g.focus >= len(g.Children) {
		return Ignored()
	}
	return DispatchPasteEvent(g.Children[g.focus], e)
}

// ConsumeMouse routes to the child whose drawn cell contains the event.
func (g *Grid) ConsumeMouse(e MouseEvent) EventResult {
	if len(g.Children) == 0 {
		return Ignored()
	}
	x, y := e.X+g.lastRect.X, e.Y+g.lastRect.Y
	for i, rect := range g.childRects {
		if rect.Contains(x, y) {
			if e.Action == MousePress && i >= 0 && i < len(g.Children) {
				g.setFocus(i)
			}
			e.X = x - rect.X
			e.Y = y - rect.Y
			if i >= 0 && i < len(g.Children) {
				return g.Children[i].ConsumeMouse(e)
			}
			return Ignored()
		}
	}
	return Ignored()
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

// ApplyTheme updates the focus background color and forwards the theme to all
// children that implement Themeable.
func (g *Grid) ApplyTheme(theme ThemeColors) {
	g.FocusBG = theme.FocusBGColor()
	for _, child := range g.Children {
		if themeable, ok := child.(Themeable); ok {
			themeable.ApplyTheme(theme)
		}
	}
}
