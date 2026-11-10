package loom_test

import (
	"fmt"
	"strings"
	"ubunatic.com/loom"
)

func ExampleNewCanvas() {
	c := loom.NewCanvas(5, 3)
	c.Fill(loom.Rect{X: 1, Y: 1, W: 3, H: 1}, loom.Cell{Text: "*"})

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
	//
	//  ***
}
