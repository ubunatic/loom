// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"ubunatic.com/loom/measure"
)

// TextEntryOptions configures a TextEntry. The zero value gives a plain entry
// that grows without a height limit, submits on Enter and breaks lines on
// Shift+Enter, Ctrl+J or Alt+Enter.
type TextEntryOptions struct {
	Placeholder string
	Prompt      string   // first-row prefix, e.g. "> "
	MaxHeight   int      // visual rows before scrolling; 0 = unbounded
	SubmitKeys  []string // default ["enter"]
	NewlineKeys []string // default ["shift-enter", "ctrl-j", "alt-enter"]
	OnChange    func(value string)
	OnSubmit    func(value string) EventResult
}

// Default key bindings, used when TextEntryOptions leaves a list empty.
var (
	textEntrySubmitDefault  = []string{"enter"}
	textEntryNewlineDefault = []string{"shift-enter", "ctrl-j", "alt-enter"}
)

// TextEntry is a plain-text input for chat-style lines. It wraps its text to
// the width it is given, grows with its content up to MaxHeight and then
// scrolls, and edits with mouse and keyboard selection, cut, copy and paste
// (internal clipboard, OSC 52 through the host clipboard hook, bracketed
// paste), and undo and redo.
//
// The host keeps its own keys: TextEntry returns Ignored for every key that is
// not an editing key, so Tab, Ctrl+O, Ctrl+S and similar host shortcuts reach
// the host. Cursor keys are always consumed, except Up and Down (and page keys)
// at the first or last row, which are ignored so the host can use them for
// history. Edits and clipboard actions that change nothing are also ignored.
// The Submit and Newline bindings decide between submitting and inserting a
// line break; Enter without OnSubmit is ignored as well.
//
// Text is plain: tabs become four spaces and CR or CRLF become LF.
type TextEntry struct {
	placeholder string
	prompt      string
	maxHeight   int
	submitKeys  []string
	newlineKeys []string
	onChange    func(string)
	onSubmit    func(string) EventResult

	units   []string // one grapheme cluster per element; "\n" is a hard line break
	caret   int      // unit index in [0, len(units)]
	anchor  int      // the selection runs between anchor and caret
	goalCol int      // display column kept across Up and Down
	goalSet bool
	scroll  int // first visible visual row

	focused     bool
	dragging    bool
	typingRun   bool
	followCaret bool // Draw scrolls the caret into view once after a caret move
	undo        []textEntrySnapshot
	redo        []textEntrySnapshot

	clipboard      string
	writeClipboard func(string)

	layoutWidth  int // width of the latest Draw or Measure, 0 before either
	layoutHeight int // visible rows at the latest Draw
	clickCount   int
	lastClick    time.Time
	lastX, lastY int
	now          func() time.Time
}

type textEntrySnapshot struct {
	units  []string
	caret  int
	anchor int
}

type textEntryRow struct {
	start, end int  // unit range shown on the row; end excludes a line break
	hard       bool // the row ends a line (line break or end of text)
}

// textEntryUnbounded is the layout width used when no width is known.
const textEntryUnbounded = 1 << 30

// NewTextEntry creates an empty entry configured by opts.
func NewTextEntry(opts TextEntryOptions) *TextEntry {
	e := &TextEntry{
		placeholder: opts.Placeholder,
		prompt:      opts.Prompt,
		maxHeight:   opts.MaxHeight,
		submitKeys:  slices.Clone(opts.SubmitKeys),
		newlineKeys: slices.Clone(opts.NewlineKeys),
		onChange:    opts.OnChange,
		onSubmit:    opts.OnSubmit,
		now:         time.Now,
	}
	if len(e.submitKeys) == 0 {
		e.submitKeys = slices.Clone(textEntrySubmitDefault)
	}
	if len(e.newlineKeys) == 0 {
		e.newlineKeys = slices.Clone(textEntryNewlineDefault)
	}
	return e
}

// Value returns the entry text with line breaks as "\n".
func (e *TextEntry) Value() string {
	return strings.Join(e.units, "")
}

// SetValue replaces the text, puts the caret at the end and clears the
// selection. The change is one undo step. It does not call OnChange.
func (e *TextEntry) SetValue(s string) {
	next := textEntryUnits(s)
	if !slices.Equal(e.units, next) {
		e.pushUndo(e.snapshot())
	}
	e.units = next
	e.caret = len(next)
	e.anchor = e.caret
	e.typingRun = false
	e.goalSet = false
	e.followCaret = true
}

// Cursor returns the caret cell relative to the widget's visible top-left
// corner, after soft wrap and scrolling. The wrap width is that of the latest
// Draw or Measure call.
func (e *TextEntry) Cursor() (row, col int) {
	rows := e.rows(e.layoutWidth)
	r, c := e.locate(rows, e.caret)
	return r - e.scroll, e.promptWidth() + c
}

// Measure reports the soft-wrapped size at width, with height between one row
// and MaxHeight. A width of zero or less measures the text on its own lines.
func (e *TextEntry) Measure(width int) measure.Size {
	if width > 0 {
		e.layoutWidth = width
	}
	rows := e.rows(width)
	height := len(rows)
	if e.maxHeight > 0 {
		height = min(height, e.maxHeight)
	}
	if width > 0 {
		return measure.Size{Width: width, Height: height}
	}
	widest := 0
	for _, row := range rows {
		widest = max(widest, e.widthBetween(row.start, row.end))
	}
	return measure.Size{Width: e.promptWidth() + widest, Height: height}
}

// Focused reports whether the entry shows the caret.
func (e *TextEntry) Focused() bool { return e.focused }

// SetFocus sets whether the entry shows the caret and takes editing input. Losing
// focus ends a mouse drag and the current typing run.
func (e *TextEntry) SetFocus(focused bool) {
	e.focused = focused
	if !focused {
		e.dragging = false
		e.typingRun = false
	}
}

// setClipboardWriter receives copy and cut text; Pane binds it to OSC 52.
func (e *TextEntry) setClipboardWriter(write func(string)) { e.writeClipboard = write }

// ConsumePaste inserts pasted text at the caret as one undo step.
func (e *TextEntry) ConsumePaste(event PasteEvent) EventResult {
	return e.insert(event.Text, false)
}

// Draw paints the prompt, the wrapped text or the placeholder into r. When
// focused, it places the hardware cursor at the caret and keeps the caret row
// visible. Mouse wheel scrolling is kept until the caret moves again.
func (e *TextEntry) Draw(c *Canvas, r Rect) {
	c.PaintSurface(r, Style{})
	e.layoutWidth, e.layoutHeight = r.W, r.H
	rows := e.rows(r.W)
	caretRow, caretCol := e.locate(rows, e.caret)
	if e.followCaret && r.H > 0 {
		if caretRow < e.scroll {
			e.scroll = caretRow
		}
		if caretRow >= e.scroll+r.H {
			e.scroll = caretRow - r.H + 1
		}
	}
	e.followCaret = false
	e.scroll = min(max(0, e.scroll), max(0, len(rows)-r.H))

	promptW := e.promptWidth()
	right := r.X + r.W
	for y := 0; y < r.H && e.scroll+y < len(rows); y++ {
		row := rows[e.scroll+y]
		x := r.X
		if e.scroll+y == 0 {
			x = e.writeClipped(c, x, r.Y+y, right, e.prompt, Style{})
		}
		x = r.X + promptW
		if len(e.units) == 0 && e.scroll+y == 0 {
			e.writeClipped(c, x, r.Y+y, right, e.placeholder, Style{Dim: true})
			continue
		}
		for i := row.start; i < row.end; i++ {
			style := Style{}
			if e.selected(i) {
				style.Invert = true
			}
			x = e.writeClipped(c, x, r.Y+y, right, e.units[i], style)
		}
	}
	if e.focused {
		cy := caretRow - e.scroll
		cx := r.X + promptW + caretCol
		if cy >= 0 && cy < r.H && cx < right {
			c.CursorX, c.CursorY = cx, r.Y+cy
			c.CursorShape, c.CursorShapeSet = CursorShapeBar, true
		}
	}
}

// ConsumeKey applies an editing key. Keys it does not use are returned Ignored.
func (e *TextEntry) ConsumeKey(ke KeyEvent) EventResult {
	if ke.Report {
		return Ignored()
	}
	if ke.Key == "" {
		return e.insert(ke.Text, true)
	}
	name := normalizeTextEntryKey(ke.Key)
	e.typingRun = false
	if !isTextEntryVerticalKey(name) {
		e.goalSet = false
	}
	if textEntryBound(name, e.submitKeys) {
		if e.onSubmit == nil {
			return Ignored()
		}
		return e.onSubmit(e.Value())
	}
	if textEntryBound(name, e.newlineKeys) {
		return e.insert("\n", false)
	}
	extend := strings.HasPrefix(name, "shift-") || strings.HasPrefix(name, "ctrl-shift-")
	page := max(1, e.layoutHeight)
	switch name {
	case "left", "shift-left":
		return e.step(-1, extend)
	case "right", "shift-right":
		return e.step(1, extend)
	case "ctrl-left", "ctrl-shift-left":
		e.setCaret(e.wordLeft(e.caret), extend)
		return Consumed()
	case "ctrl-right", "ctrl-shift-right":
		e.setCaret(e.wordRight(e.caret), extend)
		return Consumed()
	case "up", "shift-up":
		return e.vertical(-1, extend)
	case "down", "shift-down":
		return e.vertical(1, extend)
	case "pgup", "shift-pgup":
		return e.vertical(-page, extend)
	case "pgdown", "shift-pgdown":
		return e.vertical(page, extend)
	case "home", "shift-home":
		rows := e.rows(e.layoutWidth)
		row, _ := e.locate(rows, e.caret)
		e.setCaret(rows[row].start, extend)
		return Consumed()
	case "end", "shift-end":
		rows := e.rows(e.layoutWidth)
		row, _ := e.locate(rows, e.caret)
		e.setCaret(rows[row].end, extend)
		return Consumed()
	case "ctrl-home", "ctrl-shift-home":
		e.setCaret(0, extend)
		return Consumed()
	case "ctrl-end", "ctrl-shift-end":
		e.setCaret(len(e.units), extend)
		return Consumed()
	case "backspace":
		return e.removeNear(true)
	case "delete":
		return e.removeNear(false)
	case "ctrl-a":
		if len(e.units) == 0 {
			return Ignored()
		}
		e.anchor, e.caret = 0, len(e.units)
		return Consumed()
	case "ctrl-c":
		if !e.hasSelection() {
			return Ignored()
		}
		e.setClipboard(e.selectedText())
		return Consumed()
	case "ctrl-x":
		if !e.hasSelection() {
			return Ignored()
		}
		e.setClipboard(e.selectedText())
		e.change(false, func() { e.deleteSelection() })
		return Consumed()
	case "ctrl-v":
		if e.clipboard == "" {
			return Ignored()
		}
		return e.insert(e.clipboard, false)
	case "ctrl-z":
		return e.undoStep()
	case "ctrl-y", "ctrl-shift-z":
		return e.redoStep()
	}
	return Ignored()
}

// ConsumeMouse places the caret on a press, extends the selection on a drag,
// selects a word on a double click and a line on a triple click, and scrolls
// the wheel. Coordinates are 0-based and relative to the widget.
func (e *TextEntry) ConsumeMouse(me MouseEvent) EventResult {
	switch me.Action {
	case MouseScrollUp:
		return e.scrollBy(-1)
	case MouseScrollDown:
		return e.scrollBy(1)
	case MousePress:
		if me.Button != MouseLeft {
			return Ignored()
		}
		e.goalSet = false
		e.typingRun = false
		pos := e.positionAt(me.X, me.Y)
		switch e.registerClick(me) {
		case 2:
			e.dragging = false
			e.anchor, e.caret = e.wordAt(pos)
		case 3:
			e.dragging = false
			e.anchor, e.caret = e.lineAt(pos)
		default:
			e.setCaret(pos, me.Shift)
			e.dragging = true
		}
		e.followCaret = false
		return Consumed()
	case MouseDrag:
		if !e.dragging {
			return Ignored()
		}
		e.setCaret(e.positionAt(me.X, me.Y), true)
		return Consumed()
	case MouseRelease:
		if !e.dragging {
			return Ignored()
		}
		e.dragging = false
		return Consumed()
	}
	return Ignored()
}

func (e *TextEntry) promptWidth() int { return StringWidth(e.prompt) }

// rows lays the text out at width. The last display cell is kept free so the
// caret at the end of a full row stays visible. A width of zero or less means
// the text is not wrapped.
func (e *TextEntry) rows(width int) []textEntryRow {
	avail := textEntryUnbounded
	if width > 0 {
		avail = max(1, width-e.promptWidth()-1)
	}
	var out []textEntryRow
	start := 0
	for {
		used, i := 0, start
		for i < len(e.units) && e.units[i] != "\n" {
			w := measure.ClusterWidth(e.units[i])
			if used+w > avail && i > start {
				break
			}
			used += w
			i++
		}
		if i == len(e.units) {
			return append(out, textEntryRow{start: start, end: i, hard: true})
		}
		if e.units[i] == "\n" {
			out = append(out, textEntryRow{start: start, end: i, hard: true})
			start = i + 1
			continue
		}
		out = append(out, textEntryRow{start: start, end: i})
		start = i
	}
}

// locate returns the visual row and display column of the caret at pos. A
// caret at a soft wrap belongs to the start of the next row.
func (e *TextEntry) locate(rows []textEntryRow, pos int) (row, col int) {
	for k, r := range rows {
		if pos < r.end || (pos == r.end && r.hard) {
			return k, e.widthBetween(r.start, pos)
		}
	}
	last := rows[len(rows)-1]
	return len(rows) - 1, e.widthBetween(last.start, last.end)
}

// positionAt maps a widget-relative cell to the nearest caret position on
// that cell's visual row.
func (e *TextEntry) positionAt(x, y int) int {
	rows := e.rows(e.layoutWidth)
	row := min(max(0, e.scroll+y), len(rows)-1)
	return e.indexAtColumn(rows[row], max(0, x-e.promptWidth()))
}

// indexAtColumn returns the caret position nearest to display column col on r.
// The left half of a cluster puts the caret before it, the right half after it.
func (e *TextEntry) indexAtColumn(r textEntryRow, col int) int {
	acc := 0
	for i := r.start; i < r.end; i++ {
		w := measure.ClusterWidth(e.units[i])
		if col < acc+w {
			if col-acc < (w+1)/2 {
				return i
			}
			return i + 1
		}
		acc += w
	}
	return r.end
}

func (e *TextEntry) widthBetween(from, to int) int {
	w := 0
	for i := from; i < to; i++ {
		w += measure.ClusterWidth(e.units[i])
	}
	return w
}

func (e *TextEntry) selected(i int) bool {
	lo, hi := e.selectionBounds()
	return lo <= i && i < hi
}

func (e *TextEntry) hasSelection() bool { return e.anchor != e.caret }

func (e *TextEntry) selectionBounds() (int, int) {
	if e.anchor < e.caret {
		return e.anchor, e.caret
	}
	return e.caret, e.anchor
}

func (e *TextEntry) selectedText() string {
	lo, hi := e.selectionBounds()
	return strings.Join(e.units[lo:hi], "")
}

// setCaret moves the caret and, unless extend is set, collapses the selection.
func (e *TextEntry) setCaret(pos int, extend bool) {
	e.caret = min(max(0, pos), len(e.units))
	if !extend {
		e.anchor = e.caret
	}
	e.followCaret = true
}

// step moves one cell left or right. Without extend, a selection collapses to
// its near edge instead of moving.
func (e *TextEntry) step(dir int, extend bool) EventResult {
	if !extend && e.hasSelection() {
		lo, hi := e.selectionBounds()
		if dir < 0 {
			e.setCaret(lo, false)
		} else {
			e.setCaret(hi, false)
		}
		return Consumed()
	}
	e.setCaret(e.caret+dir, extend)
	return Consumed()
}

// vertical moves the caret delta visual rows, keeping its display column. It
// is ignored when the caret is already on the first or last row.
func (e *TextEntry) vertical(delta int, extend bool) EventResult {
	rows := e.rows(e.layoutWidth)
	row, col := e.locate(rows, e.caret)
	target := min(max(0, row+delta), len(rows)-1)
	if target == row {
		return Ignored()
	}
	if !e.goalSet {
		e.goalCol, e.goalSet = col, true
	}
	e.setCaret(e.indexAtColumn(rows[target], e.goalCol), extend)
	return Consumed()
}

func (e *TextEntry) scrollBy(delta int) EventResult {
	rows := len(e.rows(e.layoutWidth))
	maxScroll := max(0, rows-max(1, e.layoutHeight))
	next := min(max(0, e.scroll+delta), maxScroll)
	if next == e.scroll {
		return Ignored()
	}
	e.scroll = next
	return Consumed()
}

// registerClick counts presses on the same cell within the double-click
// interval and tolerance: 1, 2, 3 and then back to 1.
func (e *TextEntry) registerClick(me MouseEvent) int {
	now := e.now()
	if !e.lastClick.IsZero() && now.Sub(e.lastClick) <= SpeccedDefaults.Mouse.DoubleClickInterval &&
		withinMovementTolerance(me.X, me.Y, e.lastX, e.lastY, SpeccedDefaults.Mouse.MovementTolerance) {
		e.clickCount = e.clickCount%3 + 1
	} else {
		e.clickCount = 1
	}
	e.lastClick, e.lastX, e.lastY = now, me.X, me.Y
	return e.clickCount
}

// wordAt returns the run of same-class units (word, space or other) around pos.
func (e *TextEntry) wordAt(pos int) (int, int) {
	n := len(e.units)
	k := pos
	if k >= n || e.units[k] == "\n" {
		k = pos - 1
	}
	if k < 0 || k >= n || e.units[k] == "\n" {
		return pos, pos
	}
	class := textEntryClass(e.units[k])
	lo, hi := k, k+1
	for lo > 0 && textEntryClass(e.units[lo-1]) == class {
		lo--
	}
	for hi < n && textEntryClass(e.units[hi]) == class {
		hi++
	}
	return lo, hi
}

// lineAt returns the logical line around pos, without its line break.
func (e *TextEntry) lineAt(pos int) (int, int) {
	lo := pos
	for lo > 0 && e.units[lo-1] != "\n" {
		lo--
	}
	hi := pos
	for hi < len(e.units) && e.units[hi] != "\n" {
		hi++
	}
	return lo, hi
}

func (e *TextEntry) wordLeft(pos int) int {
	i := pos
	for i > 0 && !isTextEntryWord(e.units[i-1]) {
		i--
	}
	for i > 0 && isTextEntryWord(e.units[i-1]) {
		i--
	}
	return i
}

func (e *TextEntry) wordRight(pos int) int {
	i := pos
	for i < len(e.units) && isTextEntryWord(e.units[i]) {
		i++
	}
	for i < len(e.units) && !isTextEntryWord(e.units[i]) {
		i++
	}
	return i
}

// insert replaces the selection with text. Typed runs join one undo step.
func (e *TextEntry) insert(text string, typing bool) EventResult {
	ins := textEntryUnits(text)
	if len(ins) == 0 {
		return Ignored()
	}
	e.change(typing, func() {
		e.deleteSelection()
		next := make([]string, 0, len(e.units)+len(ins))
		next = append(next, e.units[:e.caret]...)
		next = append(next, ins...)
		next = append(next, e.units[e.caret:]...)
		e.units = next
		e.caret += len(ins)
		e.anchor = e.caret
	})
	return Consumed()
}

// removeNear deletes the selection, or the unit before or after the caret.
func (e *TextEntry) removeNear(backward bool) EventResult {
	if !e.change(false, func() {
		if e.deleteSelection() {
			return
		}
		if backward && e.caret > 0 {
			e.units = slices.Delete(e.units, e.caret-1, e.caret)
			e.caret--
		}
		if !backward && e.caret < len(e.units) {
			e.units = slices.Delete(e.units, e.caret, e.caret+1)
		}
		e.anchor = e.caret
	}) {
		return Ignored()
	}
	return Consumed()
}

func (e *TextEntry) deleteSelection() bool {
	if !e.hasSelection() {
		return false
	}
	lo, hi := e.selectionBounds()
	e.units = slices.Delete(e.units, lo, hi)
	e.caret, e.anchor = lo, lo
	return true
}

// change runs edit and records one undo step when the text changed. A typed
// edit that follows another typed edit joins its step. It reports whether the
// text changed and notifies OnChange when it did.
func (e *TextEntry) change(typing bool, edit func()) bool {
	before := e.snapshot()
	edit()
	if slices.Equal(before.units, e.units) {
		return false
	}
	if !(typing && e.typingRun) {
		e.pushUndo(before)
	}
	e.redo = nil
	e.typingRun = typing
	e.followCaret = true
	if e.onChange != nil {
		e.onChange(e.Value())
	}
	return true
}

func (e *TextEntry) pushUndo(snap textEntrySnapshot) {
	e.undo = append(e.undo, snap)
	if limit := SpeccedDefaults.RichTextEdit.UndoLimit; len(e.undo) > limit {
		e.undo = e.undo[len(e.undo)-limit:]
	}
}

func (e *TextEntry) undoStep() EventResult {
	return e.restoreFrom(&e.undo, &e.redo)
}

func (e *TextEntry) redoStep() EventResult {
	return e.restoreFrom(&e.redo, &e.undo)
}

// restoreFrom moves the newest snapshot from history onto other and restores it.
func (e *TextEntry) restoreFrom(history, other *[]textEntrySnapshot) EventResult {
	if len(*history) == 0 {
		return Ignored()
	}
	before := e.Value()
	prev := (*history)[len(*history)-1]
	*history = (*history)[:len(*history)-1]
	*other = append(*other, e.snapshot())
	e.units, e.caret, e.anchor = prev.units, prev.caret, prev.anchor
	e.typingRun = false
	e.followCaret = true
	if after := e.Value(); after != before && e.onChange != nil {
		e.onChange(after)
	}
	return Consumed()
}

func (e *TextEntry) snapshot() textEntrySnapshot {
	return textEntrySnapshot{units: slices.Clone(e.units), caret: e.caret, anchor: e.anchor}
}

// setClipboard stores text in the internal clipboard and the host clipboard hook.
func (e *TextEntry) setClipboard(text string) {
	e.clipboard = text
	if e.writeClipboard != nil {
		e.writeClipboard(text)
	}
}

// writeClipped writes text starting at x and stops before column right. It
// returns the column after the last cell written.
func (e *TextEntry) writeClipped(c *Canvas, x, y, right int, text string, style Style) int {
	for _, cluster := range measure.Clusters(text) {
		w := measure.ClusterWidth(cluster)
		if x+w > right {
			break
		}
		x += c.Write(x, y, cluster, style)
	}
	return x
}

// textEntryUnits splits text into clusters, keeping "\n" as its own unit.
// CR and CRLF become LF, tabs become four spaces and other controls are dropped.
// Lines are split before clustering because measure.Clusters drops controls.
func textEntryUnits(text string) []string {
	text = textEntryNormalizer.Replace(text)
	if text == "" {
		return nil
	}
	var units []string
	for i, line := range strings.Split(text, "\n") {
		if i > 0 {
			units = append(units, "\n")
		}
		line = strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return -1
			}
			return r
		}, line)
		units = append(units, measure.Clusters(line)...)
	}
	return units
}

var (
	textEntryNormalizer = strings.NewReplacer("\r\n", "\n", "\r", "\n", "\t", "    ")
	textEntryKeyAliases = strings.NewReplacer("pagedown", "pgdown", "pgdn", "pgdown", "pageup", "pgup")
)

// normalizeTextEntryKey maps page-key aliases to the names the decoder emits.
func normalizeTextEntryKey(key string) string {
	return textEntryKeyAliases.Replace(key)
}

func textEntryBound(name string, bindings []string) bool {
	for _, b := range bindings {
		if normalizeTextEntryKey(b) == name {
			return true
		}
	}
	return false
}

func isTextEntryVerticalKey(name string) bool {
	switch strings.TrimPrefix(name, "shift-") {
	case "up", "down", "pgup", "pgdown":
		return true
	}
	return false
}

// textEntryClass groups units for word selection: 0 word, 1 space, 2 other.
func textEntryClass(unit string) int {
	if unit == "\n" {
		return 3
	}
	if isTextEntryWord(unit) {
		return 0
	}
	if r, _ := utf8.DecodeRuneInString(unit); unicode.IsSpace(r) {
		return 1
	}
	return 2
}

func isTextEntryWord(unit string) bool {
	r, _ := utf8.DecodeRuneInString(unit)
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
