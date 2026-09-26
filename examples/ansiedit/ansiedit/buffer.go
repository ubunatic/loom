// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansiedit

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/measure"
)

// EditMode represents character insertion vs overtype behavior.
type EditMode int

const (
	ModeOvertype EditMode = iota
	ModeInsert
)

func (m EditMode) String() string {
	switch m {
	case ModeInsert:
		return "Insert"
	default:
		return "Overtype"
	}
}

// BufferCell represents a single styled character cell in the ANSI canvas buffer.
type BufferCell struct {
	Rune      rune
	FG        loom.Color
	BG        loom.Color
	Bold      bool
	Dim       bool
	Underline bool
	Invert    bool
}

// BlankCell returns a default empty cell.
func BlankCell() BufferCell {
	return BufferCell{
		Rune: ' ',
		FG:   loom.ColorReset(),
		BG:   loom.ColorReset(),
	}
}

// IsBlank reports whether the cell contains whitespace and default colors.
func (c BufferCell) IsBlank() bool {
	return (c.Rune == 0 || c.Rune == ' ') &&
		c.FG == loom.ColorReset() &&
		c.BG == loom.ColorReset() &&
		!c.Bold && !c.Dim && !c.Underline && !c.Invert
}

// Style converts BufferCell formatting to loom.Style.
func (c BufferCell) Style() loom.Style {
	fg := c.FG
	bg := c.BG
	if c.Invert {
		fg, bg = bg, fg
	}
	return loom.Style{
		FG:        fg,
		BG:        bg,
		Bold:      c.Bold,
		Dim:       c.Dim,
		Underline: c.Underline,
	}
}

// ToLoomCell converts BufferCell to loom.Cell.
func (c BufferCell) ToLoomCell() loom.Cell {
	r := c.Rune
	if r == 0 {
		r = ' '
	}
	return loom.Cell{
		Text:  string(r),
		Style: c.Style(),
	}
}

// Buffer is an in-memory 2D cell grid for ANSI art editing.
type Buffer struct {
	cols      int
	rows      int
	cells     [][]BufferCell
	clipboard BufferCell
	hasClip   bool
	modified  bool
	path      string
}

// NewBuffer creates an empty ANSI buffer with the specified columns and rows.
func NewBuffer(cols, rows int) *Buffer {
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	buf := &Buffer{
		cols:  cols,
		rows:  rows,
		cells: make([][]BufferCell, rows),
	}
	for y := 0; y < rows; y++ {
		buf.cells[y] = make([]BufferCell, cols)
		for x := 0; x < cols; x++ {
			buf.cells[y][x] = BlankCell()
		}
	}
	return buf
}

// Cols returns the buffer width.
func (b *Buffer) Cols() int { return b.cols }

// Rows returns the buffer height.
func (b *Buffer) Rows() int { return b.rows }

// Modified returns true if the buffer has unsaved changes.
func (b *Buffer) Modified() bool { return b.modified }

// SetModified sets the modified state.
func (b *Buffer) SetModified(m bool) { b.modified = m }

// Path returns the buffer file path.
func (b *Buffer) Path() string { return b.path }

// SetPath sets the buffer file path.
func (b *Buffer) SetPath(p string) { b.path = p }

// Get returns the cell at (x, y), or BlankCell if out of bounds.
func (b *Buffer) Get(x, y int) BufferCell {
	if x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return BlankCell()
	}
	return b.cells[y][x]
}

// Set places a cell at (x, y) if within bounds.
func (b *Buffer) Set(x, y int, cell BufferCell) {
	if x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return
	}
	if cell.Rune == 0 {
		cell.Rune = ' '
	}
	b.cells[y][x] = cell
	b.modified = true
}

// Put writes a character with given styles at (x, y).
func (b *Buffer) Put(x, y int, r rune, fg, bg loom.Color, bold, dim, underline, invert bool) {
	if x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return
	}
	b.Set(x, y, BufferCell{
		Rune:      r,
		FG:        fg,
		BG:        bg,
		Bold:      bold,
		Dim:       dim,
		Underline: underline,
		Invert:    invert,
	})
}

// PutChar writes a character at (x, y) respecting the current edit mode.
// In insert mode, characters to the right on the same row are shifted right.
func (b *Buffer) PutChar(x, y int, r rune, fg, bg loom.Color, bold, dim, underline, invert bool, mode EditMode) {
	if x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return
	}
	cell := BufferCell{
		Rune:      r,
		FG:        fg,
		BG:        bg,
		Bold:      bold,
		Dim:       dim,
		Underline: underline,
		Invert:    invert,
	}
	if mode == ModeInsert {
		for i := b.cols - 1; i > x; i-- {
			b.cells[y][i] = b.cells[y][i-1]
		}
	}
	b.cells[y][x] = cell
	b.modified = true
}

// Erase resets the cell at (x, y) to blank.
func (b *Buffer) Erase(x, y int) {
	if x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return
	}
	b.cells[y][x] = BlankCell()
	b.modified = true
}

// Delete deletes the character at (x, y).
// In insert mode, shifts remaining characters on the line left by 1.
// In overtype mode, replaces with blank.
func (b *Buffer) Delete(x, y int, mode EditMode) {
	if x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return
	}
	if mode == ModeInsert {
		for i := x; i < b.cols-1; i++ {
			b.cells[y][i] = b.cells[y][i+1]
		}
		b.cells[y][b.cols-1] = BlankCell()
	} else {
		b.cells[y][x] = BlankCell()
	}
	b.modified = true
}

// Backspace handles backspace at cursor (x, y).
// Returns the new cursor x coordinate.
func (b *Buffer) Backspace(x, y int, mode EditMode) int {
	if x <= 0 || y < 0 || y >= b.rows {
		return x
	}
	newX := x - 1
	if mode == ModeInsert {
		for i := newX; i < b.cols-1; i++ {
			b.cells[y][i] = b.cells[y][i+1]
		}
		b.cells[y][b.cols-1] = BlankCell()
	} else {
		b.cells[y][newX] = BlankCell()
	}
	b.modified = true
	return newX
}

// Copy copies the cell at (x, y) to the internal clipboard.
func (b *Buffer) Copy(x, y int) BufferCell {
	c := b.Get(x, y)
	b.clipboard = c
	b.hasClip = true
	return c
}

// Cut copies the cell at (x, y) and erases it.
func (b *Buffer) Cut(x, y int) BufferCell {
	c := b.Copy(x, y)
	b.Erase(x, y)
	return c
}

// Paste pastes the clipboard cell at (x, y) if clipboard is non-empty.
func (b *Buffer) Paste(x, y int) bool {
	if !b.hasClip || x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return false
	}
	b.Set(x, y, b.clipboard)
	return true
}

// Clipboard returns the current clipboard content and whether it is populated.
func (b *Buffer) Clipboard() (BufferCell, bool) {
	return b.clipboard, b.hasClip
}

// Resize resizes the buffer grid, preserving existing cells.
func (b *Buffer) Resize(newCols, newRows int) {
	if newCols < 1 {
		newCols = 1
	}
	if newRows < 1 {
		newRows = 1
	}
	newCells := make([][]BufferCell, newRows)
	for y := 0; y < newRows; y++ {
		newCells[y] = make([]BufferCell, newCols)
		for x := 0; x < newCols; x++ {
			if y < b.rows && x < b.cols {
				newCells[y][x] = b.cells[y][x]
			} else {
				newCells[y][x] = BlankCell()
			}
		}
	}
	b.cols = newCols
	b.rows = newRows
	b.cells = newCells
}

// NextWord finds the next word boundary to the right on line y starting from x.
func (b *Buffer) NextWord(x, y int) int {
	if y < 0 || y >= b.rows {
		return x
	}
	col := x
	// If currently on a non-space, advance to space
	for col < b.cols && b.cells[y][col].Rune != ' ' && b.cells[y][col].Rune != 0 {
		col++
	}
	// Then advance past spaces to next non-space
	for col < b.cols && (b.cells[y][col].Rune == ' ' || b.cells[y][col].Rune == 0) {
		col++
	}
	if col >= b.cols {
		return b.cols - 1
	}
	return col
}

// PrevWord finds the previous word boundary to the left on line y starting from x.
func (b *Buffer) PrevWord(x, y int) int {
	if y < 0 || y >= b.rows || x <= 0 {
		return 0
	}
	col := x - 1
	// Skip trailing spaces
	for col > 0 && (b.cells[y][col].Rune == ' ' || b.cells[y][col].Rune == 0) {
		col--
	}
	// Walk back to beginning of word
	for col > 0 && b.cells[y][col-1].Rune != ' ' && b.cells[y][col-1].Rune != 0 {
		col--
	}
	if col < 0 {
		return 0
	}
	return col
}

// NextObjectRow finds the next non-empty block / object row below y.
func (b *Buffer) NextObjectRow(y int) int {
	if y >= b.rows-1 {
		return b.rows - 1
	}
	isRowEmpty := func(r int) bool {
		for x := 0; x < b.cols; x++ {
			if !b.cells[r][x].IsBlank() {
				return false
			}
		}
		return true
	}

	curEmpty := isRowEmpty(y)
	target := y + 1
	// If starting on non-empty, first find empty row
	if !curEmpty {
		for target < b.rows && !isRowEmpty(target) {
			target++
		}
	}
	// Then find next non-empty row
	for target < b.rows && isRowEmpty(target) {
		target++
	}
	if target >= b.rows {
		return b.rows - 1
	}
	return target
}

// PrevObjectRow finds the previous non-empty block / object row above y.
func (b *Buffer) PrevObjectRow(y int) int {
	if y <= 0 {
		return 0
	}
	isRowEmpty := func(r int) bool {
		for x := 0; x < b.cols; x++ {
			if !b.cells[r][x].IsBlank() {
				return false
			}
		}
		return true
	}

	target := y - 1
	// Skip empty rows upwards
	for target > 0 && isRowEmpty(target) {
		target--
	}
	// Find top of this non-empty block
	for target > 0 && !isRowEmpty(target-1) {
		target--
	}
	if target < 0 {
		return 0
	}
	return target
}

// ── ANSI Serializer ─────────────────────────────────────────────────────────

// Serialize converts the buffer contents to an ANSI string.
func (b *Buffer) Serialize() string {
	var buf strings.Builder
	var lastStyle BufferCell

	writeStyleTransition := func(target BufferCell) {
		if target.FG == lastStyle.FG &&
			target.BG == lastStyle.BG &&
			target.Bold == lastStyle.Bold &&
			target.Dim == lastStyle.Dim &&
			target.Underline == lastStyle.Underline &&
			target.Invert == lastStyle.Invert {
			return
		}

		if target.IsBlank() {
			buf.WriteString("\x1b[0m")
			lastStyle = BlankCell()
			return
		}

		buf.WriteString(target.Style().ANSI())
		lastStyle = target
	}

	for y := 0; y < b.rows; y++ {
		// Find last non-blank cell on this row to avoid trailing spaces
		lastNonBlank := -1
		for x := b.cols - 1; x >= 0; x-- {
			if !b.cells[y][x].IsBlank() {
				lastNonBlank = x
				break
			}
		}

		if lastNonBlank >= 0 {
			for x := 0; x <= lastNonBlank; x++ {
				c := b.cells[y][x]
				writeStyleTransition(c)
				r := c.Rune
				if r == 0 {
					r = ' '
				}
				buf.WriteRune(r)
			}
		}

		if lastStyle != BlankCell() {
			buf.WriteString("\x1b[0m")
			lastStyle = BlankCell()
		}

		if y < b.rows-1 {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

// SaveFile writes the ANSI buffer contents to path.
func (b *Buffer) SaveFile(path string) error {
	data := b.Serialize()
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		return fmt.Errorf("ansiedit: save file %s: %w", path, err)
	}
	b.path = path
	b.modified = false
	return nil
}

// ── ANSI Parser & Loader ───────────────────────────────────────────────────

// LoadBuffer reads an ANSI file and parses it into a 2D Buffer.
// If the file does not exist, an empty buffer with default dimensions is returned.
func LoadBuffer(path string, defaultCols, defaultRows int) (*Buffer, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		buf := NewBuffer(defaultCols, defaultRows)
		buf.path = path
		buf.modified = false
		return buf, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ansiedit: read %s: %w", path, err)
	}

	buf, err := ParseBuffer(string(data), defaultCols, defaultRows)
	if err != nil {
		return nil, err
	}
	buf.path = path
	buf.modified = false
	return buf, nil
}

// ParseBuffer parses an ANSI string into a 2D Buffer.
func ParseBuffer(content string, minCols, minRows int) (*Buffer, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	maxCols := minCols
	maxRows := max(minRows, len(lines))

	for _, line := range lines {
		w := measure.StringWidth(stripANSIEscapes(line))
		if w > maxCols {
			maxCols = w
		}
	}

	buf := NewBuffer(maxCols, maxRows)

	var curStyle BufferCell
	curStyle.Rune = ' '
	curStyle.FG = loom.ColorReset()
	curStyle.BG = loom.ColorReset()

	x, y := 0, 0
	rs := []rune(content)
	n := len(rs)

	for i := 0; i < n && y < maxRows; {
		// Escape sequences
		if rs[i] == '\x1b' && i+1 < n && rs[i+1] == '[' {
			j := i + 2
			for j < n && (rs[j] < '@' || rs[j] > '~') {
				j++
			}
			if j < n {
				params := string(rs[i+2 : j])
				cmd := rs[j]
				switch cmd {
				case 'm':
					curStyle = parseSGR(curStyle, params)
				case 'C':
					num := parseCSINum(params, 1)
					x += num
				case 'H', 'f':
					row, col := parseCSIPos(params)
					y = row - 1
					x = col - 1
				case 'G':
					x = parseCSINum(params, 1) - 1
				case 'd':
					y = parseCSINum(params, 1) - 1
				case 'J':
					if params == "2" || params == "3" {
						for cy := 0; cy < buf.rows; cy++ {
							for cx := 0; cx < buf.cols; cx++ {
								buf.cells[cy][cx] = BlankCell()
							}
						}
						x, y = 0, 0
					}
				}
				i = j + 1
				continue
			}
		}

		if rs[i] == '\x1b' && i+2 < n && rs[i+1] == '(' {
			i += 3
			continue
		}

		if rs[i] == '\r' {
			x = 0
			i++
			continue
		}

		if rs[i] == '\n' {
			y++
			x = 0
			i++
			continue
		}

		r := rs[i]
		w := measure.RuneWidth(r)
		if w > 0 && x < buf.cols && y < buf.rows {
			cell := curStyle
			cell.Rune = r
			buf.cells[y][x] = cell
			x += w
		}
		i++
	}

	return buf, nil
}

func parseCSINum(params string, def int) int {
	n, err := strconv.Atoi(params)
	if err != nil || n < 1 {
		return def
	}
	return n
}

func parseCSIPos(params string) (int, int) {
	parts := strings.Split(params, ";")
	r, c := 1, 1
	if len(parts) > 0 && parts[0] != "" {
		r = parseCSINum(parts[0], 1)
	}
	if len(parts) > 1 && parts[1] != "" {
		c = parseCSINum(parts[1], 1)
	}
	return r, c
}

func parseSGR(style BufferCell, params string) BufferCell {
	if params == "" || params == "0" {
		return BufferCell{
			Rune: ' ',
			FG:   loom.ColorReset(),
			BG:   loom.ColorReset(),
		}
	}

	parts := strings.Split(params, ";")
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		code := 0
		if p != "" {
			var err error
			code, err = strconv.Atoi(p)
			if err != nil {
				continue
			}
		}

		switch {
		case code == 0:
			style.FG = loom.ColorReset()
			style.BG = loom.ColorReset()
			style.Bold = false
			style.Dim = false
			style.Underline = false
			style.Invert = false
		case code == 1:
			style.Bold = true
		case code == 2:
			style.Dim = true
		case code == 4:
			style.Underline = true
		case code == 7:
			style.Invert = true
		case code == 22:
			style.Bold = false
			style.Dim = false
		case code == 24:
			style.Underline = false
		case code == 27:
			style.Invert = false
		case code >= 30 && code <= 37:
			style.FG = loom.ColorIndex(uint8(code - 30))
		case code >= 90 && code <= 97:
			style.FG = loom.ColorIndex(uint8(code - 90 + 8))
		case code >= 40 && code <= 47:
			style.BG = loom.ColorIndex(uint8(code - 40))
		case code >= 100 && code <= 107:
			style.BG = loom.ColorIndex(uint8(code - 100 + 8))
		case code == 39:
			style.FG = loom.ColorReset()
		case code == 49:
			style.BG = loom.ColorReset()
		case code == 38 && i+2 < len(parts) && parts[i+1] == "5":
			v, err := strconv.Atoi(parts[i+2])
			if err == nil && v >= 0 && v <= 255 {
				style.FG = loom.ColorIndex(uint8(v))
			}
			i += 2
		case code == 48 && i+2 < len(parts) && parts[i+1] == "5":
			v, err := strconv.Atoi(parts[i+2])
			if err == nil && v >= 0 && v <= 255 {
				style.BG = loom.ColorIndex(uint8(v))
			}
			i += 2
		case code == 38 && i+4 < len(parts) && parts[i+1] == "2":
			r, errR := strconv.Atoi(parts[i+2])
			g, errG := strconv.Atoi(parts[i+3])
			b, errB := strconv.Atoi(parts[i+4])
			if errR == nil && errG == nil && errB == nil &&
				r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
				style.FG = loom.ColorRGB(uint8(r), uint8(g), uint8(b))
			}
			i += 4
		case code == 48 && i+4 < len(parts) && parts[i+1] == "2":
			r, errR := strconv.Atoi(parts[i+2])
			g, errG := strconv.Atoi(parts[i+3])
			b, errB := strconv.Atoi(parts[i+4])
			if errR == nil && errG == nil && errB == nil &&
				r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
				style.BG = loom.ColorRGB(uint8(r), uint8(g), uint8(b))
			}
			i += 4
		}
	}
	return style
}

func stripANSIEscapes(s string) string {
	var b bytes.Buffer
	inEsc := false
	for _, r := range s {
		if inEsc {
			if (r >= '@' && r <= '~') || r == 'm' {
				inEsc = false
			}
			continue
		}
		if r == '\x1b' {
			inEsc = true
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
