// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strings"
	"time"

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
	// Claim marks the cell as foreground-owned even when it is blank and has
	// the default style. This is useful for widgets whose layout owns a cell
	// without relying on text or style inference.
	Claim bool
	// Surface marks an explicit background painted by PaintSurface. Surface
	// cells remain eligible for decoration while their background is inherited
	// by later foreground writes.
	Surface bool
}

// blank is the default empty cell.
var blank = Cell{Text: " "}

// Canvas is a 2-D frame buffer. Widgets draw into it; Pane flushes it to the terminal.
type Canvas struct {
	cols, rows int
	cells      [][]Cell
	claimed    [][]bool
	composing  bool
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
	claimed := make([][]bool, rows)
	for y := range cells {
		cells[y] = make([]Cell, cols)
		claimed[y] = make([]bool, cols)
		for x := range cells[y] {
			cells[y][x] = blank
		}
	}
	return &Canvas{cols: cols, rows: rows, cells: cells, claimed: claimed, CursorX: -1, CursorY: -1}
}

// Cols returns the canvas width.
func (c *Canvas) Cols() int { return c.cols }

// Rows returns the canvas height.
func (c *Canvas) Rows() int { return c.rows }

// Bounds returns a Rect covering the full canvas.
func (c *Canvas) Bounds() Rect { return Rect{W: c.cols, H: c.rows} }

// PaintSurface paints a background surface without claiming its cells from
// the decoration layer.
func (c *Canvas) PaintSurface(r Rect, style Style) {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			c.set(x, y, Cell{Text: " ", Style: style, Surface: true}, false)
		}
	}
}

// PaintForeground paints content that owns its cell, including an explicitly
// blank cell via Claim.
func (c *Canvas) PaintForeground(x, y int, cell Cell) {
	cell.Claim = true
	cell.Surface = false
	c.set(x, y, cell, true)
}

// PaintDecoration paints only into cells available to the decoration layer.
// It is safe to call during or outside ComposeBackground.
func (c *Canvas) PaintDecoration(x, y int, cell Cell) {
	if !c.IsEligibleBackground(x, y) {
		return
	}
	cell.Claim = false
	cell.Surface = false
	c.set(x, y, cell, false)
}

// Set places a single cell at (x, y). Out-of-bounds writes are silently dropped.
func (c *Canvas) Set(x, y int, cell Cell) {
	c.set(x, y, cell, true)
}

func (c *Canvas) set(x, y int, cell Cell, claim bool) {
	if x < 0 || x >= c.cols || y < 0 || y >= c.rows {
		return
	}
	if c.composing && !c.IsEligibleBackground(x, y) {
		return
	}
	// A blank cell without an explicit background inherits the surface
	// already present at this coordinate, including its decoration
	// eligibility. This must happen before the claim check below so a blank
	// cell merged up from a nested canvas (e.g. Box -> child widget) keeps
	// the Surface marker its inherited color came from, instead of looking
	// like claimed foreground content one level up.
	if !cell.Claim && (cell.Text == "" || cell.Text == " ") &&
		cell.Style.BG == ColorReset() && c.cells[y][x].Style.BG != ColorReset() {
		cell.Style.BG = c.cells[y][x].Style.BG
		cell.Surface = cell.Surface || c.cells[y][x].Surface
	}
	// A blank cell with the default background is transparent foreground
	// surface. Containers commonly paint these cells to establish their
	// bounds, but they must not hide a composited background. Text, explicit
	// background colors, and attributes remain foreground-owned.
	if claim && !cell.Surface {
		c.claimed[y][x] = true
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
	// A foreground cell without an explicit background inherits the surface
	// already present at this coordinate. This is what lets a child canvas
	// paint decoration without erasing its parent's colored surface.
	if cell.Style.BG == ColorReset() && c.cells[y][x].Style.BG != ColorReset() {
		cell.Style.BG = c.cells[y][x].Style.BG
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

// IsEligibleBackground reports whether a cell may be filled by a background
// compositor. Cells written by the foreground, wide-rune continuations, and
// the active cursor are protected.
func (c *Canvas) IsEligibleBackground(x, y int) bool {
	if x < 0 || x >= c.cols || y < 0 || y >= c.rows {
		return false
	}
	if c.claimed[y][x] || (x == c.CursorX && y == c.CursorY) {
		return false
	}
	return !c.Get(x, y).Continuation
}

// ComposeBackground renders a background only into cells not claimed by the
// foreground pass. The canvas cursor is restored even if the effect changes
// it accidentally, keeping cursor ownership with the foreground renderer.
func (c *Canvas) ComposeBackground(background Background, area Rect, now time.Time) {
	if c == nil || background == nil {
		return
	}
	cursorX, cursorY := c.CursorX, c.CursorY
	c.composing = true
	defer func() {
		c.composing = false
		c.CursorX, c.CursorY = cursorX, cursorY
	}()
	if animated, ok := background.(AnimatedBackground); ok {
		animated.DrawBackgroundAt(c, area, now)
		return
	}
	background.DrawBackground(c, area)
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

// ClearRect resets cells in r without inheriting their previous surface style.
func (c *Canvas) ClearRect(r Rect) {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			if x < 0 || x >= c.cols || y < 0 || y >= c.rows {
				continue
			}
			c.cells[y][x] = blank
			c.claimed[y][x] = false
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
	c.FlushWithClear(out, startRow, 0)
}

// FlushWithClear writes the complete canvas as one synchronized terminal
// update using the default resize rendering configuration. clearRows requests
// additional rows below the canvas to be erased; this is used when a resize
// shrinks the pane without exposing a blank intermediate frame.
func (c *Canvas) FlushWithClear(out interface{ WriteString(string) (int, error) }, startRow, clearRows int) {
	c.FlushWithConfig(out, startRow, clearRows, DefaultResizeConfig())
}

// FlushWithConfig writes the canvas to out with the provided resize rendering
// switches (atomic buffered flush, per-row clearing, synchronized output).
func (c *Canvas) FlushWithConfig(out interface{ WriteString(string) (int, error) }, startRow, clearRows int, cfg ResizeConfig) {
	var b strings.Builder
	var sink interface{ WriteString(string) (int, error) } = &b
	if !cfg.AtomicFlush {
		sink = out
	}

	if cfg.SynchronizedOutput {
		sink.WriteString("\x1b[?2026h") //nolint:errcheck
	}
	for y := 0; y < c.rows; y++ {
		sink.WriteString(fmt.Sprintf("\x1b[%d;1H", startRow+y)) // move to row //nolint:errcheck
		sink.WriteString(c.Row(y))                              //nolint:errcheck
		if cfg.RowClear {
			sink.WriteString("\x1b[K") //nolint:errcheck
		}
	}
	for y := 0; y < clearRows; y++ {
		sink.WriteString(fmt.Sprintf("\x1b[%d;1H\x1b[2K", startRow+c.rows+y)) //nolint:errcheck
	}
	if c.CursorX >= 0 && c.CursorY >= 0 {
		// Position and show the cursor only when a widget asked for it (a prompt).
		sink.WriteString(fmt.Sprintf("\x1b[%d;%dH", startRow+c.CursorY, c.CursorX+1)) //nolint:errcheck
		sink.WriteString("\x1b[?25h")                                                 //nolint:errcheck
	} else {
		// No prompt on this frame: hide the hardware cursor so it doesn't linger
		// as a stray block after the last drawn cell.
		sink.WriteString("\x1b[?25l") //nolint:errcheck
	}
	if cfg.SynchronizedOutput {
		sink.WriteString("\x1b[?2026l") //nolint:errcheck
	}
	if cfg.AtomicFlush {
		out.WriteString(b.String()) //nolint:errcheck
	}
}

// Clear resets every cell in the canvas to blank and hides the cursor. The
// cursor is reset each frame so a widget that wants it visible must set it
// during Draw; otherwise a prompt-less page (e.g. a View) would inherit a stale
// cursor position from the previous frame and leave a stray block on screen.
func (c *Canvas) Clear() {
	for y := range c.cells {
		for x := range c.cells[y] {
			c.cells[y][x] = blank
			c.claimed[y][x] = false
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
