// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"ubunatic.com/loom"
)

func markupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "markup <file>",
		Short: "Convert between .tui markup and ANSI",
		Long: "Convert a .tui markup file to ANSI, or any other file (read as ANSI) to .tui markup.\n" +
			"Output goes to stdout, e.g. loom markup art.tui > art.ansi && loom eval art.ansi.",
		Args: cobra.ExactArgs(1), SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return convertMarkup(cmd.OutOrStdout(), args[0])
		},
	}
}

func convertMarkup(out io.Writer, path string) error {
	data, err := readAsset(path)
	if err != nil {
		return err
	}
	if !loom.IsMarkupPath(path) {
		var doc loom.RichDocument
		doc.FromANSI(string(data))
		_, err = io.WriteString(out, doc.ToMarkup())
		return err
	}
	doc, err := loom.ParseMarkup(string(data))
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	_, err = io.WriteString(out, doc.ToANSI()+"\n")
	return err
}
