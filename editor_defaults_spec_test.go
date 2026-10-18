package loom

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
	"ubunatic.com/loom/spec"
)

func TestEditorEmbeddedDefaults(t *testing.T) {
	data, err := spec.DefaultsYAML()
	if err != nil {
		t.Fatal(err)
	}
	var loaded LibDefaults
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.Editor, SpeccedDefaults.Editor) || !reflect.DeepEqual(loaded.RichTextEdit, SpeccedDefaults.RichTextEdit) {
		t.Fatal("embedded defaults differ from runtime")
	}
	if loaded.Editor.BrowserMaxWidth <= 0 || loaded.RichTextEdit.UndoLimit <= 0 || len(loaded.RichTextEdit.Bindings) == 0 {
		t.Fatal("editor defaults missing")
	}
	e := NewRichTextEdit(nil)
	if e.ShowCursor != loaded.RichTextEdit.ShowCursor || e.ShowPopover != loaded.RichTextEdit.ShowPopover {
		t.Fatal("constructor ignored spec")
	}
}

func TestEditorSpecDrivesHistoryAndBinding(t *testing.T) {
	before := SpeccedDefaults.RichTextEdit
	t.Cleanup(func() { SpeccedDefaults.RichTextEdit = before })
	d := &SpeccedDefaults.RichTextEdit
	d.UndoLimit = 2
	d.Bindings = map[string][]string{}
	for k, v := range before.Bindings {
		d.Bindings[k] = append([]string(nil), v...)
	}
	d.Bindings["bold"] = []string{"ctrl-alt-b"}
	e := NewRichTextEdit(nil)
	if !e.ConsumeKey(KeyEvent{Key: "ctrl-alt-b"}).Consumed || !e.ActiveStyle.Bold {
		t.Fatal("spec binding ignored")
	}
	for range d.UndoLimit + 3 {
		e.ConsumeKey(KeyEvent{Text: "a "})
	}
	if len(e.undoStack) != d.UndoLimit {
		t.Fatalf("undo history length = %d", len(e.undoStack))
	}
}

func TestEditorSpecDrivesFilesAndClickTiming(t *testing.T) {
	before := SpeccedDefaults.RichTextEdit
	t.Cleanup(func() { SpeccedDefaults.RichTextEdit = before })
	d := &SpeccedDefaults.RichTextEdit
	d.FinalNewline = false
	d.FileMode = 0o600
	d.CreateMissingFile = false
	path := filepath.Join(t.TempDir(), "new.ansi")
	if _, err := NewRichTextEditFromFile(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing-file policy: %v", err)
	}
	e := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "text"}}}}})
	if err := e.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// ANSI serialization retains its terminal reset; only the trailing newline is configurable.
	if string(data) != e.Document.ToANSI() {
		t.Fatalf("final newline policy ignored: %q", data)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != os.FileMode(d.FileMode) {
		t.Fatalf("file mode = %v", info.Mode())
	}
	d.MultiClickWindow = time.Second
	now := time.Unix(1, 0)
	e.now = func() time.Time { return now }
	if e.registerClick(RichPosition{}) != 1 {
		t.Fatal("first click")
	}
	now = now.Add(d.MultiClickWindow / 2)
	if e.registerClick(RichPosition{}) != 2 {
		t.Fatal("spec click interval ignored")
	}
	now = now.Add(d.MultiClickWindow + time.Millisecond)
	if e.registerClick(RichPosition{}) != 1 {
		t.Fatal("expired click interval ignored")
	}
}

func TestEditorSpecHelpPageBindingsAndMouse(t *testing.T) {
	before := SpeccedDefaults.RichTextEdit
	t.Cleanup(func() { SpeccedDefaults.RichTextEdit = before })
	SpeccedDefaults.RichTextEdit.HelpPageStep = 3
	for _, key := range []string{"pgdown", "pgdn", "pagedown"} {
		h := &richTextEditHelp{maxScroll: 100}
		h.ConsumeKey(KeyEvent{Key: key})
		if h.scroll != SpeccedDefaults.RichTextEdit.HelpPageStep {
			t.Fatalf("%s scroll = %d", key, h.scroll)
		}
	}
	for _, key := range []string{"pgup", "pageup"} {
		h := &richTextEditHelp{scroll: 10, maxScroll: 100}
		h.ConsumeKey(KeyEvent{Key: key})
		if h.scroll != 10-SpeccedDefaults.RichTextEdit.HelpPageStep {
			t.Fatalf("%s scroll = %d", key, h.scroll)
		}
	}
	h := &richTextEditHelp{maxScroll: 100}
	if !h.ConsumeMouse(MouseEvent{Action: MouseScrollDown}).Consumed || h.scroll != 1 {
		t.Fatal("mouse help scrolling")
	}
}

func TestEditorSpecConfigPath(t *testing.T) {
	before := SpeccedDefaults.Editor
	t.Cleanup(func() { SpeccedDefaults.Editor = before })
	SpeccedDefaults.Editor.ConfigDirectory = "editor-test"
	SpeccedDefaults.Editor.ConfigFile = "preferences.yaml"
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if got := DefaultEditorConfigPath(); got != filepath.Join(dir, SpeccedDefaults.Editor.ConfigDirectory, SpeccedDefaults.Editor.ConfigFile) {
		t.Fatalf("config path = %q", got)
	}
}
