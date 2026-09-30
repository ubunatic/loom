package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestDialogCentersClearsAndTruncatesTitle(t *testing.T) {
	c := loom.NewCanvas(30, 10)
	c.Fill(c.Bounds(), loom.Cell{Text: "x"})
	d := loom.NewDialog("A title longer than the box", "Message", "OK")
	d.Width, d.Height = 16, 5
	d.Draw(c, c.Bounds())
	rows := loom.Render(d, 30, 10)
	if !strings.Contains(strings.Join(rows, "\n"), "Message") {
		t.Fatal("dialog body was not rendered")
	}
	if !strings.Contains(strings.Join(rows, "\n"), "▶ OK") {
		t.Fatal("selected button was not rendered")
	}
	if got := c.Get(8, 3).Text; got != "M" {
		t.Fatalf("dialog body cell = %q, want M (background must be cleared)", got)
	}
}

func TestDialogPlacedAndSelectsButton(t *testing.T) {
	d := loom.NewDialog("Confirm", "Delete?", "Cancel", "Delete")
	d.Rect = loom.Rect{X: 3, Y: 2, W: 20, H: 5}
	d.ConsumeKey(loom.KeyEvent{Key: "right"})
	var selected string
	d.OnSelect = func(label string) { selected = label }
	c := loom.NewCanvas(30, 10)
	d.Draw(c, c.Bounds())
	if got := loom.Render(d, 30, 10)[2]; !strings.Contains(got, "┌") {
		t.Fatalf("placed dialog row = %q, missing border", got)
	}
	if quit := d.ConsumeKey(loom.KeyEvent{Key: "enter"}).Quit; quit || selected != "Delete" || d.Open {
		t.Fatalf("enter: quit=%v selected=%q open=%v", quit, selected, d.Open)
	}
}

func TestDialogEscapeDismisses(t *testing.T) {
	d := loom.NewDialog("Alert", "Done", "OK")
	d.ConsumeKey(loom.KeyEvent{Key: "esc"})
	if d.Open {
		t.Fatal("Escape left dialog open")
	}
}

func TestDialogMouseUsesDrawLocalCoordinates(t *testing.T) {
	for _, placed := range []bool{false, true} {
		t.Run(map[bool]string{false: "centered", true: "placed"}[placed], func(t *testing.T) {
			d := loom.NewDialog("Save changes", "Keep edits?", "Discard", "Save")
			d.Width, d.Height = 36, 7
			if placed {
				d.Rect = loom.Rect{X: 12, Y: 8, W: 36, H: 7}
			}
			allocation := loom.Rect{X: 5, Y: 3, W: 60, H: 20}
			c := loom.NewCanvas(80, 30)
			d.Draw(c, allocation)
			var selected string
			d.OnSelect = func(label string) { selected = label }
			for y := 0; y < 30; y++ {
				var row strings.Builder
				for x := 0; x < 80; x++ {
					row.WriteString(c.Get(x, y).Text)
				}
				// The button row also contains Discard; the title contains Save.
				if !strings.Contains(row.String(), "Discard") {
					continue
				}
				text := []rune(row.String())
				for x := 0; x+4 <= len(text); x++ {
					if string(text[x:x+4]) != "Save" {
						continue
					}
					res := d.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: x - allocation.X, Y: y - allocation.Y})
					if !res.Consumed || d.Open || selected != "Save" {
						t.Fatalf("button click: result=%+v open=%v selected=%q", res, d.Open, selected)
					}
					return
				}
			}
			t.Fatal("Save button not visible")
		})
	}
}
