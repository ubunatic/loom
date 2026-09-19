// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command winch runs the interactive resize diagnostics application.
package main

import (
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom/examples/winch/winch"
)

func main() {
	if err := winch.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
