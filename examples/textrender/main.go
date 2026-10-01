// Command textrender demonstrates non-ASCII text rendering in loom widgets.
package main

import (
	"fmt"
	"os"

	"ubunatic.com/loom/examples/textrender/textrender"
)

func main() {
	if err := textrender.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
