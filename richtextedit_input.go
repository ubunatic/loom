// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"unicode"
)

// ConsumeKey applies navigation, text editing, selection, and inline formatting.
// It fires OnStateChange when the key changed the document state.
func (e *RichTextEdit) ConsumeKey(key KeyEvent) EventResult {
	defer e.notifyStateChange()
	return e.consumeKey(key)
}

func (e *RichTextEdit) consumeKey(key KeyEvent) EventResult {
	if e.helpPopup != nil {
		result := e.helpPopup.ConsumeKey(key)
		if !e.helpPopup.Open {
			e.helpPopup = nil
		}
		return result
	}
	if e.savePopup != nil {
		popup := e.savePopup
		result := popup.ConsumeKey(key)
		if e.savePopup != nil && !e.savePopup.Open {
			if e.savePicker != nil && e.savePicker.options.OnCancel != nil {
				e.savePicker.options.OnCancel()
			}
			e.savePicker = nil
			e.savePopup = nil
		}
		return result
	}
	if !e.ShowFileBar && key.Is(SpeccedDefaults.RichTextEdit.HotkeyOpenBinding) {
		e.RequestOpen()
		return Handled()
	}
	if !e.ShowFileBar && key.Is(SpeccedDefaults.RichTextEdit.HotkeyCloseBinding) {
		e.RequestClose()
		return Handled()
	}
	if !e.ShowFileBar && key.Is(SpeccedDefaults.RichTextEdit.HotkeySaveBinding) {
		if err := e.Save(); err != nil {
			e.LastSaveError = err
		}
		return Handled()
	}
	if key.Is("f1") {
		e.ensureFileBar()
		e.fileBar.menu.Open = false
		e.fileBar.menu.SetFocus(false)
		e.helpPopup = NewPopup("RichTextEdit Help", newRichTextEditHelp(e.helpSections))
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
	preserveSelection := !e.popoverFocusSet && e.popoverOpen() && key.Is("left", "right", "up", "down") && !e.PopoverDefaults.UnselectOnCancel
	if result, handled := e.consumePopoverKey(key); handled {
		return result
	}
	if preserveSelection {
		from, to := e.SelectionFrom, e.SelectionTo
		defer func() {
			e.SelectionFrom, e.SelectionTo, e.HasSelection = from, to, true
			e.popoverSuppressed = e.PopoverDefaults.ArrowsCancelUnfocusedMenu
		}()
	}
	if key.Is(SpeccedDefaults.Editor.HotkeyBoxBinding) {
		if e.HasSelection {
			e.wrapSelectionInBox()
		} else {
			e.toggleBoxMode()
		}
		return Handled()
	}
	if e.BoxMode {
		switch {
		case key.Is("esc", "enter", "return"):
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
	case key.Is("ctrl-i") || (key.Is("tab") && e.HasSelection):
		e.toggleAttribute(func(s *Style, on bool) { s.Italic = on }, func(s Style) bool { return s.Italic })
	case key.Is("ctrl-u"):
		e.toggleAttribute(func(s *Style, on bool) { s.Underline = on }, func(s Style) bool { return s.Underline })
	case key.Is(e.PopoverDefaults.PopupOpenKey):
		e.dragSelecting = false
		e.clearPopoverKeyboardState(false)
		e.popoverSuppressed = false
		e.popoverAtCursor = false
		e.popoverExplicit = true
		if e.HasSelection {
			e.ShowPopover = true
		} else if !e.cursorInVoid && e.selectBoxAtCursor() {
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
