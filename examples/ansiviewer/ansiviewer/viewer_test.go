package ansiviewer

import (
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestBrowserListsTextAndANSI(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "plain.txt"), []byte("hello\nworld\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "color.ansi"), []byte("\x1b[31mred\x1b[0m\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.files) != 2 || b.files[0].Name() != "color.ansi" {
		t.Fatalf("files = %v", b.files)
	}
	b.selectFile(0)
	c := loom.NewCanvas(30, 4)
	b.Draw(c, c.Bounds())
	if got := c.Row(0); got == "" {
		t.Fatal("empty rendered row")
	}
}

func TestANSIWriteClipsToBounds(t *testing.T) {
	c := loom.NewCanvas(5, 1)
	writeANSI(c, loom.Rect{X: 0, Y: 0, W: 5, H: 1}, "abcdef\x1b[31m!")
	if got := c.Row(0); got == "" {
		t.Fatal("empty row")
	}
}
