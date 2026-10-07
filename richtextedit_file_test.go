// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRichTextEditSaveWithoutPathOpensSavePicker(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
	if err := edit.Save(); err != nil {
		t.Fatalf("Save without path: %v", err)
	}
	if edit.savePicker == nil || edit.savePicker.options.Mode != FilePickerSave || edit.savePopup == nil {
		t.Fatalf("save picker = %#v, want save destination picker", edit.savePicker)
	}
	if edit.savePopup.Inner != edit.savePicker || edit.savePopup.Width != SpeccedDefaults.RichTextEdit.SavePopupMaxWidth || edit.savePopup.Height != SpeccedDefaults.RichTextEdit.SavePopupMaxHeight {
		t.Fatalf("save popup = %+v, want bounded popup containing picker", edit.savePopup)
	}
	if edit.FilePath != "" {
		t.Fatalf("path changed before destination selection: %q", edit.FilePath)
	}
	if edit.savePicker.nameFocus {
		t.Fatal("pathless save picker opened with filename focus instead of directory search")
	}
}

func TestRichTextEditSaveWritesAssociatedPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.ansi")
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello", Style: Style{Bold: true}}}}}})
	edit.FilePath = path
	if err := edit.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), edit.Document.ToANSI()+"\n"; got != want {
		t.Fatalf("saved data = %q, want ANSI serialization %q", got, want)
	}
	if edit.FilePath != path {
		t.Fatalf("associated path = %q, want %q", edit.FilePath, path)
	}
}

func TestRichTextEditSaveEndsWithNewlineAndRoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.ansi")

	// 1. Single line document
	doc1 := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "first line"}}}}}
	edit1 := NewRichTextEdit(doc1)
	edit1.FilePath = path
	if err := edit1.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	data1, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(data1), "\x1b[0m\n") {
		t.Fatalf("saved output %q does not end with \\x1b[0m\\n", string(data1))
	}

	// Load->Save roundtrip
	var doc1Loaded RichDocument
	doc1Loaded.FromANSI(string(data1))
	edit1Loaded := NewRichTextEdit(&doc1Loaded)
	edit1Loaded.FilePath = path
	if err := edit1Loaded.Save(); err != nil {
		t.Fatalf("Save loaded: %v", err)
	}
	data1RoundTrip, _ := os.ReadFile(path)
	if string(data1RoundTrip) != string(data1) {
		t.Fatalf("roundtrip bytes changed: got %q, want %q", string(data1RoundTrip), string(data1))
	}

	// 2. Document ending in an empty line
	doc2 := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "line 1"}}}, {}}}
	edit2 := NewRichTextEdit(doc2)
	edit2.FilePath = path
	if err := edit2.Save(); err != nil {
		t.Fatalf("Save doc ending in empty line: %v", err)
	}
	data2, _ := os.ReadFile(path)
	if !strings.HasSuffix(string(data2), "\x1b[0m\n") {
		t.Fatalf("saved empty-line output %q does not end with \\x1b[0m\\n", string(data2))
	}

	var doc2Loaded RichDocument
	doc2Loaded.FromANSI(string(data2))
	if len(doc2Loaded.Lines) != 2 {
		t.Fatalf("loaded doc2 lines = %d, want 2 (including trailing empty line)", len(doc2Loaded.Lines))
	}
	edit2Loaded := NewRichTextEdit(&doc2Loaded)
	edit2Loaded.FilePath = path
	if err := edit2Loaded.Save(); err != nil {
		t.Fatalf("Save loaded doc2: %v", err)
	}
	data2RoundTrip, _ := os.ReadFile(path)
	if string(data2RoundTrip) != string(data2) {
		t.Fatalf("empty line roundtrip bytes changed: got %q, want %q", string(data2RoundTrip), string(data2))
	}

	// 3. Custom SerializeDocument is written unchanged
	editCustom := NewRichTextEdit(doc1)
	editCustom.FilePath = path
	editCustom.SerializeDocument = func(*RichDocument) ([]byte, error) {
		return []byte("custom-no-newline"), nil
	}
	if err := editCustom.Save(); err != nil {
		t.Fatalf("Save custom: %v", err)
	}
	dataCustom, _ := os.ReadFile(path)
	if string(dataCustom) != "custom-no-newline" {
		t.Fatalf("custom serialization data = %q, want custom-no-newline", string(dataCustom))
	}
}

func TestRichTextEditSaveAsChangesAssociationOnlyAfterSuccessfulWrite(t *testing.T) {
	oldPath := filepath.Join(t.TempDir(), "old.ansi")
	newPath := filepath.Join(t.TempDir(), "new.ansi")
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "new"}}}}})
	edit.FilePath = oldPath
	if err := edit.SaveAs(newPath); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
	if edit.FilePath != newPath {
		t.Fatalf("associated path = %q, want %q", edit.FilePath, newPath)
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("SaveAs did not create destination: %v", err)
	}

	badPath := filepath.Join(t.TempDir(), "missing", "document.ansi")
	if err := edit.SaveAs(badPath); err == nil {
		t.Fatal("SaveAs to missing directory succeeded")
	}
	if edit.FilePath != newPath {
		t.Fatalf("failed SaveAs changed associated path to %q", edit.FilePath)
	}
}

func TestRichTextEditSaveAsCancellationPreservesAssociation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "existing.ansi")
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
	edit.FilePath = path
	if err := edit.openSavePicker(); err != nil {
		t.Fatalf("open Save as picker: %v", err)
	}
	edit.savePicker.ConsumeKey(KeyEvent{Key: "esc"})
	if edit.savePicker != nil {
		t.Fatal("cancel did not close Save as picker")
	}
	if edit.savePopup != nil {
		t.Fatal("cancel did not close Save as popup")
	}
	if edit.FilePath != path {
		t.Fatalf("cancel changed associated path to %q", edit.FilePath)
	}
}

func TestRichTextEditSaveReportsSerializationAndFilesystemErrors(t *testing.T) {
	wantErr := errors.New("serialize failed")
	path := filepath.Join(t.TempDir(), "document.ansi")
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
	edit.FilePath = path
	edit.SerializeDocument = func(*RichDocument) ([]byte, error) { return nil, wantErr }
	if err := edit.Save(); !errors.Is(err, wantErr) {
		t.Fatalf("Save serialization error = %v, want %v", err, wantErr)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("serialization failure created destination: %v", err)
	}

	edit.SerializeDocument = nil
	edit.FilePath = filepath.Join(t.TempDir(), "missing", "document.ansi")
	if err := edit.Save(); err == nil {
		t.Fatal("Save to missing directory succeeded")
	}
}

func TestRichTextEditSaveAsReportsSerializerErrorAndPreservesPath(t *testing.T) {
	wantErr := errors.New("serialize failed")
	oldPath := filepath.Join(t.TempDir(), "old.ansi")
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
	edit.FilePath = oldPath
	edit.SerializeDocument = func(*RichDocument) ([]byte, error) { return nil, wantErr }
	if err := edit.SaveAs(filepath.Join(t.TempDir(), "new.ansi")); !errors.Is(err, wantErr) {
		t.Fatalf("SaveAs serialization error = %v, want %v", err, wantErr)
	}
	if edit.FilePath != oldPath {
		t.Fatalf("serializer failure changed associated path to %q", edit.FilePath)
	}
}

func TestRichTextEditCustomSerializerCanPreserveDocumentMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.rich")
	metadata := &RichPill{Kind: "mention", ID: "ada"}
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "@ada", PillData: metadata}}}}}
	edit := NewRichTextEdit(doc)
	edit.SerializeDocument = func(got *RichDocument) ([]byte, error) {
		if got.Lines[0].Spans[0].PillData != metadata {
			t.Fatal("serializer did not receive the original rich document metadata")
		}
		return []byte("rich-format:@ada:mention:ada"), nil
	}
	if err := edit.SaveAs(path); err != nil {
		t.Fatalf("SaveAs with custom serializer: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "rich-format:@ada:mention:ada"; got != want {
		t.Fatalf("saved custom data = %q, want %q", got, want)
	}
}

func TestRichTextEditSavePickerKeepsAssociationAndReportsSaveError(t *testing.T) {
	wantErr := errors.New("serialize failed")
	oldPath := filepath.Join(t.TempDir(), "old.ansi")
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
	edit.FilePath = oldPath
	edit.SerializeDocument = func(*RichDocument) ([]byte, error) { return nil, wantErr }
	if err := edit.openSavePicker(); err != nil {
		t.Fatalf("open Save as picker: %v", err)
	}
	edit.savePicker.fileName.SetValue("new.ansi")
	edit.savePicker.ConsumeKey(KeyEvent{Key: "tab"})
	edit.savePicker.ConsumeKey(KeyEvent{Key: "enter"})
	if edit.savePicker == nil {
		t.Fatal("save picker closed after failed save")
	}
	if !errors.Is(edit.LastSaveError, wantErr) {
		t.Fatalf("last save error = %v, want %v", edit.LastSaveError, wantErr)
	}
	if edit.FilePath != oldPath {
		t.Fatalf("failed picker save changed associated path to %q", edit.FilePath)
	}
}

func TestRichTextEditSavePickerReopensAfterSave(t *testing.T) {
	t.Chdir(t.TempDir())
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})

	// Open save picker and save
	if err := edit.openSavePicker(); err != nil {
		t.Fatalf("open Save as picker: %v", err)
	}
	edit.savePicker.fileName.SetValue("first.ansi")
	edit.savePicker.ConsumeKey(KeyEvent{Key: "tab"})
	edit.ConsumeKey(KeyEvent{Key: "enter"})

	if edit.savePicker != nil || edit.savePopup != nil {
		t.Fatal("save picker/popup not nil after save")
	}
	if edit.FilePath == "" {
		t.Fatal("FilePath is empty after save")
	}

	// Reopen save picker via openSavePicker / Save as
	if err := edit.openSavePicker(); err != nil {
		t.Fatalf("reopen Save as picker: %v", err)
	}
	if edit.savePicker == nil || edit.savePopup == nil || !edit.savePopup.Open {
		t.Fatal("Save as picker failed to reopen after saving once")
	}

	// Escape closes the reopened picker
	edit.ConsumeKey(KeyEvent{Key: "esc"})
	if edit.savePicker != nil || edit.savePopup != nil {
		t.Fatal("Escape did not clear savePicker and savePopup")
	}

	// Can open again after Escape
	if err := edit.openSavePicker(); err != nil {
		t.Fatalf("reopen Save as picker after Escape: %v", err)
	}
	if edit.savePicker == nil || edit.savePopup == nil {
		t.Fatal("Save as picker failed to reopen after Escape")
	}
}

func TestNewRichTextEditFromFileAndIsModified(t *testing.T) {
	dir := t.TempDir()
	existingPath := filepath.Join(dir, "existing.ansi")
	if err := os.WriteFile(existingPath, []byte("Hello Loom\n"), 0600); err != nil {
		t.Fatal(err)
	}

	// 1. Existing file
	edit1, err := NewRichTextEditFromFile(existingPath)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile existing: %v", err)
	}
	if edit1.FilePath != existingPath {
		t.Fatalf("edit1.FilePath = %q, want %q", edit1.FilePath, existingPath)
	}
	if !edit1.ShowFileBar {
		t.Fatal("edit1.ShowFileBar = false, want true")
	}
	if edit1.IsModified() {
		t.Fatal("newly loaded document reports IsModified() = true")
	}
	edit1.ConsumeKey(KeyEvent{Text: "!"})
	if !edit1.IsModified() {
		t.Fatal("modified document reports IsModified() = false")
	}

	// 2. Missing file
	missingPath := filepath.Join(dir, "missing.ansi")
	edit2, err := NewRichTextEditFromFile(missingPath)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile missing: %v", err)
	}
	if edit2.FilePath != missingPath {
		t.Fatalf("edit2.FilePath = %q, want %q", edit2.FilePath, missingPath)
	}
	if !edit2.ShowFileBar {
		t.Fatal("edit2.ShowFileBar = false, want true")
	}
	if edit2.IsModified() {
		t.Fatal("new empty document for missing path reports IsModified() = true")
	}

	// 3. Read error (e.g. reading a directory)
	_, err = NewRichTextEditFromFile(dir)
	if err == nil {
		t.Fatal("expected error when reading a directory")
	}
}
