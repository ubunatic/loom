package loom_test

import (
	"fmt"
	"strings"
	"ubunatic.com/loom"
)

func ExampleMouseEvent() {
	var w loom.Widget = loom.NewTabs(loom.Tab{Title: "A"}, loom.Tab{Title: "B"})
	w.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 5, Y: 0})

	c := loom.NewCanvas(8, 1)
	w.Draw(c, loom.Rect{W: 8, H: 1})

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
