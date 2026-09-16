// Command split demonstrates independent scrolling and keyboard focus in a Frame.
package main

import (
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom/examples/split/split"
)

func main() {
	if err := split.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
