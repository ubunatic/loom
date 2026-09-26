// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"codeberg.org/ubunatic/loom/measure"
)

// AnsiEditMode represents character insertion vs overtype behavior in an ANSI buffer.
type AnsiEditMode int

const (
	AnsiModeOvertype AnsiEditMode = iota
	AnsiModeInsert
)

// Aliases for convenience and compatibility.
const (
	ModeOvertype = AnsiModeOvertype
	ModeInsert   = AnsiModeInsert
)

func (m AnsiEditMode) String() string {
	switch m {
	case AnsiModeInsert:
		return "Insert"
	default:
		return "Overtype"
	}
}

// AnsiCell represents a single styled character cell in a 2D ANSI buffer grid.
type AnsiCell struct {
	Rune      rune
	FG        Color
	BG        Color
	Bold      bool
	Dim       bool
	Underline bool
	Invert    bool
}

// BlankAnsiCell returns a default empty cell with reset styling.
func BlankAnsiCell() AnsiCell {
	return AnsiCell{
		Rune: ' ',
		FG:   ColorReset(),
		BG:   ColorReset(),
	}
}

// IsBlank reports whether the cell contains whitespace and default colors without attributes.
func (c AnsiCell) IsBlank() bool {
	return (c.Rune == 0 || c.Rune == ' ') &&
		c.FG == ColorReset() &&
		c.BG == ColorReset() &&
		!c.Bold && !c.Dim && !c.Underline && !c.Invert
}

// Style converts AnsiCell formatting to Style.
func (c AnsiCell) Style() Style {
	fg := c.FG
	bg := c.BG
	if c.Invert {
		fg, bg = bg, fg
	}
	return Style{
		FG:        fg,
		BG:        bg,
		Bold:      c.Bold,
		Dim:       c.Dim,
		Underline: c.Underline,
	}
}

// ToCell converts AnsiCell to Cell.
func (c AnsiCell) ToCell() Cell {
	r := c.Rune
	if r == 0 {
		r = ' '
	}
	return Cell{
		Text:  string(r),
		Style: c.Style(),
	}
}

// ToLoomCell is an alias for ToCell for backwards compatibility.
func (c AnsiCell) ToLoomCell() Cell {
	return c.ToCell()
}

// AnsiBuffer is an in-memory 2D cell grid for ANSI art editing and terminal graphics.
type AnsiBuffer struct {
	cols      int
	rows      int
	cells     [][]AnsiCell
	clipboard AnsiCell
	hasClip   bool
	modified  bool
	path      string
}

// NewAnsiBuffer creates an empty ANSI buffer with the specified columns and rows.
func NewAnsiBuffer(cols, rows int) *AnsiBuffer {
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	buf := &AnsiBuffer{
		cols:  cols,
		rows:  rows,
		cells: make([][]AnsiCell, rows),
	}
	for y := 0; y < rows; y++ {
		buf.cells[y] = make([]AnsiCell, cols)
		for x := 0; x < cols; x++ {
			buf.cells[y][x] = BlankAnsiCell()
		}
	}
	return buf
}

// Cols returns the buffer width.
func (b *AnsiBuffer) Cols() int { return b.cols }

// Rows returns the buffer height.
func (b *AnsiBuffer) Rows() int { return b.rows }

// Modified returns true if the buffer has unsaved changes.
func (b *AnsiBuffer) Modified() bool { return b.modified }

// SetModified sets the modified state.
func (b *AnsiBuffer) SetModified(m bool) { b.modified = m }

// Path returns the buffer file path.
func (b *AnsiBuffer) Path() string { return b.path }

// SetPath sets the buffer file path.
func (b *AnsiBuffer) SetPath(p string) { b.path = p }

// Get returns the cell at (x, y), or BlankAnsiCell if out of bounds.
func (b *AnsiBuffer) Get(x, y int) AnsiCell {
	if x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return BlankAnsiCell()
	}
	return b.cells[y][x]
}

// Set places a cell at (x, y) if within bounds.
func (b *AnsiBuffer) Set(x, y int, cell AnsiCell) {
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
func (b *AnsiBuffer) Put(x, y int, r rune, fg, bg Color, bold, dim, underline, invert bool) {
	if x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return
	}
	b.Set(x, y, AnsiCell{
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
func (b *AnsiBuffer) PutChar(x, y int, r rune, fg, bg Color, bold, dim, underline, invert bool, mode AnsiEditMode) {
	if x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return
	}
	cell := AnsiCell{
		Rune:      r,
		FG:        fg,
		BG:        bg,
		Bold:      bold,
		Dim:       dim,
		Underline: underline,
		Invert:    invert,
	}
	if mode == AnsiModeInsert {
		for i := b.cols - 1; i > x; i-- {
			b.cells[y][i] = b.cells[y][i-1]
		}
	}
	b.cells[y][x] = cell
	b.modified = true
}

// Erase resets the cell at (x, y) to blank.
func (b *AnsiBuffer) Erase(x, y int) {
	if x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return
	}
	b.cells[y][x] = BlankAnsiCell()
	b.modified = true
}

// Delete deletes the character at (x, y).
// In insert mode, shifts remaining characters on the line left by 1.
// In overtype mode, replaces with blank.
func (b *AnsiBuffer) Delete(x, y int, mode AnsiEditMode) {
	if x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return
	}
	if mode == AnsiModeInsert {
		for i := x; i < b.cols-1; i++ {
			b.cells[y][i] = b.cells[y][i+1]
		}
		b.cells[y][b.cols-1] = BlankAnsiCell()
	} else {
		b.cells[y][x] = BlankAnsiCell()
	}
	b.modified = true
}

// Backspace handles backspace at cursor (x, y).
// Returns the new cursor x coordinate.
func (b *AnsiBuffer) Backspace(x, y int, mode AnsiEditMode) int {
	if x <= 0 || y < 0 || y >= b.rows {
		return x
	}
	newX := x - 1
	if mode == AnsiModeInsert {
		for i := newX; i < b.cols-1; i++ {
			b.cells[y][i] = b.cells[y][i+1]
		}
		b.cells[y][b.cols-1] = BlankAnsiCell()
	} else {
		b.cells[y][newX] = BlankAnsiCell()
	}
	b.modified = true
	return newX
}

// Copy copies the cell at (x, y) to the internal clipboard.
func (b *AnsiBuffer) Copy(x, y int) AnsiCell {
	c := b.Get(x, y)
	b.clipboard = c
	b.hasClip = true
	return c
}

// Cut copies the cell at (x, y) and erases it.
func (b *AnsiBuffer) Cut(x, y int) AnsiCell {
	c := b.Copy(x, y)
	b.Erase(x, y)
	return c
}

// Paste pastes the clipboard cell at (x, y) if clipboard is non-empty.
func (b *AnsiBuffer) Paste(x, y int) bool {
	if !b.hasClip || x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return false
	}
	b.Set(x, y, b.clipboard)
	return true
}

// Clipboard returns the current clipboard content and whether it is populated.
func (b *AnsiBuffer) Clipboard() (AnsiCell, bool) {
	return b.clipboard, b.hasClip
}

// SetClipboard sets the buffer's clipboard cell directly.
func (b *AnsiBuffer) SetClipboard(cell AnsiCell) {
	b.clipboard = cell
	b.hasClip = true
}

// Resize resizes the buffer grid, preserving existing cells.
func (b *AnsiBuffer) Resize(newCols, newRows int) {
	if newCols < 1 {
		newCols = 1
	}
	if newRows < 1 {
		newRows = 1
	}
	newCells := make([][]AnsiCell, newRows)
	for y := 0; y < newRows; y++ {
		newCells[y] = make([]AnsiCell, newCols)
		for x := 0; x < newCols; x++ {
			if y < b.rows && x < b.cols {
				newCells[y][x] = b.cells[y][x]
			} else {
				newCells[y][x] = BlankAnsiCell()
			}
		}
	}
	b.cols = newCols
	b.rows = newRows
	b.cells = newCells
}

// NextWord finds the next word boundary to the right on line y starting from x.
func (b *AnsiBuffer) NextWord(x, y int) int {
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
func (b *AnsiBuffer) PrevWord(x, y int) int {
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
func (b *AnsiBuffer) NextObjectRow(y int) int {
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
func (b *AnsiBuffer) PrevObjectRow(y int) int {
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

// Serialize converts the buffer contents to an ANSI formatted string.
func (b *AnsiBuffer) Serialize() string {
	var buf strings.Builder
	var lastStyle AnsiCell

	writeStyleTransition := func(target AnsiCell) {
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
			lastStyle = BlankAnsiCell()
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

		if lastStyle != BlankAnsiCell() {
			buf.WriteString("\x1b[0m")
			lastStyle = BlankAnsiCell()
		}

		if y < b.rows-1 {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

// SerializeAnsiBuffer converts an AnsiBuffer to an ANSI formatted string.
func SerializeAnsiBuffer(buf *AnsiBuffer) string {
	if buf == nil {
		return ""
	}
	return buf.Serialize()
}

// SaveFile writes the ANSI buffer contents to path.
func (b *AnsiBuffer) SaveFile(path string) error {
	data := b.Serialize()
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		return fmt.Errorf("ansibuffer: save file %s: %w", path, err)
	}
	b.path = path
	b.modified = false
	return nil
}

// SaveAnsiBuffer writes an AnsiBuffer to path.
func SaveAnsiBuffer(buf *AnsiBuffer, path string) error {
	if buf == nil {
		return fmt.Errorf("ansibuffer: nil buffer")
	}
	return buf.SaveFile(path)
}

// ── ANSI Parser & Loader ───────────────────────────────────────────────────

// LoadAnsiBuffer reads an ANSI file and parses it into a 2D AnsiBuffer.
// If the file does not exist, an empty buffer with default dimensions is returned.
func LoadAnsiBuffer(path string, defaultCols, defaultRows int) (*AnsiBuffer, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		buf := NewAnsiBuffer(defaultCols, defaultRows)
		buf.path = path
		buf.modified = false
		return buf, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ansibuffer: read %s: %w", path, err)
	}

	buf, err := ParseAnsiBuffer(string(data), defaultCols, defaultRows)
	if err != nil {
		return nil, err
	}
	buf.path = path
	buf.modified = false
	return buf, nil
}

// ParseAnsiBuffer parses an ANSI string into a 2D AnsiBuffer.
func ParseAnsiBuffer(content string, minCols, minRows int) (*AnsiBuffer, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	maxCols := minCols
	maxRows := max(minRows, len(lines))

	for _, line := range lines {
		w := measure.StringWidth(stripANSI(line))
		if w > maxCols {
			maxCols = w
		}
	}

	buf := NewAnsiBuffer(maxCols, maxRows)

	var curStyle AnsiCell
	curStyle.Rune = ' '
	curStyle.FG = ColorReset()
	curStyle.BG = ColorReset()

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
					curStyle = parseAnsiSGR(curStyle, params)
				case 'C':
					num := parseAnsiCSINum(params, 1)
					x += num
				case 'H', 'f':
					row, col := parseAnsiCSIPos(params)
					y = row - 1
					x = col - 1
				case 'G':
					x = parseAnsiCSINum(params, 1) - 1
				case 'd':
					y = parseAnsiCSINum(params, 1) - 1
				case 'J':
					if params == "2" || params == "3" {
						for cy := 0; cy < buf.rows; cy++ {
							for cx := 0; cx < buf.cols; cx++ {
								buf.cells[cy][cx] = BlankAnsiCell()
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

func parseAnsiCSINum(params string, def int) int {
	n, err := strconv.Atoi(params)
	if err != nil || n < 1 {
		return def
	}
	return n
}

func parseAnsiCSIPos(params string) (int, int) {
	parts := strings.Split(params, ";")
	r, c := 1, 1
	if len(parts) > 0 && parts[0] != "" {
		r = parseAnsiCSINum(parts[0], 1)
	}
	if len(parts) > 1 && parts[1] != "" {
		c = parseAnsiCSINum(parts[1], 1)
	}
	return r, c
}

func parseAnsiSGR(style AnsiCell, params string) AnsiCell {
	if params == "" || params == "0" {
		return AnsiCell{
			Rune: ' ',
			FG:   ColorReset(),
			BG:   ColorReset(),
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
			style.FG = ColorReset()
			style.BG = ColorReset()
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
			style.FG = ColorIndex(uint8(code - 30))
		case code >= 90 && code <= 97:
			style.FG = ColorIndex(uint8(code - 90 + 8))
		case code >= 40 && code <= 47:
			style.BG = ColorIndex(uint8(code - 40))
		case code >= 100 && code <= 107:
			style.BG = ColorIndex(uint8(code - 100 + 8))
		case code == 39:
			style.FG = ColorReset()
		case code == 49:
			style.BG = ColorReset()
		case code == 38 && i+2 < len(parts) && parts[i+1] == "5":
			v, err := strconv.Atoi(parts[i+2])
			if err == nil && v >= 0 && v <= 255 {
				style.FG = ColorIndex(uint8(v))
			}
			i += 2
		case code == 48 && i+2 < len(parts) && parts[i+1] == "5":
			v, err := strconv.Atoi(parts[i+2])
			if err == nil && v >= 0 && v <= 255 {
				style.BG = ColorIndex(uint8(v))
			}
			i += 2
		case code == 38 && i+4 < len(parts) && parts[i+1] == "2":
			r, errR := strconv.Atoi(parts[i+2])
			g, errG := strconv.Atoi(parts[i+3])
			b, errB := strconv.Atoi(parts[i+4])
			if errR == nil && errG == nil && errB == nil &&
				r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
				style.FG = ColorRGB(uint8(r), uint8(g), uint8(b))
			}
			i += 4
		case code == 48 && i+4 < len(parts) && parts[i+1] == "2":
			r, errR := strconv.Atoi(parts[i+2])
			g, errG := strconv.Atoi(parts[i+3])
			b, errB := strconv.Atoi(parts[i+4])
			if errR == nil && errG == nil && errB == nil &&
				r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
				style.BG = ColorRGB(uint8(r), uint8(g), uint8(b))
			}
			i += 4
		}
	}
	return style
}
