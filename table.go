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

// DefaultTableStyle returns a minimal monochrome style.
func DefaultTableStyle() TableStyle {
	return TableStyle{
		Header:     Style{Bold: true},
		SortHeader: Style{Bold: true, Underline: true},
		Normal:     Reset,
		Selected:   Style{Bold: true},
		Prompt:     Reset,
	}
}

// Table is a filterable, sortable, keyboard-navigable table widget.
//
// Key bindings:
//   - tab        cycle sort column forward (or complete command in command mode)
//   - !          toggle sort direction (asc ▲ / desc ▼)
//   - up/down    move row selection; wraps at boundaries
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

	Prompt   string
	Controls string // right-aligned hint; auto-computed from SortCol when empty
	OnSelect func(Row)
	OnSort   func(col int, desc bool) // called when Tab or ! changes sort

	Style TableStyle

	cmd        *cmdBar
	cmdNav     Nav
	query      string
	sel        int
	viewOffset int
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

	// Header row.
	x := r.X
	for i, col := range t.Columns {
		if i > 0 {
			cv.Write(x, r.Y, sep, t.Style.Header)
			x += sepW
		}
		style := t.Style.Header
		if i == t.SortCol {
			style = t.Style.SortHeader
		}
		w := t.colWidths[i]
		cv.Write(x, r.Y, padCol(t.headerText(i), w, col.Align), style)
		x += w
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
		if fi == t.sel {
			style = t.Style.Selected
		}
		x := r.X
		for i, col := range t.Columns {
			if i > 0 {
				cv.Write(x, y, sep, style)
				x += sepW
			}
			w := t.colWidths[i]
			cell := ""
			if i < len(t.filtered[fi].Cells) {
				cell = t.filtered[fi].Cells[i]
			}
			cv.Write(x, y, padCol(cell, w, col.Align), style)
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
}

// HandleKey drives navigation, filtering, and sort controls.
// ':' or '/' activates command mode; Tab completes commands when active.
func (t *Table) HandleKey(e KeyEvent) (quit bool) {
	if consumed, result := t.cmd.HandleKey(e); consumed {
		switch result {
		case cmdBack:
			t.aborted = true
			t.done = true
			return true
		case cmdHome:
			t.cmdNav = NavHome
			t.aborted = true
			t.done = true
			return true
		}
		return false
	}
	switch e.Key {
	case "esc", "ctrl-c", "ctrl-d", "ctrl-q":
		t.aborted = true
		t.done = true
		return true
	case "enter":
		if t.OnSelect != nil && len(t.filtered) > 0 {
			t.OnSelect(t.filtered[t.sel])
			return false
		}
		t.done = true
		return true
	case "up":
		if t.sel > 0 {
			t.sel--
		} else {
			t.sel = max(0, len(t.filtered)-1)
		}
	case "down":
		if t.sel < len(t.filtered)-1 {
			t.sel++
		} else {
			t.sel = 0
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
	return false
}

// HandleMouse is a no-op placeholder (mouse support is optional/future).
func (t *Table) HandleMouse(_ MouseEvent) (quit bool) { return false }

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
