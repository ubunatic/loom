#!/bin/bash
cat << 'INNER_EOF' > example_keyevent_test.go
package loom_test

import (
	"fmt"
	"strings"
	"ubunatic.com/loom"
)

func ExampleKeyEvent() {
	var w loom.Widget = loom.NewTextEntry(loom.TextEntryOptions{})
	w.ConsumeKey(loom.KeyEvent{Key: "a"})
	c := loom.NewCanvas(5, 1)
	w.Draw(c, loom.Rect{W: 5, H: 1})

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
	// a
}
INNER_EOF

cat << 'INNER_EOF' > example_style_test.go
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
INNER_EOF

cat << 'INNER_EOF' > example_rect_test.go
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
INNER_EOF

cat << 'INNER_EOF' > example_eventresult_test.go
package loom_test

import (
	"fmt"
	"strings"
	"ubunatic.com/loom"
)

func ExampleEventResult() {
	var w loom.Widget = loom.NewTextEntry(loom.TextEntryOptions{})
	res := w.ConsumeKey(loom.KeyEvent{Key: "O"})

	c := loom.NewCanvas(5, 1)
	if res.Consumed {
		c.Write(0, 0, "OK", loom.Style{})
	}
	w.Draw(c, loom.Rect{W: 5, H: 1})

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
	// O
}
INNER_EOF

cat << 'INNER_EOF' > example_mouseevent_test.go
package loom_test

import (
	"fmt"
	"strings"
	"ubunatic.com/loom"
)

func ExampleMouseEvent() {
	var w loom.Widget = loom.NewTabs(loom.Tab{Title: "A"}, loom.Tab{Title: "B"})
	w.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft}, 5, 0) // click on B

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
INNER_EOF

cat << 'INNER_EOF' > example_new_test.go
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
	// No TTY
}
INNER_EOF

cat << 'INNER_EOF' > example_panerequest_test.go
package loom_test

import (
	"fmt"
	"strings"
	"ubunatic.com/loom"
)

func ExamplePaneRequest() {
	var w loom.Widget = loom.NewTextEntry(loom.TextEntryOptions{})

	c := loom.NewCanvas(8, 1)
	if req, ok := w.(loom.PaneRequester); ok {
		c.Write(0, 0, fmt.Sprintf("%v", req.PaneRequest().Resizeable), loom.Style{})
	} else {
		c.Write(0, 0, "No Req", loom.Style{})
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
	// false
}
INNER_EOF
