package loom_test

import (
	"fmt"
	"strings"
	"ubunatic.com/loom"
)

func ExampleRect() {
	c := loom.NewCanvas(5, 2)
	c.Fill(loom.Rect{X: 1, Y: 0, W: 3, H: 2}, loom.Cell{Text: "X"})

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
	//  XXX
	//  XXX
}
