// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command monitor prints the embedded static shell once, without opening a TTY.
package main

import (
	"embed"
	"fmt"
	"io"
	"os"
	"strings"

	"codeberg.org/ubunatic/loom"
)

//go:embed spec/monitor.yaml
var documents embed.FS

func run(out io.Writer) error {
	document, err := documents.Open("spec/monitor.yaml")
	if err != nil {
		return err
	}
	defer document.Close()
	root, cfg, err := loom.BuildWidget(document)
	if err != nil {
		return err
	}
	for _, row := range loom.Render(root, cfg.MaxWidth(), cfg.Height(0)) {
		// This monochrome shell has no styles; omit Render's row reset so
		// redirected output is plain terminal text, with no cursor controls.
		if _, err := fmt.Fprintln(out, strings.TrimSuffix(row, "\x1b[0m")); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
