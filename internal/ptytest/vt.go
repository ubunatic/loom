// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package ptytest is a deliberately small PTY test harness: a Session that
// runs a real binary on a pseudo-terminal, and a minimal VT that replays the
// captured bytes into a screen grid so tests can assert what a terminal would
// actually show (final screen, and the screen at every synchronized-output
// frame boundary).
//
// The VT implements only what Loom emits: printable text, CR/LF, CUP, CUU/CUD/
// CUF/CUB, EL, ED, SGR (ignored), and the ?7 (auto-wrap), ?25 (cursor) and
// ?2026 (synchronized output) private modes. Anything else is ignored.
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

	cells    [][]rune // 0 marks the trailing half of a wide rune
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

func (v *VT) blank(cols, rows int) [][]rune {
	g := make([][]rune, rows)
	for i := range g {
		g[i] = make([]rune, cols)
		for j := range g[i] {
			g[i][j] = ' '
		}
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
		for _, r := range row {
			if r != 0 {
				b.WriteRune(r)
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
	v.cells = append(v.cells[1:], v.blank(v.Cols, 1)[0])
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
	if v.x < 0 {
		return
	}
	v.cells[v.y][v.x] = r
	if w == 2 && v.x+1 < v.Cols {
		v.cells[v.y][v.x+1] = 0
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
	var out []int
	for _, f := range strings.Split(s, ";") {
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
			}
		}
	}
}

func (v *VT) eraseLine(mode int) {
	from, to := v.x, v.Cols
	switch mode {
	case 1:
		from, to = 0, v.x+1
	case 2:
		from, to = 0, v.Cols
	}
	for x := from; x < min(to, v.Cols); x++ {
		v.cells[v.y][x] = ' '
	}
}

func (v *VT) eraseDisplay(mode int) {
	switch mode {
	case 0:
		v.eraseLine(0)
		for y := v.y + 1; y < v.Rows; y++ {
			copy(v.cells[y], v.blank(v.Cols, 1)[0])
		}
	case 2, 3:
		v.cells = v.blank(v.Cols, v.Rows)
	}
}
