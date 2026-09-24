package filebrowser

import (
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestNavigationPaneReadsAndSelectsEntries(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "alpha.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "beta"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	selected := ""
	pane, err := NewNavigationPane(dir, NavigationPaneOptions{
		OnSelection: func(entry loom.FileEntry) { selected = entry.Name },
		OnActivate:  func(entry loom.FileEntry) { selected = "activated:" + entry.Name },
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := pane.Directory().Path; got != dir {
		t.Fatalf("directory = %q, want %q", got, dir)
	}
	if got, ok := pane.Selected(); !ok || got.Name != "alpha.txt" {
		t.Fatalf("initial selection = %+v, ok=%v; want alpha.txt", got, ok)
	}
	pane.HandleKey(loom.KeyEvent{Text: "beta"})
	if got := pane.List().Query(); got != "beta" {
		t.Fatalf("filter = %q, want beta", got)
	}
	if selected != "beta" {
		t.Fatalf("selection callback name = %q, want beta", selected)
	}
	if pane.HandleKey(loom.KeyEvent{Key: "enter"}) {
		t.Fatal("activating an entry should not quit")
	}
	if selected != "activated:beta" {
		t.Fatalf("activation callback name = %q, want activated:beta", selected)
	}
}

func TestNavigationPaneOpensDirectoryAndRestoresParentSelection(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	pane, err := NewNavigationPane(root, NavigationPaneOptions{})
	if err != nil {
		t.Fatal(err)
	}
	pane.HandleKey(loom.KeyEvent{Key: "enter"})
	if got := pane.Directory().Path; got != child {
		t.Fatalf("opened directory = %q, want %q", got, child)
	}
	if quit, consumed := pane.ConsumeKey(loom.KeyEvent{Key: "esc"}); quit || !consumed {
		t.Fatalf("escape = quit:%v consumed:%v, want false,true", quit, consumed)
	}
	if got := pane.Directory().Path; got != root {
		t.Fatalf("returned directory = %q, want %q", got, root)
	}
	if got, ok := pane.Selected(); !ok || got.Name != "child" {
		t.Fatalf("restored selection = %+v, ok=%v; want child", got, ok)
	}
}

func TestNavigationPaneEscapeAtRootRequestsQuit(t *testing.T) {
	root := t.TempDir()
	pane, err := NewNavigationPane(root, NavigationPaneOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if quit, consumed := pane.ConsumeKey(loom.KeyEvent{Key: "esc"}); !quit || !consumed {
		t.Fatalf("root escape = quit:%v consumed:%v, want true,true", quit, consumed)
	}
}

func TestNavigationPaneMouseSelectionUsesChildLocalCoordinates(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"alpha", "beta"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	pane, err := NewNavigationPane(dir, NavigationPaneOptions{})
	if err != nil {
		t.Fatal(err)
	}
	canvas := loom.NewCanvas(40, 10)
	pane.Draw(canvas, loom.Rect{X: 7, Y: 5, W: 30, H: 6})
	pane.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 2, Y: 2})
	if got, ok := pane.Selected(); !ok || got.Name != "beta" {
		t.Fatalf("selection after local row 1 click = %+v, ok=%v; want beta", got, ok)
	}
}
