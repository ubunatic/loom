// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command monitor prints the embedded static shell once, without opening a TTY.
package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"strings"

	"codeberg.org/ubunatic/loom"
	"github.com/spf13/cobra"
)

//go:embed spec/*.yaml
var documents embed.FS

func run(out io.Writer) error {
	return runWidth(out, 0)
}

func runWidth(out io.Writer, width int) error {
	document, err := documents.Open("spec/monitor.yaml")
	if err != nil {
		return err
	}
	defer document.Close()
	root, cfg, err := loom.BuildWidget(document)
	if err != nil {
		return err
	}
	if width == 0 {
		width = cfg.MaxWidth()
	}
	height := cfg.Height(0)
	if responsive, ok := root.(loom.WidthHeighter); ok {
		height = responsive.HeightForWidth(width)
	}
	for _, row := range loom.Render(root, width, height) {
		// This monochrome shell has no styles; omit Render's row reset so
		// redirected output is plain terminal text, with no cursor controls.
		if _, err := fmt.Fprintln(out, strings.TrimSuffix(row, "\x1b[0m")); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := execute(context.Background(), os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func execute(ctx context.Context, args []string, out io.Writer) error {
	spec, err := loadWatch()
	if err != nil {
		return err
	}
	var watch bool
	var width int
	cmd := &cobra.Command{Use: spec.Command, Short: spec.Description, Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Flags().Changed("width") && width <= 0 {
				return fmt.Errorf("width must be positive")
			}
			if watch && cmd.Flags().Changed("width") {
				return fmt.Errorf("--width is for show-once; watch uses terminal width")
			}
			if watch {
				return runWatch(cmd.Context(), spec)
			}
			return runWidth(out, width)
		},
	}
	cmd.Flags().BoolVar(&watch, "watch", false, spec.WatchHelp)
	cmd.Flags().IntVar(&width, "width", 0, spec.WidthHelp)
	cmd.SetArgs(args)
	cmd.SetOut(out)
	cmd.SetErr(out)
	return cmd.ExecuteContext(ctx)
}
