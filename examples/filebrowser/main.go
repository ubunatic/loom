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
	configurePane(pane)
	if err := pane.Run(app); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func configurePane(pane *loom.Pane) {
	pane.Resizeable = true
	pane.MaxCols = 0               // Use the terminal width; Loom's default cap is 50 columns.
	pane.DisableDefaultQuit = true // q remains available as a file-list filter.
}
