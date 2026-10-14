// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command suspend opens $EDITOR on Ctrl+G by suspending the pane.
package main

import (
	"fmt"
	"os"

	"ubunatic.com/loom/examples/suspend/suspend"
)

func main() {
	if err := suspend.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
