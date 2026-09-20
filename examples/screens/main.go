// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command screens demonstrates switching a small inline TUI to full screen and
// to the alternate screen, with automatic quasi-fullscreen detection.
package main

import (
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom/examples/screens/screens"
)

func main() {
	if err := screens.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
