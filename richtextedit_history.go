// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"reflect"
)

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
	e.Document.normalize()
	before := e.snapshot()
	edit()
	e.Document.normalize()
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
