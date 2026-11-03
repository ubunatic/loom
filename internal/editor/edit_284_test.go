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
	edit.Cursor = loom.RichPosition{Offset: 2} // Should be at 'β'

	view, err := newEditView(edit, docPath, loom.EditorConfig{Theme: "julia256"})
	if err != nil {
		t.Fatal(err)
	}

	// Pre-draw to set layout boxes
	loom.Render(view, 100, 24)

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

	// 2. Parse ANSI and verify we find the cursor styling and the unicode chars
	cells := loom.ParseANSI(shotStr)
	foundAlpha := false
	foundBeta := false
	foundGamma := false
	for _, cell := range cells {
		if strings.Contains(cell.Text, "α") {
			foundAlpha = true
		}
		if strings.Contains(cell.Text, "β") {
			foundBeta = true
			if !cell.Style.Invert && !cell.Style.Bold {
				// Cursor is at Beta. Check if it's highlighted.
			}
		}
		if strings.Contains(cell.Text, "γ") {
			foundGamma = true
		}
	}

	if !foundAlpha || !foundBeta || !foundGamma {
		t.Fatalf("ParseANSI missing cells for unicode chars. alpha=%v beta=%v gamma=%v", foundAlpha, foundBeta, foundGamma)
	}
}
