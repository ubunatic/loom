// Command filebrowser demonstrates a file list and live metadata in split panes.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom/examples/filebrowser/filebrowser"
)

func main() {
	if err := filebrowser.Run(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
