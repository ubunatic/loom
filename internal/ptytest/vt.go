// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package ptytest is a deliberately small PTY test harness: a Session that
// runs a real binary on a pseudo-terminal, and a minimal VT that replays the
// captured bytes into a screen grid so tests can assert what a terminal would
// actually show (final screen, and the screen at every synchronized-output
// frame boundary).
//
// The VT implements only what Loom emits: printable text, CR/LF, CUP, CUU/CUD/
// CUF/CUB, EL, ED, SGR (per-cell colors and styles), and the ?7 (auto-wrap),
// ?25 (cursor) and ?2026 (synchronized output) private modes. Anything else
// is ignored.
package ptytest

import (
	"strings"
	"unicode/utf8"

	"codeberg.org/ubunatic/loom/measure"
)

// VT is a minimal terminal emulator. It is not safe for concurrent use.
type VT struct {
	Cols, Rows int
	// CursorVisible reflects the last ?25 mode.
	CursorVisible bool
	// Frames holds a screen snapshot for every completed ?2026 frame.
	Frames [][]string
	// CellFrames holds a cell grid snapshot for every completed ?2026 frame.
	CellFrames [][][]Cell

	cells    [][]Cell // Rune=0 marks the trailing half of a wide rune
	pen      Style
	x, y     int
	autoWrap bool
	pending  []byte // incomplete escape or UTF-8 tail carried between Write calls
}

// NewVT returns a blank VT of the given size with auto-wrap enabled.
func NewVT(cols, rows int) *VT {
	v := &VT{Cols: cols, Rows: rows, autoWrap: true, CursorVisible: true}
	v.cells = v.blank(cols, rows)
	return v
}

func (v *VT) blankRow(cols int, s Style) []Cell {
	row := make([]Cell, cols)
	for i := range row {
		row[i] = Cell{Rune: ' ', Style: s}
	}
	return row
}

func (v *VT) blank(cols, rows int) [][]Cell {
	g := make([][]Cell, rows)
	for i := range g {
		g[i] = v.blankRow(cols, Style{})
	}
	return g
}

// Resize changes the grid size, keeping the top-left content, like a terminal
// that does not reflow.
func (v *VT) Resize(cols, rows int) {
	g := v.blank(cols, rows)
	for y := 0; y < min(rows, v.Rows); y++ {
		copy(g[y], v.cells[y][:min(cols, v.Cols)])
	}
	v.cells, v.Cols, v.Rows = g, cols, rows
	v.x, v.y = min(v.x, cols-1), min(v.y, rows-1)
}

// Screen returns the visible rows with trailing blanks trimmed.
func (v *VT) Screen() []string {
	out := make([]string, v.Rows)
	for y, row := range v.cells {
		var b strings.Builder
		for _, cell := range row {
			if cell.Rune != 0 {
				b.WriteRune(cell.Rune)
			}
		}
		out[y] = strings.TrimRight(b.String(), " ")
	}
	return out
}

// Text returns Screen joined by newlines.
func (v *VT) Text() string { return strings.Join(v.Screen(), "\n") }

// Cursor returns the zero-based cursor position.
func (v *VT) Cursor() (x, y int) { return v.x, v.y }

// Pen returns the current SGR pen style.
func (v *VT) Pen() Style { return v.pen }

// Cells returns a snapshot of the current cell grid.
func (v *VT) Cells() [][]Cell {
	out := make([][]Cell, v.Rows)
	for y := range v.cells {
		out[y] = make([]Cell, v.Cols)
		copy(out[y], v.cells[y])
	}
	return out
}

// Cell returns the cell at (x, y), or a zero Cell if out of bounds.
func (v *VT) Cell(x, y int) Cell {
	if x < 0 || x >= v.Cols || y < 0 || y >= v.Rows {
		return Cell{}
	}
	return v.cells[y][x]
}

// FrameCells returns a copy of the cell snapshots taken at each ?2026 frame end.
func (v *VT) FrameCells() [][][]Cell {
	out := make([][][]Cell, len(v.CellFrames))
	for i, frame := range v.CellFrames {
		out[i] = make([][]Cell, len(frame))
		for y, row := range frame {
			out[i][y] = make([]Cell, len(row))
			copy(out[i][y], row)
		}
	}
	return out
}

// Write feeds terminal output into the VT.
func (v *VT) Write(p []byte) (int, error) {
	data := append(v.pending, p...)
	v.pending = nil
	for i := 0; i < len(data); {
		b := data[i]
		switch {
		case b == 0x1b:
			n, ok := v.escape(data[i:])
			if !ok {
				v.pending = append([]byte(nil), data[i:]...)
				return len(p), nil
			}
			i += n
		case b == '\r':
			v.x = 0
			i++
		case b == '\n':
			v.lineFeed()
			i++
		case b < 0x20:
			i++
		default:
			if !utf8.FullRune(data[i:]) {
				v.pending = append([]byte(nil), data[i:]...)
				return len(p), nil
			}
			r, n := utf8.DecodeRune(data[i:])
			v.put(r)
			i += n
		}
	}
	return len(p), nil
}

func (v *VT) lineFeed() {
	if v.y < v.Rows-1 {
		v.y++
		return
	}
	v.cells = append(v.cells[1:], v.blankRow(v.Cols, Style{}))
}

func (v *VT) put(r rune) {
	w := max(1, measure.RuneWidth(r))
	if v.x+w > v.Cols {
		if v.autoWrap {
			v.x = 0
			v.lineFeed()
		} else {
			v.x = v.Cols - w
		}
	}
	if v.x < 0 || v.x >= v.Cols || v.y < 0 || v.y >= v.Rows {
		return
	}
	v.cells[v.y][v.x] = Cell{Rune: r, Style: v.pen}
	if w == 2 && v.x+1 < v.Cols {
		v.cells[v.y][v.x+1] = Cell{Rune: 0, Style: v.pen}
	}
	v.x += w
}

// escape consumes one escape sequence from data. ok is false when data ends
// before the sequence does.
func (v *VT) escape(data []byte) (n int, ok bool) {
	if len(data) < 2 {
		return 0, false
	}
	if data[1] != '[' {
		return 2, true // two-byte escape such as ESC 7; ignored
	}
	for i := 2; i < len(data); i++ {
		if data[i] >= 0x40 && data[i] <= 0x7e {
			v.csi(string(data[2:i]), data[i])
			return i + 1, true
		}
	}
	return 0, false
}

func csiParams(s string) []int {
	s = strings.TrimPrefix(s, "?")
	if s == "" {
		return []int{0}
	}
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ';' || r == ':'
	})
	if len(fields) == 0 {
		return []int{0}
	}
	var out []int
	for _, f := range fields {
		n := 0
		for _, c := range f {
			if c >= '0' && c <= '9' {
				n = n*10 + int(c-'0')
			}
		}
		out = append(out, n)
	}
	return out
}

func (v *VT) csi(params string, final byte) {
	p := csiParams(params)
	arg := func(i, def int) int {
		if i < len(p) && p[i] > 0 {
			return p[i]
		}
		return def
	}
	switch final {
	case 'm':
		v.sgr(params)
	case 'H', 'f':
		v.y = min(max(arg(0, 1)-1, 0), v.Rows-1)
		v.x = min(max(arg(1, 1)-1, 0), v.Cols-1)
	case 'A':
		v.y = max(v.y-arg(0, 1), 0)
	case 'B':
		v.y = min(v.y+arg(0, 1), v.Rows-1)
	case 'C':
		v.x = min(v.x+arg(0, 1), v.Cols-1)
	case 'D':
		v.x = max(v.x-arg(0, 1), 0)
	case 'K':
		v.eraseLine(p[0])
	case 'J':
		v.eraseDisplay(p[0])
	case 'h', 'l':
		on := final == 'h'
		if !strings.HasPrefix(params, "?") {
			return
		}
		switch p[0] {
		case 7:
			v.autoWrap = on
		case 25:
			v.CursorVisible = on
		case 2026:
			if !on {
				v.Frames = append(v.Frames, v.Screen())
				v.CellFrames = append(v.CellFrames, v.Cells())
			}
		}
	}
}

func (v *VT) sgr(params string) {
	p := csiParams(params)
	for i := 0; i < len(p); i++ {
		code := p[i]
		switch {
		case code == 0:
			v.pen = Style{}
		case code == 1:
			v.pen.Bold = true
		case code == 2:
			v.pen.Dim = true
		case code == 3:
			v.pen.Italic = true
		case code == 4:
			v.pen.Underline = true
		case code == 5 || code == 6:
			v.pen.Blink = true
		case code == 7:
			v.pen.Reverse = true
		case code == 8:
			v.pen.Hidden = true
		case code == 9:
			v.pen.Strike = true
		case code == 21:
			v.pen.Bold = false
		case code == 22:
			v.pen.Bold = false
			v.pen.Dim = false
		case code == 23:
			v.pen.Italic = false
		case code == 24:
			v.pen.Underline = false
		case code == 25:
			v.pen.Blink = false
		case code == 27:
			v.pen.Reverse = false
		case code == 28:
			v.pen.Hidden = false
		case code == 29:
			v.pen.Strike = false
		case code >= 30 && code <= 37:
			v.pen.FG = ColorIndex(uint8(code - 30))
		case code == 38:
			if i+1 < len(p) {
				switch p[i+1] {
				case 5: // 38;5;n
					if i+2 < len(p) {
						v.pen.FG = ColorIndex(uint8(p[i+2]))
						i += 2
					}
				case 2: // 38;2;r;g;b
					if i+4 < len(p) {
						v.pen.FG = ColorRGB(uint8(p[i+2]), uint8(p[i+3]), uint8(p[i+4]))
						i += 4
					}
				}
			}
		case code == 39:
			v.pen.FG = ColorReset()
		case code >= 40 && code <= 47:
			v.pen.BG = ColorIndex(uint8(code - 40))
		case code == 48:
			if i+1 < len(p) {
				switch p[i+1] {
				case 5: // 48;5;n
					if i+2 < len(p) {
						v.pen.BG = ColorIndex(uint8(p[i+2]))
						i += 2
					}
				case 2: // 48;2;r;g;b
					if i+4 < len(p) {
						v.pen.BG = ColorRGB(uint8(p[i+2]), uint8(p[i+3]), uint8(p[i+4]))
						i += 4
					}
				}
			}
		case code == 49:
			v.pen.BG = ColorReset()
		case code >= 90 && code <= 97:
			v.pen.FG = ColorIndex(uint8(code - 90 + 8))
		case code >= 100 && code <= 107:
			v.pen.BG = ColorIndex(uint8(code - 100 + 8))
		}
	}
}

func (v *VT) eraseLine(mode int) {
	if v.y < 0 || v.y >= v.Rows {
		return
	}
	from, to := v.x, v.Cols
	switch mode {
	case 1:
		from, to = 0, v.x+1
	case 2:
		from, to = 0, v.Cols
	}
	eraseStyle := Style{BG: v.pen.BG}
	for x := max(0, from); x < min(to, v.Cols); x++ {
		v.cells[v.y][x] = Cell{Rune: ' ', Style: eraseStyle}
	}
}

func (v *VT) eraseDisplay(mode int) {
	eraseStyle := Style{BG: v.pen.BG}
	switch mode {
	case 0:
		v.eraseLine(0)
		for y := v.y + 1; y < v.Rows; y++ {
			v.cells[y] = v.blankRow(v.Cols, eraseStyle)
		}
	case 1:
		for y := 0; y < v.y; y++ {
			v.cells[y] = v.blankRow(v.Cols, eraseStyle)
		}
		v.eraseLine(1)
	case 2, 3:
		for y := 0; y < v.Rows; y++ {
			v.cells[y] = v.blankRow(v.Cols, eraseStyle)
		}
	}
}
