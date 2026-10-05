// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRichTextEditSaveWithoutPathOpensSavePicker(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
	if err := edit.Save(); err != nil {
		t.Fatalf("Save without path: %v", err)
	}
	if edit.savePicker == nil || edit.savePicker.options.Mode != FilePickerSave {
		t.Fatalf("save picker = %#v, want save destination picker", edit.savePicker)
	}
	if edit.FilePath != "" {
		t.Fatalf("path changed before destination selection: %q", edit.FilePath)
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
	if got, want := string(data), edit.Document.ToANSI(); got != want {
		t.Fatalf("saved data = %q, want ANSI serialization %q", got, want)
	}
	if edit.FilePath != path {
		t.Fatalf("associated path = %q, want %q", edit.FilePath, path)
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
