// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "codeberg.org/ubunatic/loom/layout"

// StackDir controls the direction a Stack or Split arranges its children.
type StackDir int

const (
	Vertical   StackDir = iota // children stacked top-to-bottom
	Horizontal                 // children arranged left-to-right
)

// Stack arranges a fixed list of widgets vertically or horizontally.
// Each child receives an equal share of the available space; fractional
// pixels are added to the last child so dimensions always sum correctly.
// Focus cycles through children on Tab.
type Stack struct {
	Dir      StackDir
	Children []Widget
	// Measured enables constraint-based allocation for this stack. When false,
	// the historical equal-share behavior is preserved.
	Measured    bool
	Constraints []layout.Constraint
	Gap         int
	focus       int // index of focused child
	childRects  []Rect
}

// NewStack creates a Stack with the given children and direction.
func NewStack(dir StackDir, children ...Widget) *Stack {
	return &Stack{Dir: dir, Children: children}
}

// PaneRequest merges the terminal requirements of all children.
func (s *Stack) PaneRequest() (request PaneRequest) {
	first := true
	for _, child := range s.Children {
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

// Draw tiles all children across r.
func (s *Stack) Draw(c *Canvas, r Rect) {
	n := len(s.Children)
	if n == 0 {
		return
	}
	for i, child := range s.Children {
		if f, ok := child.(Focusable); ok {
			f.SetFocus(i == s.focus)
		}
		cr := s.childRect(r, i, n)
		if len(s.childRects) != n {
			s.childRects = make([]Rect, n)
		}
		s.childRects[i] = cr
		child.Draw(c, cr)
		if Debug {
			drawDebugBorder(c, cr)
		}
	}
}

// childRect computes the Rect for child i of n within the parent Rect r.
func (s *Stack) childRect(r Rect, i, n int) Rect {
	if s.Measured && len(s.Constraints) == n && s.Gap >= 0 {
		items := make([]layout.Item, n)
		for j, constraint := range s.Constraints {
			items[j] = layout.Item{Constraint: constraint, Visible: true}
		}
		total := r.W
		if s.Dir == Vertical {
			total = r.H
		}
		allocations, err := layout.Plan(total, s.Gap, items)
		if err == nil {
			a := allocations[i]
			if s.Dir == Horizontal {
				return Rect{X: r.X + a.Offset, Y: r.Y, W: a.Size, H: r.H}
			}
			return Rect{X: r.X, Y: r.Y + a.Offset, W: r.W, H: a.Size}
		}
	}
	if s.Dir == Horizontal {
		unit := r.W / n
		x := r.X + i*unit
		w := unit
		if i == n-1 {
			w = r.W - i*unit // last child absorbs remainder
		}
		return Rect{X: x, Y: r.Y, W: w, H: r.H}
	}
	// Vertical
	unit := r.H / n
	y := r.Y + i*unit
	h := unit
	if i == n-1 {
		h = r.H - i*unit
	}
	return Rect{X: r.X, Y: y, W: r.W, H: h}
}

// HandleKey forwards to the focused child; Tab cycles focus.
func (s *Stack) HandleKey(e KeyEvent) (quit bool) {
	if len(s.Children) == 0 {
		return false
	}
	if e.Key == "tab" {
		s.focus = (s.focus + 1) % len(s.Children)
		return false
	}
	if quit, consumed := s.ConsumeKey(e); consumed {
		return quit
	}
	return s.Children[s.focus].HandleKey(e)
}

func (s *Stack) ConsumeKey(e KeyEvent) (quit, consumed bool) {
	if len(s.Children) == 0 {
		return false, false
	}
	if c, ok := s.Children[s.focus].(KeyConsumer); ok {
		if quit, consumed = c.ConsumeKey(e); consumed {
			return quit, true
		}
	}
	if e.Key == "tab" {
		s.focus = (s.focus + 1) % len(s.Children)
		return false, true
	}
	return false, false
}

// HandleMouse forwards to the child whose rect contains the event.
func (s *Stack) HandleMouse(e MouseEvent) (quit bool) {
	if len(s.Children) == 0 {
		return false
	}
	for i, rect := range s.childRects {
		if rect.Contains(e.X, e.Y) {
			return s.Children[i].HandleMouse(e)
		}
	}
	return s.Children[s.focus].HandleMouse(e)
}

// ContentHeight estimates the required height for this stack layout.
func (s *Stack) ContentHeight() int {
	if len(s.Children) == 0 {
		return 0
	}
	if s.Measured && len(s.Constraints) == len(s.Children) {
		height := 0
		for _, constraint := range s.Constraints {
			preferred := constraint.Preferred
			if preferred == 0 {
				preferred = constraint.Min
			}
			if s.Dir == Vertical {
				height += preferred
			} else {
				height = max(height, preferred)
			}
		}
		if s.Dir == Vertical && len(s.Children) > 1 {
			height += s.Gap * (len(s.Children) - 1)
		}
		return height
	}
	if s.Dir == Horizontal {
		maxH := 0
		for _, child := range s.Children {
			ch := 1
			if chWidget, ok := child.(ContentHeighter); ok {
				ch = chWidget.ContentHeight()
			}
			if ch > maxH {
				maxH = ch
			}
		}
		return maxH
	}
	// Vertical
	totalH := 0
	for _, child := range s.Children {
		ch := 1
		if chWidget, ok := child.(ContentHeighter); ok {
			ch = chWidget.ContentHeight()
		}
		totalH += ch
	}
	return totalH
}

// ContentWidth estimates the widest child width for horizontal composition.
func (s *Stack) ContentWidth() int {
	if len(s.Children) == 0 {
		return 0
	}
	width := 0
	for _, child := range s.Children {
		w := 1
		if measured, ok := child.(ContentWidther); ok {
			w = measured.ContentWidth()
		}
		if s.Dir == Horizontal {
			width += w
		} else {
			width = max(width, w)
		}
	}
	if s.Dir == Horizontal && len(s.Children) > 1 {
		width += s.Gap * (len(s.Children) - 1)
	}
	return width
}

// ApplyTheme forwards the theme to all children that implement Themeable.
func (s *Stack) ApplyTheme(theme ThemeColors) {
	for _, child := range s.Children {
		if themeable, ok := child.(Themeable); ok {
			themeable.ApplyTheme(theme)
		}
	}
}
