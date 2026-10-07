// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// ModalMouseProvider exposes the topmost active modal mouse recipient. The
// rectangle's origin is relative to the provider's allocation, not the canvas;
// events outside that rectangle still belong exclusively to the modal. Return
// nil when no modal is active. A modal may return itself with a zero origin.
type ModalMouseProvider interface {
	ModalMouseTarget() (Widget, Rect)
}

// ActiveModalMouse resolves a widget's modal contract, including transparent
// WidgetUnwrapper wrappers. Containers expose only their visible children.
func ActiveModalMouse(root Widget) (Widget, Rect) {
	for root != nil {
		if provider, ok := root.(ModalMouseProvider); ok {
			return provider.ModalMouseTarget()
		}
		wrapper, ok := root.(WidgetUnwrapper)
		if !ok {
			break
		}
		root = wrapper.Unwrap()
	}
	return nil, Rect{}
}

// ModalMouseCapture routes events before ordinary hit-testing or focus changes.
// Its zero value is ready to use. Hosts and containers retain one instance across
// events so a modal press owns its drag/release even after dismissing the modal.
type ModalMouseCapture struct {
	target Widget
	origin Rect
}

// Dispatch returns captured=true whenever the modal layer owns the event. It
// preserves Done and Quit and always marks a captured event Consumed, even when
// the modal's handler ignores it. Coordinates reaching the target are 0-based
// and target-local; a backdrop event may have negative coordinates.
func (c *ModalMouseCapture) Dispatch(root Widget, e MouseEvent) (result EventResult, captured bool) {
	if e.Action == MousePress {
		// A new press starts a new gesture, including terminals that omit release.
		c.target = nil
	}
	if c.target != nil && (e.Action == MouseDrag || e.Action == MouseRelease) {
		target, origin := c.target, c.origin
		if e.Action == MouseRelease {
			c.target = nil
		}
		if active, _ := ActiveModalMouse(target); active == nil {
			return Handled(), true
		}
		e.X -= origin.X
		e.Y -= origin.Y
		result = target.ConsumeMouse(e)
		result.Consumed = true
		return result, true
	}
	target, origin := ActiveModalMouse(root)
	if target == nil {
		return Ignored(), false
	}
	if e.Action == MousePress {
		c.target, c.origin = target, origin
	}
	e.X -= origin.X
	e.Y -= origin.Y
	result = target.ConsumeMouse(e)
	result.Consumed = true
	return result, true
}

func childModalMouse(child Widget, allocation, parent Rect) (Widget, Rect) {
	if allocation.W <= 0 || allocation.H <= 0 || !rectsOverlap(allocation, parent) {
		return nil, Rect{}
	}
	target, origin := ActiveModalMouse(child)
	if target != nil {
		origin.X += allocation.X - parent.X
		origin.Y += allocation.Y - parent.Y
	}
	return target, origin
}

func rectsOverlap(a, b Rect) bool {
	return a.X < b.X+b.W && a.Y < b.Y+b.H && b.X < a.X+a.W && b.Y < a.Y+a.H
}

// ModalMouseTarget returns the last drawn active modal in the stack.
func (s *Stack) ModalMouseTarget() (Widget, Rect) {
	return childrenModalMouse(s.Children, s.childRects, s.lastRect)
}

// ModalMouseTarget returns the last visible active modal in the grid.
func (g *Grid) ModalMouseTarget() (Widget, Rect) {
	return childrenModalMouse(g.Children, g.childRects, g.lastRect)
}

func childrenModalMouse(children []Widget, allocations []Rect, parent Rect) (Widget, Rect) {
	for i := min(len(children), len(allocations)) - 1; i >= 0; i-- {
		if target, origin := childModalMouse(children[i], allocations[i], parent); target != nil {
			return target, origin
		}
	}
	return nil, Rect{}
}

// ModalMouseTarget returns the topmost active modal in either split child.
func (s *Split) ModalMouseTarget() (Widget, Rect) {
	if target, origin := childModalMouse(s.Second, s.secondRect, s.lastRect); target != nil {
		return target, origin
	}
	return childModalMouse(s.First, s.firstRect, s.lastRect)
}

// ModalMouseTarget exposes only the active tab's visible modal.
func (t *Tabs) ModalMouseTarget() (Widget, Rect) {
	if !t.drawn {
		return nil, Rect{}
	}
	return childModalMouse(t.active(), t.childRect(t.lastRect), t.lastRect)
}

// ModalMouseTarget translates the child's content origin through scrolling.
func (v *Viewport) ModalMouseTarget() (Widget, Rect) {
	if v.content.W <= 0 || v.content.H <= 0 {
		return nil, Rect{}
	}
	target, origin := ActiveModalMouse(v.Child)
	if target != nil {
		origin.X -= v.ScrollX
		origin.Y -= v.ScrollY
	}
	return target, origin
}

// ModalMouseTarget returns the last drawn modal in a visible frame box.
func (f *Frame) ModalMouseTarget() (Widget, Rect) {
	allocations := f.Layout(f.lastRect.W, f.lastRect.H)
	parent := Rect{W: f.lastRect.W, H: f.lastRect.H}
	for i := len(allocations) - 1; i >= 0; i-- {
		box, area := f.Boxes[i], allocations[i]
		if box.Hidden || area.W < 2 || area.H < 2 {
			continue
		}
		padding := max(0, box.Padding)
		inner := Rect{X: area.X + 1 + padding, Y: area.Y + 1 + padding,
			W: area.W - 2 - 2*padding, H: area.H - 2 - 2*padding}
		if target, origin := childModalMouse(box.Child, inner, parent); target != nil {
			return target, origin
		}
	}
	return nil, Rect{}
}

// ModalMouseTarget returns the last visible modal in a form field.
func (f *Form) ModalMouseTarget() (Widget, Rect) {
	for i := min(len(f.Fields), len(f.fieldRects)) - 1; i >= 0; i-- {
		child, _ := f.Fields[i].Widget.(Widget)
		if target, origin := childModalMouse(child, f.fieldRects[i], f.lastRect); target != nil {
			return target, origin
		}
	}
	return nil, Rect{}
}

// ModalMouseTarget exposes the choice's command help popup, when open.
func (c *Choice) ModalMouseTarget() (Widget, Rect) {
	if commandHelpOpen(c.cmd) {
		return c, Rect{}
	}
	return nil, Rect{}
}

// ModalMouseTarget exposes the table's command help popup, when open.
func (t *Table) ModalMouseTarget() (Widget, Rect) {
	if commandHelpOpen(t.cmd) {
		return t, Rect{}
	}
	return nil, Rect{}
}

// ModalMouseTarget exposes the search bar's command help popup, when open.
func (s *SearchBar) ModalMouseTarget() (Widget, Rect) {
	if commandHelpOpen(s.cmd) {
		return s, Rect{}
	}
	return nil, Rect{}
}

func commandHelpOpen(commands *cmdBar) bool {
	return commands != nil && commands.help != nil && commands.help.Open
}
