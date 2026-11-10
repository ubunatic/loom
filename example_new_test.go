package loom_test

import (
	"fmt"
	"strings"
	"ubunatic.com/loom"
)

func ExampleNew() {
	_, err := loom.New(5)

	c := loom.NewCanvas(8, 1)
	if err != nil {
		c.Write(0, 0, "No TTY", loom.Style{})
	} else {
		c.Write(0, 0, "TTY", loom.Style{})
	}

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
	// TTY
}
