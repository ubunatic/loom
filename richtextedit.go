// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"time"
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
	// Highlights are display-only ranges, independent of editing selection.
	Highlights  []RichTextHighlight
	ActiveStyle Style
	ScrollX     int
	ScrollY     int
	ShowCursor  bool
	ShowPopover bool
	// PopoverDefaults configures popup behavior; initialized from the embedded spec.
	PopoverDefaults      RichTextEditDefaults
	popoverExplicit      bool
	popoverSelectionFrom RichPosition
	popoverSelectionTo   RichPosition
	popoverSelectionAt   time.Time
	// ViewMode makes the editor read-only: navigation, selection and copy work;
	// every edit, style key and the format popover are disabled and edit keys
	// return Ignored so the app can use them.
	ViewMode bool
	// BoxMode draws box glyphs into adjacent document cells with arrow keys.
	BoxMode bool
	// GhostCursorEnabled allows vertical and horizontal navigation into empty
	// document space. The default comes from spec/defaults.yaml.
	GhostCursorEnabled bool
	// ShowFileBar reserves the final row for the File menu and document status.
	// It is disabled by default to preserve the viewport.
	ShowFileBar bool
	// FilePath is the path associated with the document by SaveAs or Save.
	FilePath string
	// LastSaveError holds the most recent save failure, including an error
	// returned after choosing a destination in the asynchronous save picker.
	LastSaveError error
	// OnStateChange, when set, is called once per DocState transition, after
	// the key, mouse event or save that caused it.
	OnStateChange func(state DocState)
	// OnOpenRequest, when set, is called when the user asks to open a file
	// (Open menu item or its hotkey). The host decides how to present files.
	OnOpenRequest func()
	// OnCloseRequest, when set, is called when the user asks to close the
	// document (Close menu item or its hotkey). The host runs its unsaved-changes
	// guard and then calls Close. Without it, RequestClose calls Close directly.
	OnCloseRequest func()
	// SerializeDocument optionally serializes the document for saving. A nil
	// callback uses RichDocument.ToANSI.
	SerializeDocument func(*RichDocument) ([]byte, error)

	focused                 bool
	lastRect                Rect
	savePicker              *FilePicker
	savePopup               *Popup
	helpPopup               *Popup
	helpSections            []HelpSection
	fileBar                 *richTextEditFileBar
	chromeTheme             *ThemeColors
	savedDocument           []RichLine
	notifiedState           DocState
	selectionAnchor         RichPosition
	selectionExtending      bool
	dragSelecting           bool
	activeStyleSet          bool
	popoverButtons          []richPopoverButton
	popoverColorPicker      *ColorPicker
	popoverColorRect        Rect
	popoverSubmenu          string
	popoverBoxChoices       []richPopoverBoxChoice
	popoverHamburgerChoices []richPopoverHamburgerChoice
	popoverRelease          bool
	popoverFocus            int
	popoverFocusSet         bool
	popoverSubmenuFocus     int
	popoverSubmenuFocusSet  bool
	popoverSuppressed       bool
	popoverAtCursor         bool
	cursorInVoid            bool
	voidLine                int
	voidColumn              int
	clipboard               []RichLine
	undoStack               []richSnapshot
	redoStack               []richSnapshot
	typingRun               bool
	clickCount              int
	lastClickAt             time.Time
	lastClickPos            RichPosition
	now                     func() time.Time
	boxStrokeBefore         *richSnapshot
	boxSelection            *richBoxSelection
	boxStrokeFG             Color
	boxStrokeFGSet          bool
}

// NewRichTextEdit creates a rich text view for doc. A nil document is treated
// as an empty document.
func NewRichTextEdit(doc *RichDocument) *RichTextEdit {
	e := &RichTextEdit{Document: doc, PopoverDefaults: SpeccedDefaults.RichTextEdit, ShowCursor: SpeccedDefaults.RichTextEdit.ShowCursor, ShowPopover: SpeccedDefaults.RichTextEdit.ShowPopover, GhostCursorEnabled: SpeccedDefaults.RichTextEdit.GhostCursorEnabled, focused: true}
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

// ModalOpen reports whether a popup, menu or visible format popover captures
// input. ModalMouseTarget exposes that state to hosts and containers. A toolbar
// appearing during document drag selection acquires capture after the release.
func (e *RichTextEdit) ModalOpen() bool {
	return (e.helpPopup != nil && e.helpPopup.Open) || (e.savePopup != nil && e.savePopup.Open) ||
		(e.ShowFileBar && e.fileBar != nil && e.fileBar.menu.Open) || (e.popoverOpen() && !e.dragSelecting)
}

// ModalMouseTarget captures mouse input while an editor overlay is active.
func (e *RichTextEdit) ModalMouseTarget() (Widget, Rect) {
	if e != nil && e.ModalOpen() {
		return e, Rect{}
	}
	return nil, Rect{}
}

func (e *RichTextEdit) popoverOpen() bool {
	if e.popoverColorPicker != nil {
		return true
	}
	_, _, _, _, open := e.popoverGeometry(e.lastRect, e.documentLines())
	return open
}

// HotkeyHint returns the most useful file and help shortcuts for the available width.
func (e *RichTextEdit) HotkeyHint(width int) string {
	return e.HotkeyBar().Text(width)
}

// HotkeyBar returns structured, clickable hints for the editor's key actions.
// Draw the bar in the host's hint row and route row-local mouse events to it.
func (e *RichTextEdit) HotkeyBar() *HintBar {
	mode := SpeccedDefaults.RichTextEdit.ViewLabel
	if e.ViewMode {
		mode = SpeccedDefaults.RichTextEdit.EditLabel
	}
	defs := SpeccedDefaults.RichTextEdit
	entry := func(binding, label string) HintEntry {
		return HintEntry{Binding: binding, Label: label, Action: func() EventResult { return e.ConsumeKey(KeyEvent{Key: binding}) }}
	}
	bar := NewHintBar(
		entry(defs.HotkeyHelpBinding, defs.HotkeyHelpLabel),
		entry(defs.HotkeySaveBinding, defs.HotkeySaveLabel),
		entry(defs.HotkeyViewEditBinding, mode),
	)
	bar.ApplyTheme(e.toolbarTheme())
	return bar
}

// ApplyTheme updates the editor's file menu, status and hotkey chrome.
func (e *RichTextEdit) ApplyTheme(theme ThemeColors) { e.chromeTheme = &theme }

func (e *RichTextEdit) toolbarTheme() ThemeColors {
	if e.chromeTheme != nil {
		return *e.chromeTheme
	}
	return DefaultTheme()
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

// AddHelpSection appends a host-specific section to the F1 help, after the
// built-in sections.
func (e *RichTextEdit) AddHelpSection(section HelpSection) {
	e.helpSections = append(e.helpSections, section)
}
