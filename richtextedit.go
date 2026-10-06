// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"reflect"
	"strings"
	"time"
	"unicode"

	"ubunatic.com/loom/measure"
)

// RichPosition identifies a rune offset in a logical rich document line.
type RichPosition struct {
	Line   int
	Offset int
}

// RichTextEdit edits a RichDocument with styled spans, selection, and a
// scrolling viewport.
type RichTextEdit struct {
	Document       *RichDocument
	Cursor         RichPosition
	SelectionFrom  RichPosition
	SelectionTo    RichPosition
	HasSelection   bool
	SelectionStyle Style
	ActiveStyle    Style
	ScrollX        int
	ScrollY        int
	ShowCursor     bool
	ShowPopover    bool
	// ViewMode makes the editor read-only: navigation, selection and copy work;
	// every edit, style key and the format popover are disabled and edit keys
	// return Ignored so the app can use them.
	ViewMode bool
	// BoxMode draws box glyphs into adjacent document cells with arrow keys.
	BoxMode bool
	// GhostCursorEnabled allows vertical and horizontal navigation into empty
	// document space. The default comes from spec/defaults.yaml.
	GhostCursorEnabled bool
	// ShowFileBar reserves the final row for Save, Save as, document status,
	// and keyboard hints. It is disabled by default to preserve the viewport.
	ShowFileBar bool
	// FilePath is the path associated with the document by SaveAs or Save.
	FilePath string
	// LastSaveError holds the most recent save failure, including an error
	// returned after choosing a destination in the asynchronous save picker.
	LastSaveError error
	// SerializeDocument optionally serializes the document for saving. A nil
	// callback uses RichDocument.ToANSI.
	SerializeDocument func(*RichDocument) ([]byte, error)

	focused                bool
	lastRect               Rect
	savePicker             *FilePicker
	savePopup              *Popup
	helpPopup              *Popup
	fileBar                *richTextEditFileBar
	savedDocument          []RichLine
	selectionAnchor        RichPosition
	selectionExtending     bool
	dragSelecting          bool
	activeStyleSet         bool
	popoverButtons         []richPopoverButton
	popoverPalette         string
	popoverSwatches        []richPopoverSwatch
	popoverSubmenu         string
	popoverBoxChoices      []richPopoverBoxChoice
	popoverRelease         bool
	popoverFocus           int
	popoverFocusSet        bool
	popoverSubmenuFocus    int
	popoverSubmenuFocusSet bool
	popoverSuppressed      bool
	popoverAtCursor        bool
	cursorInVoid           bool
	voidLine               int
	voidColumn             int
	clipboard              []RichLine
	undoStack              []richSnapshot
	redoStack              []richSnapshot
	typingRun              bool
	clickCount             int
	lastClickAt            time.Time
	lastClickPos           RichPosition
	now                    func() time.Time
	boxStrokeBefore        *richSnapshot
	boxSelection           *richBoxSelection
	boxStrokeFG            Color
	boxStrokeFGSet         bool
}

const (
	richUndoLimit        = 100
	richMultiClickWindow = 400 * time.Millisecond
)

type richPopoverButton struct {
	label   string
	action  string
	rect    Rect
	enabled bool
}

type richPopoverSwatch struct {
	color Color
	rect  Rect
}

type richPopoverBoxChoice struct {
	label string
	style BoxBorderStyle
	rect  Rect
}

var richPopoverActions = []string{"B", "I", "U", "S", "Link", "#FG", "#BG", "Box", "Draw"}

var _ Widget = (*RichTextEdit)(nil)

// NewRichTextEdit creates a rich text view for doc. A nil document is treated
// as an empty document.
func NewRichTextEdit(doc *RichDocument) *RichTextEdit {
	e := &RichTextEdit{Document: doc, ShowCursor: true, ShowPopover: true, GhostCursorEnabled: SpeccedDefaults.RichTextEdit.GhostCursorEnabled, focused: true}
	e.savedDocument = cloneRichDocumentLines(doc)
	return e
}

// CursorShape reports the terminal cursor style required by the editor mode.
func (e *RichTextEdit) CursorShape() CursorShape {
	if e.BoxMode {
		return CursorShapeBlock
	}
	return CursorShapeBar
}

// NewRichTextView creates a read-only RichTextEdit for doc (ViewMode, no popover).
func NewRichTextView(doc *RichDocument) *RichTextEdit {
	e := NewRichTextEdit(doc)
	e.ViewMode, e.ShowPopover = true, false
	return e
}

// Focused reports whether the editor currently has keyboard focus.
func (e *RichTextEdit) Focused() bool { return e.focused }

// HotkeyHint returns the most useful file and help shortcuts for the available width.
func (e *RichTextEdit) HotkeyHint(width int) string {
	full := "F1 Help · Alt+F File · Ctrl+S Save · Ctrl+Shift+S Save as"
	if StringWidth(full) <= width {
		return full
	}
	compact := "F1 Help · Ctrl+S Save · Ctrl+Shift+S As"
	if StringWidth(compact) <= width {
		return compact
	}
	return TruncateText("F1 Help · Ctrl+S Save", max(0, width), "")
}

// SetFocus sets the editor focus state.
func (e *RichTextEdit) SetFocus(focused bool) {
	if e.focused && !focused && e.BoxMode {
		e.toggleBoxMode()
	}
	if !focused {
		e.clearPopoverKeyboardState(false)
	}
	e.focused = focused
}

// SetSelection sets a half-open selection range in rune offsets.
func (e *RichTextEdit) SetSelection(from, to RichPosition) {
	e.clearPopoverKeyboardState(false)
	e.popoverSuppressed = false
	e.cursorInVoid = false
	e.boxSelection = nil
	e.SelectionFrom, e.SelectionTo = from, to
	e.HasSelection = from != to
	e.selectionExtending = false
}

// ClearSelection removes the current selection.
func (e *RichTextEdit) ClearSelection() {
	e.clearPopoverKeyboardState(false)
	e.popoverSuppressed = false
	e.boxSelection = nil
	e.HasSelection = false
	e.selectionExtending = false
}

// Draw renders visible lines and styled spans, scrolls the cursor into view,
// paints selected text, and places the terminal cursor when focused.
func (e *RichTextEdit) Draw(c *Canvas, r Rect) {
	if c == nil {
		return
	}
	if e.ViewMode {
		e.clearPopoverKeyboardState(false)
	}
	fullRect := r
	if e.ShowFileBar && r.H > 0 {
		r.H--
	}
	e.lastRect = r
	e.popoverButtons = nil
	if r.W <= 0 || fullRect.H <= 0 {
		return
	}
	if r.H <= 0 {
		if e.ShowFileBar {
			e.ensureFileBar().Draw(c, fullRect)
		}
		return
	}
	c.PaintSurface(r, Style{})
	lines := e.documentLines()
	if !e.cursorInVoid {
		e.clampPosition(&e.Cursor, lines)
		e.Cursor = e.normalizePosition(e.Cursor, 0)
	}
	cursorLine, cursorCol := e.cursorCoordinates(lines)
	if cursorLine < e.ScrollY {
		e.ScrollY = cursorLine
	}
	if cursorLine >= e.ScrollY+r.H {
		e.ScrollY = cursorLine - r.H + 1
	}
	if cursorCol < e.ScrollX {
		e.ScrollX = cursorCol
	}
	if cursorCol >= e.ScrollX+r.W {
		e.ScrollX = cursorCol - r.W + 1
	}
	e.ScrollX = max(0, e.ScrollX)
	maxScrollY := max(0, len(lines)-r.H)
	if e.cursorInVoid {
		maxScrollY = max(maxScrollY, cursorLine-r.H+1)
	}
	e.ScrollY = min(max(0, e.ScrollY), maxScrollY)

	for row := 0; row < r.H && e.ScrollY+row < len(lines); row++ {
		e.drawLine(c, r, lines[e.ScrollY+row], e.ScrollY+row)
	}
	e.drawPopover(c, r, lines)
	if e.ShowCursor && e.focused && !e.ViewMode {
		x := cursorCol - e.ScrollX
		y := cursorLine - e.ScrollY
		if x >= 0 && x < r.W && y >= 0 && y < r.H {
			c.CursorX = r.X + x
			c.CursorY = r.Y + y
			c.CursorShape = e.CursorShape()
			c.CursorShapeSet = true
		}
	}
	if e.ShowFileBar {
		e.ensureFileBar().Draw(c, fullRect)
	}
	if e.savePopup != nil {
		e.savePopup.Width = min(SpeccedDefaults.RichTextEdit.SavePopupMaxWidth, max(4, r.W-4))
		e.savePopup.Height = min(SpeccedDefaults.RichTextEdit.SavePopupMaxHeight, max(3, r.H-2))
		e.savePopup.Draw(c, r)
	}
	if e.helpPopup != nil {
		e.helpPopup.Width = min(72, max(4, r.W-2))
		e.helpPopup.Height = min(18, max(3, r.H))
		e.helpPopup.Draw(c, r)
	}
}

func (e *RichTextEdit) documentLines() []RichLine {
	if e.Document == nil || len(e.Document.Lines) == 0 {
		return []RichLine{{}}
	}
	return e.Document.Lines
}

func (e *RichTextEdit) cursorCoordinates(lines []RichLine) (int, int) {
	if e.cursorInVoid {
		return e.voidLine, e.voidColumn
	}
	return e.Cursor.Line, richLineColumn(lines[e.Cursor.Line], e.Cursor.Offset)
}

func (e *RichTextEdit) materializeVoidCursor() {
	if !e.cursorInVoid {
		return
	}
	e.ensureDocument()
	for len(e.Document.Lines) <= e.voidLine {
		e.Document.Lines = append(e.Document.Lines, RichLine{})
	}
	line := e.Document.Lines[e.voidLine]
	width := richLineColumn(line, richLineRuneCount(line))
	if width < e.voidColumn {
		line.Spans = mergeRichSpans(append(line.Spans, RichSpan{Text: strings.Repeat(" ", e.voidColumn-width)}))
	}
	e.Document.Lines[e.voidLine] = line
	e.Cursor = RichPosition{Line: e.voidLine, Offset: richLineOffsetAtColumn(line, e.voidColumn)}
	e.cursorInVoid = false
}

func (e *RichTextEdit) moveToVoid(line, column int, extend bool) {
	if !e.GhostCursorEnabled {
		return
	}
	if line < 0 || column < 0 {
		return
	}
	if !extend {
		e.ClearSelection()
	}
	e.cursorInVoid = true
	e.voidLine, e.voidColumn = line, column
	e.boxSelection = nil
}

func (e *RichTextEdit) clampPosition(pos *RichPosition, lines []RichLine) {
	pos.Line = min(max(0, pos.Line), len(lines)-1)
	lineRunes := 0
	for _, span := range lines[pos.Line].Spans {
		lineRunes += len([]rune(span.Text))
	}
	pos.Offset = min(max(0, pos.Offset), lineRunes)
}

func richLineColumn(line RichLine, offset int) int {
	col, consumed := 0, 0
	for _, span := range line.Spans {
		for _, cluster := range measure.Clusters(span.Text) {
			clusterRunes := len([]rune(cluster))
			if consumed+clusterRunes > offset {
				return col
			}
			consumed += clusterRunes
			col += measure.StringWidth(cluster)
		}
	}
	return col
}

func (e *RichTextEdit) drawLine(c *Canvas, rect Rect, line RichLine, lineIndex int) {
	start, end := e.selectionLineBounds(lineIndex)
	col, offset := 0, 0
	for _, span := range line.Spans {
		for _, cluster := range measure.Clusters(span.Text) {
			w := measure.StringWidth(cluster)
			runeCount := len([]rune(cluster))
			selected := e.HasSelection && offset < end && offset+runeCount > start
			x := rect.X + col - e.ScrollX
			if w > 0 && x >= rect.X && x+w <= rect.X+rect.W {
				style := e.spanStyle(span)
				if selected {
					style = e.selectionStyle(style)
				}
				c.Set(x, rect.Y+lineIndex-e.ScrollY, Cell{Text: cluster, Style: style})
			}
			col += w
			offset += runeCount
		}
	}
}

func (e *RichTextEdit) selectionLineBounds(line int) (int, int) {
	if e.boxSelection != nil {
		for _, slice := range e.boxSelection.rows {
			if slice.line == line {
				return slice.start, slice.end
			}
		}
		return 0, 0
	}
	from, to := e.selectionBounds()
	if line < from.Line || line > to.Line {
		return 0, 0
	}
	return richSelectionLineBounds(line, from, to)
}

func (e *RichTextEdit) selectionLineBoundsForRange(line int, from, to RichPosition) (int, int) {
	if e.boxSelection != nil {
		return e.selectionLineBounds(line)
	}
	if line < from.Line || line > to.Line {
		return 0, 0
	}
	return richSelectionLineBounds(line, from, to)
}

// spanStyle returns the draw style of span; link spans get the spec link style
// for any attribute the span leaves unset, in edit and view mode, focused or not.
func (e *RichTextEdit) spanStyle(span RichSpan) Style {
	style := span.Style
	if span.Link == "" {
		return style
	}
	defs := SpeccedDefaults.RichTextEdit
	if style.FG == ColorReset() {
		style.FG = ColorIndex(uint8(defs.LinkFG))
	}
	style.Underline = style.Underline || defs.LinkUnderline
	return style
}

func (e *RichTextEdit) selectionBounds() (RichPosition, RichPosition) {
	from, to := e.SelectionFrom, e.SelectionTo
	if richPositionBefore(to, from) {
		from, to = to, from
	}
	return from, to
}

func (e *RichTextEdit) clearPopoverKeyboardState(suppress bool) {
	e.popoverFocusSet = false
	e.popoverSubmenuFocusSet = false
	e.popoverPalette = ""
	e.popoverSwatches = nil
	e.popoverSubmenu = ""
	e.popoverBoxChoices = nil
	e.popoverAtCursor = false
	if suppress {
		e.popoverSuppressed = true
	}
}

func (e *RichTextEdit) consumePopoverKey(key KeyEvent) (EventResult, bool) {
	if e.ViewMode || !e.ShowPopover || (!e.HasSelection && !e.popoverAtCursor) || e.popoverSuppressed {
		return Ignored(), false
	}
	submenuOpen := e.popoverPalette != "" || e.popoverSubmenu != ""
	popoverArrow := key.Is("up", "down", "left", "right")
	if popoverArrow && !submenuOpen && !e.popoverFocusSet && len(e.popoverButtons) == 0 {
		return Ignored(), false
	}
	if !key.Is("tab", "shift-tab", "backtab", "enter", "return", "space", " ", "esc", "up", "down", "left", "right") {
		if e.popoverAtCursor || e.popoverFocusSet || e.popoverPalette != "" {
			e.clearPopoverKeyboardState(true)
		}
		return Ignored(), false
	}
	if !e.popoverCanShow() {
		e.clearPopoverKeyboardState(true)
		return Handled(), true
	}
	if submenuOpen && !e.popoverSubmenuCanShow() {
		e.clearPopoverKeyboardState(true)
		return Handled(), true
	}
	firstTab := !e.popoverFocusSet && key.Is("tab")
	if !e.popoverFocusSet {
		e.popoverFocus = e.initialPopoverFocus()
		e.popoverFocusSet = true
	}
	if submenuOpen {
		return e.consumePopoverSubmenuKey(key)
	}
	switch {
	case key.Is("esc"):
		if e.popoverPalette != "" || e.popoverSubmenu != "" {
			e.popoverPalette = ""
			e.popoverSwatches = nil
			e.popoverSubmenu = ""
			e.popoverBoxChoices = nil
		} else {
			e.clearPopoverKeyboardState(true)
		}
		return Handled(), true
	case key.Is("tab", "shift-tab", "backtab"):
		if firstTab {
			return Handled(), true
		}
		direction := 1
		if key.Is("shift-tab", "backtab") {
			direction = -1
		}
		e.popoverFocus = e.nextPopoverFocus(e.popoverFocus, direction)
		return Handled(), true
	case key.Is("left", "right"):
		direction := 1
		if key.Is("left") {
			direction = -1
		}
		e.popoverFocus = e.nextPopoverFocus(e.popoverFocus, direction)
		return Handled(), true
	case key.Is("up", "down"):
		return Handled(), true
	case key.Is("enter", "return", "space", " "):
		if e.popoverFocus < 0 || e.popoverFocus >= len(richPopoverActions) || !e.popoverActionEnabled(e.popoverFocus) {
			return Handled(), true
		}
		action := richPopoverActions[e.popoverFocus]
		if action == "Link" {
			return Handled(), true
		}
		if action == "#FG" || action == "#BG" {
			e.applyPopoverAction(action)
			e.popoverSubmenuFocus = e.initialPopoverSubmenuFocus()
			e.popoverSubmenuFocusSet = true
			if !e.popoverSubmenuCanShow() {
				e.clearPopoverKeyboardState(true)
			}
			return Handled(), true
		}
		e.applyPopoverAction(action)
		if action == "Box" {
			e.popoverSubmenuFocus = e.initialPopoverSubmenuFocus()
			e.popoverSubmenuFocusSet = true
			if !e.popoverSubmenuCanShow() {
				e.clearPopoverKeyboardState(true)
			}
			return Handled(), true
		}
		e.clearPopoverKeyboardState(action != "Draw" && e.HasSelection)
		return Handled(), true
	default:
		return Ignored(), false
	}
}

func (e *RichTextEdit) consumePopoverSubmenuKey(key KeyEvent) (EventResult, bool) {
	if !e.popoverSubmenuFocusSet {
		e.popoverSubmenuFocus = e.initialPopoverSubmenuFocus()
		e.popoverSubmenuFocusSet = true
	}
	if key.Is("esc") {
		e.popoverPalette = ""
		e.popoverSwatches = nil
		e.popoverSubmenu = ""
		e.popoverBoxChoices = nil
		e.popoverSubmenuFocusSet = false
		return Handled(), true
	}
	if key.Is("tab", "shift-tab", "backtab", "up", "down", "left", "right") {
		direction := 1
		if key.Is("shift-tab", "backtab", "up", "left") {
			direction = -1
		}
		count := e.popoverSubmenuChoiceCount()
		if count > 0 {
			e.popoverSubmenuFocus = (e.popoverSubmenuFocus + direction + count) % count
		}
		return Handled(), true
	}
	if key.Is("enter", "return", "space", " ") {
		if e.popoverPalette != "" {
			palette := e.popoverPalette
			e.applyPopoverColor(palette, ColorIndex(uint8(e.popoverSubmenuFocus)))
			e.clearPopoverKeyboardState(e.HasSelection)
			return Handled(), true
		}
		labels := SpeccedDefaults.RichTextEdit.BoxStyleLabels
		if e.popoverSubmenu == "Box" && e.popoverSubmenuFocus >= 0 && e.popoverSubmenuFocus < len(labels) {
			e.applyBoxStyleChoice(labels[e.popoverSubmenuFocus])
			return Handled(), true
		}
		return Handled(), true
	}
	return Ignored(), false
}

func (e *RichTextEdit) popoverSubmenuChoiceCount() int {
	if e.popoverPalette != "" {
		return 16
	}
	if e.popoverSubmenu == "Box" {
		return len(SpeccedDefaults.RichTextEdit.BoxStyleLabels)
	}
	return 0
}

func (e *RichTextEdit) initialPopoverSubmenuFocus() int { return 0 }

func (e *RichTextEdit) popoverSubmenuCanShow() bool {
	if !e.popoverCanShow() {
		return false
	}
	r := e.lastRect
	if e.popoverPalette != "" && r.W < 16 {
		return false
	}
	if e.popoverSubmenu == "Box" {
		width := 2 + len(SpeccedDefaults.RichTextEdit.BoxStyleLabels) - 1
		for _, label := range SpeccedDefaults.RichTextEdit.BoxStyleLabels {
			width += len(label) + 2
		}
		if width > r.W {
			return false
		}
	}
	return true
}

func (e *RichTextEdit) nextPopoverFocus(current, direction int) int {
	for range len(richPopoverActions) {
		current = (current + direction + len(richPopoverActions)) % len(richPopoverActions)
		if e.popoverActionEnabled(current) {
			return current
		}
	}
	return current
}

func (e *RichTextEdit) popoverActionEnabled(index int) bool {
	if index < 0 || index >= len(richPopoverActions) || richPopoverActions[index] == "Link" {
		return false
	}
	return e.HasSelection || richPopoverActions[index] == "Draw"
}

func (e *RichTextEdit) initialPopoverFocus() int {
	if !e.HasSelection && e.popoverAtCursor {
		return 8
	}
	if e.boxSelection != nil {
		return 7
	}
	for i, active := range []func(Style) bool{
		func(s Style) bool { return s.Bold },
		func(s Style) bool { return s.Italic },
		func(s Style) bool { return s.Underline },
	} {
		if e.selectionUniformlyHas(active) {
			return i
		}
	}
	return 0
}

func (e *RichTextEdit) selectionUniformlyHas(active func(Style) bool) bool {
	lines := e.documentLines()
	from, to := e.selectionBounds()
	seen := false
	for line := from.Line; line <= to.Line && line < len(lines); line++ {
		start, end := e.selectionLineBounds(line)
		if end <= start {
			continue
		}
		offset := 0
		for _, span := range lines[line].Spans {
			n := len([]rune(span.Text))
			lo, hi := max(start, offset), min(end, offset+n)
			if lo < hi {
				seen = true
				if !active(e.spanStyle(span)) {
					return false
				}
			}
			offset += n
		}
	}
	return seen
}

func (e *RichTextEdit) popoverCanShow() bool {
	r := e.lastRect
	if r.W < 18 || r.H < 3 || (!e.HasSelection && !e.popoverAtCursor) || e.ViewMode || !e.ShowPopover {
		return false
	}
	lines := e.documentLines()
	anchorLine, anchorCol := e.cursorCoordinates(lines)
	if !e.popoverAtCursor {
		from, _ := e.selectionBounds()
		e.clampPosition(&from, lines)
		anchorLine, anchorCol = from.Line, richLineColumn(lines[from.Line], from.Offset)
	}
	anchorX := anchorCol - e.ScrollX
	anchorY := anchorLine - e.ScrollY
	if anchorX < 0 || anchorX >= r.W || anchorY < 0 || anchorY >= r.H {
		return false
	}
	width := 2
	for i, label := range SpeccedDefaults.RichTextEdit.PopoverLabels {
		width += len(label) + 1
		if i > 0 {
			width++
		}
	}
	barY := 0
	if anchorY >= 2 {
		barY = anchorY - 2
	} else if anchorY+2 < r.H {
		barY = anchorY + 2
	} else {
		return false
	}
	barX := min(max(0, anchorX-width/2), r.W-width)
	return barX >= 0 && barY >= 0 && barY < r.H
}

func richPositionBefore(a, b RichPosition) bool {
	return a.Line < b.Line || a.Line == b.Line && a.Offset < b.Offset
}

func (e *RichTextEdit) selectionStyle(base Style) Style {
	selection := e.SelectionStyle
	if selection == (Style{}) {
		base.BG = ColorIndex(uint8(SpeccedDefaults.RichTextEdit.SelectionBG))
		return base
	}
	if selection.FG != ColorReset() {
		base.FG = selection.FG
	}
	if selection.BG != ColorReset() {
		base.BG = selection.BG
	}
	base.Bold = base.Bold || selection.Bold
	base.Dim = base.Dim || selection.Dim
	base.Italic = base.Italic || selection.Italic
	base.Underline = base.Underline || selection.Underline
	base.Strike = base.Strike || selection.Strike
	base.Invert = base.Invert || selection.Invert
	return base
}

// ConsumeKey applies navigation, text editing, selection, and inline formatting.
func (e *RichTextEdit) ConsumeKey(key KeyEvent) EventResult {
	if e.helpPopup != nil {
		result := e.helpPopup.ConsumeKey(key)
		if !e.helpPopup.Open {
			e.helpPopup = nil
		}
		return result
	}
	if e.savePopup != nil {
		if key.Is("esc") {
			return e.savePicker.ConsumeKey(key)
		}
		return e.savePopup.ConsumeKey(key)
	}
	if key.Is("f1") {
		e.ensureFileBar()
		e.fileBar.menu.Open = false
		e.fileBar.menu.SetFocus(false)
		e.helpPopup = NewPopup("RichTextEdit Help", newRichTextEditHelp())
		return Handled()
	}
	if key.Is("f10") {
		return Ignored()
	}
	if e.ShowFileBar {
		if result := e.ensureFileBar().ConsumeKey(key); result.Consumed {
			return result
		}
	}
	e.ensureDocument()
	if !e.cursorInVoid {
		e.clampPosition(&e.Cursor, e.documentLines())
		e.Cursor = e.normalizePosition(e.Cursor, 0)
	}
	if !key.Is("enter", "return") && (key.Key != "" || key.Text == "") {
		e.typingRun = false
	}
	if e.ViewMode && !richViewModeKey(key) {
		e.clearPopoverKeyboardState(false)
		return Ignored()
	}
	if e.ViewMode {
		e.clearPopoverKeyboardState(false)
	}
	if e.ViewMode && e.BoxMode {
		e.toggleBoxMode()
	}
	if result, handled := e.consumePopoverKey(key); handled {
		return result
	}
	if key.Is("f5") {
		if e.HasSelection {
			e.wrapSelectionInBox()
		} else {
			e.toggleBoxMode()
		}
		return Handled()
	}
	if e.BoxMode {
		switch {
		case key.Is("esc"):
			e.toggleBoxMode()
			return Handled()
		case key.Is("up"):
			e.drawBoxStep(-1, 0)
			return Handled()
		case key.Is("right"):
			e.drawBoxStep(0, 1)
			return Handled()
		case key.Is("down"):
			e.drawBoxStep(1, 0)
			return Handled()
		case key.Is("left"):
			e.drawBoxStep(0, -1)
			return Handled()
		case key.Is("ctrl-z", "ctrl-y", "ctrl-r", "ctrl-shift-y", "ctrl-shift-z"):
			e.toggleBoxMode()
		default:
			return Ignored()
		}
	}
	switch {
	case key.Is("ctrl-b"):
		e.toggleAttribute(func(s *Style, on bool) { s.Bold = on }, func(s Style) bool { return s.Bold })
	case key.Is("ctrl-i"):
		e.toggleAttribute(func(s *Style, on bool) { s.Italic = on }, func(s Style) bool { return s.Italic })
	case key.Is("ctrl-u"):
		e.toggleAttribute(func(s *Style, on bool) { s.Underline = on }, func(s Style) bool { return s.Underline })
	case key.Is("ctrl-space"):
		e.clearPopoverKeyboardState(false)
		e.popoverSuppressed = false
		e.popoverAtCursor = false
		if !e.cursorInVoid && e.selectBoxAtCursor() {
			e.ShowPopover = true
		} else if !e.cursorInVoid && e.selectWord() {
			e.ShowPopover = true
		} else {
			e.ClearSelection()
			e.popoverAtCursor = true
			e.ShowPopover = true
		}
	case key.Is("ctrl-c", "ctrl-insert"):
		if e.cursorInVoid {
			return Ignored()
		}
		if !e.copySelectionOrWord() {
			return Ignored()
		}
	case key.Is("ctrl-x", "shift-delete"):
		if e.cursorInVoid {
			return Ignored()
		}
		e.mutate(false, false, func() {
			if e.copySelectionOrWord() {
				e.deleteSelection()
			}
		})
	case key.Is("ctrl-v", "shift-insert"):
		e.mutate(false, false, func() { e.insertRich(e.clipboard) })
	case key.Is("ctrl-z", "ctrl-y"):
		e.undo()
	case key.Is("ctrl-r", "ctrl-shift-y", "ctrl-shift-z"):
		e.redo()
	case key.Is("left", "shift-left"):
		e.moveHorizontal(-1, key.Is("shift-left"))
	case key.Is("right", "shift-right"):
		e.moveHorizontal(1, key.Is("shift-right"))
	case key.Is("up", "shift-up"):
		e.moveVertical(-1, key.Is("shift-up"))
	case key.Is("down", "shift-down"):
		e.moveVertical(1, key.Is("shift-down"))
	case key.Is("home", "shift-home", "ctrl-a", "ctrl-shift-a"):
		e.moveCursor(RichPosition{Line: e.Cursor.Line}, key.Is("shift-home", "ctrl-shift-a"))
	case key.Is("end", "shift-end", "ctrl-e", "ctrl-shift-e"):
		e.moveCursor(RichPosition{Line: e.Cursor.Line, Offset: richLineRuneCount(e.documentLines()[e.Cursor.Line])}, key.Is("shift-end", "ctrl-shift-e"))
	case key.Is("ctrl-left"):
		if e.cursorInVoid {
			return Ignored()
		}
		if e.HasSelection {
			from, _ := e.selectionBounds()
			e.moveCursor(from, false)
		} else {
			e.moveCursor(e.wordPosition(-1), false)
		}
	case key.Is("ctrl-right"):
		if e.cursorInVoid {
			return Ignored()
		}
		if e.HasSelection {
			_, to := e.selectionBounds()
			e.moveCursor(to, false)
		} else {
			e.moveCursor(e.wordPosition(1), false)
		}
	case key.Is("enter", "return"):
		e.mutate(true, true, func() { e.insertText("\n") })
	case key.Is("backspace"):
		if e.cursorInVoid {
			return Ignored()
		}
		e.mutate(false, false, func() { e.deleteDirection(-1) })
	case key.Is("delete"):
		if e.cursorInVoid {
			return Ignored()
		}
		e.mutate(false, false, func() { e.deleteDirection(1) })
	default:
		if key.Text == "" {
			return Ignored()
		}
		boundary := strings.ContainsFunc(key.Text, func(r rune) bool {
			return !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_')
		})
		e.mutate(true, boundary, func() { e.insertText(key.Text) })
	}
	return Handled()
}

// richViewModeKey reports whether key is navigation, selection or copy.
func richViewModeKey(key KeyEvent) bool {
	return key.Is("ctrl-c", "ctrl-insert", "left", "shift-left", "right", "shift-right", "up", "shift-up",
		"down", "shift-down", "home", "shift-home", "end", "shift-end", "ctrl-a", "ctrl-shift-a", "ctrl-e", "ctrl-shift-e", "ctrl-left", "ctrl-right")
}

// ConsumeMouse positions the cursor and supports click-drag selection.
func (e *RichTextEdit) ConsumeMouse(mouse MouseEvent) EventResult {
	if e.helpPopup != nil {
		result := e.helpPopup.ConsumeMouse(mouse)
		if !e.helpPopup.Open {
			e.helpPopup = nil
		}
		return result
	}
	if e.savePopup != nil {
		result := e.savePopup.ConsumeMouse(mouse)
		if !e.savePopup.Open {
			e.savePicker = nil
			e.savePopup = nil
		}
		return result
	}
	if e.ShowFileBar {
		menu := e.ensureFileBar().menu
		if menu.Open || menu.Focused() {
			result := menu.ConsumeMouse(mouse)
			if !menu.Open {
				menu.SetFocus(false)
			}
			if result.Consumed {
				return result
			}
			return Handled()
		}
		if result := menu.ConsumeMouse(mouse); result.Consumed {
			return result
		}
	}
	e.ensureDocument()
	if e.popoverRelease && mouse.Action == MouseRelease {
		e.popoverRelease = false
		return Handled()
	}
	if e.popoverHit(mouse.X, mouse.Y) {
		if mouse.Action == MousePress && mouse.Button == MouseLeft {
			for _, choice := range e.popoverBoxChoices {
				if choice.rect.Contains(mouse.X, mouse.Y) {
					e.applyBoxStyleChoice(choice.label)
					e.popoverRelease = true
					return Handled()
				}
			}
			for _, swatch := range e.popoverSwatches {
				if swatch.rect.Contains(mouse.X, mouse.Y) {
					e.applyPopoverColor(e.popoverPalette, swatch.color)
					e.popoverRelease = true
					return Handled()
				}
			}
			for _, button := range e.popoverButtons {
				if button.enabled && button.rect.Contains(mouse.X, mouse.Y) {
					e.popoverFocusSet = false
					e.applyPopoverAction(button.action)
					e.popoverRelease = true
					break
				}
			}
		}
		return Handled()
	}
	if mouse.Action == MousePress {
		e.popoverPalette = ""
		e.popoverSwatches = nil
		e.popoverSubmenu = ""
		e.popoverBoxChoices = nil
	}
	switch mouse.Action {
	case MouseScrollUp:
		e.ScrollY = max(0, e.ScrollY-1)
		return Handled()
	case MouseScrollDown:
		e.ScrollY++
		return Handled()
	case MousePress:
		if mouse.Button != MouseLeft || !e.mouseInBounds(mouse) {
			return Ignored()
		}
		e.typingRun = false
		position := e.mousePosition(mouse)
		e.cursorInVoid = false
		e.Cursor, e.selectionAnchor = position, position
		e.ClearSelection()
		switch e.registerClick(position) {
		case 2:
			e.dragSelecting = false
			e.selectWord()
		case 3:
			e.dragSelecting = false
			e.selectLine()
		default:
			e.dragSelecting = true
		}
		return Handled()
	case MouseDrag:
		if !e.dragSelecting {
			return Ignored()
		}
		e.extendMouseSelection(mouse)
		return Handled()
	case MouseRelease:
		if !e.dragSelecting || mouse.Button != MouseLeft {
			return Ignored()
		}
		e.extendMouseSelection(mouse)
		e.dragSelecting = false
		return Handled()
	default:
		return Ignored()
	}
}

func (e *RichTextEdit) drawPopover(c *Canvas, r Rect, lines []RichLine) {
	e.popoverButtons = nil
	e.popoverSwatches = nil
	e.popoverBoxChoices = nil
	if (e.popoverPalette != "" || e.popoverSubmenu != "") && !e.popoverSubmenuCanShow() {
		e.clearPopoverKeyboardState(true)
	}
	if e.ViewMode || !e.ShowPopover || (!e.HasSelection && !e.popoverAtCursor) || e.popoverSuppressed || r.W < 18 || r.H < 3 {
		return
	}
	anchorLine, anchorCol := e.cursorCoordinates(lines)
	if !e.popoverAtCursor {
		from, _ := e.selectionBounds()
		e.clampPosition(&from, lines)
		anchorLine, anchorCol = from.Line, richLineColumn(lines[from.Line], from.Offset)
	}
	anchorX := anchorCol - e.ScrollX
	anchorY := anchorLine - e.ScrollY
	if anchorX < 0 || anchorX >= r.W || anchorY < 0 || anchorY >= r.H {
		return
	}
	width := 2
	labels := SpeccedDefaults.RichTextEdit.PopoverLabels
	for i, label := range labels {
		width += len(label) + 1
		if i > 0 {
			width++
		}
	}
	defs := SpeccedDefaults.RichTextEdit
	barY, pointerY, pointer := 0, 0, defs.PointerUpGlyph
	if anchorY >= 2 {
		barY, pointerY, pointer = anchorY-2, anchorY-1, defs.PointerDownGlyph
	} else if anchorY+2 < r.H {
		barY, pointerY = anchorY+2, anchorY+1
	} else {
		return
	}
	barX := min(max(0, anchorX-width/2), r.W-width)
	if barX < 0 { // Keep the complete action row visible in very narrow widgets.
		return
	}
	toolbarStyle := Style{FG: ColorIndex(uint8(defs.ToolbarFG)), BG: ColorIndex(uint8(defs.ToolbarBG)), Bold: true}
	for x := 0; x < width; x++ {
		c.Set(r.X+barX+x, r.Y+barY, Cell{Text: " ", Style: toolbarStyle})
	}
	c.Set(r.X+barX, r.Y+barY, Cell{Text: "[", Style: toolbarStyle})
	c.Set(r.X+barX+width-1, r.Y+barY, Cell{Text: "]", Style: toolbarStyle})
	x := barX + 1
	for i, label := range labels {
		button := Rect{X: x, Y: barY, W: len(label) + 1, H: 1}
		action := richPopoverActions[i]
		enabled := e.popoverActionEnabled(i)
		e.popoverButtons = append(e.popoverButtons, richPopoverButton{label: label, action: action, rect: button, enabled: enabled})
		for j, ch := range " " + label {
			style := toolbarStyle
			if !enabled {
				style.Dim = true
			}
			if label == "B" && ch == 'B' {
				style.Bold = true
			}
			if e.popoverFocusSet && i == e.popoverFocus {
				style.FG = ColorIndex(uint8(defs.PopoverFocusFG))
				style.BG = ColorIndex(uint8(defs.PopoverFocusBG))
				style.Dim = false
			}
			c.Set(r.X+x+j, r.Y+barY, Cell{Text: string(ch), Style: style})
		}
		x += button.W
		if i < len(labels)-1 {
			c.Set(r.X+x, r.Y+barY, Cell{Text: defs.SeparatorGlyph, Style: Style{FG: ColorIndex(uint8(defs.SeparatorFG)), BG: ColorIndex(uint8(defs.SeparatorBG))}})
			x++
		}
	}
	c.Set(r.X+barX+min(max(0, anchorX-barX), width-1), r.Y+pointerY, Cell{Text: pointer, Style: Style{FG: ColorIndex(uint8(defs.PointerFG))}})
	e.drawPopoverPalette(c, r, barX, barY)
	e.drawPopoverBoxStyles(c, r, barY)
}

func (e *RichTextEdit) drawPopoverPalette(c *Canvas, r Rect, barX, barY int) {
	if e.popoverPalette == "" || r.W < 16 {
		return
	}
	var button Rect
	for _, item := range e.popoverButtons {
		if item.label == e.popoverPalette {
			button = item.rect
			break
		}
	}
	if button.W == 0 {
		e.popoverPalette = ""
		return
	}
	paletteY := barY + 1
	if paletteY >= r.H {
		paletteY = barY - 1
	}
	if paletteY < 0 || paletteY >= r.H {
		return
	}
	paletteX := min(max(0, button.X+button.W/2-8), r.W-16)
	for i := 0; i < 16; i++ {
		color := ColorIndex(uint8(i))
		cellStyle := Style{BG: color}
		if e.popoverSubmenuFocusSet && i == e.popoverSubmenuFocus {
			cellStyle.Invert = true
		}
		c.Set(r.X+paletteX+i, r.Y+paletteY, Cell{Text: " ", Style: cellStyle})
		e.popoverSwatches = append(e.popoverSwatches, richPopoverSwatch{color: color, rect: Rect{X: paletteX + i, Y: paletteY, W: 1, H: 1}})
	}
}

func (e *RichTextEdit) drawPopoverBoxStyles(c *Canvas, r Rect, barY int) {
	if e.popoverSubmenu != "Box" {
		return
	}
	var button Rect
	for _, item := range e.popoverButtons {
		if item.action == "Box" {
			button = item.rect
			break
		}
	}
	if button.W == 0 {
		e.popoverSubmenu = ""
		return
	}
	labels := SpeccedDefaults.RichTextEdit.BoxStyleLabels
	width := 2 + len(labels) - 1
	for _, label := range labels {
		width += len(label) + 2
	}
	y := barY + 1
	if y >= r.H {
		y = barY - 1
	}
	if y < 0 || y >= r.H || width > r.W {
		return
	}
	x := min(max(0, button.X+button.W/2-width/2), r.W-width)
	style := Style{FG: ColorIndex(uint8(SpeccedDefaults.RichTextEdit.ToolbarFG)), BG: ColorIndex(uint8(SpeccedDefaults.RichTextEdit.ToolbarBG)), Bold: true}
	for col := 0; col < width; col++ {
		c.Set(r.X+x+col, r.Y+y, Cell{Text: " ", Style: style})
	}
	c.Set(r.X+x, r.Y+y, Cell{Text: "[", Style: style})
	c.Set(r.X+x+width-1, r.Y+y, Cell{Text: "]", Style: style})
	col := x + 1
	for i, label := range labels {
		choiceWidth := len(label) + 2
		choice := richPopoverBoxChoice{label: label, style: BoxBorderStyle(i), rect: Rect{X: col, Y: y, W: choiceWidth, H: 1}}
		e.popoverBoxChoices = append(e.popoverBoxChoices, choice)
		choiceStyle := style
		if e.popoverSubmenuFocusSet && i == e.popoverSubmenuFocus {
			choiceStyle.FG = ColorIndex(uint8(SpeccedDefaults.RichTextEdit.PopoverFocusFG))
			choiceStyle.BG = ColorIndex(uint8(SpeccedDefaults.RichTextEdit.PopoverFocusBG))
		}
		for j, ch := range " " + label + " " {
			c.Set(r.X+col+j, r.Y+y, Cell{Text: string(ch), Style: choiceStyle})
		}
		col += choiceWidth
		if i < len(labels)-1 {
			c.Set(r.X+col, r.Y+y, Cell{Text: SpeccedDefaults.RichTextEdit.SeparatorGlyph, Style: Style{FG: ColorIndex(uint8(SpeccedDefaults.RichTextEdit.SeparatorFG)), BG: ColorIndex(uint8(SpeccedDefaults.RichTextEdit.SeparatorBG))}})
			col++
		}
	}
}

func (e *RichTextEdit) popoverHit(x, y int) bool {
	if e.popoverSuppressed || !e.ShowPopover || (!e.HasSelection && !e.popoverAtCursor) || e.ViewMode {
		return false
	}
	for _, swatch := range e.popoverSwatches {
		if swatch.rect.Contains(x, y) {
			return true
		}
	}
	for _, choice := range e.popoverBoxChoices {
		if choice.rect.Contains(x, y) {
			return true
		}
	}
	for _, button := range e.popoverButtons {
		if button.enabled && button.rect.Contains(x, y) {
			return true
		}
	}
	if len(e.popoverButtons) == 0 {
		return false
	}
	first, last := e.popoverButtons[0].rect, e.popoverButtons[len(e.popoverButtons)-1].rect
	return y == first.Y && x >= first.X-1 && x <= last.X+last.W
}

func (e *RichTextEdit) applyPopoverAction(label string) {
	switch label {
	case "Box":
		e.popoverPalette = ""
		e.popoverSwatches = nil
		e.popoverSubmenuFocusSet = false
		if e.popoverSubmenu == "Box" {
			e.popoverSubmenu = ""
			e.popoverBoxChoices = nil
		} else {
			e.popoverSubmenu = "Box"
		}
	case "Draw":
		e.toggleBoxMode()
	case "B":
		e.toggleAttribute(func(s *Style, on bool) { s.Bold = on }, func(s Style) bool { return s.Bold })
	case "I":
		e.toggleAttribute(func(s *Style, on bool) { s.Italic = on }, func(s Style) bool { return s.Italic })
	case "U":
		e.toggleAttribute(func(s *Style, on bool) { s.Underline = on }, func(s Style) bool { return s.Underline })
	case "S":
		e.toggleAttribute(func(s *Style, on bool) { s.Strike = on }, func(s Style) bool { return s.Strike })
	case "#FG", "#BG":
		e.popoverSubmenu = ""
		e.popoverBoxChoices = nil
		e.popoverSubmenuFocusSet = false
		if e.popoverPalette == label {
			e.popoverPalette = ""
			e.popoverSwatches = nil
		} else {
			e.popoverPalette = label
		}
	}
}

func (e *RichTextEdit) applyBoxStyleChoice(label string) {
	labels := SpeccedDefaults.RichTextEdit.BoxStyleLabels
	style := BoxBorderStyleSharp
	for index, candidate := range labels {
		if strings.EqualFold(candidate, label) {
			style = BoxBorderStyle(index)
			break
		}
	}
	if e.boxSelection != nil {
		e.restyleSelectedBox(style)
	} else {
		e.wrapSelectionInBoxStyle(style)
	}
	e.popoverSubmenu = ""
	e.popoverBoxChoices = nil
	if e.HasSelection {
		e.clearPopoverKeyboardState(true)
	} else {
		e.clearPopoverKeyboardState(false)
	}
}

func (e *RichTextEdit) wrapSelectionInBox() {
	style := BoxBorderStyleSharp
	if strings.EqualFold(SpeccedDefaults.RichTextEdit.BoxStyleDefault, "rounded") {
		style = BoxBorderStyleRounded
	}
	e.wrapSelectionInBoxStyle(style)
}

func (e *RichTextEdit) wrapSelectionInBoxStyle(boxStyle BoxBorderStyle) {
	if !e.HasSelection || e.boxSelection != nil {
		return
	}
	e.mutate(false, false, func() {
		lines := e.documentLines()
		from, to := e.selectionBounds()
		e.clampPosition(&from, lines)
		e.clampPosition(&to, lines)
		from, to = expandRangeForPills(lines, from, to)
		if from == to {
			return
		}
		prefix, firstRest := splitRichLine(lines[from.Line], from.Offset)
		firstLength := richLineRuneCount(RichLine{Spans: firstRest})
		if from.Line == to.Line {
			firstLength = to.Offset - from.Offset
		}
		first, _ := splitRichLine(RichLine{Spans: firstRest}, firstLength)
		_, suffix := splitRichLine(lines[to.Line], to.Offset)
		content := make([]RichLine, 0, to.Line-from.Line+1)
		content = append(content, RichLine{Spans: first})
		for i := from.Line + 1; i < to.Line; i++ {
			content = append(content, RichLine{Spans: append([]RichSpan(nil), lines[i].Spans...)})
		}
		last, _ := splitRichLine(lines[to.Line], to.Offset)
		if to.Line != from.Line {
			content = append(content, RichLine{Spans: last})
		} else {
			content = []RichLine{{Spans: first}}
		}
		width := 0
		for _, line := range content {
			width = max(width, richLineColumn(line, richLineRuneCount(line)))
		}
		border, err := getBoxBorderGlyphs(boxStyle)
		if err != nil {
			return
		}
		style := Style{}
		if len(content) > 0 && len(content[0].Spans) > 0 {
			style = content[0].Spans[0].Style
		}
		boxed := []RichLine{{Spans: []RichSpan{{Text: border.TopLeft + strings.Repeat(border.Horizontal, width) + border.TopRight, Style: style}}}}
		for _, line := range content {
			padding := width - richLineColumn(line, richLineRuneCount(line))
			spans := []RichSpan{{Text: border.Vertical, Style: style}}
			spans = append(spans, line.Spans...)
			if padding > 0 {
				spans = append(spans, RichSpan{Text: strings.Repeat(" ", padding), Style: style})
			}
			spans = append(spans, RichSpan{Text: border.Vertical, Style: style})
			boxed = append(boxed, RichLine{Spans: mergeRichSpans(spans)})
		}
		boxed = append(boxed, RichLine{Spans: []RichSpan{{Text: border.BottomLeft + strings.Repeat(border.Horizontal, width) + border.BottomRight, Style: style}}})
		lastIndex := len(boxed) - 1
		updated := append([]RichLine(nil), lines[:from.Line]...)
		if richLineRuneCount(RichLine{Spans: prefix}) > 0 {
			updated = append(updated, RichLine{Spans: prefix})
		}
		boxLine := from.Line + len(updated) - len(lines[:from.Line])
		updated = append(updated, boxed...)
		if richLineRuneCount(RichLine{Spans: suffix}) > 0 {
			updated = append(updated, RichLine{Spans: suffix})
		}
		updated = append(updated, lines[to.Line+1:]...)
		e.Document.Lines = updated
		e.Cursor = RichPosition{Line: boxLine + lastIndex, Offset: richLineRuneCount(boxed[lastIndex])}
		e.ClearSelection()
	})
}

func (e *RichTextEdit) restyleSelectedBox(boxStyle BoxBorderStyle) {
	if e.boxSelection == nil || (boxStyle != BoxBorderStyleSharp && boxStyle != BoxBorderStyleRounded) {
		return
	}
	border, err := getBoxBorderGlyphs(boxStyle)
	if err != nil {
		return
	}
	e.mutate(false, false, func() {
		selection := e.boxSelection
		corners := []struct {
			line, column int
			arms         BoxArms
			glyph        string
		}{
			{selection.top, selection.left, BoxArmRight | BoxArmDown, border.TopLeft},
			{selection.top, selection.right, BoxArmLeft | BoxArmDown, border.TopRight},
			{selection.bottom, selection.left, BoxArmRight | BoxArmUp, border.BottomLeft},
			{selection.bottom, selection.right, BoxArmLeft | BoxArmUp, border.BottomRight},
		}
		for _, corner := range corners {
			line := e.Document.Lines[corner.line]
			if BoxGlyphArms(richLineCell(line, corner.column)) != corner.arms {
				continue // Preserve crossing/shared-edge junction arms.
			}
			e.Document.Lines[corner.line] = richLineSetCell(line, corner.column, corner.glyph)
		}
	})
}

func (e *RichTextEdit) applyPopoverColor(palette string, color Color) {
	if !e.HasSelection || (palette != "#FG" && palette != "#BG") {
		e.popoverPalette = ""
		e.popoverSwatches = nil
		return
	}
	e.mutate(false, false, func() {
		lines := e.documentLines()
		from, to := e.selectionBounds()
		e.clampPosition(&from, lines)
		e.clampPosition(&to, lines)
		from, to = expandRangeForPills(lines, from, to)
		for lineIndex := from.Line; lineIndex <= to.Line; lineIndex++ {
			start, end := e.selectionLineBoundsForRange(lineIndex, from, to)
			lines[lineIndex].Spans = colorRichLine(lines[lineIndex], start, end, palette, color)
		}
	})
	e.popoverPalette = ""
	e.popoverSwatches = nil
}

func colorRichLine(line RichLine, start, end int, field string, color Color) []RichSpan {
	var result []RichSpan
	offset := 0
	for _, span := range line.Spans {
		length := len([]rune(span.Text))
		spanEnd := offset + length
		if length == 0 || offset >= end || spanEnd <= start {
			result = append(result, span)
			offset = spanEnd
			continue
		}
		localStart, localEnd := max(0, start-offset), min(length, end-offset)
		prefix, rest := splitRichLine(RichLine{Spans: []RichSpan{span}}, localStart)
		selected, suffix := splitRichLine(RichLine{Spans: rest}, localEnd-localStart)
		result = append(result, prefix...)
		for i := range selected {
			if field == "#FG" {
				selected[i].Style.FG = color
			} else {
				selected[i].Style.BG = color
			}
		}
		result = append(result, selected...)
		result = append(result, suffix...)
		offset = spanEnd
	}
	return mergeRichSpans(result)
}

func (e *RichTextEdit) ensureDocument() {
	if e.Document == nil {
		e.Document = &RichDocument{}
	}
	if len(e.Document.Lines) == 0 {
		e.Document.Lines = []RichLine{{}}
	}
}

func richLineRuneCount(line RichLine) int {
	n := 0
	for _, span := range line.Spans {
		n += len([]rune(span.Text))
	}
	return n
}

func richLineUnits(line RichLine) []richTextUnit {
	var units []richTextUnit
	offset := 0
	for _, span := range line.Spans {
		if span.PillData != nil {
			n := len([]rune(span.Text))
			if n > 0 {
				units = append(units, richTextUnit{start: offset, end: offset + n, text: span.Text, pill: true})
				offset += n
			}
			continue
		}
		for _, cluster := range measure.Clusters(span.Text) {
			n := len([]rune(cluster))
			units = append(units, richTextUnit{start: offset, end: offset + n, text: cluster})
			offset += n
		}
	}
	return units
}

type richTextUnit struct {
	start, end int
	text       string
	pill       bool
}

func richLineStops(line RichLine) []int {
	stops := []int{0}
	for _, unit := range richLineUnits(line) {
		stops = append(stops, unit.end)
	}
	return stops
}

func (e *RichTextEdit) normalizePosition(pos RichPosition, direction int) RichPosition {
	lines := e.documentLines()
	e.clampPosition(&pos, lines)
	line := lines[pos.Line]
	for _, unit := range richLineUnits(line) {
		if pos.Offset > unit.start && pos.Offset < unit.end {
			if direction < 0 {
				pos.Offset = unit.start
			} else if direction > 0 {
				pos.Offset = unit.end
			} else if pos.Offset-unit.start < unit.end-pos.Offset {
				pos.Offset = unit.start
			} else {
				pos.Offset = unit.end
			}
			break
		}
	}
	return pos
}

func (e *RichTextEdit) moveCursor(target RichPosition, extend bool) {
	e.cursorInVoid = false
	e.boxSelection = nil
	target = e.normalizePosition(target, 0)
	if extend {
		if !e.selectionExtending {
			e.selectionAnchor = e.Cursor
			e.selectionExtending = true
		}
		e.Cursor = target
		e.SelectionFrom, e.SelectionTo = e.selectionAnchor, target
		e.HasSelection = e.SelectionFrom != e.SelectionTo
		return
	}
	if e.HasSelection && target.Line == e.Cursor.Line && target.Offset == e.Cursor.Offset {
		from, to := e.selectionBounds()
		if target.Offset <= from.Offset {
			target = from
		} else if target.Offset >= to.Offset {
			target = to
		}
	}
	e.Cursor = target
	e.ClearSelection()
}

func (e *RichTextEdit) moveHorizontal(direction int, extend bool) {
	if e.cursorInVoid {
		line, column := e.voidLine, e.voidColumn+direction
		if column < 0 {
			return
		}
		lines := e.documentLines()
		if line < len(lines) && column <= richLineColumn(lines[line], richLineRuneCount(lines[line])) {
			e.moveCursor(RichPosition{Line: line, Offset: richLineOffsetAtColumn(lines[line], column)}, extend)
			return
		}
		e.moveToVoid(line, column, extend)
		return
	}
	if e.HasSelection && !extend {
		from, to := e.selectionBounds()
		if direction < 0 {
			e.moveCursor(from, false)
		} else {
			e.moveCursor(to, false)
		}
		return
	}
	lines := e.documentLines()
	stops := richLineStops(lines[e.Cursor.Line])
	if direction < 0 {
		for i := len(stops) - 1; i >= 0; i-- {
			if stops[i] < e.Cursor.Offset {
				e.moveCursor(RichPosition{Line: e.Cursor.Line, Offset: stops[i]}, extend)
				return
			}
		}
		if e.Cursor.Line > 0 {
			line := e.Cursor.Line - 1
			e.moveCursor(RichPosition{Line: line, Offset: richLineRuneCount(lines[line])}, extend)
		}
		return
	}
	for _, stop := range stops {
		if stop > e.Cursor.Offset {
			e.moveCursor(RichPosition{Line: e.Cursor.Line, Offset: stop}, extend)
			return
		}
	}
	if e.Cursor.Line+1 < len(lines) {
		e.moveCursor(RichPosition{Line: e.Cursor.Line + 1}, extend)
	}
}

func (e *RichTextEdit) moveVertical(direction int, extend bool) {
	lines := e.documentLines()
	line, column := e.cursorCoordinates(lines)
	targetLine := line + direction
	if targetLine < 0 {
		return
	}
	if !e.GhostCursorEnabled {
		if targetLine >= len(lines) {
			return
		}
		best, distance := 0, int(^uint(0)>>1)
		for _, stop := range richLineStops(lines[targetLine]) {
			d := richTextAbs(richLineColumn(lines[targetLine], stop) - column)
			if d < distance {
				best, distance = stop, d
			}
		}
		e.moveCursor(RichPosition{Line: targetLine, Offset: best}, extend)
		return
	}
	if targetLine >= len(lines) {
		e.moveToVoid(targetLine, column, extend)
		return
	}
	lineWidth := richLineColumn(lines[targetLine], richLineRuneCount(lines[targetLine]))
	if column > lineWidth {
		e.moveToVoid(targetLine, column, extend)
		return
	}
	best, distance := 0, int(^uint(0)>>1)
	for _, stop := range richLineStops(lines[targetLine]) {
		d := richTextAbs(richLineColumn(lines[targetLine], stop) - column)
		if d < distance {
			best, distance = stop, d
		}
	}
	e.moveCursor(RichPosition{Line: targetLine, Offset: best}, extend)
}

func (e *RichTextEdit) wordPosition(direction int) RichPosition {
	lines := e.documentLines()
	line := lines[e.Cursor.Line]
	units := richLineUnits(line)
	index := 0
	for index < len(units) && units[index].end <= e.Cursor.Offset {
		index++
	}
	if direction < 0 {
		if index == 0 {
			if e.Cursor.Line > 0 {
				previous := e.Cursor.Line - 1
				return RichPosition{Line: previous, Offset: richLineRuneCount(lines[previous])}
			}
			return RichPosition{}
		}
		index--
		for index >= 0 && richUnitSpace(units[index]) {
			index--
		}
		if index >= 0 && richUnitWord(units[index]) {
			for index > 0 && richUnitWord(units[index-1]) {
				index--
			}
		}
		if index < 0 {
			return RichPosition{Line: e.Cursor.Line}
		}
		return RichPosition{Line: e.Cursor.Line, Offset: units[index].start}
	}
	if index < len(units) && richUnitWord(units[index]) {
		for index < len(units) && richUnitWord(units[index]) {
			index++
		}
	} else if index < len(units) {
		index++
	}
	for index < len(units) && richUnitSpace(units[index]) {
		index++
	}
	if index == len(units) && e.Cursor.Line+1 < len(lines) {
		return RichPosition{Line: e.Cursor.Line + 1}
	}
	if index == len(units) {
		return RichPosition{Line: e.Cursor.Line, Offset: richLineRuneCount(line)}
	}
	return RichPosition{Line: e.Cursor.Line, Offset: units[index].start}
}

func richUnitSpace(unit richTextUnit) bool {
	for _, r := range unit.text {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return unit.text != ""
}

func richUnitWord(unit richTextUnit) bool {
	if unit.pill {
		return true
	}
	for _, r := range unit.text {
		if !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_') {
			return false
		}
	}
	return unit.text != ""
}

func richTextAbs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func (e *RichTextEdit) insertText(text string) {
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	if text == "" {
		return
	}
	e.materializeVoidCursor()
	e.deleteSelection()
	e.Cursor = e.normalizePosition(e.Cursor, 0)
	lines := e.Document.Lines
	left, right := splitRichLine(lines[e.Cursor.Line], e.Cursor.Offset)
	parts := strings.Split(text, "\n")
	style := e.insertionStyle(left, right)
	if len(parts) == 1 {
		spans := append([]RichSpan(nil), left...)
		if parts[0] != "" {
			spans = append(spans, RichSpan{Text: parts[0], Style: style})
		}
		spans = append(spans, right...)
		lines[e.Cursor.Line].Spans = mergeRichSpans(spans)
		e.Cursor.Offset += len([]rune(parts[0]))
		return
	}
	updated := make([]RichLine, 0, len(lines)+len(parts)-1)
	updated = append(updated, lines[:e.Cursor.Line]...)
	first := append([]RichSpan(nil), left...)
	if parts[0] != "" {
		first = append(first, RichSpan{Text: parts[0], Style: style})
	}
	updated = append(updated, RichLine{Spans: mergeRichSpans(first)})
	for _, part := range parts[1 : len(parts)-1] {
		var spans []RichSpan
		if part != "" {
			spans = append(spans, RichSpan{Text: part, Style: style})
		}
		updated = append(updated, RichLine{Spans: spans})
	}
	last := make([]RichSpan, 0, len(right)+1)
	if lastText := parts[len(parts)-1]; lastText != "" {
		last = append(last, RichSpan{Text: lastText, Style: style})
	}
	last = append(last, right...)
	updated = append(updated, RichLine{Spans: mergeRichSpans(last)})
	updated = append(updated, lines[e.Cursor.Line+1:]...)
	e.Document.Lines = updated
	e.Cursor.Line += len(parts) - 1
	e.Cursor.Offset = len([]rune(parts[len(parts)-1]))
}

func (e *RichTextEdit) insertionStyle(left, right []RichSpan) Style {
	if e.activeStyleSet {
		return e.ActiveStyle
	}
	if len(left) > 0 {
		return left[len(left)-1].Style
	}
	if len(right) > 0 {
		return right[0].Style
	}
	return e.ActiveStyle
}

func splitRichLine(line RichLine, offset int) (left, right []RichSpan) {
	consumed := 0
	for _, span := range line.Spans {
		length := len([]rune(span.Text))
		end := consumed + length
		if end <= offset {
			left = append(left, span)
		} else if consumed >= offset {
			right = append(right, span)
		} else if span.PillData != nil {
			if offset-consumed < end-offset {
				right = append(right, span)
			} else {
				left = append(left, span)
			}
		} else {
			cut := offset - consumed
			runes := []rune(span.Text)
			before, after := span, span
			before.Text, after.Text = string(runes[:cut]), string(runes[cut:])
			if before.Text != "" {
				left = append(left, before)
			}
			if after.Text != "" {
				right = append(right, after)
			}
		}
		consumed = end
	}
	return mergeRichSpans(left), mergeRichSpans(right)
}

func mergeRichSpans(spans []RichSpan) []RichSpan {
	merged := make([]RichSpan, 0, len(spans))
	for _, span := range spans {
		if span.Text == "" {
			continue
		}
		n := len(merged)
		if n > 0 && merged[n-1].PillData == nil && span.PillData == nil && merged[n-1].Style == span.Style && merged[n-1].Link == span.Link && merged[n-1].Code == span.Code {
			merged[n-1].Text += span.Text
		} else {
			merged = append(merged, span)
		}
	}
	return merged
}

func expandRangeForPills(lines []RichLine, from, to RichPosition) (RichPosition, RichPosition) {
	for lineIndex := from.Line; lineIndex <= to.Line; lineIndex++ {
		offset := 0
		for _, span := range lines[lineIndex].Spans {
			end := offset + len([]rune(span.Text))
			if span.PillData != nil {
				if lineIndex == from.Line && from.Offset > offset && from.Offset < end {
					from.Offset = offset
				}
				if lineIndex == to.Line && to.Offset > offset && to.Offset < end {
					to.Offset = end
				}
				if lineIndex == from.Line && from.Offset < end && to.Line > lineIndex {
					from.Offset = min(from.Offset, offset)
				}
				if lineIndex == to.Line && to.Offset > offset && from.Line < lineIndex {
					to.Offset = max(to.Offset, end)
				}
			}
			offset = end
		}
	}
	return from, to
}

func (e *RichTextEdit) deleteSelection() bool {
	if !e.HasSelection {
		return false
	}
	if e.boxSelection != nil {
		selection := e.boxSelection
		for _, row := range selection.rows {
			line := e.Document.Lines[row.line]
			left, rest := splitRichLine(line, row.start)
			_, right := splitRichLine(RichLine{Spans: rest}, row.end-row.start)
			e.Document.Lines[row.line].Spans = mergeRichSpans(append(left, right...))
		}
		e.Cursor = RichPosition{Line: selection.top, Offset: richLineOffsetAtColumn(e.Document.Lines[selection.top], selection.left)}
		e.ClearSelection()
		return true
	}
	from, to := e.selectionBounds()
	e.clampPosition(&from, e.Document.Lines)
	e.clampPosition(&to, e.Document.Lines)
	from, to = expandRangeForPills(e.Document.Lines, from, to)
	e.deleteRange(from, to)
	e.Cursor = from
	e.ClearSelection()
	return true
}

func (e *RichTextEdit) deleteRange(from, to RichPosition) {
	lines := e.Document.Lines
	if richPositionBefore(to, from) {
		from, to = to, from
	}
	if from.Line == to.Line {
		left, _ := splitRichLine(lines[from.Line], from.Offset)
		_, right := splitRichLine(lines[to.Line], to.Offset)
		lines[from.Line].Spans = mergeRichSpans(append(left, right...))
		return
	}
	left, _ := splitRichLine(lines[from.Line], from.Offset)
	_, right := splitRichLine(lines[to.Line], to.Offset)
	replacement := RichLine{Spans: mergeRichSpans(append(left, right...))}
	updated := make([]RichLine, 0, len(lines)-(to.Line-from.Line))
	updated = append(updated, lines[:from.Line]...)
	updated = append(updated, replacement)
	updated = append(updated, lines[to.Line+1:]...)
	e.Document.Lines = updated
}

func (e *RichTextEdit) deleteDirection(direction int) {
	if e.deleteSelection() {
		return
	}
	lines := e.Document.Lines
	if direction < 0 {
		stops := richLineStops(lines[e.Cursor.Line])
		for i := len(stops) - 1; i >= 0; i-- {
			if stops[i] < e.Cursor.Offset {
				e.deleteRange(RichPosition{Line: e.Cursor.Line, Offset: stops[i]}, e.Cursor)
				e.Cursor.Offset = stops[i]
				return
			}
		}
		if e.Cursor.Line > 0 {
			previous := e.Cursor.Line - 1
			offset := richLineRuneCount(lines[previous])
			lines[previous].Spans = mergeRichSpans(append(lines[previous].Spans, lines[e.Cursor.Line].Spans...))
			e.Document.Lines = append(lines[:e.Cursor.Line], lines[e.Cursor.Line+1:]...)
			e.Cursor = RichPosition{Line: previous, Offset: offset}
		}
		return
	}
	for _, stop := range richLineStops(lines[e.Cursor.Line]) {
		if stop > e.Cursor.Offset {
			e.deleteRange(e.Cursor, RichPosition{Line: e.Cursor.Line, Offset: stop})
			return
		}
	}
	if e.Cursor.Line+1 < len(lines) {
		line := &lines[e.Cursor.Line]
		line.Spans = mergeRichSpans(append(line.Spans, lines[e.Cursor.Line+1].Spans...))
		e.Document.Lines = append(lines[:e.Cursor.Line+1], lines[e.Cursor.Line+2:]...)
	}
}

func (e *RichTextEdit) toggleAttribute(set func(*Style, bool), enabled func(Style) bool) {
	if e.HasSelection {
		from, to := e.selectionBounds()
		e.mutate(false, false, func() { e.toggleRange(from, to, set, enabled) })
		return
	}
	if from, to, ok := e.wordRange(e.Cursor); ok {
		e.mutate(false, false, func() { e.toggleRange(from, to, set, enabled) })
		return
	}
	set(&e.ActiveStyle, !enabled(e.ActiveStyle))
	e.activeStyleSet = true
}

func (e *RichTextEdit) toggleRange(from, to RichPosition, set func(*Style, bool), enabled func(Style) bool) {
	e.clampPosition(&from, e.Document.Lines)
	e.clampPosition(&to, e.Document.Lines)
	from, to = expandRangeForPills(e.Document.Lines, from, to)
	allEnabled := true
	for lineIndex := from.Line; lineIndex <= to.Line; lineIndex++ {
		start, end := e.selectionLineBoundsForRange(lineIndex, from, to)
		offset := 0
		for _, span := range e.Document.Lines[lineIndex].Spans {
			spanEnd := offset + len([]rune(span.Text))
			if offset < end && spanEnd > start && !enabled(span.Style) {
				allEnabled = false
			}
			offset = spanEnd
		}
	}
	state := !allEnabled
	for lineIndex := from.Line; lineIndex <= to.Line; lineIndex++ {
		start, end := e.selectionLineBoundsForRange(lineIndex, from, to)
		e.Document.Lines[lineIndex].Spans = styleRichLine(e.Document.Lines[lineIndex], start, end, state, set)
	}
}

func richSelectionLineBounds(line int, from, to RichPosition) (start, end int) {
	if line == from.Line {
		start = from.Offset
	}
	if line == to.Line {
		end = to.Offset
	} else {
		end = int(^uint(0) >> 1)
	}
	return start, end
}

func styleRichLine(line RichLine, start, end int, state bool, set func(*Style, bool)) []RichSpan {
	var result []RichSpan
	offset := 0
	for _, span := range line.Spans {
		length := len([]rune(span.Text))
		spanEnd := offset + length
		if length == 0 || offset >= end || spanEnd <= start {
			result = append(result, span)
			offset = spanEnd
			continue
		}
		localStart, localEnd := max(0, start-offset), min(length, end-offset)
		prefix, rest := splitRichLine(RichLine{Spans: []RichSpan{span}}, localStart)
		selected, suffix := splitRichLine(RichLine{Spans: rest}, localEnd-localStart)
		result = append(result, prefix...)
		for i := range selected {
			set(&selected[i].Style, state)
		}
		result = append(result, selected...)
		result = append(result, suffix...)
		offset = spanEnd
	}
	return mergeRichSpans(result)
}

func (e *RichTextEdit) mouseInBounds(mouse MouseEvent) bool {
	return e.lastRect.W > 0 && e.lastRect.H > 0 && mouse.X >= 0 && mouse.X < e.lastRect.W && mouse.Y >= 0 && mouse.Y < e.lastRect.H
}

func (e *RichTextEdit) mousePosition(mouse MouseEvent) RichPosition {
	lines := e.documentLines()
	lineIndex := min(max(0, e.ScrollY+mouse.Y), len(lines)-1)
	targetColumn := e.ScrollX + mouse.X
	best, distance := 0, int(^uint(0)>>1)
	for _, stop := range richLineStops(lines[lineIndex]) {
		d := richTextAbs(richLineColumn(lines[lineIndex], stop) - targetColumn)
		if d < distance {
			best, distance = stop, d
		}
	}
	return RichPosition{Line: lineIndex, Offset: best}
}

func (e *RichTextEdit) extendMouseSelection(mouse MouseEvent) {
	if !e.mouseInBounds(mouse) {
		return
	}
	position := e.mousePosition(mouse)
	e.Cursor = position
	e.SelectionTo = position
	e.SelectionFrom = e.selectionAnchor
	e.boxSelection = nil
	e.HasSelection = e.SelectionFrom != e.SelectionTo
}

// wordRange returns the word containing pos on its line. Whitespace, punctuation,
// and the line end have no word.
func (e *RichTextEdit) wordRange(pos RichPosition) (from, to RichPosition, ok bool) {
	lines := e.documentLines()
	e.clampPosition(&pos, lines)
	units := richLineUnits(lines[pos.Line])
	index := -1
	for i, unit := range units {
		if unit.start <= pos.Offset && pos.Offset < unit.end {
			index = i
			break
		}
	}
	if index < 0 || !richUnitWord(units[index]) {
		return pos, pos, false
	}
	first, last := index, index
	for first > 0 && richUnitWord(units[first-1]) {
		first--
	}
	for last+1 < len(units) && richUnitWord(units[last+1]) {
		last++
	}
	return RichPosition{Line: pos.Line, Offset: units[first].start}, RichPosition{Line: pos.Line, Offset: units[last].end}, true
}

func (e *RichTextEdit) selectWord() bool {
	from, to, ok := e.wordRange(e.Cursor)
	if !ok {
		return false
	}
	e.selectRange(from, to)
	return true
}

func (e *RichTextEdit) selectLine() {
	line := e.Cursor.Line
	e.selectRange(RichPosition{Line: line}, RichPosition{Line: line, Offset: richLineRuneCount(e.documentLines()[line])})
}

func (e *RichTextEdit) selectRange(from, to RichPosition) {
	e.SetSelection(from, to)
	e.selectionAnchor = from
	e.Cursor = to
}

// registerClick counts presses at the same position within the multi-click
// window, cycling 1, 2, 3, 1.
func (e *RichTextEdit) registerClick(pos RichPosition) int {
	now := time.Now
	if e.now != nil {
		now = e.now
	}
	at := now()
	if e.clickCount > 0 && e.clickCount < 3 && pos == e.lastClickPos && at.Sub(e.lastClickAt) <= richMultiClickWindow {
		e.clickCount++
	} else {
		e.clickCount = 1
	}
	e.lastClickAt, e.lastClickPos = at, pos
	return e.clickCount
}

// copySelectionOrWord stores the selection, or the word at the cursor, in the
// internal clipboard. It reports whether anything was copied; a word copy
// selects the word so that cut can delete it.
func (e *RichTextEdit) copySelectionOrWord() bool {
	if !e.HasSelection && !e.selectWord() {
		return false
	}
	from, to := e.selectionBounds()
	e.clampPosition(&from, e.Document.Lines)
	e.clampPosition(&to, e.Document.Lines)
	from, to = expandRangeForPills(e.Document.Lines, from, to)
	if from == to {
		return false
	}
	e.clipboard = nil
	for lineIndex := from.Line; lineIndex <= to.Line; lineIndex++ {
		line := e.Document.Lines[lineIndex]
		start, end := e.selectionLineBounds(lineIndex)
		end = min(end, richLineRuneCount(line))
		_, rest := splitRichLine(line, start)
		part, _ := splitRichLine(RichLine{Spans: rest}, end-start)
		e.clipboard = append(e.clipboard, RichLine{Spans: part})
	}
	return true
}

// insertRich inserts styled lines at the cursor, replacing any selection.
func (e *RichTextEdit) insertRich(clip []RichLine) {
	if len(clip) == 0 {
		return
	}
	e.materializeVoidCursor()
	e.deleteSelection()
	e.Cursor = e.normalizePosition(e.Cursor, 0)
	lines := e.Document.Lines
	left, right := splitRichLine(lines[e.Cursor.Line], e.Cursor.Offset)
	last := len(clip) - 1
	origin := e.Cursor.Line
	updated := make([]RichLine, 0, len(lines)+last)
	updated = append(updated, lines[:origin]...)
	first := append(append([]RichSpan(nil), left...), clip[0].Spans...)
	if last == 0 {
		e.Cursor.Offset += richLineRuneCount(clip[0])
		updated = append(updated, RichLine{Spans: mergeRichSpans(append(first, right...))})
	} else {
		updated = append(updated, RichLine{Spans: mergeRichSpans(first)})
		for _, line := range clip[1:last] {
			updated = append(updated, RichLine{Spans: append([]RichSpan(nil), line.Spans...)})
		}
		tail := append(append([]RichSpan(nil), clip[last].Spans...), right...)
		updated = append(updated, RichLine{Spans: mergeRichSpans(tail)})
		e.Cursor.Line += last
		e.Cursor.Offset = richLineRuneCount(clip[last])
	}
	updated = append(updated, lines[origin+1:]...)
	e.Document.Lines = updated
}

type richSnapshot struct {
	lines        []RichLine
	cursor       RichPosition
	cursorInVoid bool
	voidLine     int
	voidColumn   int
	from, to     RichPosition
	hasSelection bool
	boxSelection *richBoxSelection
}

func (e *RichTextEdit) snapshot() richSnapshot {
	lines := make([]RichLine, len(e.Document.Lines))
	for i, line := range e.Document.Lines {
		lines[i] = RichLine{Spans: append([]RichSpan(nil), line.Spans...)}
	}
	return richSnapshot{lines: lines, cursor: e.Cursor, cursorInVoid: e.cursorInVoid, voidLine: e.voidLine, voidColumn: e.voidColumn, from: e.SelectionFrom, to: e.SelectionTo, hasSelection: e.HasSelection, boxSelection: cloneRichBoxSelection(e.boxSelection)}
}

func (e *RichTextEdit) restore(snap richSnapshot) {
	e.Document.Lines = snap.lines
	e.Cursor = snap.cursor
	e.cursorInVoid, e.voidLine, e.voidColumn = snap.cursorInVoid, snap.voidLine, snap.voidColumn
	e.SelectionFrom, e.SelectionTo, e.HasSelection = snap.from, snap.to, snap.hasSelection
	e.selectionExtending = false
	e.boxSelection = cloneRichBoxSelection(snap.boxSelection)
	e.revalidateBoxSelection()
	e.typingRun = false
}

// mutate runs an edit and records one undo step when it changed the document.
// Typing edits join the open typing run until a boundary character ends it.
func (e *RichTextEdit) mutate(typing, boundary bool, edit func()) {
	before := e.snapshot()
	edit()
	if reflect.DeepEqual(before.lines, e.Document.Lines) {
		return
	}
	e.revalidateBoxSelection()
	e.redoStack = nil
	if !(typing && e.typingRun) {
		e.undoStack = append(e.undoStack, before)
		if len(e.undoStack) > richUndoLimit {
			e.undoStack = e.undoStack[len(e.undoStack)-richUndoLimit:]
		}
	}
	e.typingRun = typing && !boundary
}

func (e *RichTextEdit) undo() {
	if len(e.undoStack) == 0 {
		return
	}
	e.redoStack = append(e.redoStack, e.snapshot())
	e.restore(e.undoStack[len(e.undoStack)-1])
	e.undoStack = e.undoStack[:len(e.undoStack)-1]
}

func (e *RichTextEdit) redo() {
	if len(e.redoStack) == 0 {
		return
	}
	e.undoStack = append(e.undoStack, e.snapshot())
	e.restore(e.redoStack[len(e.redoStack)-1])
	e.redoStack = e.redoStack[:len(e.redoStack)-1]
}

func (e *RichTextEdit) toggleBoxMode() {
	if !e.BoxMode {
		e.ClearSelection()
		e.boxStrokeFGSet = false
		lines := e.documentLines()
		line, column := e.cursorCoordinates(lines)
		if !e.cursorInVoid && line < len(lines) {
			if glyph := richBoxCell(lines[line], column); glyph != "" && BoxGlyphArms(glyph) != 0 {
				if style, ok := richLineStyleAtColumn(lines[line], column); ok {
					e.boxStrokeFG, e.boxStrokeFGSet = style.FG, true
				}
			}
		}
		e.BoxMode = true
		before := e.snapshot()
		e.boxStrokeBefore = &before
		e.typingRun = false
		return
	}
	e.BoxMode = false
	e.finishBoxStroke()
}

func (e *RichTextEdit) finishBoxStroke() {
	if e.boxStrokeBefore == nil {
		e.boxStrokeFGSet = false
		return
	}
	defer func() { e.boxStrokeFGSet = false }()
	before := *e.boxStrokeBefore
	e.boxStrokeBefore = nil
	if reflect.DeepEqual(before.lines, e.Document.Lines) {
		return
	}
	e.undoStack = append(e.undoStack, before)
	if len(e.undoStack) > richUndoLimit {
		e.undoStack = e.undoStack[len(e.undoStack)-richUndoLimit:]
	}
	e.redoStack = nil
}

func (e *RichTextEdit) drawBoxStep(dy, dx int) {
	e.boxSelection = nil
	e.materializeVoidCursor()
	e.ensureDocument()
	lines := e.Document.Lines
	fromLine, fromColumn := e.Cursor.Line, richLineColumn(lines[e.Cursor.Line], e.Cursor.Offset)
	toLine, toColumn := fromLine+dy, fromColumn+dx
	if toColumn < 0 {
		return
	}
	if toLine < 0 {
		prepend := make([]RichLine, -toLine)
		e.Document.Lines = append(prepend, lines...)
		fromLine += len(prepend)
		toLine = 0
	} else if toLine >= len(lines) {
		for len(e.Document.Lines) <= toLine {
			e.Document.Lines = append(e.Document.Lines, RichLine{})
		}
	}
	lines = e.Document.Lines
	for len(lines[fromLine].Spans) > 0 && richLineColumn(lines[fromLine], richLineRuneCount(lines[fromLine])) < fromColumn {
		lines[fromLine].Spans = append(lines[fromLine].Spans, RichSpan{Text: " "})
	}
	for len(lines[toLine].Spans) > 0 && richLineColumn(lines[toLine], richLineRuneCount(lines[toLine])) < toColumn {
		lines[toLine].Spans = append(lines[toLine].Spans, RichSpan{Text: " "})
	}
	fromArm, toArm := BoxArmRight, BoxArmLeft
	if dx < 0 {
		fromArm, toArm = BoxArmLeft, BoxArmRight
	} else if dy < 0 {
		fromArm, toArm = BoxArmUp, BoxArmDown
	} else if dy > 0 {
		fromArm, toArm = BoxArmDown, BoxArmUp
	}
	fromCell := richLineCell(lines[fromLine], fromColumn)
	fromMask := boxNeighbourArms(lines, fromLine, fromColumn) | fromArm
	lines[fromLine] = richLineSetCell(lines[fromLine], fromColumn, boxGlyphForStroke(fromCell, fromMask))
	if e.boxStrokeFGSet {
		lines[fromLine] = richLineSetCellForeground(lines[fromLine], fromColumn, e.boxStrokeFG)
	}
	toCell := richLineCell(lines[toLine], toColumn)
	toMask := boxNeighbourArms(lines, toLine, toColumn) | toArm
	lines[toLine] = richLineSetCell(lines[toLine], toColumn, boxGlyphForStroke(toCell, toMask))
	if e.boxStrokeFGSet {
		lines[toLine] = richLineSetCellForeground(lines[toLine], toColumn, e.boxStrokeFG)
	}
	e.Document.Lines = lines
	e.Cursor = RichPosition{Line: toLine, Offset: richLineOffsetAtColumn(lines[toLine], toColumn)}
	e.typingRun = false
}

func boxGlyphForStroke(existing string, arms BoxArms) string {
	if existing == "╭" || existing == "╮" || existing == "╰" || existing == "╯" {
		if BoxGlyphArms(existing) == arms {
			return existing
		}
	}
	return BoxGlyph(arms)
}

func boxNeighbourArms(lines []RichLine, line, column int) BoxArms {
	var arms BoxArms
	if line > 0 && BoxGlyphArms(richLineCell(lines[line-1], column))&BoxArmDown != 0 {
		arms |= BoxArmUp
	}
	if line+1 < len(lines) && BoxGlyphArms(richLineCell(lines[line+1], column))&BoxArmUp != 0 {
		arms |= BoxArmDown
	}
	if column > 0 && BoxGlyphArms(richLineCell(lines[line], column-1))&BoxArmRight != 0 {
		arms |= BoxArmLeft
	}
	if BoxGlyphArms(richLineCell(lines[line], column+1))&BoxArmLeft != 0 {
		arms |= BoxArmRight
	}
	return arms | BoxGlyphArms(richLineCell(lines[line], column))
}

func richLineCell(line RichLine, column int) string {
	for _, unit := range richLineUnits(line) {
		start := richLineColumn(line, unit.start)
		if column >= start && column < start+measure.StringWidth(unit.text) {
			return unit.text
		}
	}
	return " "
}

func richLineStyleAtColumn(line RichLine, column int) (Style, bool) {
	cellX := 0
	for _, span := range line.Spans {
		for _, cluster := range measure.Clusters(span.Text) {
			width := measure.StringWidth(cluster)
			if column >= cellX && column < cellX+width {
				return span.Style, span.PillData == nil
			}
			cellX += width
		}
	}
	return Style{}, false
}

func richLineOffsetAtColumn(line RichLine, column int) int {
	for _, unit := range richLineUnits(line) {
		start := richLineColumn(line, unit.start)
		if column < start+measure.StringWidth(unit.text) {
			return unit.start
		}
	}
	return richLineRuneCount(line)
}

func richLineSetCell(line RichLine, column int, glyph string) RichLine {
	var result []RichSpan
	cellX, replaced := 0, false
	for _, span := range line.Spans {
		for _, cluster := range measure.Clusters(span.Text) {
			width := measure.StringWidth(cluster)
			if width <= 0 {
				result = appendRichCellSpan(result, span, cluster)
				continue
			}
			if !replaced && column >= cellX && column < cellX+width {
				if span.PillData != nil {
					return line
				}
				before := strings.Repeat(" ", column-cellX)
				after := strings.Repeat(" ", cellX+width-column-1)
				result = appendRichCellSpan(result, span, before+glyph+after)
				replaced = true
			} else {
				result = appendRichCellSpan(result, span, cluster)
			}
			cellX += width
		}
	}
	if !replaced {
		if column > cellX {
			result = append(result, RichSpan{Text: strings.Repeat(" ", column-cellX)})
		}
		result = append(result, RichSpan{Text: glyph})
	}
	return RichLine{Spans: mergeRichSpans(result)}
}

func richLineSetCellForeground(line RichLine, column int, foreground Color) RichLine {
	for _, unit := range richLineUnits(line) {
		start := richLineColumn(line, unit.start)
		if column < start || column >= start+measure.StringWidth(unit.text) {
			continue
		}
		if unit.pill {
			return line
		}
		styled := styleRichLine(line, unit.start, unit.end, true, func(style *Style, _ bool) {
			style.FG = foreground
		})
		return RichLine{Spans: styled}
	}
	return line
}

func appendRichCellSpan(spans []RichSpan, source RichSpan, text string) []RichSpan {
	if text == "" {
		return spans
	}
	span := source
	span.Text = text
	span.PillData = nil
	return append(spans, span)
}
