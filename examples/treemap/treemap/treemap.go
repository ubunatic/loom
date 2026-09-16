// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command treemap renders the live process CPU-usage tree (via `ps`) as a
// graph.RenderTreemap box layout, sized to the terminal by default. It is a
// thin, standalone demo of codeberg.org/ubunatic/loom/graph's
// AggregateTreemap + RenderTreemap: all process-tree reading lives here in
// the example, not in the dependency-free graph package.
//
// --watch keeps redrawing in place on an interval (like `watch ps`) until
// interrupted, instead of loom's declarative Pane/View widgets, so --ansi
// coloring survives (a loom.View strips inline ANSI from its lines). It
// still quits on loom's own default keys (q, Esc, Ctrl-C, Ctrl-Q, Ctrl-D)
// via a small raw-mode /dev/tty reader -- see watchForQuitKey.
package treemap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/graph"
	"github.com/spf13/cobra"
)

// buildOptions is the copyable library setup; demoPalette supplies only
// this example's color choices.
func buildOptions(width, height, theme int, ansi, showValues bool) graph.TreemapOptions {
	opts := graph.TreemapOptions{
		Width: width, Height: height,
		Theme:      treemapTheme(theme),
		ShowValues: showValues, ValuePrecision: 0, ValueSuffix: "%",
		ANSI: ansi,
	}
	if ansi {
		opts.BackgroundANSI, opts.ForegroundANSI = demoPalette()
	}
	return opts
}

// renderOnce reads a fresh process tree and renders it. Width/height are
// re-resolved against the current terminal size on every call, so watch
// mode picks up terminal resizes between redraws.
func renderOnce(ctx context.Context, width, height, maxNodes, theme int, ansi, showValues bool, legendPosition string, legendRows, legendWidth int, legendMinValue float64, excludeSelf bool) ([]string, error) {
	root, err := readProcTree(ctx, excludeSelf)
	if err != nil {
		return nil, err
	}
	w, h := resolveDimensions(width, height)
	segments := graph.AggregateTreemap(root, maxNodes)
	opts := buildOptions(w, h, theme, ansi, showValues)
	opts.LegendRows = legendRows
	opts.LegendWidth = legendWidth
	opts.LegendMinValue = legendMinValue
	if legendPosition == "right" {
		opts.LegendPosition = graph.TreemapLegendRight
	}
	return graph.RenderTreemap(segments, opts), nil
}

// run prints the current process tree once and exits. It uses loom.WriteRows
// (plain, sequential, cursor-position-agnostic output), not loom.RawScreen:
// this is a one-shot "print and exit" command and must never touch cursor
// position or clear the screen -- that would clobber whatever the shell
// already has on screen above it. See loom.RawScreen's doc comment for why
// that distinction matters (it was a real regression here, caught by PTY
// testing).
func run(width, height, maxNodes, theme int, ansi, showValues bool, legendPosition string, legendRows, legendWidth int, legendMinValue float64, excludeSelf bool) error {
	rows, err := renderOnce(context.Background(), width, height, maxNodes, theme, ansi, showValues, legendPosition, legendRows, legendWidth, legendMinValue, excludeSelf)
	if err != nil {
		return err
	}
	return loom.WriteRows(os.Stdout, rows)
}

// runWatch redraws in place on interval until ctx is cancelled (SIGINT/
// SIGTERM), like a colored `watch ps`. It draws via loom.RawScreen rather
// than a loom.Pane/View, because loom.View strips inline ANSI from its
// lines in favor of a single uniform Style -- which would silently drop
// --ansi coloring. RawScreen is loom's "final render loop" for exactly this
// case: it clips every row to the terminal's live width and disables
// auto-wrap so an over-wide or stale-sized row can never wrap and cascade
// into a whole-screen scramble, and it positions each row absolutely so a
// bad row can only ever corrupt its own line. See rawscreen.go.
func runWatch(ctx context.Context, width, height, maxNodes, theme int, ansi, showValues bool, legendPosition string, legendRows, legendWidth int, legendMinValue float64, excludeSelf bool, interval time.Duration) error {
	if interval <= 0 {
		return fmt.Errorf("treemap: --interval must be positive")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	cleanupQuitKey := watchForQuitKey(cancel)
	defer cleanupQuitKey()

	screen := loom.OpenRawScreen(os.Stdout)
	defer screen.Close()

	draw := func() error {
		rows, err := renderOnce(ctx, width, height, maxNodes, theme, ansi, showValues, legendPosition, legendRows, legendWidth, legendMinValue, excludeSelf)
		if err != nil {
			return err
		}
		return screen.Draw(rows)
	}
	if err := draw(); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := draw(); err != nil {
				return err
			}
		}
	}
}

// Run parses args and runs the treemap example, matching the
// Run(args []string) error signature shared by the other examples for
// loom-demo/loom-bench registration.
func Run(args []string) error {
	var width, height, maxNodes, theme int
	var legendRows, legendWidth int
	var legendMinValue float64
	var legendPosition string
	var ansi, showValues, watch, excludeSelf bool
	var interval time.Duration
	cmd := &cobra.Command{
		Use:           "treemap",
		Short:         "Render the live process CPU-usage tree as a graph.RenderTreemap box layout",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := parseTheme(theme, ansi); err != nil {
				return err
			}
			if legendPosition != "bottom" && legendPosition != "right" {
				return fmt.Errorf("treemap: --legend must be bottom or right")
			}
			if legendMinValue < 0 {
				return fmt.Errorf("treemap: --legend-min-value must be non-negative")
			}
			if !watch {
				return run(width, height, maxNodes, theme, ansi, showValues, legendPosition, legendRows, legendWidth, legendMinValue, excludeSelf)
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runWatch(ctx, width, height, maxNodes, theme, ansi, showValues, legendPosition, legendRows, legendWidth, legendMinValue, excludeSelf, interval)
		},
	}
	cmd.Flags().IntVarP(&width, "width", "W", 0, "canvas width in columns (default: terminal width)")
	cmd.Flags().IntVarP(&height, "height", "H", 0, "canvas height in rows (default: terminal height - 2)")
	cmd.Flags().IntVar(&maxNodes, "max-nodes", graph.MaxTreemapNodes, "node budget passed to AggregateTreemap")
	cmd.Flags().BoolVar(&ansi, "ansi", false, "color each box with a cycling ANSI background")
	cmd.Flags().BoolVar(&showValues, "values", true, "append each segment's %CPU to its label")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "keep redrawing in place on an interval until interrupted (Ctrl-C)")
	cmd.Flags().BoolVar(&excludeSelf, "exclude-self", false, "exclude this process, its children, and a direct go run launcher")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "redraw interval in --watch mode")
	cmd.Flags().IntVar(&theme, "theme", 1, "visual style: 1 (bordered boxes) or 2 (thin edges + a corner number on every box, requires --ansi)")
	cmd.Flags().StringVar(&legendPosition, "legend", "bottom", "legend position: bottom or right")
	cmd.Flags().IntVar(&legendRows, "legend-rows", 2, "bottom legend rows (0: library default of two, negative: unlimited)")
	cmd.Flags().IntVar(&legendWidth, "legend-width", 0, "right legend width in columns (default: one third of total width)")
	cmd.Flags().Float64Var(&legendMinValue, "legend-min-value", 0, "omit legend entries below this value (percent CPU; boxes remain visible)")
	cmd.SetArgs(args)
	return cmd.Execute()
}
