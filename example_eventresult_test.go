package loom_test

import (
	"fmt"
	"strings"
	"ubunatic.com/loom"
)

func ExampleEventResult() {
	c := loom.NewCanvas(5, 1)
	res := loom.EventResult{Consumed: true}
	if res.Consumed {
		c.Write(0, 0, "OK", loom.Style{})
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
	// OK
}
