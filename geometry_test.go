// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestGeometryArtifacts(t *testing.T) {
	root, _, err := BuildWidget(strings.NewReader(shellFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	f := root.(*Frame)
	for i := range f.Boxes {
		f.Boxes[i].Child = geometryChild{}
	}
	for _, tc := range []struct {
		name string
		w, h int
	}{{"wide", 64, 9}, {"slim", 40, 18}} {
		rows := Render(f, tc.w, tc.h)
		replay := ScreenshotScript(rows, fmt.Sprintf("geometry %s: %dx%d; ANSI replay, not raster", tc.name, tc.w, tc.h))
		artifact := struct {
			Name          string
			Width, Height int
			Rows          []string
			Replay        string
		}{tc.name, tc.w, tc.h, rows, replay}
		if os.Getenv("LOOM_GEOMETRY_EXPORT") == "1" {
			data, err := json.Marshal(artifact)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("GEOMETRY_EXPORT " + string(data))
			continue
		}
		data, err := os.ReadFile("testdata/geometry/" + tc.name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var saved struct {
			Width, Height int
			Rows          []string
		}
		if err := json.Unmarshal(data, &saved); err != nil {
			t.Fatal(err)
		}
		if saved.Width != tc.w || saved.Height != tc.h || strings.Join(saved.Rows, "\n") != strings.Join(rows, "\n") {
			t.Fatalf("%s golden changed", tc.name)
		}
		script, err := os.ReadFile("testdata/geometry/" + tc.name + ".sh")
		if err != nil {
			t.Fatal(err)
		}
		if string(script) != replay {
			t.Fatalf("%s replay changed", tc.name)
		}
	}
}

// oracleCells is deliberately corpus-bounded. It interprets emitted SGR and
// explicit, independently specified glyph widths, never Loom width helpers.
func oracleCells(row string) ([]string, error) {
	var cells []string
	runes := []rune(row)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == 27 {
			i++
			if i >= len(runes) || runes[i] != '[' {
				return nil, fmt.Errorf("non-SGR escape")
			}
			for i++; i < len(runes) && (runes[i] >= '0' && runes[i] <= '9' || runes[i] == ';'); i++ {
			}
			if i >= len(runes) || runes[i] != 'm' {
				return nil, fmt.Errorf("non-SGR instruction")
			}
			continue
		}
		if r == '\u0301' {
			if len(cells) == 0 {
				return nil, fmt.Errorf("unattached mark")
			}
			cells[len(cells)-1] += string(r)
			continue
		}
		switch {
		case r >= ' ' && r <= '~', strings.ContainsRune("┌┐└┘─│█░", r), r >= '\u2800' && r <= '\u28ff':
			cells = append(cells, string(r))
		case r == '界' || r == '中' || r == '🔍':
			cells = append(cells, string(r), "")
		default:
			return nil, fmt.Errorf("glyph outside oracle corpus: %U", r)
		}
	}
	return cells, nil
}

func checkGeometry(rows []string, width int, boxes []Rect) error {
	grid := make([][]string, len(rows))
	for y, row := range rows {
		var err error
		grid[y], err = oracleCells(row)
		if err != nil {
			return err
		}
		if len(grid[y]) != width {
			return fmt.Errorf("row %d width=%d want=%d", y, len(grid[y]), width)
		}
	}
	for _, r := range boxes {
		if r.W == 0 || r.H == 0 {
			continue
		}
		for y := r.Y; y < r.Y+r.H; y++ {
			left, right := "│", "│"
			if y == r.Y {
				left, right = "┌", "┐"
			}
			if y == r.Y+r.H-1 {
				left, right = "└", "┘"
			}
			if grid[y][r.X] != left || grid[y][r.X+r.W-1] != right {
				return fmt.Errorf("broken border at row %d", y)
			}
			if y == r.Y+r.H-1 {
				for x := r.X + 1; x < r.X+r.W-1; x++ {
					if grid[y][x] != "─" {
						return fmt.Errorf("broken bottom border at %d,%d", x, y)
					}
				}
			}
			if y > r.Y && y < r.Y+r.H-1 && r.W >= 4 && r.H >= 4 {
				for x := r.X + 1; x < r.X+r.W-1; x++ {
					if (x == r.X+1 || x == r.X+r.W-2 || y == r.Y+1 || y == r.Y+r.H-2) && grid[y][x] != " " {
						return fmt.Errorf("broken padding at %d,%d", x, y)
					}
				}
			}
		}
	}
	return nil
}

type geometryChild struct{}

func (geometryChild) HandleKey(KeyEvent) bool     { return false }
func (geometryChild) HandleMouse(MouseEvent) bool { return false }
func (geometryChild) Draw(c *Canvas, r Rect) {
	c.Fill(Rect{-20, -20, 100, 100}, Cell{Text: "."})
	c.Write(-1, 0, "界ASCII e\u0301 中 ⣿⣀█░", Style{Bold: true, FG: ColorIndex(2)})
	c.Write(0, 1, "\x1b[31mstyled\x1b[0m", Style{Underline: true})
	c.Write(c.Cols()-1, 1, "界", Style{})
	c.Set(0, 2, Cell{Text: "\x1b[2Joversized\ncell"})
	c.Write(0, c.Rows(), "outside", Style{})
}

func TestGeometryFinalOutput(t *testing.T) {
	root, _, err := BuildWidget(strings.NewReader(shellFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	f := root.(*Frame)
	for i := range f.Boxes {
		f.Boxes[i].Child = geometryChild{}
	}
	for _, tc := range []struct {
		w, h  int
		boxes []Rect
	}{
		{64, 9, []Rect{{0, 1, 31, 7}, {33, 1, 31, 7}}},
		{40, 18, []Rect{{0, 1, 31, 7}, {0, 10, 31, 7}}},
		{8, 7, []Rect{{0, 1, 8, 5}}},
		{2, 4, []Rect{{0, 1, 2, 2}}},
		{1, 2, nil},
	} {
		rows := Render(f, tc.w, tc.h)
		if err := checkGeometry(rows, tc.w, tc.boxes); err != nil {
			t.Fatalf("%dx%d: %v", tc.w, tc.h, err)
		}
	}
	rows := Render(f, 64, 9)
	for _, broken := range []string{
		strings.Replace(rows[2], "│", "X", 1),
		"X" + rows[2],
		"\x1b[2J" + rows[2],
	} {
		mutated := append([]string(nil), rows...)
		mutated[2] = broken
		if checkGeometry(mutated, 64, []Rect{{0, 1, 31, 7}, {33, 1, 31, 7}}) == nil {
			t.Fatal("gate accepted corrupted border/width/control")
		}
	}
	for i := 0; i < 8; i++ {
		f.HandleKey(KeyEvent{Text: "u"})
		f.HandleKey(KeyEvent{Text: "l"})
		for _, w := range []int{64, 40, 64} {
			var boxes []Rect
			if i%2 == 1 {
				boxes = []Rect{{0, 1, 31, 7}, {33, 1, 31, 7}}
				if w < 64 {
					boxes[1] = Rect{0, 10, 31, 7}
				}
			}
			if err := checkGeometry(Render(f, w, f.HeightForWidth(w)), w, boxes); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestCanvasUnicodeBoundaries(t *testing.T) {
	for _, tc := range []struct {
		text string
		x    int
		want string
	}{
		{"e\u0301界⣿", 0, "e\u0301界⣿ "},
		{"界", 4, "     "},
		{"界A", -1, " A   "},
		{"\u0301A", 0, "A    "},
		{"\x1b[31mA\x1b[0m\x1b]0;title\aB", 0, "AB   "},
	} {
		c := NewCanvas(5, 1)
		c.Write(tc.x, 0, tc.text, Style{})
		got := strings.TrimSuffix(c.Row(0), "\x1b[0m")
		if got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.text, got, tc.want)
		}
		if cells, err := oracleCells(c.Row(0)); err != nil || len(cells) != 5 {
			t.Fatalf("%v %v", cells, err)
		}
	}
	c := NewCanvas(5, 1)
	c.Write(0, 0, "界中", Style{})
	c.Set(1, 0, Cell{Text: "A"})
	c.Set(2, 0, Cell{Text: "B"})
	if got := strings.TrimSuffix(c.Row(0), "\x1b[0m"); got != " AB  " {
		t.Fatal(got)
	}
}
