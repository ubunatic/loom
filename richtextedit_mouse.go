// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"time"
)

// ConsumeMouse positions the cursor and supports click-drag selection.
// It fires OnStateChange when the event changed the document state.
func (e *RichTextEdit) ConsumeMouse(mouse MouseEvent) EventResult {
	defer e.notifyStateChange()
	return e.consumeMouse(mouse)
}

func (e *RichTextEdit) consumeMouse(mouse MouseEvent) EventResult {
	if e.ModalOpen() {
		e.dragSelecting = false
	}
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
			if e.savePicker != nil && e.savePicker.options.OnCancel != nil {
				e.savePicker.options.OnCancel()
			}
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
	if e.popoverRelease {
		switch mouse.Action {
		case MouseRelease:
			e.popoverRelease = false
			return Handled()
		case MouseDrag:
			return Handled()
		case MousePress:
			e.popoverRelease = false
		}
	}
	if e.popoverOpen() && !e.dragSelecting {
		if !e.popoverHit(mouse.X, mouse.Y) {
			if mouse.Action == MousePress && mouse.Button == MouseLeft {
				e.cancelPopover()
				e.dragSelecting = false
				e.popoverRelease = true
			}
			return Handled()
		}
		if mouse.Action == MouseScrollUp || mouse.Action == MouseScrollDown {
			if e.PopoverDefaults.DismissOnScroll {
				e.cancelPopover()
				return Handled()
			}
		}
		if e.PopoverDefaults.MouseHoverFocus && mouse.Action == MouseHover {
			for i, button := range e.popoverButtons {
				if button.enabled && button.rect.Contains(mouse.X, mouse.Y) {
					e.popoverFocus, e.popoverFocusSet = i, true
				}
			}
			for i, swatch := range e.popoverSwatches {
				if swatch.rect.Contains(mouse.X, mouse.Y) {
					e.popoverFocusSet, e.popoverSubmenuFocusSet, e.popoverSubmenuFocus = true, true, i
				}
			}
			for i, choice := range e.popoverBoxChoices {
				if choice.rect.Contains(mouse.X, mouse.Y) {
					e.popoverFocusSet, e.popoverSubmenuFocusSet, e.popoverSubmenuFocus = true, true, i
				}
			}
		}
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
			for i, button := range e.popoverButtons {
				if button.enabled && button.rect.Contains(mouse.X, mouse.Y) {
					e.popoverFocus, e.popoverFocusSet = i, true
					e.applyPopoverAction(button.action)
					e.popoverRelease = true
					break
				}
			}
		}
		return Handled()
	}
	if mouse.Action == MouseScrollUp || mouse.Action == MouseScrollDown {
		if mouse.Action == MouseScrollUp {
			e.ScrollY = max(0, e.ScrollY-1)
		} else {
			e.ScrollY++
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
