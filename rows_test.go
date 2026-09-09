// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
)

func TestRowsStableColumns(t *testing.T) {
	rows := &Rows{Gap: 1, Ellipsis: "...", Columns: []RowColumn{{Width: 8, Bold: true}, {Width: 4, Align: "right"}, {Width: 7}}, Values: [][]string{{"usage", "99%", "2d20h"}, {"long name", "100%", "123d20h"}}}
	for _, width := range []int{1, 2, 8, 12, 21} {
		output := Render(rows, width, 2)
		for y, line := range output {
			cells, err := oracleCells(line)
			if err != nil || len(cells) != width {
				t.Fatalf("width %d row %d: %v %v", width, y, cells, err)
			}
			if width == 21 {
				want := " 99%"
				if y == 1 {
					want = "100%"
				}
				if strings.Join(cells[9:13], "") != want {
					t.Fatalf("value shifted: %v", cells)
				}
				if strings.Join(cells[14:21], "") != rows.Values[y][2]+strings.Repeat(" ", 7-len(rows.Values[y][2])) {
					t.Fatal("duration shifted")
				}
			}
		}
	}
}

func TestDeclaredDummyRowsGeometry(t *testing.T) {
	root, _, err := BuildWidget(strings.NewReader(monitorFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	f := root.(*Frame)
	for _, width := range []int{64, 40, 8} {
		boxes := []Rect{{0, 1, 31, 8}, {33, 1, 31, 8}}
		if width < 64 {
			boxes[1] = Rect{0, 11, 31, 8}
		}
		if width == 8 {
			boxes[0].W = 8
			boxes[1].W = 8
		}
		if err := checkGeometry(Render(f, width, f.HeightForWidth(width)), width, boxes); err != nil {
			t.Fatal(err)
		}
	}
	if !strings.Contains(strings.Join(Render(f, 64, 10), "\n"), "Claude") {
		t.Fatal("dummy values missing")
	}
	for _, tc := range []struct{ old, new string }{
		{"width: 8", "width: 0"}, {"align: left", "align: center"},
		{"['Claude', '[⣿⣿  ]', '60%', '[    ]']", "['Claude']"},
	} {
		if _, _, err := BuildWidget(strings.NewReader(strings.Replace(monitorFixture(t), tc.old, tc.new, 1))); err == nil {
			t.Fatalf("accepted invalid rows %s", tc.new)
		}
	}
}

func TestRowsSetValuesDynamicBinding(t *testing.T) {
	root, _, err := BuildWidget(strings.NewReader(monitorFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	f := root.(*Frame)
	usageBox := f.Box("usage")
	if usageBox == nil || usageBox.Rows == nil {
		t.Fatal("usage box or rows missing")
	}

	// Verify GetValues returns initial values
	initial := usageBox.Rows.GetValues()
	if len(initial) != 4 || initial[0][0] != "Claude" {
		t.Fatalf("unexpected initial values: %v", initial)
	}

	// Update dynamically
	dynamicValues := [][]string{
		{"Updated", "[⣿⣿⣿⣿]", "100%", "[⣿⣿  ]"},
		{"WorkerB", "[⣀⣀⣀⣀]", "0%", "[    ]"},
	}
	f.Box("usage").SetRowsValues(dynamicValues)

	rendered := strings.Join(Render(f, 64, 10), "\n")
	if !strings.Contains(rendered, "Updated") || !strings.Contains(rendered, "100%") {
		t.Fatalf("rendered output did not reflect dynamic values:\n%s", rendered)
	}
	if strings.Contains(rendered, "Claude") {
		t.Fatalf("rendered output still contains old value:\n%s", rendered)
	}
}
