// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command usage runs Loom's compact colored usage watch example.
package main

import (
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom/examples/usage/usage"
)

func main() {
	if err := usage.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
