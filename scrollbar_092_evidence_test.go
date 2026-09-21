package loom

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateScrollbar092Evidence(t *testing.T) {
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate ticket 092 evidence")
	}
	dir := filepath.Join("docs", "progress", "092")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name string, widget Widget) {
		t.Helper()
		var buf bytes.Buffer
		if err := RenderTo(&buf, widget, 32, 10); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	lines := make([]string, 40)
	lines[0] = "VERTICAL VIEW BEFORE"
	for i := 1; i < len(lines); i++ {
		lines[i] = "view line " + string(rune('A'+i-1))
	}
	v := NewView(lines)
	v.Draw(NewCanvas(32, 10), Rect{W: 32, H: 10})
	write("M2-vertical-before.ansi", v)
	v.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 32, Y: 1})
	v.HandleMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 32, Y: 5})
	write("M2-vertical-mid.ansi", v)
	v.HandleMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 32, Y: 9})
	v.HandleMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 32, Y: 9})
	write("M2-vertical-after.ansi", v)

	c := NewChoice(makeItems(40))
	c.Items[0].Name = "CHOICE BEFORE"
	c.Draw(NewCanvas(32, 10), Rect{W: 32, H: 10})
	write("M2-choice-before.ansi", c)
	c.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 31, Y: 1})
	c.HandleMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 31, Y: 5})
	write("M2-choice-mid.ansi", c)
	c.HandleMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 31, Y: 8})
	c.HandleMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 31, Y: 8})
	write("M2-choice-after.ansi", c)
}

func makeItems(n int) []Item {
	items := make([]Item, n)
	for i := range items {
		items[i] = Item{Name: fmt.Sprintf("choice line %02d", i)}
	}
	return items
}
