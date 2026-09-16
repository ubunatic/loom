// Command splash demonstrates the Harnez startup splash screen and transition lifecycle.
package main

import (
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom/examples/splash/splash"
)

func main() {
	if err := splash.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
