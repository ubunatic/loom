package gallery

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/ptytest"
)

func TestAllTableMatchesStandaloneAndRoutesCells(t *testing.T) {
	g := newAllDemo()
	table := g.Children[18].(*loom.Table)
	standalone := demos["Table"]().(*loom.Table)
	if !table.CellCursor || table.CellCursor != standalone.CellCursor || table.FrozenCols != standalone.FrozenCols || table.Controls != standalone.Controls || !reflect.DeepEqual(table.Columns, standalone.Columns) || !reflect.DeepEqual(table.Rows, standalone.Rows) {
		t.Fatal("All table differs from standalone table")
	}
	c := loom.NewCanvas(150, 64)
	allocation := loom.Rect{X: 3, Y: 2, W: 144, H: 60}
	g.Draw(c, allocation)
	r := g.ChildRect(18)
	row, col := -1, -1
	table.OnCellSelect = func(y, x int) { row, col = y, x }
	// Click the second data column using the declared display width.
	res := g.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: r.X - allocation.X + table.Columns[0].Width + 2, Y: r.Y - allocation.Y + 1})
	if !res.Consumed || res.Quit || g.Focus() != 18 || row != 0 || col != 1 {
		t.Fatalf("click: %+v focus=%d cell=(%d,%d)", res, g.Focus(), row, col)
	}
	for _, tc := range []struct {
		key        string
		focus, col int
	}{
		{"right", 18, 2}, {"right", 19, 2}, {"shift-left", 18, 2},
		{"left", 18, 1}, {"left", 18, 0}, {"left", 17, 0},
	} {
		res := g.ConsumeKey(loom.KeyEvent{Key: tc.key})
		if !res.Consumed || g.Focus() != tc.focus || col != tc.col {
			t.Fatalf("%s: %+v focus=%d col=%d, want %d/%d", tc.key, res, g.Focus(), col, tc.focus, tc.col)
		}
	}
}

func TestAllDialogAndTableNavigationPTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	if out, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	s := ptytest.Start(t, 150, 64, bin, "widgets", "--show", "--theme", "plain", "-W", "144", "-H", "60", "All")
	s.WaitFor("Keep edits?", 5*time.Second)
	point := func(text string) (int, int) {
		t.Helper()
		want := []rune(text)
		for y, line := range s.Screen() {
			runes := []rune(line)
			for x := 0; x+len(want) <= len(runes); x++ {
				if string(runes[x:x+len(want)]) == text {
					return x, y
				}
			}
		}
		t.Fatalf("%q absent", text)
		return 0, 0
	}
	click := func(x, y int) { s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", x+1, y+1, x+1, y+1))) }
	x, y := point("Save?")
	click(x, y) // Focus the dialog via its border, leaving it open.
	s.Send("\x1b[C")
	s.WaitFor("▶ Yes", 5*time.Second)
	s.Send("\x1b[D")
	s.WaitFor("▶ No", 5*time.Second)
	// Confirm mouse activation closes only the child; Enter reopens its demo.
	x, y = point("▶ No")
	click(x+2, y)
	deadline := time.Now().Add(5 * time.Second)
	for strings.Contains(strings.Join(s.Screen(), "\n"), "Keep edits?") && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if strings.Contains(strings.Join(s.Screen(), "\n"), "Keep edits?") {
		t.Fatal("mouse did not close dialog")
	}
	s.Send("\r")
	s.WaitFor("Keep edits?", 5*time.Second)

	x, y = point("Compile")
	sx, sy := point("done")
	click(x, y)
	wait := func(check func() bool) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if check() {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatal("table cell highlight did not move")
	}
	wait(func() bool { return s.Cell(x, y).Style != s.Cell(sx, sy).Style })
	first, normal := s.Cell(x, y).Style, s.Cell(sx, sy).Style
	s.Send("\x1b[C")
	wait(func() bool { return s.Cell(x, y).Style == normal && s.Cell(sx, sy).Style == first })
	s.Send("\x1b[D")
	wait(func() bool { return s.Cell(x, y).Style == first && s.Cell(sx, sy).Style == normal })
}

func TestAllSharedControlsMatchStandalone(t *testing.T) {
	g := newAllDemo()
	if !g.Children[9].(*loom.Stopwatch).Controls || !g.Children[10].(*loom.Timer).Controls {
		t.Fatal("All time widgets omit standalone controls")
	}
	if got, want := loom.Render(g.Children[15], 60, 5), loom.Render(demos["KeyHelp"](), 60, 5); !reflect.DeepEqual(got, want) {
		t.Fatal("All KeyHelp differs from standalone")
	}
}
