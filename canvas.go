// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strings"

	"codeberg.org/ubunatic/loom/measure"
)

// Rect describes a rectangular region within the canvas (0-based, top-left origin).
type Rect struct {
	X, Y int // top-left corner
	W, H int // width and height
}

// Contains reports whether (x, y) falls inside r.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// Cell is a single terminal cell with printable content and visual style.
// Text should be a single printable character (or empty for a blank cell).
// Wide characters (emoji) occupy two consecutive cells; the right half
// should be stored as an empty Cell with the same style.
type Cell struct {
	Text         string
	Style        Style
	Continuation bool
}

// blank is the default empty cell.
var blank = Cell{Text: " "}

// Canvas is a 2-D frame buffer. Widgets draw into it; Pane flushes it to the terminal.
type Canvas struct {
	cols, rows int
	cells      [][]Cell
	CursorX    int // 0-based column index, -1 if hidden
	CursorY    int // 0-based row index, -1 if hidden
}

// NewCanvas allocates a cols×rows canvas filled with blank cells.
func NewCanvas(cols, rows int) *Canvas {
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	cells := make([][]Cell, rows)
	for y := range cells {
		cells[y] = make([]Cell, cols)
		for x := range cells[y] {
			cells[y][x] = blank
		}
	}
	return &Canvas{cols: cols, rows: rows, cells: cells, CursorX: -1, CursorY: -1}
}

// Cols returns the canvas width.
func (c *Canvas) Cols() int { return c.cols }

// Rows returns the canvas height.
func (c *Canvas) Rows() int { return c.rows }

// Bounds returns a Rect covering the full canvas.
func (c *Canvas) Bounds() Rect { return Rect{W: c.cols, H: c.rows} }

// Set places a single cell at (x, y). Out-of-bounds writes are silently dropped.
func (c *Canvas) Set(x, y int, cell Cell) {
	if x < 0 || x >= c.cols || y < 0 || y >= c.rows {
		return
	}
	if cell.Continuation {
		// Only a real wide lead may own a continuation cell.
		if x > 0 && StringWidth(c.cells[y][x-1].Text) == 2 {
			c.cells[y][x] = cell
		}
		return
	}
	clusters := textClusters(cell.Text)
	cell.Text = " "
	if len(clusters) > 0 {
		cell.Text = clusters[0]
	}
	w := StringWidth(cell.Text)
	if w == 2 && x+1 >= c.cols {
		return
	}
	// Erase both halves of any previous wide glyph touched by this write.
	clear := func(col int) {
		if c.cells[y][col].Continuation && col > 0 {
			c.cells[y][col-1] = blank
		}
		if col+1 < c.cols && c.cells[y][col+1].Continuation {
			c.cells[y][col+1] = blank
		}
		c.cells[y][col] = blank
	}
	clear(x)
	if w == 2 {
		clear(x + 1)
	}
	c.cells[y][x] = cell
	if w == 2 {
		c.cells[y][x+1] = Cell{Style: cell.Style, Continuation: true}
	}
}

// Get returns the cell at (x, y), or blank for out-of-bounds coordinates.
func (c *Canvas) Get(x, y int) Cell {
	if x >= 0 && x < c.cols && y >= 0 && y < c.rows {
		return c.cells[y][x]
	}
	return blank
}

// Fill fills the rectangle r with cell. Clips to canvas bounds.
func (c *Canvas) Fill(r Rect, cell Cell) {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			c.Set(x, y, cell)
		}
	}
}

// Write renders text starting at (x, y) using style, advancing x for each rune.
// Returns the number of columns consumed. Clips at canvas right edge.
// Correctly handles wide characters (emojis) by creating continuation cells.
func (c *Canvas) Write(x, y int, text string, style Style) int {
	col := x
	for _, cluster := range textClusters(text) {
		w := StringWidth(cluster)
		if col+w > c.cols {
			break
		}
		if col >= 0 {
			c.Set(col, y, Cell{Text: cluster, Style: style})
		}
		col += w
	}
	return col - x
}

// Row renders row y as an ANSI string, resetting style at the end.
// Returns an empty string for out-of-range rows.
func (c *Canvas) Row(y int) string {
	if y < 0 || y >= c.rows {
		return ""
	}
	var b strings.Builder
	var cur Style
	for _, cell := range c.cells[y] {
		if cell.Continuation {
			continue
		}
		if cell.Style != cur {
			b.WriteString(cell.Style.ANSI())
			cur = cell.Style
		}
		if cell.Text == "" {
			b.WriteByte(' ')
		} else {
			b.WriteString(cell.Text)
		}
	}
	b.WriteString("\x1b[0m") // reset after every row so colors don't bleed
	return b.String()
}

// Flush writes all rows to out using absolute cursor positioning.
// startRow is the 1-based terminal row of the canvas top-left corner.
func (c *Canvas) Flush(out interface{ WriteString(string) (int, error) }, startRow int) {
	var b strings.Builder
	for y := 0; y < c.rows; y++ {
		b.WriteString(fmt.Sprintf("\x1b[%d;1H", startRow+y)) // move to row
		b.WriteString(c.Row(y))
	}
	if c.CursorX >= 0 && c.CursorY >= 0 {
		// Position and show the cursor only when a widget asked for it (a prompt).
		b.WriteString(fmt.Sprintf("\x1b[%d;%dH", startRow+c.CursorY, c.CursorX+1))
		b.WriteString("\x1b[?25h")
	} else {
		// No prompt on this frame: hide the hardware cursor so it doesn't linger
		// as a stray block after the last drawn cell.
		b.WriteString("\x1b[?25l")
	}
	out.WriteString(b.String()) //nolint:errcheck
}

// Clear resets every cell in the canvas to blank and hides the cursor. The
// cursor is reset each frame so a widget that wants it visible must set it
// during Draw; otherwise a prompt-less page (e.g. a View) would inherit a stale
// cursor position from the previous frame and leave a stray block on screen.
func (c *Canvas) Clear() {
	for y := range c.cells {
		for x := range c.cells[y] {
			c.cells[y][x] = blank
		}
	}
	c.CursorX = -1
	c.CursorY = -1
}

// RuneWidth returns the visual column width of a single rune.
func RuneWidth(r rune) int {
	return measure.RuneWidth(r)
}

// StringWidth returns the visual column width of a string.
func StringWidth(s string) int {
	return measure.StringWidth(s)
}

// plainTerminalText removes terminal instructions from untrusted text. Styles
// are supplied through Style, not embedded control sequences. Unterminated
// control strings consume the remainder rather than leaking terminal commands.
func plainTerminalText(s string) string {
	return strings.Join(measure.Clusters(s), "")
}

// textClusters supports base runes with combining marks, not emoji ZWJ clusters.
// Leading combining marks are dropped: they must not attach outside the region.
func textClusters(text string) []string {
	return measure.Clusters(text)
}
