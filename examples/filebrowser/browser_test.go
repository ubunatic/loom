package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestBrowserSelectionAndNavigation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir)
	if err != nil {
		t.Fatal(err)
	}
	b.HandleKey(loom.KeyEvent{Key: "down"}) // parent -> a.txt
	if item, ok := b.list.Selected(); !ok || item.Name != "a.txt" {
		t.Fatalf("selected item = %+v, ok=%v", item, ok)
	}
	if got := strings.Join(b.details.Lines, "\n"); !strings.Contains(got, "Size: 5 bytes") || !strings.Contains(got, "Type: file") {
		t.Fatalf("file metadata missing: %s", got)
	}
	b.HandleKey(loom.KeyEvent{Key: "down"}) // sub
	b.HandleKey(loom.KeyEvent{Key: "enter"})
	if b.dir != filepath.Join(dir, "sub") || b.frame.Boxes[0].Child != b.list {
		t.Fatalf("directory navigation failed: %q", b.dir)
	}
	if item, ok := b.list.Selected(); !ok || item.Name != ".." {
		t.Fatalf("parent entry missing: %+v, ok=%v", item, ok)
	}
	b.HandleKey(loom.KeyEvent{Key: "enter"})
	if b.dir != dir {
		t.Fatalf("parent navigation returned %q, want %q", b.dir, dir)
	}
}

func TestBrowserDetailScrollingAndFilterKey(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "q.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir)
	if err != nil {
		t.Fatal(err)
	}
	if b.HandleKey(loom.KeyEvent{Text: "q"}) {
		t.Fatal("typing q into the file filter quit the app")
	}
	if item, ok := b.list.Selected(); !ok || item.Name != "q.txt" {
		t.Fatalf("filter did not select q.txt: %+v, ok=%v", item, ok)
	}
	b.details.Draw(loom.NewCanvas(40, 2), loom.Rect{W: 40, H: 2})
	b.HandleKey(loom.KeyEvent{Key: "tab"})
	b.HandleKey(loom.KeyEvent{Key: "pgdown"})
	if b.details.Scroll != 2 || !b.details.Focused() {
		t.Fatalf("detail pane did not scroll with focus: scroll=%d focused=%v", b.details.Scroll, b.details.Focused())
	}
	if !b.HandleKey(loom.KeyEvent{Key: "ctrl-q"}) {
		t.Fatal("Ctrl-Q did not quit")
	}
}

func TestBrowserShowsSideBySidePanes(t *testing.T) {
	b, err := newBrowser(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	pane := &loom.Pane{MaxCols: loom.DefaultMaxCols}
	configurePane(pane)
	cols := 80
	if pane.MaxCols > 0 && cols > pane.MaxCols {
		cols = pane.MaxCols
	}
	rects := b.frame.Layout(cols, 20)
	if rects[0].W < 20 || rects[1].W < 25 || rects[1].X <= rects[0].X+rects[0].W {
		t.Fatalf("expected two side-by-side panes at terminal width 80 (canvas %d), got %+v", cols, rects)
	}
	if b.frame.Boxes[0].Border.Vertical == "" || b.frame.Boxes[1].Border.Vertical == "" {
		t.Fatal("pane borders are invisible")
	}
}
