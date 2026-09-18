// Command background demonstrates an image painted behind a Loom widget.
package main

import (
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom/examples/background/background"
)

func main() {
	if err := background.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
