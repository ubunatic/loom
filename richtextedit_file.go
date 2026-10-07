// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

// DocState is the save state of a RichTextEdit document.
type DocState int

const (
	// DocStateUntitled means the document has no file path and no changes.
	DocStateUntitled DocState = iota
	// DocStateSaved means the document matches its file.
	DocStateSaved
	// DocStateModified means the document has unsaved changes.
	DocStateModified
	// DocStateError means the last save failed.
	DocStateError
)

// Label returns the spec-defined display label for the state.
func (s DocState) Label() string {
	defs := SpeccedDefaults.RichTextEdit
	switch s {
	case DocStateSaved:
		return defs.StateSavedLabel
	case DocStateModified:
		return defs.StateModifiedLabel
	case DocStateError:
		return defs.StateErrorLabel
	}
	return defs.StateUntitledLabel
}

// DocState reports the single source of truth for the document's save state.
// A failed save wins over modified; unchanged documents without a path are untitled.
func (e *RichTextEdit) DocState() DocState {
	if e == nil {
		return DocStateUntitled
	}
	switch {
	case e.LastSaveError != nil:
		return DocStateError
	case e.IsModified():
		return DocStateModified
	case e.FilePath == "":
		return DocStateUntitled
	}
	return DocStateSaved
}

// notifyStateChange calls OnStateChange when the state differs from the last
// state reported; the baseline is Untitled until a state was reported.
func (e *RichTextEdit) notifyStateChange() {
	state := e.DocState()
	if state == e.notifiedState {
		return
	}
	e.notifiedState = state
	if e.OnStateChange != nil {
		e.OnStateChange(state)
	}
}

// IsModified reports whether the document has unsaved changes compared to its last saved state.
func (e *RichTextEdit) IsModified() bool {
	if e == nil {
		return false
	}
	e.ensureDocument()
	return !reflect.DeepEqual(e.savedDocument, cloneRichDocumentLines(e.Document))
}

// NewRichTextEditFromFile creates a RichTextEdit populated with content from path.
// If path does not exist, it initializes an empty document bound to path.
// It sets FilePath = path and ShowFileBar = true.
func NewRichTextEditFromFile(path string) (*RichTextEdit, error) {
	doc := &RichDocument{}
	data, err := os.ReadFile(path)
	if err == nil {
		doc.FromANSI(string(data))
	} else if os.IsNotExist(err) {
		doc.Lines = []RichLine{{}}
	} else {
		return nil, fmt.Errorf("read RichTextEdit file %q: %w", path, err)
	}
	edit := NewRichTextEdit(doc)
	edit.FilePath = path
	edit.ShowFileBar = true
	edit.notifiedState = edit.DocState()
	return edit, nil
}

// RequestOpen asks the host to open a file by calling OnOpenRequest.
func (e *RichTextEdit) RequestOpen() {
	if e.OnOpenRequest != nil {
		e.OnOpenRequest()
	}
}

// RequestClose asks to close the document. The host's OnCloseRequest owns the
// unsaved-changes guard; without one the document closes immediately.
func (e *RichTextEdit) RequestClose() {
	if e.OnCloseRequest != nil {
		e.OnCloseRequest()
		return
	}
	e.Close()
}

// Close discards the document and its file association and leaves an empty,
// untitled buffer. It does not ask about unsaved changes; hosts do that first.
func (e *RichTextEdit) Close() {
	e.Document = &RichDocument{Lines: []RichLine{{}}}
	e.FilePath = ""
	e.LastSaveError = nil
	e.Cursor = RichPosition{}
	e.ClearSelection()
	e.undoStack, e.redoStack = nil, nil
	e.savedDocument = cloneRichDocumentLines(e.Document)
	e.Highlights = nil
	e.notifyStateChange()
}

// Save writes the document to its associated path. If no path is associated,
// it opens a FilePicker in save mode and returns while the user chooses one.
func (e *RichTextEdit) Save() error {
	e.ensureDocument()
	if e.FilePath == "" {
		return e.openSavePicker()
	}
	return e.SaveAs(e.FilePath)
}

// SaveAs serializes the document, writes it to path, and associates path only
// after both serialization and writing succeed.
func (e *RichTextEdit) SaveAs(path string) error {
	defer e.notifyStateChange()
	if path == "" {
		e.LastSaveError = fmt.Errorf("loom: save document: destination path is empty")
		return e.LastSaveError
	}
	e.ensureDocument()
	e.Document.normalize()
	var data []byte
	var err error
	if e.SerializeDocument == nil {
		data = []byte(e.Document.ToANSI() + "\n")
	} else {
		data, err = e.SerializeDocument(e.Document)
		if err != nil {
			e.LastSaveError = fmt.Errorf("loom: serialize rich document: %w", err)
			return e.LastSaveError
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		e.LastSaveError = fmt.Errorf("loom: write rich document %q: %w", path, err)
		return e.LastSaveError
	}
	e.FilePath = path
	e.LastSaveError = nil
	e.savedDocument = cloneRichDocumentLines(e.Document)
	return nil
}

func cloneRichDocumentLines(doc *RichDocument) []RichLine {
	if doc == nil || len(doc.Lines) == 0 {
		return []RichLine{{}}
	}
	lines := make([]RichLine, len(doc.Lines))
	for i, line := range doc.Lines {
		lines[i].Spans = mergeRichSpans(line.Spans)
		for j := range lines[i].Spans {
			pill := lines[i].Spans[j].PillData
			if pill == nil {
				continue
			}
			copyPill := *pill
			if pill.Metadata != nil {
				copyPill.Metadata = make(map[string]string, len(pill.Metadata))
				for key, value := range pill.Metadata {
					copyPill.Metadata[key] = value
				}
			}
			lines[i].Spans[j].PillData = &copyPill
		}
	}
	return lines
}

func (e *RichTextEdit) openSavePicker() error {
	directory, name := ".", ""
	if e.FilePath != "" {
		directory, name = filepath.Dir(e.FilePath), filepath.Base(e.FilePath)
	}
	picker, err := NewFilePicker(directory, FilePickerOptions{
		Mode:     FilePickerSave,
		FileName: name,
		OnSave: func(path string) error {
			if err := e.SaveAs(path); err != nil {
				return err
			}
			e.savePicker = nil
			e.savePopup = nil
			return nil
		},
		OnCancel: func() {
			e.savePicker = nil
			e.savePopup = nil
		},
	})
	if err != nil {
		return fmt.Errorf("loom: open save destination picker: %w", err)
	}
	e.savePicker = picker
	e.savePopup = NewPopup("Save as", picker)
	e.savePopup.Width = SpeccedDefaults.RichTextEdit.SavePopupMaxWidth
	e.savePopup.Height = SpeccedDefaults.RichTextEdit.SavePopupMaxHeight
	e.savePopup.Style = DefaultMenuStyle().Normal
	e.savePopup.DismissOnOutsideClick = true
	return nil
}
