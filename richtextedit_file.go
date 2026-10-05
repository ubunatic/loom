// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"os"
	"path/filepath"
)

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
	if path == "" {
		e.LastSaveError = fmt.Errorf("loom: save document: destination path is empty")
		return e.LastSaveError
	}
	e.ensureDocument()
	var data []byte
	var err error
	if e.SerializeDocument == nil {
		data = []byte(e.Document.ToANSI())
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
		lines[i].Spans = append([]RichSpan(nil), line.Spans...)
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
	return nil
}
