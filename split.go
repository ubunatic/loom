// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "math"

// Orientation is the axis on which a Split arranges its children. It aliases
// StackDir so Horizontal and Vertical are shared consistently by containers.
type Orientation = StackDir

// DividerStyle controls the glyph and appearance of a Split divider.
type DividerStyle struct {
	Glyph string
	Style Style
}

// Split lays out two widgets along one axis. Ratio is the preferred fraction
// assigned to First after Gap has been reserved. Minimums are measured on the
// split axis and are honored whenever the available space permits.
type Split struct {
	Orientation Orientation
	Ratio       float64
	Gap         int
	MinFirst    int
	MinSecond   int
	Divider     DividerStyle
	First       Widget
	Second      Widget

	focused    bool
	focus      int
	lastRect   Rect
	firstRect  Rect
	secondRect Rect
	dragging   bool
}

// NewSplit returns an evenly divided horizontal split with a one-cell divider.
func NewSplit(first, second Widget) *Split {
	return &Split{First: first, Second: second, Orientation: Horizontal, Ratio: 0.5, Gap: 1}
}

// SetRatio clamps and stores the preferred first-child ratio.
func (s *Split) SetRatio(ratio float64) {
	if math.IsNaN(ratio) {
		ratio = 0.5
	}
	s.Ratio = clampRatio(ratio)
}

// Layout returns the child rectangles for bounds r. The space between them is
// the divider. When both minimums cannot fit, First is bounded first and Second
// receives the remainder.
func (s *Split) Layout(r Rect) (Rect, Rect) {
	axis := r.W
	vertical := s.Orientation == Vertical
	if vertical {
		axis = r.H
	}
	gap := min(max(0, s.Gap), max(0, axis))
	available := max(0, axis-gap)
	ratio := s.Ratio
	if math.IsNaN(ratio) {
		ratio = 0.5
	}
	ratio = clampRatio(ratio)
	first := int(math.Round(float64(available) * ratio))
	minFirst := min(max(0, s.MinFirst), available)
	minSecond := min(max(0, s.MinSecond), available)
	if minFirst+minSecond <= available {
		first = min(max(first, minFirst), available-minSecond)
	} else {
		first = minFirst
	}
	second := available - first
	if vertical {
		return Rect{X: r.X, Y: r.Y, W: r.W, H: first}, Rect{X: r.X, Y: r.Y + first + gap, W: r.W, H: second}
	}
	return Rect{X: r.X, Y: r.Y, W: first, H: r.H}, Rect{X: r.X + first + gap, Y: r.Y, W: second, H: r.H}
}

func clampRatio(ratio float64) float64 {
	if ratio < 0 {
		return 0
	}
	if ratio > 1 {
		return 1
	}
	return ratio
}

// Draw renders both children and the divider in r.
func (s *Split) Draw(c *Canvas, r Rect) {
	s.lastRect = r
	s.firstRect, s.secondRect = s.Layout(r)
	if s.First != nil && s.firstRect.W > 0 && s.firstRect.H > 0 {
		paintClipped(c, s.firstRect, func(child *Canvas) { s.First.Draw(child, child.Bounds()) })
	}
	if s.Second != nil && s.secondRect.W > 0 && s.secondRect.H > 0 {
		paintClipped(c, s.secondRect, func(child *Canvas) { s.Second.Draw(child, child.Bounds()) })
	}
	s.drawDivider(c)
}

func (s *Split) drawDivider(c *Canvas) {
	glyph := s.Divider.Glyph
	if glyph == "" {
		if s.Orientation == Vertical {
			glyph = "─"
		} else {
			glyph = "│"
		}
	}
	if s.Orientation == Vertical {
		start, end := s.firstRect.Y+s.firstRect.H, s.secondRect.Y
		for y := start; y < end; y++ {
			for x := s.lastRect.X; x < s.lastRect.X+s.lastRect.W; x++ {
				c.Set(x, y, Cell{Text: glyph, Style: s.Divider.Style})
			}
		}
		return
	}
	start, end := s.firstRect.X+s.firstRect.W, s.secondRect.X
	for x := start; x < end; x++ {
		for y := s.lastRect.Y; y < s.lastRect.Y+s.lastRect.H; y++ {
			c.Set(x, y, Cell{Text: glyph, Style: s.Divider.Style})
		}
	}
}

// HandleKey traverses children with Tab and Shift-Tab, then delegates other
// keys to the focused child.
func (s *Split) HandleKey(e KeyEvent) bool {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	s.ensureFocus()
	switch key {
	case "tab":
		if !s.FocusNext() {
			s.focusFirst()
		}
		return false
	case "shift-tab":
		if !s.FocusPrevious() {
			s.focusLast()
		}
		return false
	case "[":
		s.SetRatio(s.Ratio - 0.05)
		return false
	case "]":
		s.SetRatio(s.Ratio + 0.05)
		return false
	}
	if child := s.focusedChild(); child != nil {
		return child.HandleKey(e)
	}
	return false
}

// HandleMouse forwards an event to the child under the pointer, translating
// terminal coordinates to that child's local 1-based coordinates.
func (s *Split) HandleMouse(e MouseEvent) bool {
	x, y := e.X-1, e.Y-1
	if e.Action == MousePress && e.Button == MouseLeft && s.dividerContains(x, y) {
		s.dragging = true
		return false
	}
	if e.Action == MouseRelease && s.dragging {
		s.dragging = false
		return false
	}
	if e.Action == MouseDrag && s.dragging {
		s.setRatioFromPosition(x, y)
		return false
	}
	for index, target := range []struct {
		widget Widget
		rect   Rect
	}{{s.First, s.firstRect}, {s.Second, s.secondRect}} {
		if target.widget == nil || !target.rect.Contains(x, y) {
			continue
		}
		if e.Action == MousePress {
			s.setFocusedChild(index)
		}
		e.X = x - target.rect.X + 1
		e.Y = y - target.rect.Y + 1
		return target.widget.HandleMouse(e)
	}
	return false
}

func (s *Split) dividerContains(x, y int) bool {
	if s.Orientation == Vertical {
		return x >= s.lastRect.X && x < s.lastRect.X+s.lastRect.W &&
			y >= s.firstRect.Y+s.firstRect.H && y < s.secondRect.Y
	}
	return y >= s.lastRect.Y && y < s.lastRect.Y+s.lastRect.H &&
		x >= s.firstRect.X+s.firstRect.W && x < s.secondRect.X
}

func (s *Split) setRatioFromPosition(x, y int) {
	axis, position := s.lastRect.W, x-s.lastRect.X
	if s.Orientation == Vertical {
		axis, position = s.lastRect.H, y-s.lastRect.Y
	}
	available := axis - min(max(0, s.Gap), max(0, axis))
	if available <= 0 {
		return
	}
	s.SetRatio(float64(position) / float64(available))
}

// Focused reports whether the split owns focus.
func (s *Split) Focused() bool { return s.focused }

// SetFocus updates focus for the split and its selected descendant.
func (s *Split) SetFocus(focused bool) {
	s.focused = focused
	if focused {
		s.ensureFocus()
	}
	s.applyFocus()
}

// FocusNext moves to the next focusable descendant, returning false at the end.
func (s *Split) FocusNext() bool {
	s.ensureFocus()
	if child, ok := s.focusedChild().(FocusContainer); ok && child.FocusNext() {
		return true
	}
	if s.focus == 0 && isFocusable(s.Second) {
		s.setFocusedChild(1)
		focusFirstWidget(s.Second)
		return true
	}
	return false
}

// FocusPrevious moves to the previous focusable descendant, returning false at the start.
func (s *Split) FocusPrevious() bool {
	s.ensureFocus()
	if child, ok := s.focusedChild().(FocusContainer); ok && child.FocusPrevious() {
		return true
	}
	if s.focus == 1 && isFocusable(s.First) {
		s.setFocusedChild(0)
		focusLastWidget(s.First)
		return true
	}
	return false
}

func (s *Split) ensureFocus() {
	if s.focus == 1 && isFocusable(s.Second) || s.focus == 0 && isFocusable(s.First) {
		return
	}
	if isFocusable(s.First) {
		s.focus = 0
	} else if isFocusable(s.Second) {
		s.focus = 1
	}
}

func (s *Split) focusedChild() Widget {
	if s.focus == 1 {
		return s.Second
	}
	return s.First
}

func (s *Split) setFocusedChild(index int) {
	s.focus = index
	s.focused = true
	s.applyFocus()
}

func (s *Split) applyFocus() {
	if child, ok := s.First.(Focusable); ok {
		child.SetFocus(s.focused && s.focus == 0)
	}
	if child, ok := s.Second.(Focusable); ok {
		child.SetFocus(s.focused && s.focus == 1)
	}
}

func (s *Split) focusFirst() { s.setFocusedChild(firstFocusableIndex(s.First, s.Second)) }
func (s *Split) focusLast() {
	s.setFocusedChild(lastFocusableIndex(s.First, s.Second))
	focusLastWidget(s.focusedChild())
}

func isFocusable(widget Widget) bool {
	_, ok := widget.(Focusable)
	return ok
}

func firstFocusableIndex(first, second Widget) int {
	if isFocusable(first) {
		return 0
	}
	if isFocusable(second) {
		return 1
	}
	return 0
}

func lastFocusableIndex(first, second Widget) int {
	if isFocusable(second) {
		return 1
	}
	return firstFocusableIndex(first, second)
}

func focusFirstWidget(widget Widget) {
	if child, ok := widget.(Focusable); ok {
		child.SetFocus(true)
	}
	if container, ok := widget.(FocusContainer); ok {
		for container.FocusPrevious() {
		}
	}
}

func focusLastWidget(widget Widget) {
	container, ok := widget.(FocusContainer)
	if !ok {
		return
	}
	for container.FocusNext() {
	}
}
