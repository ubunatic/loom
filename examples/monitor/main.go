// Command monitor prints the embedded static shell once, without opening a TTY.
package main

import (
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom/examples/monitor/monitor"
)

func main() {
	if err := monitor.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
