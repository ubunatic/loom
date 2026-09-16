// Command tabs demonstrates the Tabs widget hosting multiple child widgets.
package main

import (
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom/examples/tabs/tabs"
)

func main() {
	if err := tabs.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
