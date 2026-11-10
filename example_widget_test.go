package loom_test

import (
	"fmt"
	"strings"
	"ubunatic.com/loom"
)

func ExampleWidget() {
	var w loom.Widget = loom.NewTabs(loom.Tab{Title: "A"}, loom.Tab{Title: "B"})

	// Send event (e.g. right key) to switch tabs
	w.ConsumeKey(loom.KeyEvent{Key: "right"})

	c := loom.NewCanvas(10, 1)
	w.Draw(c, loom.Rect{W: 10, H: 1})

	for y := 0; y < c.Rows(); y++ {
		var row strings.Builder
		for x := 0; x < c.Cols(); x++ {
			cell := c.Get(x, y)
			if cell.Text == "" {
				row.WriteString(" ")
			} else {
				row.WriteString(cell.Text)
			}
		}
		fmt.Println(strings.TrimRight(row.String(), " "))
	}

	// Output:
	// A  B
}
