// Command treemap renders the live process CPU-usage tree (via `ps`) as a
// graph.RenderTreemap box layout, sized to the terminal by default.
package main

import (
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom/examples/treemap/treemap"
)

func main() {
	if err := treemap.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
