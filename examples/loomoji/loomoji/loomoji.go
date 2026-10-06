// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package loomoji demonstrates a compact inline emoji and symbol picker.
package loomoji

import (
	"fmt"
	"os"
	"strings"

	"ubunatic.com/loom"
)

// ── Data model ────────────────────────────────────────────────────────────────

const (
	emojiGridColumns = 10
)

// ── Widget types ──────────────────────────────────────────────────────────────

type picker struct {
	query      *loom.TextInput
	split      *loom.Split
	group      int
	items      []int
	index      int
	gridFocus  bool // true = grid is focused (arrow keys move selection); false = search is focused
	chosen     string
	width      int
	searchY    int
	categoryY  int
	lastRect   loom.Rect
	gridRect   loom.Rect
	cols       int
	viewRows   int
	gridStart  int
	categories []category
	entries    []entry
	cellWidth  int
}

type gridPane struct{ picker *picker }

func (g *gridPane) Draw(c *loom.Canvas, r loom.Rect)          { g.picker.drawGrid(c, r) }
func (g *gridPane) ConsumeKey(loom.KeyEvent) loom.EventResult { return loom.Ignored() }
func (g *gridPane) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	if g.picker.handleGridMouse(e) {
		return loom.QuitResult()
	}
	return loom.Ignored()
}

func newPicker() *picker {
	p := &picker{
		query:      loom.NewTextInput(""),
		gridFocus:  true, // start with grid focused so arrow keys work immediately
		categories: categories(),
		entries:    entries(),
	}
	p.split = loom.NewSplit(&gridPane{picker: p}, nil)
	p.split.Ratio = 1.0
	p.split.MinFirst = 10
	p.split.MinSecond = 0
	p.refresh()
	return p
}

func (p *picker) refresh() {
	p.items = p.items[:0]
	q := strings.TrimSpace(strings.ToLower(p.query.Value()))
	for i, e := range p.entries {
		matchQuery := q == "" || strings.Contains(strings.ToLower(e.name), q) || strings.Contains(e.icon, q)
		if q == "" {
			if e.group == p.group {
				p.items = append(p.items, i)
			}
		} else if matchQuery {
			p.items = append(p.items, i)
		}
	}
	p.index = min(p.index, max(0, len(p.items)-1))
}

func (p *picker) selectCategory(idx int) {
	if idx >= 0 && idx < len(p.categories) {
		p.group = idx
		p.query.SetValue("")
		p.index = 0
		p.gridFocus = true
		p.refresh()
	}
}

func (p *picker) prevCategory() {
	p.selectCategory((p.group - 1 + len(p.categories)) % len(p.categories))
}

func (p *picker) nextCategory() {
	p.selectCategory((p.group + 1) % len(p.categories))
}

// ── Draw ──────────────────────────────────────────────────────────────────────

func (p *picker) Draw(c *loom.Canvas, r loom.Rect) {
	p.lastRect = r
	if r.W <= 0 || r.H <= 0 {
		return
	}
	background := loom.ColorRGB(38, 40, 43)
	panel := loom.ColorRGB(31, 33, 36)
	accent := loom.ColorRGB(32, 151, 185)

	c.PaintSurface(r, loom.Style{BG: background})
	// Left border strip
	for y := r.Y; y < r.Y+r.H; y++ {
		c.PaintForeground(r.X, y, loom.Cell{Text: " ", Style: loom.Style{BG: panel}, Claim: true})
	}
	// Top border strip
	for x := r.X; x < r.X+r.W; x++ {
		c.PaintForeground(x, r.Y, loom.Cell{Text: " ", Style: loom.Style{BG: panel}, Claim: true})
	}

	p.width = r.W

	// ── Search bar ────────────────────────────────────────────────────────────
	p.searchY = r.Y + 1
	searchRect := loom.Rect{X: r.X, Y: p.searchY, W: r.W, H: min(2, max(0, r.Y+r.H-p.searchY))}
	c.PaintSurface(searchRect, loom.Style{BG: panel})
	if r.H >= 3 {
		c.PaintSurface(loom.Rect{X: r.X + 1, Y: p.searchY, W: max(0, r.W-2), H: 1}, loom.Style{BG: loom.ColorRGB(67, 69, 72)})
		c.Write(r.X+2, p.searchY, "⌕", loom.Style{FG: accent, Bold: true})
		query := p.query.Value()
		if query == "" {
			if p.gridFocus {
				c.Write(r.X+5, p.searchY, "search…", loom.Style{FG: loom.ColorRGB(120, 120, 120), Dim: true})
			} else {
				c.Write(r.X+5, p.searchY, "search…", loom.Style{FG: loom.ColorRGB(180, 180, 180)})
			}
		} else {
			c.Write(r.X+5, p.searchY, query, loom.Style{FG: loom.ColorRGB(220, 220, 220)})
		}
		if !p.gridFocus {
			// Show cursor in search bar when search is focused
			caret := min(p.query.Caret(), len([]rune(p.query.Value())))
			c.CursorX = r.X + 5 + loom.StringWidth(string([]rune(p.query.Value())[:caret]))
			c.CursorY = p.searchY
		}
	}

	// ── Category bar ──────────────────────────────────────────────────────────
	p.categoryY = r.Y + r.H - 2
	if p.categoryY >= r.Y && p.categoryY < r.Y+r.H {
		c.PaintSurface(loom.Rect{X: r.X, Y: p.categoryY, W: r.W, H: 1}, loom.Style{BG: panel})
		x := r.X + 2
		for i, cat := range p.categories {
			w := loom.StringWidth(cat.icon)
			if x+w >= r.X+r.W-1 {
				break
			}
			style := loom.Style{FG: loom.ColorRGB(190, 190, 190)}
			if i == p.group {
				style = loom.Style{FG: loom.ColorRGB(255, 255, 255), BG: accent, Bold: true}
			}
			c.Write(x, p.categoryY, cat.icon, style)
			x += w + 2
		}
	}

	// ── Status bar ────────────────────────────────────────────────────────────
	statusY := r.Y + r.H - 1
	c.PaintSurface(loom.Rect{X: r.X, Y: statusY, W: r.W, H: 1}, loom.Style{BG: panel})
	footer := "↑↓←→ grid   Tab search   f/[ ] cycle   Enter copy   F10 Quit"
	if len(p.items) > 0 {
		e := p.entries[p.items[p.index]]
		cat := p.categories[e.group].label
		footer = fmt.Sprintf("%s  %s  [%s]  %d / %d", e.icon, e.name, cat, p.index+1, len(p.items))
	}
	c.Write(r.X+1, statusY, footer, loom.Style{FG: loom.ColorRGB(175, 178, 181), Dim: true})

	// ── Grid area ─────────────────────────────────────────────────────────────
	mainY := p.searchY + 2
	mainRect := loom.Rect{X: r.X, Y: mainY, W: r.W, H: max(0, p.categoryY-mainY)}
	p.gridRect = mainRect
	p.split.Draw(c, mainRect)
}

func (p *picker) drawGrid(c *loom.Canvas, r loom.Rect) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	accent := loom.ColorRGB(32, 151, 185)
	p.cellWidth = 1
	for _, item := range p.items {
		icon := p.entries[item].icon
		p.cellWidth = max(p.cellWidth, loom.StringWidth(icon))
	}
	p.cellWidth = max(1, p.cellWidth)
	columnWidth := p.cellWidth + 1
	p.cols = min(emojiGridColumns, max(1, (r.W-2)/columnWidth))
	p.viewRows = r.H
	selectedRow := p.index / p.cols
	startRow := 0
	if selectedRow >= p.viewRows && p.viewRows > 0 {
		startRow = selectedRow - p.viewRows + 1
	}
	start := startRow * p.cols
	p.gridStart = start
	visible := min(len(p.items)-start, p.cols*p.viewRows)
	for n := 0; n < visible; n++ {
		idx := start + n
		x := 1 + (n%p.cols)*columnWidth
		y := n / p.cols
		icon := p.entries[p.items[idx]].icon
		style := loom.Style{FG: loom.ColorRGB(220, 200, 120)}
		cellText := loom.TruncateText(icon, p.cellWidth, "")
		if idx == p.index {
			style = loom.Style{FG: loom.ColorRGB(255, 255, 255), BG: accent, Bold: true}
			if w := loom.StringWidth(cellText); w < p.cellWidth {
				cellText += strings.Repeat(" ", p.cellWidth-w)
			}
		}
		c.Write(r.X+x, r.Y+y, cellText, style)
	}
	if len(p.items) == 0 {
		c.Write(r.X+2, r.Y, "No matching entries", loom.Style{FG: loom.ColorRGB(170, 170, 170), Dim: true})
	}
}

// ── Input handling ────────────────────────────────────────────────────────────

func (p *picker) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	switch key {
	case "ctrl-c", "esc":
		return loom.QuitResult()
	case "tab":
		p.gridFocus = !p.gridFocus
	case "left":
		if !p.gridFocus {
			p.query.ConsumeKey(e)
			p.refresh()
		} else if p.index > 0 {
			p.index--
		}
	case "right":
		if !p.gridFocus {
			p.query.ConsumeKey(e)
			p.refresh()
		} else if p.index+1 < len(p.items) {
			p.index++
		}
	case "up":
		p.gridFocus = true
		p.index = max(0, p.index-p.cols)
	case "down":
		p.gridFocus = true
		p.index = min(max(0, len(p.items)-1), p.index+p.cols)
	case "pgup", "shift-left":
		p.prevCategory()
	case "pgdown", "shift-right", "ctrl-f":
		p.nextCategory()
	case "enter":
		if len(p.items) > 0 {
			p.chosen = p.entries[p.items[p.index]].icon
			return loom.QuitResult()
		}
	case "backspace":
		if !p.gridFocus {
			p.query.ConsumeKey(e)
			p.refresh()
		}
	default:
		// [ / ] or f / F cycle categories when in grid mode.
		if p.gridFocus && (key == "[" || key == "]" || key == "f" || key == "F") {
			if key == "[" {
				p.prevCategory()
			} else {
				p.nextCategory()
			}
			return loom.Ignored()
		}
		// Printable text types into search and switches to search mode if in grid.
		if e.Text != "" {
			p.gridFocus = false
			p.query.ConsumeKey(e)
			p.index = 0
			p.refresh()
		}
	}
	return loom.Ignored()
}

func (p *picker) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	if e.Action != loom.MousePress && e.Action != loom.MouseHover && e.Action != loom.MouseDrag {
		return loom.Ignored()
	}
	x, y := e.X+p.lastRect.X, e.Y+p.lastRect.Y
	// Click on search bar → switch to search focus
	if (y == p.searchY || y == p.searchY+1) && e.Action == loom.MousePress && e.Button == loom.MouseLeft {
		p.gridFocus = false
		return loom.Ignored()
	}
	// Click on category bar → switch category
	if y == p.categoryY || y == p.categoryY+1 {
		if e.Action != loom.MousePress || e.Button != loom.MouseLeft {
			return loom.Ignored()
		}
		if x < p.lastRect.X+2 || x >= p.lastRect.X+p.width {
			return loom.Ignored()
		}
		// Walk the category icons to find which one was clicked.
		cx := p.lastRect.X + 2
		for i, cat := range p.categories {
			w := loom.StringWidth(cat.icon) + 2
			if x >= cx && x < cx+w {
				p.selectCategory(i)
				return loom.Ignored()
			}
			cx += w
		}
		return loom.Ignored()
	}
	e.X, e.Y = x-p.gridRect.X, y-p.gridRect.Y
	return p.split.ConsumeMouse(e)
}

func (p *picker) handleGridMouse(e loom.MouseEvent) bool {
	if e.Action != loom.MousePress && e.Action != loom.MouseHover && e.Action != loom.MouseDrag {
		return false
	}
	if e.Y < 0 || e.Y >= p.viewRows || e.X < 1 {
		return false
	}
	if p.cellWidth < 1 {
		p.cellWidth = 1
	}
	col := (e.X - 1) / max(1, p.cellWidth+1)
	idx := p.gridStart + e.Y*p.cols + col
	if col < 0 || col >= p.cols || idx < 0 || idx >= len(p.items) {
		return false
	}
	p.index, p.gridFocus = idx, true
	if e.Action == loom.MousePress && e.Button == loom.MouseLeft {
		p.chosen = p.entries[p.items[idx]].icon
		return true
	}
	return false
}

// ── Public API ────────────────────────────────────────────────────────────────

// NewWidget constructs the picker widget for embedding in another Loom app.
func NewWidget(args []string) (loom.Widget, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("loomoji: unexpected arguments: %v", args)
	}
	return newPicker(), nil
}

// Run launches the interactive emoji and symbol picker.
func Run(args []string) error {
	cmd := newCommand(runPicker, runMeasure)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func runPicker() error {
	pane, err := loom.New(13)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	pane.EnableMouse()
	app := newPicker()
	if err := pane.Run(app); err != nil {
		return err
	}
	if app.chosen != "" {
		_, err = fmt.Fprint(os.Stdout, app.chosen)
		return err
	}
	return nil
}
