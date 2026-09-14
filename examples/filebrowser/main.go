// Command filebrowser demonstrates a file list and live metadata in split panes.
package main

import (
	"flag"
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom"
)

func main() {
	flag.Parse()
	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}
	app, err := newBrowser(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	pane, err := loom.New(20)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer pane.Close()
	pane.Resizeable = true
	pane.DisableDefaultQuit = true // q remains available as a file-list filter.
	if err := pane.Run(app); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
