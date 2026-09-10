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
	"codeberg.org/ubunatic/loom/graph"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
		width = outputWidth(out, cfg.MaxWidth())
	}
	applySnapshot(root, staticSnapshot)
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

func outputWidth(out io.Writer, fallback int) int {
	file, ok := out.(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		return fallback
	}
	cols, _, _ := term.GetSize(int(file.Fd()))
	if cols < 1 {
		return fallback
	}
	return min(cols, fallback)
}

// monitorSnapshot is application data: the declaration owns the row layout,
// while the example owns the numerical values that populate graph columns.
type monitorSnapshot struct {
	usage []float64
	load  map[string][]float64
	vram  []float64
	gtt   []float64
}

var staticSnapshot = monitorSnapshot{
	usage: []float64{60, 95, 99, 41},
	load: map[string][]float64{
		"cpu (16c)": {1, 4, 8, 12, 9, 14, 11, 7, 5, 3, 1, 2, 4, 6, 8, 10, 12, 10, 8, 6},
		"ram (45G)": {32, 34, 35, 36, 36, 37, 36, 36, 35, 36, 36, 37, 36, 36, 36, 36, 36, 36, 36, 36},
		"gpu (Phx)": {0, 3, 7, 4, 2, 0, 1, 5, 8, 4, 2, 0, 1, 3, 5, 2, 0, 1, 4, 2},
	},
	vram: []float64{4, 5, 6, 7, 6, 6, 7, 8},
	gtt:  []float64{1, 1, 2, 2, 2, 3, 2, 2},
}

func applySnapshot(root loom.Widget, snapshot monitorSnapshot) {
	frame, ok := root.(*loom.Frame)
	if !ok {
		return
	}
	for i := range frame.Boxes {
		box := &frame.Boxes[i]
		switch box.ID {
		case "usage":
			values := box.Rows.GetValues()
			for row := range values {
				if row < len(snapshot.usage) && len(values[row]) > 1 {
					values[row][1] = graph.RenderBar(snapshot.usage[row], graph.BarOptions{Width: 4, SubChar: true})
				}
			}
			box.SetRowsValues(values)
		case "load":
			values := box.Rows.GetValues()
			for row := range values {
				if len(values[row]) < 2 {
					continue
				}
				name := values[row][0]
				history, ok := snapshot.load[name]
				if !ok {
					if name == "vram/gtt" {
						values[row][1] = splitTimeline(snapshot.vram, snapshot.gtt)
					}
					continue
				}
				values[row][1] = timeline(history, 10)
			}
			box.SetRowsValues(values)
		}
	}
}

func timeline(values []float64, width int) string {
	return "[" + graph.RenderSparkline(values, graph.SparklineOptions{
		Width: width, FixedRange: true, Min: 0, Max: 100,
	}) + "]"
}

func splitTimeline(left, right []float64) string {
	return timeline(left, 4) + timeline(right, 4)
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
	cmd.Flags().IntVarP(&width, "width", "w", 0, spec.WidthHelp)
	cmd.SetArgs(args)
	cmd.SetOut(out)
	cmd.SetErr(out)
	return cmd.ExecuteContext(ctx)
}
