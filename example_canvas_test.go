package loom_test

import (
	"fmt"
	"strings"
	"ubunatic.com/loom"
)

func ExampleCanvas() {
	c := loom.NewCanvas(10, 2)
	c.Write(2, 0, "Canvas", loom.Style{})

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
	//   Canvas
}
