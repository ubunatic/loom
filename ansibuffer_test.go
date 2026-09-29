// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestValidateAnsiBox(t *testing.T) {
	for _, tc := range []struct {
		name    string
		text    string
		wantErr bool
	}{
		{name: "valid box", text: "┌───┐\n│abc│\n└───┘\n"},
		{name: "valid box with flag", text: "┌────┐\n│🇩🇪  │\n└────┘", wantErr: false},
		{name: "SGR and wide rune", text: "\x1b[31m┌───┐\x1b[0m\n│界 │\n└───┘"},
		{name: "ragged box row", text: "┌───┐\n│abc  │\n└───┘\n", wantErr: true},
		{name: "misaligned divider", text: "┌───┐\n│abc│\n├──┤\n│def│\n└───┘\n", wantErr: true},
		{name: "unboxed uneven text", text: "plain text\nshort"},
		{name: "adjacent separate boxes", text: "┌──┐\n│a │\n└──┘\n┌────┐\n│ b  │\n└────┘"},
		{name: "padding outside box", text: "┌───┐   \n│abc│\n└───┘", wantErr: true},
		{name: "cursor-addressed recording", text: "\x1b[2;1H┌───┐\x1b[3;1H│abc│"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAnsiBox(tc.text)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateAnsiBox() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestParseAnsiBufferKeepsRegionalIndicatorPairInTwoColumns(t *testing.T) {
	buf, err := ParseAnsiBuffer("🇩🇪X", 5, 1)
	if err != nil {
		t.Fatalf("ParseAnsiBuffer() error = %v", err)
	}
	if got := buf.Get(0, 0).Rune; got != '🇩' {
		t.Fatalf("cell 0 = %q, want first regional indicator", got)
	}
	if got := buf.Get(1, 0).Rune; got != '🇪' {
		t.Fatalf("cell 1 = %q, want second regional indicator", got)
	}
	if got := buf.Get(2, 0).Rune; got != 'X' {
		t.Fatalf("cell 2 = %q, want text immediately after two-column flag", got)
	}
	if got := buf.Serialize(); got != "🇩🇪X\x1b[0m" {
		t.Fatalf("Serialize() = %q, want original flag and X followed by the serializer reset", got)
	}
}

func TestAllAnsiAssetsHaveValidBoxes(t *testing.T) {
	files, err := ansiAssetFiles(".")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no ANSI assets found")
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("read %s: %v", path, err)
			continue
		}
		if err := ValidateAnsiBox(string(data)); err != nil {
			t.Errorf("%s: invalid ANSI box: %v", path, err)
		}
	}
}

func TestAnsiAssetWalkSkipsDotDirectories(t *testing.T) {
	root := t.TempDir()
	for path, contents := range map[string]string{
		"good.ansi":                 "plain",
		".loom/broken.ansi":         "┌──┐\n│x │\n└─┘",
		".cache/nested/broken.ansi": "┌──┐\n│x │\n└─┘",
	} {
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	files, err := ansiAssetFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != filepath.Join(root, "good.ansi") {
		t.Fatalf("ANSI assets = %v, want only %s", files, filepath.Join(root, "good.ansi"))
	}
}

// ansiAssetFiles walks the checkout deterministically, including untracked
// artwork while skipping repository metadata and generated/vendor trees.
func ansiAssetFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if filepath.Clean(path) != filepath.Clean(root) && len(entry.Name()) > 0 && entry.Name()[0] == '.' {
				return filepath.SkipDir
			}
			switch entry.Name() {
			case ".git", "vendor", "build", "dist", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".ansi" {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func TestAnsiBufferCreation(t *testing.T) {
	buf := NewAnsiBuffer(50, 20)
	if buf.Cols() != 50 || buf.Rows() != 20 {
		t.Fatalf("expected 50x20, got %dx%d", buf.Cols(), buf.Rows())
	}
	if buf.Modified() {
		t.Fatalf("expected unmodified")
	}
	for y := 0; y < 20; y++ {
		for x := 0; x < 50; x++ {
			c := buf.Get(x, y)
			if !c.IsBlank() {
				t.Fatalf("expected blank cell at (%d,%d), got %+v", x, y, c)
			}
		}
	}

	// Out of bounds Get returns blank
	if !buf.Get(-1, 0).IsBlank() || !buf.Get(100, 100).IsBlank() {
		t.Fatalf("expected out-of-bounds Get to return blank")
	}
}

func TestAnsiBufferEditing(t *testing.T) {
	buf := NewAnsiBuffer(10, 5)

	// Put in Overtype mode
	buf.PutChar(0, 0, 'H', ColorIndex(1), ColorReset(), true, false, false, false, AnsiModeOvertype)
	buf.PutChar(1, 0, 'i', ColorIndex(2), ColorReset(), false, false, false, false, AnsiModeOvertype)

	if !buf.Modified() {
		t.Fatalf("expected buffer to be modified")
	}

	c0 := buf.Get(0, 0)
	if c0.Rune != 'H' || !c0.Bold || c0.FG != ColorIndex(1) {
		t.Fatalf("unexpected cell 0: %+v", c0)
	}

	c1 := buf.Get(1, 0)
	if c1.Rune != 'i' || c1.FG != ColorIndex(2) {
		t.Fatalf("unexpected cell 1: %+v", c1)
	}

	// Insert mode: put 'X' at (0, 0) shifts 'H' to (1, 0) and 'i' to (2, 0)
	buf.PutChar(0, 0, 'X', ColorReset(), ColorReset(), false, false, false, false, AnsiModeInsert)
	if buf.Get(0, 0).Rune != 'X' || buf.Get(1, 0).Rune != 'H' || buf.Get(2, 0).Rune != 'i' {
		t.Fatalf("unexpected row 0 after insert: %c %c %c", buf.Get(0, 0).Rune, buf.Get(1, 0).Rune, buf.Get(2, 0).Rune)
	}

	// Delete in insert mode at (1, 0) deletes 'H' and shifts 'i' left
	buf.Delete(1, 0, AnsiModeInsert)
	if buf.Get(0, 0).Rune != 'X' || buf.Get(1, 0).Rune != 'i' || buf.Get(2, 0).Rune != ' ' {
		t.Fatalf("unexpected row 0 after delete: %c %c %c", buf.Get(0, 0).Rune, buf.Get(1, 0).Rune, buf.Get(2, 0).Rune)
	}

	// Backspace in overtype mode at (1, 0) clears (0, 0)
	newX := buf.Backspace(1, 0, AnsiModeOvertype)
	if newX != 0 || buf.Get(0, 0).Rune != ' ' {
		t.Fatalf("unexpected row 0 after backspace: newX=%d, rune=%c", newX, buf.Get(0, 0).Rune)
	}
}

func TestAnsiBufferClipboard(t *testing.T) {
	buf := NewAnsiBuffer(10, 5)
	buf.Put(3, 2, 'A', ColorIndex(9), ColorIndex(4), true, true, true, false)

	copied := buf.Copy(3, 2)
	if copied.Rune != 'A' || copied.FG != ColorIndex(9) || copied.BG != ColorIndex(4) || !copied.Bold {
		t.Fatalf("unexpected copied cell: %+v", copied)
	}

	// Paste at (0, 0)
	ok := buf.Paste(0, 0)
	if !ok {
		t.Fatalf("expected paste to succeed")
	}
	pasted := buf.Get(0, 0)
	if pasted.Rune != 'A' || pasted.FG != ColorIndex(9) || pasted.BG != ColorIndex(4) {
		t.Fatalf("unexpected pasted cell: %+v", pasted)
	}

	// Cut at (3, 2)
	cut := buf.Cut(3, 2)
	if cut.Rune != 'A' {
		t.Fatalf("unexpected cut cell: %+v", cut)
	}
	if !buf.Get(3, 2).IsBlank() {
		t.Fatalf("expected cell (3, 2) to be blank after cut")
	}

	// SetClipboard
	buf.SetClipboard(AnsiCell{Rune: 'Z', FG: ColorIndex(10)})
	clip, hasClip := buf.Clipboard()
	if !hasClip || clip.Rune != 'Z' {
		t.Fatalf("unexpected clipboard: %+v (hasClip=%v)", clip, hasClip)
	}
}

func TestAnsiBufferNavigation(t *testing.T) {
	buf := NewAnsiBuffer(20, 5)
	// Line 0: "  Hello   World  "
	buf.Put(2, 0, 'H', ColorReset(), ColorReset(), false, false, false, false)
	buf.Put(3, 0, 'e', ColorReset(), ColorReset(), false, false, false, false)
	buf.Put(4, 0, 'l', ColorReset(), ColorReset(), false, false, false, false)
	buf.Put(5, 0, 'l', ColorReset(), ColorReset(), false, false, false, false)
	buf.Put(6, 0, 'o', ColorReset(), ColorReset(), false, false, false, false)

	buf.Put(10, 0, 'W', ColorReset(), ColorReset(), false, false, false, false)
	buf.Put(11, 0, 'o', ColorReset(), ColorReset(), false, false, false, false)
	buf.Put(12, 0, 'r', ColorReset(), ColorReset(), false, false, false, false)
	buf.Put(13, 0, 'l', ColorReset(), ColorReset(), false, false, false, false)
	buf.Put(14, 0, 'd', ColorReset(), ColorReset(), false, false, false, false)

	// Next word from 0 -> 2 (start of Hello), from 2 -> 10 (start of World)
	if next := buf.NextWord(0, 0); next != 2 {
		t.Fatalf("expected next word at 2, got %d", next)
	}
	if next := buf.NextWord(2, 0); next != 10 {
		t.Fatalf("expected next word at 10, got %d", next)
	}

	// Prev word from 14 -> 10, from 10 -> 2
	if prev := buf.PrevWord(14, 0); prev != 10 {
		t.Fatalf("expected prev word at 10, got %d", prev)
	}
	if prev := buf.PrevWord(10, 0); prev != 2 {
		t.Fatalf("expected prev word at 2, got %d", prev)
	}

	// Line 2 has an object
	buf.Put(5, 2, 'X', ColorReset(), ColorReset(), false, false, false, false)
	// Object hopping
	if row := buf.NextObjectRow(0); row != 2 {
		t.Fatalf("expected next object row 2, got %d", row)
	}
	if row := buf.PrevObjectRow(2); row != 0 {
		t.Fatalf("expected prev object row 0, got %d", row)
	}
}

func TestAnsiBufferResize(t *testing.T) {
	buf := NewAnsiBuffer(5, 5)
	buf.Put(2, 2, 'C', ColorReset(), ColorReset(), false, false, false, false)
	buf.Resize(10, 10)
	if buf.Cols() != 10 || buf.Rows() != 10 {
		t.Fatalf("expected 10x10, got %dx%d", buf.Cols(), buf.Rows())
	}
	if buf.Get(2, 2).Rune != 'C' {
		t.Fatalf("cell (2,2) lost after resize: %+v", buf.Get(2, 2))
	}
}

func TestAnsiBufferSerializeAndParse(t *testing.T) {
	buf := NewAnsiBuffer(10, 3)
	buf.Put(0, 0, 'A', ColorIndex(1), ColorReset(), true, false, false, false)
	buf.Put(1, 0, 'B', ColorIndex(2), ColorIndex(4), false, false, true, false)
	buf.Put(0, 1, 'C', ColorRGB(100, 150, 200), ColorReset(), false, true, false, false)

	serialized := buf.Serialize()
	if len(serialized) == 0 {
		t.Fatalf("expected non-empty serialized ANSI")
	}

	if s := SerializeAnsiBuffer(buf); s != serialized {
		t.Fatalf("expected helper output to match method output")
	}

	parsed, err := ParseAnsiBuffer(serialized, 10, 3)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	c00 := parsed.Get(0, 0)
	if c00.Rune != 'A' || c00.FG != ColorIndex(1) || !c00.Bold {
		t.Fatalf("unexpected parsed (0,0): %+v", c00)
	}

	c10 := parsed.Get(1, 0)
	if c10.Rune != 'B' || c10.FG != ColorIndex(2) || c10.BG != ColorIndex(4) || !c10.Underline {
		t.Fatalf("unexpected parsed (1,0): %+v", c10)
	}

	c01 := parsed.Get(0, 1)
	if c01.Rune != 'C' || c01.FG != ColorRGB(100, 150, 200) || !c01.Dim {
		t.Fatalf("unexpected parsed (0,1): %+v", c01)
	}
}

func TestAnsiBufferSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.ansi")

	buf := NewAnsiBuffer(12, 4)
	buf.Put(0, 0, 'L', ColorIndex(36), ColorReset(), true, false, false, false)
	buf.Put(1, 0, 'o', ColorIndex(36), ColorReset(), true, false, false, false)
	buf.Put(2, 0, 'o', ColorIndex(36), ColorReset(), true, false, false, false)
	buf.Put(3, 0, 'm', ColorIndex(36), ColorReset(), true, false, false, false)

	if err := SaveAnsiBuffer(buf, filePath); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if buf.Modified() {
		t.Fatalf("expected unmodified after save")
	}

	loaded, err := LoadAnsiBuffer(filePath, 12, 4)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Get(0, 0).Rune != 'L' || loaded.Get(3, 0).Rune != 'm' {
		t.Fatalf("unexpected loaded buffer content: %c %c", loaded.Get(0, 0).Rune, loaded.Get(3, 0).Rune)
	}

	// Non-existent file loads as blank buffer
	nonExistent := filepath.Join(tmpDir, "non_existent.ansi")
	blankBuf, err := LoadAnsiBuffer(nonExistent, 20, 10)
	if err != nil {
		t.Fatalf("expected success for non existent file, got %v", err)
	}
	if blankBuf.Cols() != 20 || blankBuf.Rows() != 10 {
		t.Fatalf("expected 20x10, got %dx%d", blankBuf.Cols(), blankBuf.Rows())
	}
	if blankBuf.Path() != nonExistent {
		t.Fatalf("expected path %s, got %s", nonExistent, blankBuf.Path())
	}
}

func TestAnsiBufferDesign004File(t *testing.T) {
	designPath := "docs/data/ansiedit-design-004.ansi"
	if _, err := os.Stat(designPath); err != nil {
		t.Skip("design file not found")
	}

	buf, err := LoadAnsiBuffer(designPath, 80, 24)
	if err != nil {
		t.Fatalf("failed to load design file: %v", err)
	}
	if buf.Cols() < 80 || buf.Rows() < 24 {
		t.Fatalf("expected at least 80x24, got %dx%d", buf.Cols(), buf.Rows())
	}
}

// TestParseAnsiBufferMcJulia256 is a regression test for issue 163.
// mc-julia256.ansi uses cursor positioning (CSI H) and scroll regions (CSI r)
// which previously caused ParseAnsiBuffer to exit after reading only a few bytes.
func TestParseAnsiBufferMcJulia256(t *testing.T) {
	path := "docs/data/mc-julia256.ansi"
	if _, err := os.Stat(path); err != nil {
		t.Skip("mc-julia256.ansi not found, skipping")
	}

	buf, err := LoadAnsiBuffer(path, 80, 24)
	if err != nil {
		t.Fatalf("LoadAnsiBuffer(%q) error = %v", path, err)
	}

	// The file uses cursor positioning and fills at least 16 rows via CSI H sequences.
	// A correct parse must produce a buffer larger than the 13 newlines in the raw file.
	if buf.Rows() < 16 {
		t.Fatalf("expected at least 16 rows (cursor-positioned), got %d", buf.Rows())
	}

	// Collect all visible text from the buffer.
	var sb strings.Builder
	for row := 0; row < buf.Rows(); row++ {
		for col := 0; col < buf.Cols(); col++ {
			c := buf.Get(col, row)
			if c.Rune != 0 && c.Rune != ' ' {
				sb.WriteRune(c.Rune)
			}
		}
		sb.WriteByte('\n')
	}
	text := sb.String()

	for _, want := range []string{"projects/loom", "Name", "Size"} {
		if !strings.Contains(text, want) {
			t.Errorf("buffer text missing %q; got excerpt:\n%.500s", want, text)
		}
	}
}
