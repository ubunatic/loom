// Command filebrowser demonstrates a file list and live metadata in split panes.
package main

import (
	"flag"
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom"
)

func run() error {
	flag.Parse()
	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}
	app, err := newBrowser(dir)
	if err != nil {
		return err
	}
	pane, err := loom.New(20)
	if err != nil {
		return err
	}
	defer pane.Close()
	configurePane(pane)
	pane.EnableMouseClicks()
	return pane.Run(app)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func configurePane(pane *loom.Pane) {
	pane.Resizeable = true
	pane.MaxCols = 0               // Use the terminal width; Loom's default cap is 50 columns.
	pane.DisableDefaultQuit = true // q remains available as a file-list filter.
}
