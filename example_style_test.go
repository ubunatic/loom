package loom_test

import (
	"fmt"
	"strings"
	"ubunatic.com/loom"
)

func ExampleStyle() {
	c := loom.NewCanvas(5, 1)
	c.Write(0, 0, "Red", loom.Style{FG: loom.ColorIndex(1)})

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
	// Red
}
