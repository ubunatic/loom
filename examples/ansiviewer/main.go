package main

import (
	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/examples/ansiviewer/ansiviewer"
	"fmt"
	"os"
)

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	b, err := ansiviewer.New(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	p, err := loom.New(1 << 16)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer p.Close()
	if err := p.Run(b); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
