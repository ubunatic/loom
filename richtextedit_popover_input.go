// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "strings"

func (e *RichTextEdit) clearPopoverKeyboardState(suppress bool) {
	e.popoverFocusSet = false
	e.popoverSubmenuFocusSet = false
	e.popoverColorPicker = nil
	e.popoverColorRect = Rect{}
	e.popoverSubmenu = ""
	e.popoverBoxChoices = nil
	e.popoverHamburgerChoices = nil
	e.popoverAtCursor = false
	if suppress {
		e.popoverSuppressed = true
	}
}

func (e *RichTextEdit) cancelPopover() {
	if e.PopoverDefaults.UnselectOnCancel {
		e.ClearSelection()
	}
	e.clearPopoverKeyboardState(true)
}

func (e *RichTextEdit) finishPopoverAction() {
	if e.PopoverDefaults.UnselectAfterAction {
		e.ClearSelection()
	}
	if e.PopoverDefaults.ActionClosesMenu {
		e.clearPopoverKeyboardState(true)
	}
}

func (e *RichTextEdit) consumePopoverKey(key KeyEvent) (EventResult, bool) {
	if key.Is(e.PopoverDefaults.PopupOpenKey) && e.popoverOpen() {
		e.clearPopoverKeyboardState(true)
		return Handled(), true
	}
	if e.popoverColorPicker != nil {
		result := e.popoverColorPicker.ConsumeKey(key)
		if result.Done {
			return Handled(), true
		}
		if result.Consumed {
			return result, true
		}
		return Ignored(), true
	}
	if e.ViewMode || !e.ShowPopover || (!e.HasSelection && !e.popoverAtCursor) || e.popoverSuppressed || (!e.PopoverDefaults.AutoPopupOnSelection && !e.popoverExplicit && !e.popoverAtCursor) {
		return Ignored(), false
	}
	submenuOpen := e.popoverSubmenu != ""
	popoverArrow := key.Is(SpeccedDefaults.RichTextEdit.Bindings["arrows"]...)
	plainHorizontal := key.Is(SpeccedDefaults.RichTextEdit.Bindings["left"]...) || key.Is(SpeccedDefaults.RichTextEdit.Bindings["right"]...)
	if !submenuOpen && e.popoverFocusSet {
		if index := e.popoverHotkeyIndex(key); index >= 0 {
			if !e.popoverActionEnabled(index) {
				return Handled(), true
			}
			e.activatePopoverAction(index)
			return Handled(), true
		}
	}
	if !e.popoverFocusSet && !submenuOpen {
		if key.Is(SpeccedDefaults.RichTextEdit.Bindings["menu_entry"]...) {
			_, _, _, _, open := e.popoverGeometry(e.lastRect, e.documentLines())
			if open {
				e.popoverFocus = e.initialPopoverFocus()
				e.popoverFocusSet = true
				return Handled(), true
			}
		}
		if popoverArrow && !plainHorizontal {
			if e.PopoverDefaults.ArrowsCancelUnfocusedMenu {
				e.cancelPopover()
			}
			return Ignored(), false
		}
		if plainHorizontal {
			return Ignored(), false
		}
		if key.Is(SpeccedDefaults.RichTextEdit.Bindings["cancel"]...) {
			e.cancelPopover()
			return Handled(), true
		}
		if e.PopoverDefaults.MenuFocusKey == "" || !key.Is(e.PopoverDefaults.MenuFocusKey) {
			return Ignored(), false
		}
	}
	if !(key.Is(e.PopoverDefaults.MenuFocusKey) || key.Is(SpeccedDefaults.RichTextEdit.Bindings["menu_focus"]...) || key.Is(SpeccedDefaults.RichTextEdit.Bindings["activate"]...) || key.Is(SpeccedDefaults.RichTextEdit.Bindings["cancel"]...) || key.Is(SpeccedDefaults.RichTextEdit.Bindings["arrows"]...)) {
		if e.popoverAtCursor || e.popoverFocusSet {
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
	firstFocus := !e.popoverFocusSet && e.PopoverDefaults.MenuFocusKey != "" && key.Is(e.PopoverDefaults.MenuFocusKey)
	if !e.popoverFocusSet {
		e.popoverFocus = e.initialPopoverFocus()
		e.popoverFocusSet = true
	}
	if firstFocus {
		return Handled(), true
	}
	if submenuOpen {
		return e.consumePopoverSubmenuKey(key)
	}
	switch {
	case key.Is(SpeccedDefaults.RichTextEdit.Bindings["cancel"]...):
		if e.popoverSubmenu != "" {
			e.popoverSubmenu = ""
			e.popoverBoxChoices = nil
		} else {
			e.cancelPopover()
		}
		return Handled(), true
	case key.Is(SpeccedDefaults.RichTextEdit.Bindings["menu_focus"]...):
		direction := 1
		if key.Is(SpeccedDefaults.RichTextEdit.Bindings["menu_reverse"]...) {
			direction = -1
		}
		e.popoverFocus = e.nextPopoverFocus(e.popoverFocus, direction)
		return Handled(), true
	case key.Is(SpeccedDefaults.RichTextEdit.Bindings["horizontal"]...):
		direction := 1
		if key.Is(SpeccedDefaults.RichTextEdit.Bindings["left"]...) {
			direction = -1
		}
		e.popoverFocus = e.nextPopoverFocus(e.popoverFocus, direction)
		return Handled(), true
	case key.Is(SpeccedDefaults.RichTextEdit.Bindings["vertical"]...):
		_, _, _, pointer, open := e.popoverGeometry(e.lastRect, e.documentLines())
		if open && ((pointer == SpeccedDefaults.RichTextEdit.PointerUpGlyph && key.Is(SpeccedDefaults.RichTextEdit.Bindings["up"]...)) || (pointer == SpeccedDefaults.RichTextEdit.PointerDownGlyph && key.Is(SpeccedDefaults.RichTextEdit.Bindings["down"]...))) {
			e.cancelPopover()
		}
		return Handled(), true
	case key.Is(SpeccedDefaults.RichTextEdit.Bindings["activate"]...):
		e.activatePopoverAction(e.popoverFocus)
		return Handled(), true
	default:
		return Ignored(), false
	}
}

func (e *RichTextEdit) popoverHotkeyIndex(key KeyEvent) int {
	value := key.Text
	if value == "" {
		value = key.Key
	}
	if len([]rune(value)) != 1 {
		return -1
	}
	for i, hotkey := range SpeccedDefaults.RichTextEdit.PopoverHotkeys {
		if hotkey != "" && strings.EqualFold(value, hotkey) {
			return i
		}
	}
	return -1
}

func (e *RichTextEdit) activatePopoverAction(index int) {
	if index < 0 || index >= len(richPopoverActions) || !e.popoverActionEnabled(index) {
		return
	}
	action := richPopoverActions[index]
	if action == "Link" {
		return
	}
	e.applyPopoverAction(action)
	if action == "Box" {
		e.popoverSubmenuFocus = e.initialPopoverSubmenuFocus()
		e.popoverSubmenuFocusSet = true
	} else if action == "Hamburger" {
		e.popoverSubmenuFocus = 0
		e.popoverSubmenuFocusSet = true
	}
	if action == "Box" || action == "Hamburger" {
		if !e.popoverSubmenuCanShow() {
			e.clearPopoverKeyboardState(true)
		}
	}
}

func (e *RichTextEdit) consumePopoverSubmenuKey(key KeyEvent) (EventResult, bool) {
	if !e.popoverSubmenuFocusSet {
		e.popoverSubmenuFocus = e.initialPopoverSubmenuFocus()
		e.popoverSubmenuFocusSet = true
	}
	if key.Is(SpeccedDefaults.RichTextEdit.Bindings["cancel"]...) {
		e.popoverSubmenu = ""
		e.popoverBoxChoices = nil
		e.popoverHamburgerChoices = nil
		e.popoverSubmenuFocusSet = false
		return Handled(), true
	}
	if key.Is(SpeccedDefaults.RichTextEdit.Bindings["submenu_move"]...) {
		direction := 1
		if key.Is(SpeccedDefaults.RichTextEdit.Bindings["submenu_reverse"]...) {
			direction = -1
		}
		count := e.popoverSubmenuChoiceCount()
		if count > 0 {
			next := e.popoverSubmenuFocus + direction
			if e.PopoverDefaults.CycleNavigation {
				next = (next + count) % count
			}
			e.popoverSubmenuFocus = min(max(next, 0), count-1)
		}
		return Handled(), true
	}
	if key.Is(SpeccedDefaults.RichTextEdit.Bindings["activate"]...) {
		labels := e.boxSubmenuLabels()
		if e.popoverSubmenu == "Box" && e.popoverSubmenuFocus >= 0 && e.popoverSubmenuFocus < len(labels) {
			e.applyBoxStyleChoice(labels[e.popoverSubmenuFocus])
			return Handled(), true
		}
		items := SpeccedDefaults.RichTextEdit.PopoverHamburgerItems
		if e.popoverSubmenu == "Hamburger" && e.popoverSubmenuFocus >= 0 && e.popoverSubmenuFocus < len(items) {
			e.applyHamburgerItem(e.popoverSubmenuFocus)
			return Handled(), true
		}
		return Handled(), true
	}
	return Ignored(), false
}

func (e *RichTextEdit) popoverSubmenuChoiceCount() int {
	if e.popoverSubmenu == "Box" {
		return len(e.boxSubmenuLabels())
	}
	if e.popoverSubmenu == "Hamburger" {
		return len(SpeccedDefaults.RichTextEdit.PopoverHamburgerItems)
	}
	return 0
}

func (e *RichTextEdit) initialPopoverSubmenuFocus() int {
	if e.popoverSubmenu == "Box" && !e.HasSelection {
		return len(e.boxSubmenuLabels()) - 1
	}
	return 0
}

func (e *RichTextEdit) popoverSubmenuCanShow() bool {
	if !e.popoverCanShow() {
		return false
	}
	r := e.lastRect
	if e.popoverSubmenu == "Box" {
		width := 2 + len(e.boxSubmenuLabels()) - 1
		for _, label := range e.boxSubmenuLabels() {
			width += len(label) + 2
		}
		if width > r.W {
			return false
		}
	}
	if e.popoverSubmenu == "Hamburger" && e.hamburgerSubmenuWidth() > r.W {
		return false
	}
	return true
}

func (e *RichTextEdit) nextPopoverFocus(current, direction int) int {
	for range len(richPopoverActions) {
		next := current + direction
		if !e.PopoverDefaults.CycleNavigation && (next < 0 || next >= len(richPopoverActions)) {
			return current
		}
		current = (next + len(richPopoverActions)) % len(richPopoverActions)
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
	return e.HasSelection || richPopoverActions[index] == "Box" || richPopoverActions[index] == "ColorPicker" || richPopoverActions[index] == "Hamburger"
}

func (e *RichTextEdit) initialPopoverFocus() int {
	if !e.HasSelection && e.popoverAtCursor {
		return popoverActionIndex("Box")
	}
	if e.boxSelection != nil {
		return popoverActionIndex("Box")
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

func (e *RichTextEdit) popoverCanShow() bool {
	r := e.lastRect
	if r.W < SpeccedDefaults.RichTextEdit.PopoverMinWidth || r.H < 3 || (!e.HasSelection && !e.popoverAtCursor) || e.ViewMode || !e.ShowPopover {
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
	width := richPopoverBarWidth()
	if r.W < width {
		return false
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
