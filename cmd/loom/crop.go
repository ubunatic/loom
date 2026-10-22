// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ubunatic.com/loom"
)

func cropCommand() *cobra.Command {
	var regionText, output string
	cmd := &cobra.Command{
		Use: "crop <file.ansi>", Short: "Crop a cell region from an ANSI screen",
		Args: cobra.ExactArgs(1), SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			region, err := parseDiffRegion(regionText)
			if err != nil {
				return diffExitError{2, err}
			}
			data, err := readAsset(args[0])
			if err != nil {
				return diffExitError{2, err}
			}
			buf, err := loom.ParseAnsiBuffer(string(data), 1, 1)
			if err != nil {
				return diffExitError{2, err}
			}
			if region.row+region.height > buf.Rows() || region.col+region.width > buf.Cols() {
				return diffExitError{2, fmt.Errorf("crop region exceeds screen bounds (%d rows, %d columns)", buf.Rows(), buf.Cols())}
			}
			cropped := loom.NewAnsiBuffer(region.width, region.height)
			for y := 0; y < region.height; y++ {
				for x := 0; x < region.width; x++ {
					cell := buf.Get(region.col+x, region.row+y)
					cropped.Set(x, y, cell)
				}
			}
			text := cropped.Serialize() + "\n"
			if output == "" {
				_, err = fmt.Fprint(cmd.OutOrStdout(), text)
			} else {
				err = os.WriteFile(output, []byte(text), 0o644)
			}
			if err != nil {
				return diffExitError{2, err}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&regionText, "region", "", "crop row,col,width,height (1-based terminal cells)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "write cropped ANSI to file instead of stdout")
	_ = cmd.MarkFlagRequired("region")
	return cmd
}
