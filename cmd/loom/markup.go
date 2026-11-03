// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"ubunatic.com/loom"
	"ubunatic.com/loom/media"
)

func markupCommand() *cobra.Command {
	var output string
	var inPlace bool
	cmd := &cobra.Command{
		Use:   "markup <file>",
		Short: "Convert between .tui markup and ANSI",
		Long: "Convert a .tui markup file to ANSI, or any other file (read as ANSI) to .tui markup.\n" +
			"Output goes to stdout, e.g. loom markup art.tui > art.ansi && loom eval art.ansi.\n\n" +
			"Agent edit flow: draft .ansi screens directly; when an edit gets stuck,\n" +
			"convert to .tui (loom markup --in-place file.ansi), edit the readable file,\n" +
			"convert back (loom markup --in-place file.tui), and check with loom eval\n" +
			"and loom diff.",
		Args: cobra.ExactArgs(1), SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if output != "" && inPlace {
				return fmt.Errorf("cannot use both -o/--output and --in-place")
			}
			path := args[0]
			if inPlace {
				ext := strings.ToLower(filepath.Ext(path))
				if ext == ".ansi" {
					output = path[:len(path)-len(ext)] + ".tui"
				} else if ext == ".tui" {
					output = path[:len(path)-len(ext)] + ".ansi"
				} else {
					output = path + ".tui"
				}
			}
			if output != "" {
				f, err := os.Create(output)
				if err != nil {
					return err
				}
				defer f.Close()
				err = convertMarkup(f, path)
				if err != nil {
					return err
				}
				if inPlace && path != output {
					return os.Remove(path)
				}
				return nil
			}
			return convertMarkup(cmd.OutOrStdout(), path)
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "write result to file instead of stdout")
	cmd.Flags().BoolVar(&inPlace, "in-place", false, "write to x.tui for x.ansi (and vice versa) and delete the original")
	return cmd
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
	doc, err := loom.ParseMarkupWith(string(data), media.MarkupOptions(path))
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	_, err = io.WriteString(out, doc.ToANSI()+"\n")
	return err
}
