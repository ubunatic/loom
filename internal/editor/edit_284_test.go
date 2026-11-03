package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ubunatic.com/loom"
)

func TestScreenshotPreservesStylingUnicodeAndEditingState(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	docPath := filepath.Join(tempHome, "doc.txt")
	content := "α\nβ\nγ"
	if err := os.WriteFile(docPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	edit, err := loom.NewRichTextEditFromFile(docPath)
	if err != nil {
		t.Fatal(err)
	}

	// Move cursor to test editing state preservation
	edit.Cursor = loom.RichPosition{Line: 1} // at β

	view, err := newEditView(edit, docPath, loom.EditorConfig{Theme: "julia256"})
	if err != nil {
		t.Fatal(err)
	}

	// Pre-draw to set layout boxes
	loom.Render(view, 100, 24)

	before := edit.Cursor
	res := view.ConsumeKey(loom.KeyEvent{Key: "ctrl-p"})
	if !res.Consumed {
		t.Fatalf("ctrl-p not consumed")
	}

	shotDir := filepath.Join(tempHome, "Pictures", "Screenshots")
	entries, err := os.ReadDir(shotDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("ReadDir(%s) err=%v, entries=%d", shotDir, err, len(entries))
	}

	shotPath := filepath.Join(shotDir, entries[0].Name())
	shot, err := os.ReadFile(shotPath)
	if err != nil {
		t.Fatal(err)
	}

	shotStr := string(shot)

	// 1. Check Unicode content is present in the document
	if !strings.Contains(shotStr, "α") || !strings.Contains(shotStr, "β") || !strings.Contains(shotStr, "γ") {
		t.Fatalf("Screenshot missing Unicode content:\n%s", shotStr)
	}

	// 2. Unicode cells survive parsing, and the theme's styles survive too.
	cells := loom.ParseANSI(shotStr)
	found := map[string]bool{}
	styled := false
	for _, cell := range cells {
		for _, r := range []string{"α", "β", "γ"} {
			if strings.Contains(cell.Text, r) {
				found[r] = true
			}
		}
		if cell.Style != (loom.Style{}) {
			styled = true
		}
	}
	if len(found) != 3 {
		t.Fatalf("ParseANSI found %v, want α, β and γ", found)
	}
	if !styled {
		t.Fatal("screenshot has no styled cells")
	}

	// 3. Taking the screenshot must not change the editing state.
	if edit.Cursor != before || before != (loom.RichPosition{Line: 1}) {
		t.Fatalf("cursor = %+v after ctrl-p, want %+v at line 1", edit.Cursor, before)
	}
	if edit.IsModified() {
		t.Fatal("ctrl-p marked the document modified")
	}
}
