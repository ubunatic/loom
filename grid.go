// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	_ "embed"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// GridBorderMode controls whether Grid draws borders between cells and around
// its outside edge.
type GridBorderMode string

const (
	GridBorderNone  GridBorderMode = "none"
	GridBorderInner GridBorderMode = "inner"
	GridBorderFull  GridBorderMode = "full"
)

// GridBorderGlyphs contains the junction and line glyphs used by Grid borders.
type GridBorderGlyphs struct {
	Horizontal  string `yaml:"horizontal"`
	Vertical    string `yaml:"vertical"`
	Cross       string `yaml:"cross"`
	TDown       string `yaml:"t_down"`
	TUp         string `yaml:"t_up"`
	TRight      string `yaml:"t_right"`
	TLeft       string `yaml:"t_left"`
	TopLeft     string `yaml:"top_left"`
	TopRight    string `yaml:"top_right"`
	BottomLeft  string `yaml:"bottom_left"`
	BottomRight string `yaml:"bottom_right"`
}

//go:embed spec/grid.yaml
var gridSpecYAML []byte

type gridSpec struct {
	BorderModes []GridBorderMode `yaml:"border_modes"`
	Glyphs      GridBorderGlyphs `yaml:"glyphs"`
}

var speccedGrid = func() gridSpec {
	var spec gridSpec
	if err := yaml.Unmarshal(gridSpecYAML, &spec); err != nil {
		panic(fmt.Sprintf("loom: parse spec/grid.yaml: %v", err))
	}
	return spec
}()

// Grid lays out widgets in a fixed number of columns.
// Row count is inferred from len(Children) and Cols.
// Navigation: Left/Right move within a row; Up/Down move between rows.
// OnSelect is called when Enter is pressed on a cell; if nil, Enter
// delegates to the focused child widget (which may return quit=true).
type Grid struct {
	Cols        int
	Children    []Widget
	FocusBG     Color          // background color of the focused cell; zero = default
	BorderMode  GridBorderMode // none, inner, or full; zero value draws no border
	BorderStyle Style          // border appearance; zero value uses the theme border
	OnSelect    func(i int)    // called on Enter if non-nil; prevents quit propagation

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
	mode := g.BorderMode
	if !validGridBorderMode(mode) {
		mode = GridBorderNone
	}
	bordered := mode == GridBorderInner || mode == GridBorderFull
	full := mode == GridBorderFull
	verticalLines, horizontalLines := 0, 0
	if bordered {
		verticalLines, horizontalLines = g.Cols-1, rows-1
	}
	insetX, insetY := 0, 0
	if full {
		insetX, insetY = 1, 1
	}
	cellWidths := gridCellSizes(r.W-2*insetX-verticalLines, g.Cols, bordered)
	cellHeights := gridCellSizes(r.H-2*insetY-horizontalLines, rows, bordered)
	for i, child := range g.Children {
		if f, ok := child.(Focusable); ok {
			f.SetFocus(i == g.focus)
		}
		col, row := i%g.Cols, i/g.Cols
		cr := Rect{
			X: r.X + insetX + gridCellOffset(cellWidths, col, bordered),
			Y: r.Y + insetY + gridCellOffset(cellHeights, row, bordered),
			W: cellWidths[col],
			H: cellHeights[row],
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
	if bordered {
		style := g.BorderStyle
		if style == (Style{}) {
			style = Theme("plain").BoxStyle().Border
		}
		drawGridBorder(c, r, cellWidths, cellHeights, insetX, insetY, full, speccedGrid.Glyphs, style)
	}
}

func gridCellSizes(total, count int, spread bool) []int {
	if count < 1 {
		return nil
	}
	if total < 0 {
		total = 0
	}
	base, extra := total/count, total%count
	sizes := make([]int, count)
	for i := range sizes {
		sizes[i] = base
		if spread && i < extra {
			sizes[i]++
		}
	}
	return sizes
}

func gridCellOffset(sizes []int, index int, separated bool) int {
	offset := index
	if !separated {
		offset = 0
	}
	for i := 0; i < index; i++ {
		offset += sizes[i]
	}
	return offset
}

func validGridBorderMode(mode GridBorderMode) bool {
	for _, candidate := range speccedGrid.BorderModes {
		if mode == candidate {
			return true
		}
	}
	return false
}

func drawGridBorder(c *Canvas, r Rect, cellWidths, cellHeights []int, insetX, insetY int, full bool, g GridBorderGlyphs, style Style) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	left, top := r.X+insetX, r.Y+insetY
	cols, rows := len(cellWidths), len(cellHeights)
	width := sumInts(cellWidths) + cols - 1
	height := sumInts(cellHeights) + rows - 1
	if full {
		width = r.W - 2*insetX
		height = r.H - 2*insetY
	}
	if !full && (width <= 0 || height <= 0) {
		return
	}
	put := func(x, y int, text string) {
		if r.Contains(x, y) && c.Bounds().Contains(x, y) {
			c.Set(x, y, Cell{Text: text, Style: style, Claim: true})
		}
	}
	for col := 1; col < cols; col++ {
		x := left + gridCellOffset(cellWidths, col, true) - 1
		if full && (x <= r.X || x >= r.X+r.W-1) {
			continue
		}
		for y := top; y < top+height; y++ {
			put(x, y, g.Vertical)
		}
	}
	for row := 1; row < rows; row++ {
		y := top + gridCellOffset(cellHeights, row, true) - 1
		if full && (y <= r.Y || y >= r.Y+r.H-1) {
			continue
		}
		for x := left; x < left+width; x++ {
			put(x, y, g.Horizontal)
		}
	}
	if full && width > 0 && height > 0 {
		right, bottom := r.X+r.W-1, r.Y+r.H-1
		for x := r.X + 1; x < right; x++ {
			put(x, r.Y, g.Horizontal)
			put(x, bottom, g.Horizontal)
		}
		for y := r.Y + 1; y < bottom; y++ {
			put(r.X, y, g.Vertical)
			put(right, y, g.Vertical)
		}
		put(r.X, r.Y, g.TopLeft)
		put(right, r.Y, g.TopRight)
		put(r.X, bottom, g.BottomLeft)
		put(right, bottom, g.BottomRight)
		for col := 1; col < cols; col++ {
			x := left + gridCellOffset(cellWidths, col, true) - 1
			if x <= r.X || x >= right {
				continue
			}
			put(x, r.Y, g.TDown)
			put(x, bottom, g.TUp)
		}
		for row := 1; row < rows; row++ {
			y := top + gridCellOffset(cellHeights, row, true) - 1
			if y <= r.Y || y >= bottom {
				continue
			}
			put(r.X, y, g.TRight)
			put(right, y, g.TLeft)
		}
	} else if full {
		right, bottom := r.X+r.W-1, r.Y+r.H-1
		for x := r.X + 1; x < right; x++ {
			put(x, r.Y, g.Horizontal)
			put(x, bottom, g.Horizontal)
		}
		for y := r.Y + 1; y < bottom; y++ {
			put(r.X, y, g.Vertical)
			put(right, y, g.Vertical)
		}
		put(r.X, r.Y, g.TopLeft)
		put(right, r.Y, g.TopRight)
		put(r.X, bottom, g.BottomLeft)
		put(right, bottom, g.BottomRight)
	}
	for row := 1; row < rows; row++ {
		y := top + gridCellOffset(cellHeights, row, true) - 1
		for col := 1; col < cols; col++ {
			x := left + gridCellOffset(cellWidths, col, true) - 1
			if full && (x <= r.X || x >= r.X+r.W-1 || y <= r.Y || y >= r.Y+r.H-1) {
				continue
			}
			put(x, y, g.Cross)
		}
	}
}

func sumInts(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
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
	height := rows * maxH
	switch g.BorderMode {
	case GridBorderInner:
		height += rows - 1
	case GridBorderFull:
		height += rows + 1
	}
	return height
}

// ApplyTheme updates the focus background color and forwards the theme to all
// children that implement Themeable.
func (g *Grid) ApplyTheme(theme ThemeColors) {
	g.FocusBG = theme.FocusBGColor()
	g.BorderStyle = theme.BoxStyle().Border
	for _, child := range g.Children {
		if themeable, ok := child.(Themeable); ok {
			themeable.ApplyTheme(theme)
		}
	}
}
