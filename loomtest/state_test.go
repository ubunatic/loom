// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomtest

import (
	"slices"
	"testing"
)

func TestAlternateScreenRestoresGridCursorAndPen(t *testing.T) {
	v := NewVT(8, 3)
	v.Write([]byte("\x1b[1;31mhome\x1b[2;3H"))
	original := v.Cells()
	pen := v.Pen()
	v.Write([]byte("\x1b[?1049h"))
	if !v.AltScreen() || !slices.Equal(v.Screen(), []string{"", "", ""}) {
		t.Fatalf("alternate screen not blank: %q, alt=%v", v.Screen(), v.AltScreen())
	}
	v.Write([]byte("\x1b[32mapp\x1b[?1049h")) // Repeated enable must preserve the active screen.
	if v.Screen()[0] != "app" {
		t.Fatalf("repeated enable cleared screen: %q", v.Screen())
	}
	v.Write([]byte("\x1b[?1049l\x1b[?1049l"))
	if v.AltScreen() || v.Pen() != pen {
		t.Fatalf("restore: alt=%v, pen=%v", v.AltScreen(), v.Pen())
	}
	if x, y := v.Cursor(); x != 2 || y != 1 {
		t.Fatalf("restored cursor = (%d,%d)", x, y)
	}
	for y := range original {
		if !slices.Equal(v.Cells()[y], original[y]) {
			t.Fatalf("restored row %d = %v, want %v", y, v.Cells()[y], original[y])
		}
	}
	v.Write([]byte("!"))
	if c := v.Cell(2, 1); c.Rune != '!' || c.Style != pen {
		t.Fatalf("restored writing = %v", c)
	}
}

func TestResizeBothScreens(t *testing.T) {
	v := NewVT(8, 3)
	v.Write([]byte("home\x1b[3;8H\x1b[?1049happ"))
	v.Resize(5, 2)
	if v.Screen()[0] != "app" {
		t.Fatalf("resized alternate screen = %q", v.Screen())
	}
	v.Write([]byte("\x1b[?1049l"))
	if !slices.Equal(v.Screen(), []string{"home", ""}) {
		t.Fatalf("resized main = %q", v.Screen())
	}
	if x, y := v.Cursor(); x != 4 || y != 1 {
		t.Fatalf("resized cursor = (%d,%d)", x, y)
	}
	v.Resize(10, 4)
	if v.Cell(9, 3).Rune != ' ' {
		t.Fatal("new cells must be blank")
	}
}

func TestSplitANSIAndUTF8Stream(t *testing.T) {
	const stream = "\x1b[?1049h\x1b[2;3H\x1b[1;38;2;10;20;30m界!\x1b[?7;25l\x1b[?2026l"
	for split := 0; split <= len(stream); split++ {
		v := NewVT(8, 3)
		v.Write([]byte(stream[:split]))
		v.Write([]byte(stream[split:]))
		want := Style{Bold: true, FG: ColorRGB(10, 20, 30)}
		if c := v.Cell(2, 1); c.Rune != '界' || c.Style != want {
			t.Fatalf("split %d: wide cell = %v", split, c)
		}
		if v.Cell(3, 1).Rune != 0 || v.Cell(4, 1).Rune != '!' {
			t.Fatalf("split %d: grid = %v", split, v.Cells())
		}
		if x, y := v.Cursor(); x != 5 || y != 1 {
			t.Fatalf("split %d: cursor=(%d,%d)", split, x, y)
		}
		if !v.AltScreen() || v.CursorVisible || len(v.Frames) != 1 {
			t.Fatalf("split %d: incorrect terminal modes", split)
		}
	}
}

func TestControlStringsNeverAppearOnScreen(t *testing.T) {
	for _, control := range []string{
		"\x1b]0;title\a", "\x1b]8;;https://example.invalid\x1b\\", "\x1bPpayload\x1b\\",
	} {
		for split := 0; split <= len(control); split++ {
			v := NewVT(8, 1)
			v.Write([]byte("a" + control[:split]))
			v.Write([]byte(control[split:] + "b"))
			if got := v.Text(); got != "ab" {
				t.Fatalf("control %q split %d: %q", control, split, got)
			}
		}
	}
}

func TestCursorControlsAndDefaultParameters(t *testing.T) {
	v := NewVT(12, 4)
	v.Write([]byte("\x1b[;5HX\x1b[2d\x1b[3G\x1b[31m\x1b7\x1b[3;9H\x1b[32m\x1b8!"))
	if v.Cell(4, 0).Rune != 'X' || v.Cell(2, 1).Rune != '!' || v.Cell(2, 1).Style.FG != ColorIndex(1) {
		t.Fatalf("saved cursor/pen: %q", v.Screen())
	}
	v.Write([]byte("\r012\bX\tZ"))
	if got := v.Screen()[1]; got != "01X     Z" {
		t.Fatalf("CR/BS/tab = %q", got)
	}
}

func TestSnapshotsAreIndependent(t *testing.T) {
	v := NewVT(3, 1)
	v.Write([]byte("a\x1b[?2026l"))
	grid, frames := v.Cells(), v.FrameCells()
	grid[0][0].Rune = 'x'
	frames[0][0][0].Rune = 'y'
	v.Write([]byte("\rZ"))
	if v.Cell(0, 0).Rune != 'Z' || v.CellFrames[0][0][0].Rune != 'a' {
		t.Fatal("snapshots alias live or historical cells")
	}
}

func TestNonPositiveDimensions(t *testing.T) {
	v := NewVT(0, -1)
	v.Write([]byte("x"))
	v.Resize(-2, 0)
	if v.Cols != 1 || v.Rows != 1 || v.Text() != "x" {
		t.Fatalf("minimum grid = %dx%d %q", v.Cols, v.Rows, v.Text())
	}
}
