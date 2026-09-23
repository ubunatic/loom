// Command loomoji demonstrates an inline searchable emoji and symbol picker.
package main

import (
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom/examples/loomoji/loomoji"
)

func main() {
	if err := loomoji.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
