package filebrowser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestFilebrowser093Evidence(t *testing.T) {
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate issue 093 evidence")
	}
	dir := t.TempDir()
	for _, name := range []string{"alpha.txt", "beta.txt", "a-very-long-file-name.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	root := findFilebrowserRepoRoot(t)
	outDir := filepath.Join(root, "docs", "progress", "093")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}

	b, err := newBrowser(dir, "plain", loom.Theme("plain"))
	if err != nil {
		t.Fatal(err)
	}
	listRect := b.frame.Layout(80, 24)[0]
	if b.list.FilteredSel() != 0 {
		t.Fatalf("initial selection = %d, want row 0", b.list.FilteredSel())
	}
	b.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: listRect.X + listRect.W - 4, Y: listRect.Y + 3})
	if b.list.FilteredSel() != 0 {
		t.Fatalf("whitespace click changed selection to %d, want row 0", b.list.FilteredSel())
	}
	writeBrowser093Frame(t, outDir, "M2-click-after-whitespace.ansi", b, false, "M5 whitespace click on row 2: selection stays on row 0")
	b.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: listRect.X + 4, Y: listRect.Y + 3})
	if b.list.FilteredSel() != 2 {
		t.Fatalf("text click selected row %d, want row 2", b.list.FilteredSel())
	}
	writeBrowser093Frame(t, outDir, "M2-click-after-text.ansi", b, false, "M5 text click on row 2: selection moves to row 2")
}

func writeBrowser093Frame(t *testing.T, dir, name string, b *browser, overlay bool, label string) {
	t.Helper()
	const cols, rows = 80, 24
	c := loom.NewCanvas(cols, rows)
	b.Draw(c, c.Bounds())
	if overlay {
		listRect := b.frame.Layout(cols, rows)[0]
		for y := listRect.Y + 1; y < listRect.Y+listRect.H-1; y++ {
			start, end := -1, -1
			for x := listRect.X + 1; x < listRect.X+listRect.W-1; x++ {
				if strings.TrimSpace(c.Get(x, y).Text) != "" {
					if start < 0 {
						start = x
					}
					end = x
				}
			}
			if start >= 0 {
				for x := start; x <= end; x++ {
					cell := c.Get(x, y)
					cell.Style.BG = loom.ColorIndex(22)
					c.Set(x, y, cell)
				}
			}
		}
	}
	var out strings.Builder
	out.WriteString(label + "\n")
	if err := loom.RenderTo(&out, &canvasWidget{canvas: c}, cols, rows); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(out.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}
