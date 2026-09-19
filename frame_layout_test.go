// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"reflect"
	"strings"
	"testing"
)

func TestResponsiveFrameLayout(t *testing.T) {
	w, _, err := BuildWidget(strings.NewReader(shellFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	f := w.(*Frame)
	for _, tc := range []struct {
		name                     string
		width, height, preferred int
		want                     []Rect
	}{
		{"below", 63, 18, 18, []Rect{{0, 1, 31, 7}, {0, 10, 31, 7}}},
		{"at", 64, 9, 9, []Rect{{0, 1, 31, 7}, {33, 1, 31, 7}}},
		{"above", 65, 9, 9, []Rect{{0, 1, 31, 7}, {33, 1, 31, 7}}},
		{"wide", 100, 9, 9, []Rect{{0, 1, 31, 7}, {33, 1, 31, 7}}},
		{"slim", 20, 18, 18, []Rect{{0, 1, 20, 7}, {0, 10, 20, 7}}},
		{"short", 63, 6, 18, []Rect{{0, 1, 31, 4}, {}}},
		{"tiny width", 1, 18, 18, []Rect{{}, {}}},
		{"tiny height", 64, 3, 9, []Rect{{}, {}}},
		{"negative", -1, -1, 18, []Rect{{}, {}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := f.Layout(tc.width, tc.height); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %+v want %+v", got, tc.want)
			}
			if got := f.HeightForWidth(tc.width); got != tc.preferred {
				t.Fatalf("height %d", got)
			}
		})
	}
}

func TestDynamicFrameLayoutAllocatesPreferredAndRemainingWidth(t *testing.T) {
	f := Frame{Gap: 1, Boxes: []Box{
		{ID: "a", Width: 4, Height: 4, Dynamic: true, MinWidth: 2, MaxWidth: 6},
		{ID: "b", Width: 3, Height: 4, Dynamic: true, MinWidth: 2},
	}}
	got := f.Layout(12, 8)
	want := []Rect{{X: 0, Y: 1, W: 6, H: 4}, {X: 7, Y: 1, W: 5, H: 4}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dynamic layout=%v, want %v", got, want)
	}
}

func TestDynamicFrameLayoutStacksAtBreakpoint(t *testing.T) {
	f := Frame{Gap: 1, Breakpoint: 80, Boxes: []Box{
		{ID: "a", Width: 4, Height: 3, Dynamic: true, MinHeight: 2, MaxHeight: 4},
		{ID: "b", Width: 4, Height: 3, Dynamic: true, MinHeight: 2, MaxHeight: 4},
	}}
	got := f.Layout(40, 10)
	want := []Rect{{X: 0, Y: 1, W: 40, H: 4}, {X: 0, Y: 6, W: 40, H: 3}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("stacked dynamic layout=%v, want %v", got, want)
	}
}

func TestDynamicFrameLayoutReflowsShortStack(t *testing.T) {
	f := Frame{Gap: 1, Breakpoint: 65, Boxes: []Box{
		{ID: "files", Width: 18, Height: 18, Dynamic: true, MinWidth: 20},
		{ID: "metadata", Width: 25, Height: 18, Dynamic: true, MinWidth: 25},
	}}
	got := f.Layout(64, 20)
	for i, rect := range got {
		if rect.W == 0 || rect.H == 0 {
			continue
		}
		if rect.X < 0 || rect.Y < 1 || rect.X+rect.W > 64 || rect.Y+rect.H > 19 {
			t.Fatalf("box %d escapes short frame: %+v", i, rect)
		}
		if rect.H < 2 {
			t.Fatalf("box %d has incomplete border: %+v", i, rect)
		}
	}
	if got[0].H == 18 || got[1].H == 18 || got[1].Y+got[1].H > 19 {
		t.Fatalf("short stack retained preferred heights: %v", got)
	}
	if got[0].H < 8 || got[1].H < 8 {
		t.Fatalf("short stack collapsed one pane: %v", got)
	}
	if got[0].W != 64 || got[1].W != 64 {
		t.Fatalf("short stack did not stretch panes: %v", got)
	}
}

func TestDynamicBoxCanDeriveSizeFromRows(t *testing.T) {
	input := `app:
  height: 8
  max_width: 40
view:
  frame:
    boxes:
      - id: data
        dynamic: true
        rows:
          columns:
            - width: 5
          values:
            - [hello]
`
	w, _, err := BuildWidget(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	f := w.(*Frame)
	if f.Boxes[0].Width < 7 || f.Boxes[0].Height < 3 {
		t.Fatalf("derived box size=%dx%d, want content plus chrome", f.Boxes[0].Width, f.Boxes[0].Height)
	}
}

func TestDynamicFrameRendersWithinMeasuredRectangles(t *testing.T) {
	f := Frame{Gap: 1, Boxes: []Box{
		{ID: "a", Width: 4, Height: 4, Dynamic: true, MinWidth: 2, MaxWidth: 6, Border: BoxBorder{TopLeft: "+", TopRight: "+", BottomLeft: "+", BottomRight: "+", Horizontal: "-", Vertical: "|"}},
		{ID: "b", Width: 3, Height: 4, Dynamic: true, MinWidth: 2},
	}}
	f.Boxes[1].Border = f.Boxes[0].Border
	c := NewCanvas(12, 8)
	f.Draw(c, Rect{W: 12, H: 8})
	for _, rect := range f.Layout(12, 8) {
		if rect.W < 2 || rect.H < 2 {
			continue
		}
		if StringWidth(c.Get(rect.X, rect.Y).Text) != 1 || StringWidth(c.Get(rect.X+rect.W-1, rect.Y).Text) != 1 {
			t.Fatalf("missing top border at %+v", rect)
		}
		if rect.X+rect.W > c.Cols() || rect.Y+rect.H > c.Rows() {
			t.Fatalf("rectangle %+v escapes canvas", rect)
		}
	}
}

func TestDynamicFrameReflowsAfterVisibilityChange(t *testing.T) {
	f := Frame{Gap: 1, Boxes: []Box{
		{ID: "a", Width: 4, Height: 4, Dynamic: true, MinWidth: 2},
		{ID: "b", Width: 4, Height: 4, Dynamic: true, MinWidth: 2},
	}}
	wide := f.Layout(12, 8)
	f.Boxes[0].Hidden = true
	hidden := f.Layout(12, 8)
	if wide[0].W == 0 || wide[1].W == 0 || hidden[0] != (Rect{}) || hidden[1].W == 0 || hidden[1].X != 0 {
		t.Fatalf("visibility reflow wide=%v hidden=%v", wide, hidden)
	}
}

func TestResponsiveRenderPreservesState(t *testing.T) {
	w, _, err := BuildWidget(strings.NewReader(shellFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	f := w.(*Frame)
	f.Title = "clock snapshot 12:34:56"
	for _, width := range []int{64, 63, 64, 10, 64} {
		rows := Render(f, width, f.HeightForWidth(width))
		loadRow := 1
		if width < 64 {
			loadRow = 10
		}
		if !strings.Contains(rows[loadRow], "[l] Lo") {
			t.Fatalf("width %d missing load at row %d", width, loadRow)
		}
		if f.Title != "clock snapshot 12:34:56" {
			t.Fatal("layout mutated snapshot")
		}
		for i, row := range rows {
			if len([]rune(strings.TrimSuffix(row, "\x1b[0m"))) != width {
				t.Fatalf("width %d row %d escaped", width, i)
			}
		}
	}
	bad := strings.Replace(shellFixture(t), "breakpoint: 64", "breakpoint: -1", 1)
	if _, _, err := BuildWidget(strings.NewReader(bad)); err == nil {
		t.Fatal("negative breakpoint accepted")
	}
}

// TestDynamicFrameLayoutResizeSequenceStaysBoundedAndFilled sweeps a
// filebrowser-style dynamic frame through shrink and grow across the
// breakpoint (issue 072). Every step must be computed from the current size
// only (identical on the way down and back up), stay inside the frame, and
// fill the row width in the side-by-side layout or the full width when
// stacked, leaving no uncovered bands.
func TestDynamicFrameLayoutResizeSequenceStaysBoundedAndFilled(t *testing.T) {
	f := Frame{Gap: 1, Breakpoint: 65, Boxes: []Box{
		{ID: "files", Width: 18, Height: 18, Dynamic: true, MinWidth: 20},
		{ID: "metadata", Width: 25, Height: 18, Dynamic: true, MinWidth: 25},
	}}
	const height = 24
	seen := map[int][]Rect{}
	sweep := func(width int, path string) {
		got := f.Layout(width, height)
		if prev, ok := seen[width]; ok && !reflect.DeepEqual(prev, got) {
			t.Fatalf("%s: layout at width %d depends on history: %v then %v", path, width, prev, got)
		}
		seen[width] = got
		var placed []Rect
		for i, r := range got {
			if r.W == 0 || r.H == 0 {
				continue
			}
			if r.X < 0 || r.Y < 1 || r.X+r.W > width || r.Y+r.H > height {
				t.Fatalf("%s: width %d box %d escapes frame: %+v", path, width, i, r)
			}
			placed = append(placed, r)
		}
		if len(placed) != len(got) {
			return // a box was omitted because its minimum geometry cannot fit
		}
		if got[0].Y == got[1].Y { // side by side: boxes plus gap fill the row exactly
			if end := got[1].X + got[1].W; end != width {
				t.Fatalf("%s: width %d side-by-side ends at %d, leaving a gap: %v", path, width, end, got)
			}
			if got[0].X+got[0].W+f.Gap != got[1].X {
				t.Fatalf("%s: width %d boxes are not one gap apart: %v", path, width, got)
			}
		} else if got[0].W != width || got[1].W != width { // stacked: full width
			t.Fatalf("%s: width %d stacked boxes do not fill the width: %v", path, width, got)
		}
	}
	for w := 100; w >= 22; w-- {
		sweep(w, "shrink")
	}
	for w := 22; w <= 100; w++ {
		sweep(w, "grow")
	}
}
