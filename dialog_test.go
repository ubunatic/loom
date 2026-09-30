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
