// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"os"
	"strings"
	"testing"
)

func emptyShellFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/fixtures/empty-shell.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func monitorFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("examples/monitor/spec/monitor.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func shellFixture(t *testing.T) string {
	return emptyShellFixture(t)
}

func TestStaticShellGolden(t *testing.T) {
	w, cfg, err := BuildWidget(strings.NewReader(emptyShellFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	rows := Render(w, cfg.MaxWidth(), cfg.Height(0))
	// Explicit expected columns, independent of production width/layout helpers.
	want := []string{
		"Loom monitor                                                    ",
		"┌ [u] All Usage ──────────────┐  ┌ [l] Load ───────────────────┐",
		"│                             │  │                             │",
		"│                             │  │                             │",
		"│                             │  │                             │",
		"│                             │  │                             │",
		"│                             │  │                             │",
		"└─────────────────────────────┘  └─────────────────────────────┘",
		"Show once  [u]usage:on  [l]load:on  [q]quit                     ",
	}
	if len(rows) != len(want) {
		t.Fatalf("rows=%d want=%d", len(rows), len(want))
	}
	for y, row := range rows {
		got := strings.TrimSuffix(row, "\x1b[0m")
		if got != want[y] {
			t.Errorf("row %d\n got %q\nwant %q", y, got, want[y])
		}
	}
}

func TestMonitorExampleGolden(t *testing.T) {
	w, cfg, err := BuildWidget(strings.NewReader(monitorFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	rows := Render(w, cfg.MaxWidth(), cfg.Height(0))
	if len(rows) != cfg.Height(0) {
		t.Fatalf("rows=%d want=%d", len(rows), cfg.Height(0))
	}
	rendered := strings.Join(rows, "\n")
	if !strings.Contains(rendered, "Claude Code") || !strings.Contains(rendered, "cpu (16c)") {
		t.Fatalf("monitor fixture missing expected row values:\n%s", rendered)
	}
}

type overflowingChild struct{}

func (overflowingChild) Draw(c *Canvas, _ Rect) {
	c.Fill(Rect{X: -5, Y: -5, W: 30, H: 30}, Cell{Text: "X"})
}
func (overflowingChild) HandleKey(KeyEvent) bool     { return false }
func (overflowingChild) HandleMouse(MouseEvent) bool { return false }

func TestBoxPaddingAndIsolation(t *testing.T) {
	w, _, err := BuildWidget(strings.NewReader(shellFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	b := w.(*Frame).Boxes[0]
	b.Title = ""
	b.Child = overflowingChild{}
	c := NewCanvas(12, 9)
	c.Fill(c.Bounds(), Cell{Text: "."})
	b.Draw(c, Rect{X: 2, Y: 1, W: 8, H: 7})
	want := []string{
		"............",
		"..┌──────┐..",
		"..│      │..",
		"..│ XXXX │..",
		"..│ XXXX │..",
		"..│ XXXX │..",
		"..│      │..",
		"..└──────┘..",
		"............",
	}
	for y, expected := range want {
		if got := strings.TrimSuffix(c.Row(y), "\x1b[0m"); got != expected {
			t.Errorf("row %d: got %q want %q", y, got, expected)
		}
	}
}

func TestFrameTinyBounds(t *testing.T) {
	w, _, err := BuildWidget(strings.NewReader(shellFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	for _, size := range []Rect{{}, {W: 1, H: 1}, {W: 2, H: 2}, {W: 3, H: 3}, {W: 4, H: 4}, {W: 5, H: 5}} {
		c := NewCanvas(10, 10)
		c.Fill(c.Bounds(), Cell{Text: "."})
		r := Rect{X: 2, Y: 2, W: size.W, H: size.H}
		w.Draw(c, r)
		for y := 0; y < 10; y++ {
			for x := 0; x < 10; x++ {
				if !r.Contains(x, y) && c.Get(x, y).Text != "." {
					t.Fatalf("size %+v escaped at %d,%d", size, x, y)
				}
			}
		}
		if size.H >= 2 && c.Get(2, 2+size.H-1).Text != "S" {
			t.Errorf("size %+v lost status", size)
		}
	}
}

func TestShellDeclarationFidelity(t *testing.T) {
	source := shellFixture(t)
	source = strings.Replace(source, "title: Loom monitor", "title: Changed in YAML", 1)
	source = strings.Replace(source, "id: usage", "id: first", 1)
	source = strings.Replace(source, "target: usage", "target: first", 1)
	source = strings.Replace(source, "title: All Usage", "title: First", 1)
	w, _, err := BuildWidget(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	frame := w.(*Frame)
	if frame.Title != "Changed in YAML" || frame.Boxes[0].ID != "first" || frame.Boxes[0].Padding != 1 || frame.Gap != 2 {
		t.Fatalf("declaration not consumed: %+v", frame)
	}
	header, boxes, _ := strings.Cut(source, "    boxes:\n")
	parts := strings.Split(boxes, "      - id:")
	reordered := header + "    boxes:\n" + parts[0] + "      - id:" + parts[2] + "      - id:" + parts[1]
	w, _, err = BuildWidget(strings.NewReader(reordered))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(Render(w, 64, 9)[1], "┌ [l] Load ") {
		t.Fatal("box sequence ignored")
	}
}

func TestShellValidation(t *testing.T) {
	source := shellFixture(t)
	for _, tc := range []struct{ name, old, new, message string }{
		{"unknown field", "gap: 2", "gpa: 2", "line"},
		{"invalid size", "width: 31", "width: -1", "boxes[0]"},
		{"duplicate ID", "id: load", "id: usage", "duplicate"},
		{"missing ID", "id: usage", "id: ''", ".id"},
		{"padding", "padding: 1", "padding: 99", "padding"},
		{"unknown root", "height: 9", "root: absent\n  height: 9", "app.root"},
		{"missing dimensions", "height: 9", "height: 0", "positive height"},
		{"control text", "title: Loom monitor", "title: \"bad\\nline\"", "ASCII"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := strings.Replace(source, tc.old, tc.new, 1)
			_, _, err := BuildWidget(strings.NewReader(input))
			if err == nil || !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("got %v, want %q", err, tc.message)
			}
			if err := ValidateYAML(strings.NewReader(input)); err == nil {
				t.Fatal("validation/build disagree")
			}
		})
	}
}
