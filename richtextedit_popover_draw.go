// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"time"

	"ubunatic.com/loom/measure"
)

type richPopoverButton struct {
	label   string
	action  string
	rect    Rect
	enabled bool
}

type richPopoverBoxChoice struct {
	label string
	style BoxBorderStyle
	rect  Rect
}

type richPopoverHamburgerChoice struct {
	label string
	index int
	rect  Rect
}

var richPopoverActions = []string{"B", "I", "U", "S", "Link", "ColorPicker", "Box", "Hamburger"}

func richPopoverLabels() []string {
	labels := append([]string(nil), SpeccedDefaults.RichTextEdit.PopoverLabels...)
	return append(labels, SpeccedDefaults.RichTextEdit.PopoverHamburgerLabel)
}

var _ Widget = (*RichTextEdit)(nil)

// popoverGeometry is shared by drawing and modal capture so an invisible toolbar
// never blocks document input, and keyboard opening acquires capture immediately.
func (e *RichTextEdit) popoverGeometry(r Rect, lines []RichLine) (bar Rect, anchorX, pointerY int, pointer string, open bool) {
	if e.ViewMode || !e.ShowPopover || (!e.HasSelection && !e.popoverAtCursor) || e.popoverSuppressed || r.W < SpeccedDefaults.RichTextEdit.PopoverMinWidth || r.H < 3 {
		return
	}
	if !e.popoverExplicit && !e.popoverAtCursor {
		if !e.PopoverDefaults.AutoPopupOnSelection {
			return
		}
		now := time.Now()
		if e.now != nil {
			now = e.now()
		}
		if e.popoverSelectionAt.IsZero() || e.popoverSelectionFrom != e.SelectionFrom || e.popoverSelectionTo != e.SelectionTo {
			e.popoverSelectionFrom, e.popoverSelectionTo, e.popoverSelectionAt = e.SelectionFrom, e.SelectionTo, now
		}
		if now.Sub(e.popoverSelectionAt) < time.Duration(e.PopoverDefaults.PopupOpenDelayMS)*time.Millisecond {
			return
		}
	}
	anchorLine, anchorCol := e.cursorCoordinates(lines)
	if !e.popoverAtCursor {
		from, _ := e.selectionBounds()
		e.clampPosition(&from, lines)
		anchorLine, anchorCol = from.Line, richLineColumn(lines[from.Line], from.Offset)
	}
	anchorX = anchorCol - e.ScrollX
	anchorY := anchorLine - e.ScrollY
	if anchorX < 0 || anchorX >= r.W || anchorY < 0 || anchorY >= r.H {
		return
	}
	width := 2
	labels := richPopoverLabels()
	for i, label := range labels {
		width += measure.StringWidth(label) + 1
		if i > 0 {
			width++
		}
	}
	defs := SpeccedDefaults.RichTextEdit
	barY := 0
	pointer = defs.PointerUpGlyph
	firstY, lastY := anchorY, anchorY
	if e.HasSelection {
		from, to := e.selectionBounds()
		firstY, lastY = from.Line-e.ScrollY, to.Line-e.ScrollY
		if to.Offset == 0 && lastY > firstY {
			lastY--
		}
	}
	preferAbove := e.PopoverDefaults.PositionPreference != "below" || lastY+2 >= r.H
	if preferAbove && firstY >= 2 {
		barY, pointerY, pointer = firstY-2, firstY-1, defs.PointerDownGlyph
	} else if lastY+2 < r.H {
		barY, pointerY = lastY+2, lastY+1
	} else if firstY >= 2 {
		barY, pointerY, pointer = firstY-2, firstY-1, defs.PointerDownGlyph
	} else {
		return
	}
	barX := min(max(0, anchorX-width/2), r.W-width)
	if barX < 0 { // Keep the complete action row visible in very narrow widgets.
		return
	}
	return Rect{X: barX, Y: barY, W: width, H: 1}, anchorX, pointerY, pointer, true
}

func (e *RichTextEdit) drawPopover(c *Canvas, r Rect, lines []RichLine) {
	e.popoverButtons = nil
	e.popoverBoxChoices = nil
	e.popoverHamburgerChoices = nil
	e.popoverColorRect = Rect{}
	if e.popoverSubmenu != "" && !e.popoverSubmenuCanShow() {
		e.clearPopoverKeyboardState(true)
	}
	bar, anchorX, pointerY, pointer, open := e.popoverGeometry(r, lines)
	if !open {
		if e.popoverColorPicker != nil {
			e.clearPopoverKeyboardState(true)
		}
		return
	}
	barX, barY, width := bar.X, bar.Y, bar.W
	defs := SpeccedDefaults.RichTextEdit
	labels := richPopoverLabels()
	toolbarStyle := Style{FG: ColorIndex(uint8(defs.ToolbarFG)), BG: ColorIndex(uint8(defs.ToolbarBG)), Bold: true}
	for x := 0; x < width; x++ {
		c.Set(r.X+barX+x, r.Y+barY, Cell{Text: " ", Style: toolbarStyle})
	}
	c.Set(r.X+barX, r.Y+barY, Cell{Text: "[", Style: toolbarStyle})
	c.Set(r.X+barX+width-1, r.Y+barY, Cell{Text: "]", Style: toolbarStyle})
	x := barX + 1
	for i, label := range labels {
		button := Rect{X: x, Y: barY, W: measure.StringWidth(label) + 1, H: 1}
		action := richPopoverActions[i]
		enabled := e.popoverActionEnabled(i)
		e.popoverButtons = append(e.popoverButtons, richPopoverButton{label: label, action: action, rect: button, enabled: enabled})
		style := toolbarStyle
		if !enabled {
			style.Dim = true
		}
		if label == "B" {
			style.Bold = true
		}
		if e.popoverFocusSet && i == e.popoverFocus {
			style.FG = ColorIndex(uint8(defs.PopoverFocusFG))
			style.BG = ColorIndex(uint8(defs.PopoverFocusBG))
			style.Dim = false
		}
		c.Write(r.X+x, r.Y+barY, " "+label, style)
		x += button.W
		if i < len(labels)-1 {
			c.Set(r.X+x, r.Y+barY, Cell{Text: defs.SeparatorGlyph, Style: Style{FG: ColorIndex(uint8(defs.SeparatorFG)), BG: ColorIndex(uint8(defs.SeparatorBG))}})
			x++
		}
	}
	c.Set(r.X+barX+min(max(0, anchorX-barX), width-1), r.Y+pointerY, Cell{Text: pointer, Style: Style{FG: ColorIndex(uint8(defs.PointerFG))}})
	e.drawPopoverColorPicker(c, r, barX, barY)
	e.drawPopoverBoxStyles(c, r, barY)
	e.drawPopoverHamburger(c, r, barY)
}

func richPopoverBarWidth() int {
	width := 2
	for i, label := range richPopoverLabels() {
		width += measure.StringWidth(label) + 1
		if i > 0 {
			width++
		}
	}
	return width
}

func popoverActionIndex(action string) int {
	for i, candidate := range richPopoverActions {
		if candidate == action {
			return i
		}
	}
	return 0
}

func (e *RichTextEdit) drawPopoverColorPicker(c *Canvas, r Rect, barX, barY int) {
	if e.popoverColorPicker == nil {
		return
	}
	var button Rect
	for _, item := range e.popoverButtons {
		if item.action == "ColorPicker" {
			button = item.rect
			break
		}
	}
	if button.W == 0 {
		e.popoverColorPicker = nil
		return
	}
	pickerSize := e.popoverColorPicker.Measure(r.W)
	pickerW, pickerH := pickerSize.Width, pickerSize.Height
	buttonX := min(max(0, button.X+button.W/2-pickerW/2), r.W-pickerW)
	placement, ok := e.popoverSubmenuPlacement(r, buttonX, pickerW, pickerH, barY)
	if !ok {
		e.clearPopoverKeyboardState(true)
		return
	}
	e.popoverColorRect = placement
	e.popoverColorPicker.Draw(c, Rect{X: r.X + placement.X, Y: r.Y + placement.Y, W: placement.W, H: placement.H})
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
	labels := e.boxSubmenuLabels()
	width := 2 + len(labels) - 1
	for _, label := range labels {
		width += len(label) + 2
	}
	placement, ok := e.popoverSubmenuPlacement(r, button.X+button.W/2-width/2, width, 1, barY)
	if !ok {
		e.clearPopoverKeyboardState(true)
		return
	}
	x, y := placement.X, placement.Y
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

func (e *RichTextEdit) hamburgerSubmenuWidth() int {
	width := 2
	for _, label := range SpeccedDefaults.RichTextEdit.PopoverHamburgerItems {
		width += measure.StringWidth(label) + 2
	}
	return width
}

func (e *RichTextEdit) drawPopoverHamburger(c *Canvas, r Rect, barY int) {
	if e.popoverSubmenu != "Hamburger" {
		return
	}
	var button Rect
	for _, item := range e.popoverButtons {
		if item.action == "Hamburger" {
			button = item.rect
			break
		}
	}
	if button.W == 0 {
		return
	}
	labels := SpeccedDefaults.RichTextEdit.PopoverHamburgerItems
	width := e.hamburgerSubmenuWidth()
	placement, ok := e.popoverSubmenuPlacement(r, button.X+button.W/2-width/2, width, 1, barY)
	if !ok {
		e.clearPopoverKeyboardState(true)
		return
	}
	style := Style{FG: ColorIndex(uint8(SpeccedDefaults.RichTextEdit.ToolbarFG)), BG: ColorIndex(uint8(SpeccedDefaults.RichTextEdit.ToolbarBG)), Bold: true}
	for x := 0; x < width; x++ {
		c.Set(r.X+placement.X+x, r.Y+placement.Y, Cell{Text: " ", Style: style})
	}
	c.Set(r.X+placement.X, r.Y+placement.Y, Cell{Text: "[", Style: style})
	c.Set(r.X+placement.X+width-1, r.Y+placement.Y, Cell{Text: "]", Style: style})
	x := placement.X + 1
	for i, label := range labels {
		choiceWidth := measure.StringWidth(label) + 2
		choice := richPopoverHamburgerChoice{label: label, index: i, rect: Rect{X: x, Y: placement.Y, W: choiceWidth, H: 1}}
		e.popoverHamburgerChoices = append(e.popoverHamburgerChoices, choice)
		choiceStyle := style
		if e.popoverSubmenuFocusSet && i == e.popoverSubmenuFocus {
			choiceStyle.FG = ColorIndex(uint8(SpeccedDefaults.RichTextEdit.PopoverFocusFG))
			choiceStyle.BG = ColorIndex(uint8(SpeccedDefaults.RichTextEdit.PopoverFocusBG))
		}
		c.Write(r.X+x, r.Y+placement.Y, " "+label+" ", choiceStyle)
		x += choiceWidth
		if i < len(labels)-1 {
			c.Set(r.X+x, r.Y+placement.Y, Cell{Text: SpeccedDefaults.RichTextEdit.SeparatorGlyph, Style: style})
			x++
		}
	}
}

func (e *RichTextEdit) popoverSubmenuPlacement(r Rect, x, width, height, barY int) (Rect, bool) {
	if width <= 0 || height <= 0 || width > r.W || height > r.H {
		return Rect{}, false
	}
	x = min(max(0, x), r.W-width)
	first, last := 0, 0
	if e.HasSelection {
		from, to := e.selectionBounds()
		first, last = from.Line-e.ScrollY, to.Line-e.ScrollY
	} else {
		line, _ := e.cursorCoordinates(e.documentLines())
		first, last = line-e.ScrollY, line-e.ScrollY
	}
	preferredAbove := barY < first
	aboveY := min(barY-height, first-height)
	belowY := max(barY+1, last+1)
	clamp := func(y int) int { return min(max(0, y), r.H-height) }
	candidates := []int{}
	if preferredAbove {
		candidates = append(candidates, aboveY, clamp(aboveY), belowY, clamp(belowY))
	} else {
		candidates = append(candidates, belowY, clamp(belowY), aboveY, clamp(aboveY))
	}
	for _, y := range candidates {
		if y < 0 || y+height > r.H {
			continue
		}
		if y <= barY && barY < y+height {
			continue
		}
		candidate := Rect{X: x, Y: y, W: width, H: height}
		if !e.popoverRectCoversSelection(candidate) {
			return candidate, true
		}
	}
	return Rect{}, false
}

func (e *RichTextEdit) popoverRectCoversSelection(rect Rect) bool {
	if !e.HasSelection {
		return false
	}
	lines := e.documentLines()
	from, to := e.selectionBounds()
	for lineIndex := from.Line; lineIndex <= to.Line && lineIndex < len(lines); lineIndex++ {
		y := lineIndex - e.ScrollY
		if y < rect.Y || y >= rect.Y+rect.H {
			continue
		}
		start, end := e.selectionLineBounds(lineIndex)
		offset, col := 0, 0
		for _, span := range lines[lineIndex].Spans {
			for _, cluster := range measure.Clusters(span.Text) {
				runeCount, cellWidth := len([]rune(cluster)), measure.StringWidth(cluster)
				if offset < end && offset+runeCount > start && cellWidth > 0 && col-e.ScrollX < rect.X+rect.W && col+cellWidth-e.ScrollX > rect.X {
					return true
				}
				offset += runeCount
				col += cellWidth
			}
		}
	}
	return false
}

func (e *RichTextEdit) popoverHit(x, y int) bool {
	if e.popoverSuppressed || !e.ShowPopover || (!e.HasSelection && !e.popoverAtCursor) || e.ViewMode {
		return false
	}
	if e.popoverColorPicker != nil && e.popoverColorRect.Contains(x, y) {
		return true
	}
	for _, choice := range e.popoverBoxChoices {
		if choice.rect.Contains(x, y) {
			return true
		}
	}
	for _, choice := range e.popoverHamburgerChoices {
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
	if label == "B" || label == "I" || label == "U" || label == "S" || label == "Draw" {
		defer e.finishPopoverAction()
	}
	switch label {
	case "Box":
		e.popoverSubmenuFocusSet = false
		if e.popoverSubmenu == "Box" {
			e.popoverSubmenu = ""
			e.popoverBoxChoices = nil
		} else {
			e.popoverSubmenu = "Box"
		}
	case "Hamburger":
		e.popoverSubmenuFocusSet = false
		if e.popoverSubmenu == "Hamburger" {
			e.popoverSubmenu = ""
			e.popoverHamburgerChoices = nil
		} else {
			e.popoverSubmenu = "Hamburger"
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
	case "ColorPicker":
		e.popoverSubmenu = ""
		e.popoverBoxChoices = nil
		e.popoverSubmenuFocusSet = false
		e.openPopoverColorPicker()
	}
}

func (e *RichTextEdit) applyHamburgerItem(index int) {
	if index >= 0 && index < len(SpeccedDefaults.RichTextEdit.PopoverHamburgerActions) && SpeccedDefaults.RichTextEdit.PopoverHamburgerActions[index] == "select_all" {
		e.selectAll()
	}
	e.popoverSubmenu = ""
	e.popoverHamburgerChoices = nil
	e.popoverSubmenuFocusSet = false
}

func (e *RichTextEdit) boxSubmenuLabels() []string {
	d := SpeccedDefaults.RichTextEdit
	return append(append([]string(nil), d.BoxStyleLabels...), d.BoxDrawLabel)
}

func (e *RichTextEdit) applyBoxStyleChoice(label string) {
	if label == SpeccedDefaults.RichTextEdit.BoxDrawLabel {
		e.applyPopoverAction("Draw")
		return
	}
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
	e.finishPopoverAction()
}
