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
	want := []Rect{{X: 0, Y: 1, W: 4, H: 4}, {X: 0, Y: 6, W: 4, H: 3}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("stacked dynamic layout=%v, want %v", got, want)
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
