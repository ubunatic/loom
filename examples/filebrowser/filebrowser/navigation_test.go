package filebrowser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom"
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
	pane.ConsumeKey(loom.KeyEvent{Text: "beta"})
	if got := pane.List().Query(); got != "" {
		t.Fatalf("typing before / filtered to %q, want no filter", got)
	}
	pane.ConsumeKey(loom.KeyEvent{Text: "/"})
	pane.ConsumeKey(loom.KeyEvent{Text: "beta"})
	if got := pane.List().Query(); got != "beta" {
		t.Fatalf("filter = %q, want beta", got)
	}
	if selected != "beta" {
		t.Fatalf("selection callback name = %q, want beta", selected)
	}
	if pane.ConsumeKey(loom.KeyEvent{Key: "enter"}).Quit {
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
	pane.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if got := pane.Directory().Path; got != child {
		t.Fatalf("opened directory = %q, want %q", got, child)
	}
	if result := pane.ConsumeKey(loom.KeyEvent{Key: "esc"}); result.Quit || !result.Consumed {
		t.Fatalf("escape = quit:%v consumed:%v, want false,true", result.Quit, result.Consumed)
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
	if result := pane.ConsumeKey(loom.KeyEvent{Key: "esc"}); !result.Quit || !result.Consumed {
		t.Fatalf("root escape = quit:%v consumed:%v, want true,true", result.Quit, result.Consumed)
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
	pane.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 2, Y: 2})
	if got, ok := pane.Selected(); !ok || got.Name != "beta" {
		t.Fatalf("selection after local row 1 click = %+v, ok=%v; want beta", got, ok)
	}
}

func TestNavigationPaneDoubleClickActivatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "alpha.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	activated := ""
	pane, err := NewNavigationPane(dir, NavigationPaneOptions{
		OnActivate: func(entry loom.FileEntry) { activated = entry.Name },
	})
	if err != nil {
		t.Fatal(err)
	}
	pane.Draw(loom.NewCanvas(40, 8), loom.Rect{W: 40, H: 8})
	click := loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 2, Y: 1}
	pane.ConsumeMouse(click)
	if got, ok := pane.Selected(); !ok || got.Name != "alpha.txt" || activated != "" {
		t.Fatalf("first click selected %+v, ok=%v and activated %q; want alpha.txt selected only", got, ok, activated)
	}
	pane.ConsumeMouse(loom.MouseEvent{Action: loom.MouseRelease, Button: loom.MouseLeft, X: 2, Y: 1})
	time.Sleep(10 * time.Millisecond)
	pane.ConsumeMouse(click)
	if activated != "alpha.txt" {
		t.Fatalf("double-click activated %q, want alpha.txt", activated)
	}
}

func TestNavigationPaneDoubleClickEntersDirectory(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	opened := 0
	pane, err := NewNavigationPane(root, NavigationPaneOptions{
		OnOpen: func(loom.Directory) { opened++ },
	})
	if err != nil {
		t.Fatal(err)
	}
	pane.Draw(loom.NewCanvas(40, 8), loom.Rect{W: 40, H: 8})
	click := loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 2, Y: 1}
	pane.ConsumeMouse(click)
	if pane.Directory().Path != root || opened != 0 {
		t.Fatalf("first click entered %q with %d open callbacks; want root and none", pane.Directory().Path, opened)
	}
	pane.ConsumeMouse(loom.MouseEvent{Action: loom.MouseRelease, Button: loom.MouseLeft, X: 2, Y: 1})
	time.Sleep(10 * time.Millisecond)
	pane.ConsumeMouse(click)
	if pane.Directory().Path != child || opened != 1 {
		t.Fatalf("double-click entered %q with %d open callbacks; want %q and one", pane.Directory().Path, opened, child)
	}
}

func TestNavigationPaneSearchGate(t *testing.T) {
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
	if pane.ConsumeKey(loom.KeyEvent{Text: "a"}).Consumed {
		t.Fatal("a letter outside search was consumed; the app must get it")
	}
	if result := pane.ConsumeKey(loom.KeyEvent{Text: "/"}); !result.Consumed || !pane.Searching() {
		t.Fatalf("/ consumed=%v searching=%v, want both", result.Consumed, pane.Searching())
	}
	if result := pane.ConsumeKey(loom.KeyEvent{Text: "b"}); !result.Consumed || pane.List().Query() != "b" {
		t.Fatalf("typed b while searching: consumed=%v query=%q", result.Consumed, pane.List().Query())
	}
	pane.ConsumeKey(loom.KeyEvent{Key: "esc"})
	if pane.Searching() || pane.List().Query() != "" {
		t.Fatalf("esc left searching=%v query=%q, want search closed and cleared", pane.Searching(), pane.List().Query())
	}
	typing, err := NewNavigationPane(dir, NavigationPaneOptions{TypeToSearch: true})
	if err != nil {
		t.Fatal(err)
	}
	typing.ConsumeKey(loom.KeyEvent{Text: "b"})
	if got := typing.List().Query(); got != "b" {
		t.Fatalf("TypeToSearch query = %q, want b", got)
	}
}

func TestNavigationPaneSearchUIVisibility(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "alpha"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	pane, err := NewNavigationPane(dir, NavigationPaneOptions{})
	if err != nil {
		t.Fatal(err)
	}
	r := loom.Rect{W: 32, H: 4}
	draw := func() (*loom.Canvas, string) {
		canvas := loom.NewCanvas(r.W, r.H)
		pane.Draw(canvas, r)
		row := ""
		for x := 0; x < r.W; x++ {
			row += canvas.Get(x, r.H-1).Text
		}
		return canvas, row
	}
	canvas, row := draw()
	if got := strings.TrimSpace(row); got != "/ search" || canvas.CursorX != -1 {
		t.Fatalf("idle search row=%q cursor=(%d,%d), want / search and hidden cursor", got, canvas.CursorX, canvas.CursorY)
	}
	pane.ConsumeKey(loom.KeyEvent{Text: "/"})
	canvas, row = draw()
	if !strings.HasPrefix(row, "filter> ") || canvas.CursorX < 0 {
		t.Fatalf("active search row=%q cursor=(%d,%d), want editable filter and visible cursor", row, canvas.CursorX, canvas.CursorY)
	}
	pane.ConsumeKey(loom.KeyEvent{Text: "a"})
	pane.ConsumeKey(loom.KeyEvent{Key: "enter"})
	canvas, row = draw()
	if got := strings.TrimSpace(row); got != "/ search" || canvas.CursorX != -1 || pane.List().Query() != "a" {
		t.Fatalf("applied search row=%q cursor=(%d,%d) query=%q, want idle hint, hidden cursor, retained query", got, canvas.CursorX, canvas.CursorY, pane.List().Query())
	}
	pane.ConsumeKey(loom.KeyEvent{Key: "esc"})
	canvas, row = draw()
	if got := strings.TrimSpace(row); got != "/ search" || canvas.CursorX != -1 || pane.List().Query() != "" {
		t.Fatalf("cleared search row=%q cursor=(%d,%d) query=%q, want idle hint, hidden cursor, empty query", got, canvas.CursorX, canvas.CursorY, pane.List().Query())
	}
	typing, err := NewNavigationPane(dir, NavigationPaneOptions{TypeToSearch: true})
	if err != nil {
		t.Fatal(err)
	}
	canvas = loom.NewCanvas(r.W, r.H)
	typing.Draw(canvas, r)
	if canvas.CursorX < 0 || !strings.HasPrefix(canvasRow(canvas, r.H-1, r.W), "filter> ") {
		t.Fatalf("TypeToSearch row=%q cursor=(%d,%d), want visible input and cursor", canvasRow(canvas, r.H-1, r.W), canvas.CursorX, canvas.CursorY)
	}
}

func canvasRow(canvas *loom.Canvas, y, width int) string {
	row := ""
	for x := 0; x < width; x++ {
		row += canvas.Get(x, y).Text
	}
	return row
}
