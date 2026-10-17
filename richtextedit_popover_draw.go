// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"time"
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

var richPopoverActions = []string{"B", "I", "U", "S", "Link", "#FG", "#BG", "Box"}

var _ Widget = (*RichTextEdit)(nil)

// popoverGeometry is shared by drawing and modal capture so an invisible toolbar
// never blocks document input, and keyboard opening acquires capture immediately.
func (e *RichTextEdit) popoverGeometry(r Rect, lines []RichLine) (bar Rect, anchorX, pointerY int, pointer string, open bool) {
	if e.ViewMode || !e.ShowPopover || (!e.HasSelection && !e.popoverAtCursor) || e.popoverSuppressed || r.W < 18 || r.H < 3 {
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
	labels := SpeccedDefaults.RichTextEdit.PopoverLabels
	for i, label := range labels {
		width += len(label) + 1
		if i > 0 {
			width++
		}
	}
	defs := SpeccedDefaults.RichTextEdit
	barY := 0
	pointer = defs.PointerUpGlyph
	if anchorY >= 2 && (e.PopoverDefaults.PositionPreference != "below" || anchorY+2 >= r.H) {
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
	return Rect{X: barX, Y: barY, W: width, H: 1}, anchorX, pointerY, pointer, true
}

func (e *RichTextEdit) drawPopover(c *Canvas, r Rect, lines []RichLine) {
	e.popoverButtons = nil
	e.popoverSwatches = nil
	e.popoverBoxChoices = nil
	if (e.popoverPalette != "" || e.popoverSubmenu != "") && !e.popoverSubmenuCanShow() {
		e.clearPopoverKeyboardState(true)
	}
	bar, anchorX, pointerY, pointer, open := e.popoverGeometry(r, lines)
	if !open {
		return
	}
	barX, barY, width := bar.X, bar.Y, bar.W
	defs := SpeccedDefaults.RichTextEdit
	labels := defs.PopoverLabels
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
	labels := e.boxSubmenuLabels()
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
	if label == "B" || label == "I" || label == "U" || label == "S" || label == "Draw" {
		defer e.finishPopoverAction()
	}
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
