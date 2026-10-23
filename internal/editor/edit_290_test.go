package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ubunatic.com/loom"
)

func activateBrowserFile(t *testing.T, v *editView, name, input string) {
	t.Helper()
	v.filePicker.List().SearchBar().SetQuery(name)
	rows := loom.Render(v, 100, 24)
	if input == "keyboard" {
		v.ConsumeKey(loom.KeyEvent{Key: "enter"})
		return
	}
	for y, row := range rows {
		if y > v.browserRect.Y && y < v.browserRect.Y+v.browserRect.H-1 && strings.Contains(row, name) {
			mouse := loom.MouseEvent{X: v.browserRect.X + 3, Y: y, Button: loom.MouseLeft, Action: loom.MousePress}
			v.ConsumeMouse(mouse)
			mouse.Action = loom.MouseRelease
			v.ConsumeMouse(mouse)
			mouse.Action = loom.MousePress
			v.ConsumeMouse(mouse)
			return
		}
	}
	t.Fatalf("browser did not render %s", name)
}

func TestEditBrowserActivationLifecycle(t *testing.T) {
	for _, input := range []string{"keyboard", "mouse"} {
		for _, action := range []string{"Save", "Discard", "Cancel", "Save failure"} {
			t.Run(input+"/"+action, func(t *testing.T) {
				v, original := newDirtyEditView(t)
				target := filepath.Join(filepath.Dir(original), "target.txt")
				if err := os.WriteFile(target, []byte("target content\n"), 0600); err != nil {
					t.Fatal(err)
				}
				// Refresh the picker after adding the second file.
				var err error
				v, err = newEditView(v.edit, original, v.config)
				if err != nil {
					t.Fatal(err)
				}
				v.ConsumeKey(loom.KeyEvent{Key: "f2"})
				v.ConsumeKey(loom.KeyEvent{Key: "esc"})
				if v.showSidePanel || v.focused != focusEditor {
					t.Fatal("cancel did not hide browser and return editor focus")
				}
				v.ConsumeKey(loom.KeyEvent{Key: "f2"})
				v.edit.ConsumeKey(loom.KeyEvent{Text: "X"})
				if action == "Save failure" {
					if err := os.Remove(original); err != nil {
						t.Fatal(err)
					}
					if err := os.Mkdir(original, 0700); err != nil {
						t.Fatal(err)
					}
				}
				activateBrowserFile(t, v, "target.txt", input)
				if v.unsavedDialog == nil || v.filePath != original {
					t.Fatal("activation bypassed unsaved changes guard")
				}
				steps := map[string]int{"Save": 0, "Save failure": 0, "Discard": 1, "Cancel": 2}[action]
				for range steps {
					v.ConsumeKey(loom.KeyEvent{Key: "right"})
				}
				v.ConsumeKey(loom.KeyEvent{Key: "enter"})
				if v.unsavedDialog != nil || v.pendingAction != nil {
					t.Fatal("dialog left a stale pending action")
				}
				if action == "Cancel" || action == "Save failure" {
					if v.filePath != original || !v.edit.IsModified() {
						t.Fatal("cancel/failed save lost the dirty document")
					}
					if action == "Save failure" && v.statusMessage == "" {
						t.Fatal("save failure was silent")
					}
					// Retry the same picker and discard to prove it remains usable.
					activateBrowserFile(t, v, "target.txt", input)
					if v.unsavedDialog == nil {
						t.Fatal("picker stopped activating after cancel/failed save")
					}
					v.ConsumeKey(loom.KeyEvent{Key: "right"})
					v.ConsumeKey(loom.KeyEvent{Key: "enter"})
				}
				if v.filePath != target || v.edit.Document.ToPlainText() != "target content" || v.focused != focusBrowser {
					t.Fatalf("activation: path=%s text=%q focus=%v", v.filePath, v.edit.Document.ToPlainText(), v.focused)
				}
				if action != "Save failure" {
					data, err := os.ReadFile(original)
					if err != nil || strings.Contains(string(data), "X") != (action == "Save") {
						t.Fatalf("%s disk content = %q, err=%v", action, data, err)
					}
					activateBrowserFile(t, v, "doc.txt", input)
					if v.filePath != original {
						t.Fatal("picker stopped navigating after activation")
					}
				}
			})
		}
	}
}
