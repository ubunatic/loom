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
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	bin := filepath.Join(t.TempDir(), "loom")
	if out, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	s := ptytest.Start(t, 150, 64, bin, "widgets", "--show", "--theme", "plain", "-W", "144", "-H", "60", "All")
	timeout := 5 * time.Second * ptytest.RaceFactor
	contains := func(rows []string, want string) bool {
		return strings.Contains(strings.Join(rows, "\n"), want)
	}
	waitState := func(want string) {
		t.Helper()
		s.WaitForState(timeout, func(rows []string) bool { return contains(rows, want) })
	}
	waitUpdate := func(predicate func([]string) bool) {
		t.Helper()
		s.WaitForUpdateState(timeout, predicate)
	}
	waitState("Keep edits?")
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
	waitState("Keep edits?")
	click(x, y) // Focus the dialog via its border, leaving it open.
	waitUpdate(func(rows []string) bool { return contains(rows, "Keep edits?") })
	waitState("Keep edits?")
	s.Send("\x1b[C")
	waitUpdate(func(rows []string) bool { return contains(rows, "▶ Yes") })
	waitState("▶ Yes")
	s.Send("\x1b[D")
	waitUpdate(func(rows []string) bool { return contains(rows, "▶ No") })
	// Confirm mouse activation closes only the child; Enter reopens its demo.
	x, y = point("▶ No")
	waitState("▶ No")
	click(x+2, y)
	waitUpdate(func(rows []string) bool { return !contains(rows, "Keep edits?") })
	waitState("Compile")
	s.Send("\r")
	waitUpdate(func(rows []string) bool { return contains(rows, "Keep edits?") })

	x, y = point("Compile")
	sx, sy := point("done")
	// The plain theme draws the cell cursor bold; wait for each cursor move
	// after the click lands and each navigation input is processed.
	waitTableCursor := func(cursorOnCompile bool, afterInput bool) {
		t.Helper()
		check := func([]string) bool {
			return s.Cell(x, y).Style.Bold == cursorOnCompile && s.Cell(sx, sy).Style.Bold == !cursorOnCompile
		}
		if afterInput {
			waitUpdate(check)
		} else {
			s.WaitForState(timeout, check)
		}
	}
	waitState("Compile")
	waitState("done")
	click(x, y)
	waitTableCursor(true, true)
	waitTableCursor(true, false)
	s.Send("\x1b[C")
	waitTableCursor(false, true)
	waitTableCursor(false, false)
	s.Send("\x1b[D")
	waitTableCursor(true, true)
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
