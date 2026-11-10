package loom_test

import (
	"fmt"
	"strings"
	"ubunatic.com/loom"
)

func ExamplePaneRequest() {
	req := loom.PaneRequest{MaxCols: 80}

	c := loom.NewCanvas(8, 1)
	c.Write(0, 0, fmt.Sprintf("%d", req.MaxCols), loom.Style{})

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
	// 80
}
