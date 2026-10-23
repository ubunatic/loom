// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package vt

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"ubunatic.com/loom/measure"
)

const maxPendingBytes = 1024

// VT is a minimal terminal emulator. It is not safe for concurrent use.
type VT struct {
	Cols, Rows int
	// CursorVisible reflects the last ?25 mode.
	CursorVisible bool
	// MouseMode reflects the active mouse tracking mode (9, 1000, 1002, 1003, or 0 if off).
	MouseMode int
	// MouseSGR reports whether SGR mouse encoding (1006) is active.
	MouseSGR bool
	// MouseUTF8 reports whether UTF-8 mouse encoding (1005) is active.
	MouseUTF8 bool
	// MouseURXVT reports whether URXVT mouse encoding (1015) is active.
	MouseURXVT bool
	// FocusEvents reports whether focus event reporting (1004) is active.
	FocusEvents bool
	// BracketedPaste reports whether bracketed paste mode (2004) is active.
	BracketedPaste bool
	// CursorKeysApp reports whether application cursor keys (DECCKM ?1) is active.
	CursorKeysApp bool
	// KeypadApp reports whether keypad application mode (DECKPAM \x1b=) is active.
	KeypadApp bool
	// KittyKeyboard reports whether the Kitty keyboard protocol is active.
	KittyKeyboard bool
	// Frames holds a screen snapshot for every completed ?2026 frame.
	Frames [][]string
	// CellFrames holds a cell grid snapshot for every completed ?2026 frame.
	CellFrames [][][]Cell

	cells           [][]Cell // Rune=0 marks the trailing half of a wide rune
	scrollback      [][]Cell
	scrollbackLimit int
	scrollbackStart int
	pen             Style
	x, y            int
	autoWrap        bool
	pending         []byte // incomplete escape or UTF-8 tail carried between Write calls
	discardControl  byte   // unterminated control type discarded after the safety cap
	main            *terminalState
	saved           cursorState
}

type cursorState struct {
	x, y int
	pen  Style
}

type terminalState struct {
	cells [][]Cell
	cursorState
}

// AltScreen reports whether DEC mode 1049 is active.
func (v *VT) AltScreen() bool { return v.main != nil }

func (v *VT) alternate(on bool) {
	if on == v.AltScreen() {
		return
	}
	if on {
		v.main = &terminalState{cells: v.cells, cursorState: cursorState{v.x, v.y, v.pen}}
		v.cells = v.blank(v.Cols, v.Rows)
		v.x, v.y = 0, 0
	} else {
		v.cells = v.main.cells
		v.x, v.y, v.pen = v.main.x, v.main.y, v.main.pen
		v.main = nil
	}
}

// NewVT returns a blank VT of the given size with auto-wrap enabled.
func NewVT(cols, rows int) *VT {
	cols, rows = max(1, cols), max(1, rows)
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
	cols, rows = max(1, cols), max(1, rows)
	resize := func(old [][]Cell) [][]Cell {
		grid := v.blank(cols, rows)
		for y := range min(rows, v.Rows) {
			copy(grid[y], old[y][:min(cols, v.Cols)])
		}
		return grid
	}
	v.cells = resize(v.cells)
	for i, row := range v.scrollback {
		resized := v.blankRow(cols, Style{})
		copy(resized, row[:min(cols, len(row))])
		v.scrollback[i] = resized
	}
	if v.main != nil {
		v.main.cells = resize(v.main.cells)
		v.main.x, v.main.y = min(v.main.x, cols-1), min(v.main.y, rows-1)
	}
	v.Cols, v.Rows = cols, rows
	v.x, v.y = min(v.x, cols-1), min(v.y, rows-1)
	v.saved.x, v.saved.y = min(v.saved.x, cols-1), min(v.saved.y, rows-1)
}

// SetScrollbackLimit enables a bounded main-screen history. Setting zero
// disables history and clears rows already collected.
func (v *VT) SetScrollbackLimit(limit int) {
	limit = max(0, limit)
	if limit != v.scrollbackLimit && limit > 0 && len(v.scrollback) > 0 {
		rows := v.orderedScrollback()
		if len(rows) > limit {
			rows = rows[len(rows)-limit:]
		}
		v.scrollback = cloneRows(rows)
		v.scrollbackStart = 0
	}
	v.scrollbackLimit = limit
	if limit == 0 {
		v.scrollback = nil
		v.scrollbackStart = 0
		return
	}
}

// ScrollbackCells returns a copy of the main-screen history followed by the
// visible screen. While the alternate screen is active, it returns only that
// visible screen because alternate-screen rows are not part of history.
func (v *VT) ScrollbackCells() [][]Cell {
	if v.AltScreen() {
		return cloneRows(v.cells)
	}
	combined := make([][]Cell, 0, len(v.scrollback)+len(v.cells))
	combined = append(combined, v.orderedScrollback()...)
	combined = append(combined, v.cells...)
	return cloneRows(combined)
}

func (v *VT) orderedScrollback() [][]Cell {
	if v.scrollbackStart == 0 {
		return v.scrollback
	}
	rows := make([][]Cell, 0, len(v.scrollback))
	rows = append(rows, v.scrollback[v.scrollbackStart:]...)
	rows = append(rows, v.scrollback[:v.scrollbackStart]...)
	return rows
}

func cloneRows(rows [][]Cell) [][]Cell {
	cloned := make([][]Cell, len(rows))
	for i, row := range rows {
		cloned[i] = append([]Cell(nil), row...)
	}
	return cloned
}

// Screen returns the visible rows with trailing blanks trimmed.
func (v *VT) Screen() []string {
	out := make([]string, v.Rows)
	for y, row := range v.cells {
		var b strings.Builder
		for _, cell := range row {
			if cell.Rune != 0 {
				if cell.Text != "" {
					b.WriteString(cell.Text)
				} else {
					b.WriteRune(cell.Rune)
				}
			}
		}
		out[y] = strings.TrimRight(b.String(), " ")
	}
	return out
}

// Text returns Screen joined by newlines.
func (v *VT) Text() string { return strings.Join(v.Screen(), "\n") }

// ResetParser discards incomplete input without changing terminal state.
func (v *VT) ResetParser() {
	v.pending = nil
	v.discardControl = 0
}

// WriteFramed feeds p to the VT and returns bytes ending at the last complete
// parser boundary. An incomplete escape or UTF-8 tail remains buffered.
func (v *VT) WriteFramed(p []byte) ([]byte, error) {
	data := append(append([]byte(nil), v.pending...), p...)
	discarding := v.discardControl != 0
	if _, err := v.Write(p); err != nil {
		return nil, err
	}
	if discarding || v.discardControl != 0 {
		return nil, nil
	}
	complete := len(data) - len(v.pending)
	if complete < 0 {
		complete = 0
	}
	return append([]byte(nil), data[:complete]...), nil
}

// ANSI serializes the visible screen cells to ANSI SGR escape sequences and text.
// Trailing blanks on each row and trailing blank rows on the screen are trimmed.
// Each row is terminated by a newline.
func (v *VT) ANSI() string {
	lastRow := -1
	for y := len(v.cells) - 1; y >= 0; y-- {
		row := v.cells[y]
		for x := 0; x < len(row); x++ {
			if !isTrailingBlank(row[x]) {
				lastRow = y
				break
			}
		}
		if lastRow >= 0 {
			break
		}
	}
	if lastRow < 0 {
		return ""
	}

	var b strings.Builder
	for y := 0; y <= lastRow; y++ {
		row := v.cells[y]
		last := -1
		for x := len(row) - 1; x >= 0; x-- {
			if !isTrailingBlank(row[x]) {
				last = x
				break
			}
		}
		if last >= 0 {
			curStyle := Style{}
			for x := 0; x <= last; x++ {
				cell := row[x]
				if cell.Rune == 0 && cell.Text == "" {
					// Trailing half of wide character, already output
					continue
				}
				if cell.Style != curStyle {
					b.WriteString(cell.Style.SGR())
					curStyle = cell.Style
				}
				if cell.Text != "" {
					b.WriteString(cell.Text)
				} else if cell.Rune != 0 {
					b.WriteRune(cell.Rune)
				} else {
					b.WriteByte(' ')
				}
			}
			if curStyle != (Style{}) {
				b.WriteString("\x1b[0m")
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// Repaint returns an ANSI escape sequence string that completely redraws the
// visible terminal screen, including alt-screen mode, screen erasure, cell runes
// and SGR styles, cursor position, and cursor visibility.
func (v *VT) Repaint() string {
	var b strings.Builder

	if v.AltScreen() {
		b.WriteString("\x1b[?1049h")
	} else {
		b.WriteString("\x1b[?1049l")
	}

	if v.autoWrap {
		b.WriteString("\x1b[?7h")
	} else {
		b.WriteString("\x1b[?7l")
	}

	b.WriteString("\x1b[0m\x1b[H\x1b[2J")

	for y := 0; y < v.Rows; y++ {
		row := v.cells[y]
		last := -1
		for x := len(row) - 1; x >= 0; x-- {
			if !isTrailingBlank(row[x]) {
				last = x
				break
			}
		}
		if last >= 0 {
			fmt.Fprintf(&b, "\x1b[%d;1H", y+1)
			curStyle := Style{}
			for x := 0; x <= last; x++ {
				cell := row[x]
				if cell.Rune == 0 && cell.Text == "" {
					// Trailing half of wide character, already output
					continue
				}
				if cell.Style != curStyle {
					b.WriteString(cell.Style.SGR())
					curStyle = cell.Style
				}
				if cell.Text != "" {
					b.WriteString(cell.Text)
				} else if cell.Rune != 0 {
					b.WriteRune(cell.Rune)
				} else {
					b.WriteByte(' ')
				}
			}
			if curStyle != (Style{}) {
				b.WriteString("\x1b[0m")
			}
		}
	}

	fmt.Fprintf(&b, "\x1b[%d;%dH", v.y+1, v.x+1)

	if v.CursorVisible {
		b.WriteString("\x1b[?25h")
	} else {
		b.WriteString("\x1b[?25l")
	}

	if v.pen != (Style{}) {
		b.WriteString(v.pen.SGR())
	}

	return b.String()
}

func isTrailingBlank(c Cell) bool {
	if c.Text != "" && c.Text != " " {
		return false
	}
	if c.Rune != ' ' && c.Rune != 0 {
		return false
	}
	if c.Style.Reverse {
		return false
	}
	_, effBG := c.Style.Effective()
	if effBG.Type != ColorDefault {
		return false
	}
	if c.Style.Underline || c.Style.Strike {
		return false
	}
	return true
}

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
		if v.discardControl != 0 {
			if b == 0x18 || b == 0x1a || b == 0x9c || b == 7 {
				v.discardControl = 0
				i++
				continue
			}
			if b == 0x1b {
				if i+1 == len(data) {
					v.pending = []byte{b}
					return len(p), nil
				}
				if data[i+1] == '\\' {
					v.discardControl = 0
					i += 2
					continue
				}
				v.discardControl = 0
				continue
			}
			if v.discardControl == '[' && b >= 0x40 && b <= 0x7e {
				v.discardControl = 0
			}
			if v.discardControl == 'e' && b >= 0x30 && b <= 0x7e {
				v.discardControl = 0
			}
			i++
			continue
		}
		switch {
		case b == 0x18 || b == 0x1a || b == 0x9c:
			i++
		case b == 0x1b:
			n, ok := v.escape(data[i:])
			if !ok {
				v.pending = append([]byte(nil), data[i:]...)
				if len(v.pending) > maxPendingBytes {
					v.discardControl = controlType(v.pending)
					v.pending = nil
				}
				return len(p), nil
			}
			i += n
		case b == '\r':
			v.x = 0
			i++
		case b == '\n':
			v.lineFeed()
			i++
		case b == '\b':
			v.x = max(0, v.x-1)
			i++
		case b == '\t':
			v.x = min((v.x/8+1)*8, v.Cols-1)
			i++
		case b < 0x20 || b == 0x7f:
			i++
		default:
			if !utf8.FullRune(data[i:]) {
				v.pending = append([]byte(nil), data[i:]...)
				if len(v.pending) > maxPendingBytes {
					v.pending = nil
				}
				return len(p), nil
			}
			r, n := utf8.DecodeRune(data[i:])
			j := i + n
			if r >= 0x1F1E6 && r <= 0x1F1FF {
				if j < len(data) {
					if !utf8.FullRune(data[j:]) {
						v.pending = append([]byte(nil), data[i:]...)
						if len(v.pending) > maxPendingBytes {
							v.pending = nil
						}
						return len(p), nil
					}
					r2, n2 := utf8.DecodeRune(data[j:])
					if r2 >= 0x1F1E6 && r2 <= 0x1F1FF {
						j += n2
					}
				}
			} else {
				for j < len(data) {
					if data[j] == 0x1b || data[j] == '\r' || data[j] == '\n' {
						break
					}
					if !utf8.FullRune(data[j:]) {
						v.pending = append([]byte(nil), data[i:]...)
						if len(v.pending) > maxPendingBytes {
							v.pending = nil
						}
						return len(p), nil
					}
					nr, nsz := utf8.DecodeRune(data[j:])
					if nr == 0x200D {
						j += nsz
						if j < len(data) {
							if !utf8.FullRune(data[j:]) {
								v.pending = append([]byte(nil), data[i:]...)
								if len(v.pending) > maxPendingBytes {
									v.pending = nil
								}
								return len(p), nil
							}
							if jr, jsz := utf8.DecodeRune(data[j:]); measure.JoinsAfterZWJ(jr) {
								j += jsz
							}
						}
						continue
					}
					if nr == 0xFE0F || nr == 0xFE0E || unicode.Is(unicode.Mn, nr) || unicode.Is(unicode.Me, nr) || (unicode.Is(unicode.Cf, nr) && !measure.IsFormatRune(nr)) {
						j += nsz
						continue
					}
					break
				}
			}
			cluster := string(data[i:j])
			w := measure.StringWidth(cluster)
			if measure.ActiveRenderPath() == measure.RenderPathVTE && measure.VTEMode(cluster) == "pad-1" {
				w = max(1, w-1)
			}
			v.putCluster(r, cluster, w)
			i = j
		}
	}
	return len(p), nil
}

func (v *VT) lineFeed() {
	if v.y < v.Rows-1 {
		v.y++
		return
	}
	if !v.AltScreen() && v.scrollbackLimit > 0 {
		oldest := append([]Cell(nil), v.cells[0]...)
		if len(v.scrollback) < v.scrollbackLimit {
			v.scrollback = append(v.scrollback, oldest)
		} else {
			v.scrollback[v.scrollbackStart] = oldest
			v.scrollbackStart = (v.scrollbackStart + 1) % v.scrollbackLimit
		}
	}
	v.cells = append(v.cells[1:], v.blankRow(v.Cols, Style{BG: v.pen.BG}))
}

func (v *VT) put(r rune) {
	v.putCluster(r, string(r), measure.RuneWidth(r))
}

func (v *VT) putCluster(r rune, text string, w int) {
	if w <= 0 {
		return
	}
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
	v.cells[v.y][v.x] = Cell{Rune: r, Text: text, Style: v.pen}
	for k := 1; k < w && v.x+k < v.Cols; k++ {
		v.cells[v.y][v.x+k] = Cell{Rune: 0, Style: v.pen}
	}
	v.x += w
	if !v.autoWrap {
		v.x = min(v.x, v.Cols-1)
	}
}

// escape consumes one escape sequence from data. ok is false when data ends
// before the sequence does.
func (v *VT) escape(data []byte) (n int, ok bool) {
	if len(data) < 2 {
		return 0, false
	}
	switch data[1] {
	case ']', 'P', '^', '_', 'X':
		for i := 2; i < len(data); i++ {
			if data[i] == 0x18 || data[i] == 0x1a || data[i] == 0x9c {
				return i + 1, true
			}
			if data[i] == 7 {
				return i + 1, true
			}
			if data[i] == 0x1b {
				if i+1 == len(data) {
					return 0, false
				}
				if data[i+1] == '\\' {
					return i + 2, true
				}
				return i, true
			}
		}
		return 0, false
	case '7':
		v.saved = cursorState{v.x, v.y, v.pen}
		return 2, true
	case '8':
		v.x, v.y, v.pen = v.saved.x, v.saved.y, v.saved.pen
		return 2, true
	case '=':
		v.KeypadApp = true
		return 2, true
	case '>':
		v.KeypadApp = false
		return 2, true
	case '[':
		// Parse CSI below.
	default:
		if data[1] >= 0x20 && data[1] <= 0x2f {
			for i := 2; i < len(data); i++ {
				if data[i] == 0x1b {
					return i, true
				}
				if data[i] == 0x18 || data[i] == 0x1a {
					return i + 1, true
				}
				if data[i] >= 0x30 && data[i] <= 0x7e {
					return i + 1, true
				}
				if data[i] < 0x20 || data[i] > 0x2f {
					return i + 1, true
				}
			}
			return 0, false
		}
		return 2, true
	}
	for i := 2; i < len(data); i++ {
		if data[i] == 0x18 || data[i] == 0x1a {
			return i + 1, true
		}
		if data[i] == 0x1b {
			return i, true
		}
		if data[i] >= 0x40 && data[i] <= 0x7e {
			v.csi(string(data[2:i]), data[i])
			return i + 1, true
		}
	}
	return 0, false
}

func controlType(data []byte) byte {
	if len(data) > 1 && data[1] == '[' {
		return '['
	}
	if len(data) > 1 {
		switch data[1] {
		case ']', 'P', '^', '_', 'X':
			return data[1]
		default:
			return 'e'
		}
	}
	return 0
}

func csiParams(s string) []int {
	s = strings.TrimPrefix(s, "?")
	fields := strings.Split(s, ";")
	out := make([]int, len(fields))
	for i, field := range fields {
		out[i], _ = strconv.Atoi(field)
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
	case 'G':
		v.x = min(arg(0, 1)-1, v.Cols-1)
	case 'd':
		v.y = min(arg(0, 1)-1, v.Rows-1)
	case 'E':
		v.y, v.x = min(v.y+arg(0, 1), v.Rows-1), 0
	case 'F':
		v.y, v.x = max(v.y-arg(0, 1), 0), 0
	case 's':
		v.saved = cursorState{v.x, v.y, v.pen}
	case 'u':
		if strings.HasPrefix(params, ">") {
			v.KittyKeyboard = true
			return
		}
		if strings.HasPrefix(params, "<") {
			v.KittyKeyboard = false
			return
		}
		v.x, v.y, v.pen = v.saved.x, v.saved.y, v.saved.pen
	case 'K':
		v.eraseLine(p[0])
	case 'J':
		v.eraseDisplay(p[0])
	case 'h', 'l':
		on := final == 'h'
		if !strings.HasPrefix(params, "?") {
			return
		}
		for _, mode := range p {
			switch mode {
			case 1:
				v.CursorKeysApp = on
			case 7:
				v.autoWrap = on
			case 9:
				if on {
					v.MouseMode = 9
				} else if v.MouseMode == 9 {
					v.MouseMode = 0
				}
			case 25:
				v.CursorVisible = on
			case 1000:
				if on {
					v.MouseMode = 1000
				} else if v.MouseMode == 1000 {
					v.MouseMode = 0
				}
			case 1002:
				if on {
					v.MouseMode = 1002
				} else if v.MouseMode == 1002 {
					v.MouseMode = 0
				}
			case 1003:
				if on {
					v.MouseMode = 1003
				} else if v.MouseMode == 1003 {
					v.MouseMode = 0
				}
			case 1004:
				v.FocusEvents = on
			case 1005:
				v.MouseUTF8 = on
			case 1006:
				v.MouseSGR = on
			case 1015:
				v.MouseURXVT = on
			case 1049:
				v.alternate(on)
			case 2004:
				v.BracketedPaste = on
			case 2026:
				if !on {
					v.Frames = append(v.Frames, v.Screen())
					v.CellFrames = append(v.CellFrames, v.Cells())
				}
			}
		}
	}
}

func (v *VT) sgr(params string) {
	fields := strings.Split(params, ";")
	p := make([]int, len(fields))
	for i, field := range fields {
		p[i], _ = strconv.Atoi(field)
	}
	for i := 0; i < len(p); i++ {
		if strings.Contains(fields[i], ":") {
			parts := strings.Split(fields[i], ":")
			code, _ := strconv.Atoi(parts[0])
			if (code == 38 || code == 48) && len(parts) >= 3 {
				mode, _ := strconv.Atoi(parts[1])
				color := []int{}
				for _, part := range parts[2:] {
					if part == "" {
						continue
					}
					value, err := strconv.Atoi(part)
					if err == nil {
						color = append(color, value)
					}
				}
				if mode == 5 && len(color) >= 1 {
					setSGRColor(v, code, ColorIndex(uint8(color[0])))
				} else if mode == 2 && len(color) >= 3 {
					setSGRColor(v, code, ColorRGB(uint8(color[0]), uint8(color[1]), uint8(color[2])))
				}
			}
			continue
		}
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
					} else {
						i = len(p)
					}
				case 2: // 38;2;r;g;b
					if i+4 < len(p) {
						v.pen.FG = ColorRGB(uint8(p[i+2]), uint8(p[i+3]), uint8(p[i+4]))
						i += 4
					} else {
						i = len(p)
					}
				default:
					i = len(p)
				}
			} else {
				i = len(p)
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
					} else {
						i = len(p)
					}
				case 2: // 48;2;r;g;b
					if i+4 < len(p) {
						v.pen.BG = ColorRGB(uint8(p[i+2]), uint8(p[i+3]), uint8(p[i+4]))
						i += 4
					} else {
						i = len(p)
					}
				default:
					i = len(p)
				}
			} else {
				i = len(p)
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

func setSGRColor(v *VT, code int, color Color) {
	if code == 38 {
		v.pen.FG = color
	} else if code == 48 {
		v.pen.BG = color
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
