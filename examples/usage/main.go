// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command usage runs Loom's compact colored usage watch example.
package main

import (
	"fmt"
	"os"

	"ubunatic.com/loom/examples/usage/usage"
)

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println("usage: loom-usage [--collect duration] [--redraw duration] [--view loom|plain]")
		return
	}
	if err := usage.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
