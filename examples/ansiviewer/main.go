// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"codeberg.org/ubunatic/loom/examples/ansiviewer/ansiviewer"
	"fmt"
	"os"
)

func main() {
	if err := ansiviewer.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
