// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strings"

	"codeberg.org/ubunatic/loom/measure"
)

// Align controls text alignment within a column.
type Align int

const (
	AlignLeft Align = iota
	AlignRight
)

// Row is a structured table entry.
// Cells maps column-by-index to display strings; Key is the unique identifier
// returned by Selected() as Item.Name.
type Row struct {
	Cells []string
	Key   string
}

// Column defines one column in a Table.
type Column struct {
	Header string
	Width  int // 0 = auto (max of header and all cell widths)
	Align  Align
}

// TableStyle controls the visual appearance of a Table.
type TableStyle struct {
	Header     Style // unsorted header cells
	SortHeader Style // sorted column header (bold + underline by default)
	Normal     Style
	Selected   Style
	Prompt     Style
}

// DefaultTableStyle returns a minimal monochrome style, derived from the plain theme.
func DefaultTableStyle() TableStyle {
	return Theme("plain").TableStyle()
}

// Table is a filterable, sortable, keyboard-navigable table widget.
//
// Key bindings:
//   - tab        cycle sort column forward (or complete command in command mode)
//   - !          toggle sort direction (asc ▲ / desc ▼)
//   - up/down    move row selection; wraps at boundaries
//   - left/right move the cell cursor when CellCursor is enabled
//   - pgup/pgdown (also pageup/pagedown and pgdn) move by a page of rows
//   - enter      confirm selection (calls OnSelect if set, otherwise quits)
//   - esc/ctrl-c abort
//   - :  /       activate command mode (see loom.Cmd, loom.Nav)
//   - printable  append to filter (all columns searched)
//   - backspace  delete last filter character
type Table struct {
	Columns  []Column
	Rows     []Row
	SortCol  int  // index of sorted column; -1 = none
	SortDesc bool // false = ascending ▲, true = descending ▼

	// CellCursor enables cell-by-cell selection. In this mode left/right move
	// between columns and up/down move between rows; row selection remains the
	// default when it is false.
	CellCursor bool
	// FrozenCols keeps this many leading columns visible during horizontal
	// scrolling in cell-cursor mode.
	FrozenCols int
	// OnCellSelect runs when the cell cursor moves, with filtered row and
	// column indexes.
	OnCellSelect func(row, col int)

	Prompt   string
	Controls string // right-aligned hint; auto-computed from SortCol when empty
	OnSelect func(Row)
	OnSort   func(col int, desc bool) // called when Tab or ! changes sort

	Style TableStyle

	cmd        *cmdBar
	keys       *KeyMap
	cmdNav     Nav
	query      string
	sel        int
	viewOffset int
	cellCol    int
	scrollCol  int // first non-frozen column visible during cell-cursor scrolling
	pageRows   int
	filtered   []Row
	colWidths  []int
	done       bool
	aborted    bool
}

// NewTable creates a ready-to-use Table with default style.
func NewTable(cols []Column, rows []Row) *Table {
	t := &Table{
		Columns: cols,
		Rows:    rows,
		SortCol: -1,
		Style:   DefaultTableStyle(),
		Prompt:  "> ",
	}
	t.cmd = newCmdBar()
	t.keys = NewKeyMap(map[string][]string{
		"page-up":   {"pgup", "pageup"},
		"page-down": {"pgdown", "pgdn", "pagedown"},
	})
	t.refilter()
	return t
}

// Nav returns the navigation signal set by a prompt command (:home).
func (t *Table) Nav() Nav { return t.cmdNav }

// AddCmd registers a view-local command accessible via ':name' in this widget.
func (t *Table) AddCmd(cmd Cmd) { t.cmd.AddLocal(cmd) }

// SetRows replaces the row data and re-applies the current filter.
// Call this after sorting or refreshing the underlying data.
func (t *Table) SetRows(rows []Row) {
	t.Rows = rows
	t.refilter()
	if t.sel >= len(t.filtered) {
		t.sel = max(0, len(t.filtered)-1)
	}
}

// Selected returns the chosen Row (wrapped in Item for paneable compatibility)
// and whether a selection was made (not aborted).
func (t *Table) Selected() (Item, bool) {
	if t.aborted || len(t.filtered) == 0 || t.sel >= len(t.filtered) {
		return Item{}, false
	}
	return Item{Name: t.filtered[t.sel].Key}, true
}

// Aborted reports whether the user dismissed without selecting.
func (t *Table) Aborted() bool { return t.aborted }

// ContentHeight estimates the required height: header + rows + prompt.
func (t *Table) ContentHeight() int {
	h := len(t.Rows) + 2
	if h < 3 {
		return 3
	}
	return h
}

func (t *Table) refilter() {
	q := strings.ToLower(t.query)
	if q == "" {
		t.filtered = t.Rows
	} else {
		out := make([]Row, 0, len(t.Rows))
		for _, r := range t.Rows {
			for _, cell := range r.Cells {
				if strings.Contains(strings.ToLower(cell), q) {
					out = append(out, r)
					break
				}
			}
		}
		t.filtered = out
	}
	if t.sel >= len(t.filtered) {
		t.sel = max(0, len(t.filtered)-1)
	}
	t.computeWidths()
}

// computeWidths calculates column widths from all rows (not just filtered),
// so widths remain stable when the filter changes.
func (t *Table) computeWidths() {
	t.colWidths = make([]int, len(t.Columns))
	for i, col := range t.Columns {
		if col.Width > 0 {
			t.colWidths[i] = col.Width
			continue
		}
		// +1: reserve one char for the sort indicator (▲ or ▼).
		w := StringWidth(col.Header) + 1
		for _, row := range t.Rows {
			if i < len(row.Cells) {
				if cw := StringWidth(row.Cells[i]); cw > w {
					w = cw
				}
			}
		}
		t.colWidths[i] = w
	}
}

func (t *Table) headerText(i int) string {
	h := t.Columns[i].Header
	if i != t.SortCol {
		return h
	}
	if t.SortDesc {
		return h + "▼"
	}
	return h + "▲"
}

// Draw renders the table into r.
// Layout: r.Y = header row, r.Y+1 .. r.Y+H-2 = data rows, r.Y+H-1 = prompt.
func (t *Table) Draw(cv *Canvas, r Rect) {
	const sep = "  "
	const sepW = 2

	itemRows := r.H - 2
	if itemRows < 0 {
		itemRows = 0
	}
	t.pageRows = itemRows

	// Clamp viewOffset so sel is always visible.
	if t.viewOffset > t.sel {
		t.viewOffset = t.sel
	}
	if t.sel >= t.viewOffset+itemRows {
		t.viewOffset = t.sel - itemRows + 1
	}
	if t.viewOffset < 0 {
		t.viewOffset = 0
	}

	// Keep the cursor column visible in the scrollable section. Frozen columns
	// consume space first; the remaining width is used by ScrollCol onward.
	frozen := min(max(t.FrozenCols, 0), len(t.Columns))
	frozenWidth := 0
	for i := 0; i < frozen; i++ {
		if i > 0 {
			frozenWidth += sepW
		}
		frozenWidth += t.colWidths[i]
	}
	available := r.W - frozenWidth
	if frozen > 0 && frozen < len(t.Columns) {
		available -= sepW
	}
	available = max(0, available)
	if t.scrollCol < frozen || t.scrollCol >= len(t.Columns) {
		t.scrollCol = frozen
	}
	if !t.CellCursor {
		t.scrollCol = frozen
	}
	if t.CellCursor && t.cellCol >= frozen && t.cellCol < len(t.Columns) {
		if t.cellCol < t.scrollCol {
			t.scrollCol = t.cellCol
		}
		for t.scrollCol < t.cellCol {
			need := sepW
			for i := t.scrollCol; i <= t.cellCol; i++ {
				need += t.colWidths[i]
				if i > t.scrollCol {
					need += sepW
				}
			}
			if need <= available {
				break
			}
			t.scrollCol++
		}
	}
	if t.scrollCol < frozen {
		t.scrollCol = frozen
	}
	if t.scrollCol > len(t.Columns) {
		t.scrollCol = len(t.Columns)
	}
	visible := make([]int, 0, len(t.Columns))
	for i := 0; i < frozen; i++ {
		visible = append(visible, i)
	}
	for i := t.scrollCol; i < len(t.Columns); i++ {
		visible = append(visible, i)
	}

	// Header row.
	x := r.X
	for n, i := range visible {
		col := t.Columns[i]
		if n > 0 {
			if x+sepW > r.X+r.W {
				break
			}
			cv.Write(x, r.Y, sep, t.Style.Header)
			x += sepW
		}
		if x >= r.X+r.W {
			break
		}
		style := t.Style.Header
		if i == t.SortCol {
			style = t.Style.SortHeader
		}
		w := min(t.colWidths[i], r.X+r.W-x)
		cv.Write(x, r.Y, padCol(t.headerText(i), w, col.Align), style)
		x += t.colWidths[i]
	}

	// Data rows.
	for row := 0; row < itemRows; row++ {
		y := r.Y + 1 + row
		cv.PaintSurface(Rect{r.X, y, r.W, 1}, t.Style.Normal)
		fi := t.viewOffset + row
		if fi < 0 || fi >= len(t.filtered) {
			continue
		}
		style := t.Style.Normal
		if !t.CellCursor && fi == t.sel {
			style = t.Style.Selected
		}
		x := r.X
		for n, i := range visible {
			col := t.Columns[i]
			cellStyle := style
			if t.CellCursor && fi == t.sel && i == t.cellCol {
				cellStyle = t.Style.Selected
			}
			if n > 0 {
				if x+sepW > r.X+r.W {
					break
				}
				cv.Write(x, y, sep, cellStyle)
				x += sepW
			}
			if x >= r.X+r.W {
				break
			}
			w := t.colWidths[i]
			cell := ""
			if i < len(t.filtered[fi].Cells) {
				cell = t.filtered[fi].Cells[i]
			}
			cv.Write(x, y, padCol(cell, min(w, r.X+r.W-x), col.Align), cellStyle)
			x += w
		}
	}

	// Prompt row.
	promptY := r.Y + r.H - 1
	cv.PaintSurface(Rect{r.X, promptY, r.W, 1}, t.Style.Prompt)
	if prefix, hint := t.cmd.PromptParts(); prefix != "" {
		// Command mode: ":typed[completion]  dim title"
		n := cv.Write(r.X, promptY, prefix, t.Style.Prompt)
		if hint != "" {
			cv.Write(r.X+n, promptY, hint, Style{Dim: true})
		}
		cv.CursorX = r.X + n
		cv.CursorY = promptY
	} else {
		// Normal mode: base prompt + filter query + right-aligned controls.
		cv.Write(r.X, promptY, t.Prompt+t.query, t.Style.Prompt)

		controls := t.Controls
		if controls == "" && t.SortCol >= 0 && t.SortCol < len(t.Columns) {
			dir := "▲"
			if t.SortDesc {
				dir = "▼"
			}
			controls = fmt.Sprintf("[tab:%s%s !:flip]", dir, t.Columns[t.SortCol].Header)
		}
		if controls != "" {
			ctrlW := StringWidth(controls)
			promptW := StringWidth(t.Prompt) + StringWidth(t.query)
			if r.W > ctrlW+promptW {
				cv.Write(r.X+r.W-ctrlW, promptY, controls, Style{Dim: true})
			}
		}

		cv.CursorX = r.X + StringWidth(t.Prompt) + StringWidth(t.query)
		cv.CursorY = promptY
	}
	t.cmd.drawHelp(cv, r)
}

// ConsumeKey drives navigation, filtering, and sort controls.
// ':' or '/' activates command mode; Tab completes commands when active.
func (t *Table) ConsumeKey(e KeyEvent) (quit EventResult) {
	if t.cmd.handleHelp(e) {
		return Ignored()
	}
	if consumed, result := t.cmd.ConsumeKey(e); consumed {
		switch result {
		case cmdBack:
			t.aborted = true
			t.done = true
			return QuitResult()
		case cmdHome:
			t.cmdNav = NavHome
			t.aborted = true
			t.done = true
			return QuitResult()
		}
		return Ignored()
	}
	var action string
	if t.keys != nil {
		action = t.keys.Action(e)
	}
	if action != "" {
		count := t.pageRows
		if count < 1 {
			count = 10
		}
		for i := 0; i < count; i++ {
			if action == "page-up" {
				if t.sel > 0 {
					t.sel--
				}
			} else if t.sel < len(t.filtered)-1 {
				t.sel++
			}
		}
		t.cellSelectionChanged()
		return Ignored()
	}
	switch e.Key {
	case "esc", "ctrl-c", "ctrl-d", "ctrl-q":
		t.aborted = true
		t.done = true
		return QuitResult()
	case "enter":
		if t.OnSelect != nil && len(t.filtered) > 0 {
			t.OnSelect(t.filtered[t.sel])
			return Ignored()
		}
		t.done = true
		return QuitResult()
	case "up":
		if t.sel > 0 {
			t.sel--
		} else {
			t.sel = max(0, len(t.filtered)-1)
		}
		t.cellSelectionChanged()
	case "down":
		if t.sel < len(t.filtered)-1 {
			t.sel++
		} else {
			t.sel = 0
		}
		t.cellSelectionChanged()
	case "left":
		if t.CellCursor && t.cellCol > 0 {
			t.cellCol--
			t.cellSelectionChanged()
		}
	case "right":
		if t.CellCursor && t.cellCol < len(t.Columns)-1 {
			t.cellCol++
			t.cellSelectionChanged()
		}
	case "tab":
		t.cycleSortNext()
	case "backspace":
		if len(t.query) > 0 {
			runes := []rune(t.query)
			t.query = string(runes[:len(runes)-1])
			t.refilter()
		}
	default:
		if e.Text == "!" {
			t.SortDesc = !t.SortDesc
			if t.OnSort != nil && t.SortCol >= 0 {
				t.OnSort(t.SortCol, t.SortDesc)
			}
		} else if e.Text != "" {
			t.query += e.Text
			t.refilter()
		}
	}
	return Ignored()
}

func (t *Table) cellSelectionChanged() {
	if !t.CellCursor {
		return
	}
	if t.cellCol >= len(t.Columns) {
		t.cellCol = max(0, len(t.Columns)-1)
	}
	if t.OnCellSelect != nil && len(t.filtered) > 0 && len(t.Columns) > 0 {
		t.OnCellSelect(t.sel, t.cellCol)
	}
}

// ConsumeMouse is a no-op placeholder (mouse support is optional/future).
func (t *Table) ConsumeMouse(e MouseEvent) (quit EventResult) {
	if t.cmd.handleHelpMouse(e) {
		return Handled()
	}
	return Ignored()
}

func (t *Table) cycleSortNext() {
	if len(t.Columns) == 0 {
		return
	}
	if t.SortCol < 0 {
		t.SortCol = 0
	} else {
		t.SortCol = (t.SortCol + 1) % len(t.Columns)
	}
	if t.OnSort != nil {
		t.OnSort(t.SortCol, t.SortDesc)
	}
}

// padCol pads or truncates s to exactly w visual columns with the given alignment.
func padCol(s string, w int, align Align) string {
	return measure.Pad(s, w, align == AlignRight)
}

// ApplyTheme updates the Table style from the theme.
func (t *Table) ApplyTheme(theme ThemeColors) {
	t.Style = theme.TableStyle()
}
