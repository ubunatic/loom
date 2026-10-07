package loom_test

import (
	"strings"
	"testing"

	"ubunatic.com/loom"
)

func TestDialogConsumesOnlyActions(t *testing.T) {
	for _, tc := range []struct {
		name, key      string
		buttons        []string
		consumed, open bool
		selected       string
	}{
		{"right", "right", []string{"No", "Yes"}, true, true, "Yes"},
		{"left wraps", "left", []string{"No", "Yes"}, true, true, "Yes"},
		{"tab bubbles", "tab", []string{"No", "Yes"}, false, true, "No"},
		{"enter", "enter", []string{"OK"}, true, false, "OK"},
		{"escape", "esc", nil, true, false, ""},
		{"buttonless enter", "enter", nil, false, true, ""},
		{"single right", "right", []string{"OK"}, false, true, "OK"},
		{"single left", "left", []string{"OK"}, false, true, "OK"},
		{"single tab", "tab", []string{"OK"}, false, true, "OK"},
		{"unused", "down", []string{"OK"}, false, true, "OK"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := loom.NewDialog("Prompt", "Body", tc.buttons...)
			calls := 0
			d.OnSelect = func(string) { calls++ }
			res := d.ConsumeKey(loom.KeyEvent{Key: tc.key})
			if res.Consumed != tc.consumed || res.Quit || res.Done || d.Open != tc.open || d.SelectedButton() != tc.selected {
				t.Fatalf("result=%+v open=%v selected=%q", res, d.Open, d.SelectedButton())
			}
			wantCalls := 0
			if tc.key == "enter" && len(tc.buttons) > 0 {
				wantCalls = 1
			}
			if calls != wantCalls {
				t.Fatalf("callback calls=%d, want %d", calls, wantCalls)
			}
			if !d.Open && d.ConsumeKey(loom.KeyEvent{Key: tc.key}).Consumed {
				t.Fatal("closed dialog consumed key")
			}
		})
	}
}

func TestGridDialogActionsStayInChild(t *testing.T) {
	d := loom.NewDialog("Save?", "Keep edits?", "No", "Yes")
	// One column, so "down" has a sibling to move to.
	g := loom.NewGrid(1, d, loom.NewView(nil))
	parentCalls := 0
	g.OnSelect = func(int) { parentCalls++ }
	for _, key := range []string{"right", "left", "enter"} {
		if res := g.ConsumeKey(loom.KeyEvent{Key: key}); !res.Consumed || res.Quit || g.Focus() != 0 {
			t.Fatalf("%s: result=%+v focus=%d", key, res, g.Focus())
		}
	}
	if d.Open || parentCalls != 0 {
		t.Fatalf("open=%v parent callbacks=%d", d.Open, parentCalls)
	}
	d.Activate()
	if res := g.ConsumeKey(loom.KeyEvent{Key: "esc"}); !res.Consumed || d.Open {
		t.Fatalf("escape: %+v open=%v", res, d.Open)
	}
	d.Activate()
	g.ConsumeKey(loom.KeyEvent{Key: "down"})
	if g.Focus() != 1 {
		t.Fatal("unused arrow did not bubble")
	}
}

func TestGridDialogButtonClickConsumesAndCloses(t *testing.T) {
	d := loom.NewDialog("Save?", "Keep edits?", "No", "Yes")
	d.Width, d.Height = 24, 5
	g := loom.NewGrid(2, loom.NewView(nil), d)
	c := loom.NewCanvas(80, 12)
	allocation := loom.Rect{X: 4, Y: 2, W: 70, H: 8}
	g.Draw(c, allocation)
	selected := ""
	d.OnSelect = func(s string) { selected = s }
	for y := 0; y < 12; y++ {
		for x := 0; x < 78; x++ {
			if c.Get(x, y).Text == "Y" && c.Get(x+1, y).Text == "e" && c.Get(x+2, y).Text == "s" {
				res := g.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: x - allocation.X, Y: y - allocation.Y})
				// Modal capture precedes hit-testing and preserves the background
				// container's focus even when its modal lives in another cell.
				if !res.Consumed || res.Quit || d.Open || selected != "Yes" || g.Focus() != 0 {
					t.Fatalf("click: %+v open=%v selected=%q focus=%d", res, d.Open, selected, g.Focus())
				}
				return
			}
		}
	}
	t.Fatal("Yes button not drawn")
}

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

func TestDialogMeasureShowsBodyAndButtons(t *testing.T) {
	d := loom.NewDialog("Save?", "Keep edits?", "No", "Yes")
	size := d.Measure(80)
	if size.Height != 4 {
		t.Fatalf("height = %d, want 4 (border, body, buttons)", size.Height)
	}
	if out := strings.Join(loom.Render(d, size.Width, size.Height), "\n"); !strings.Contains(out, "Keep edits?") || !strings.Contains(out, "Yes") {
		t.Fatalf("dialog at measured size lost content:\n%s", out)
	}
}
