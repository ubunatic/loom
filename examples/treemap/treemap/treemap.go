// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command treemap renders the live process CPU-usage tree (via `ps`) as a
// graph.RenderTreemap box layout, sized to the terminal by default. It is a
// thin, standalone demo of ubunatic.com/loom/graph's
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
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"ubunatic.com/loom"
	"ubunatic.com/loom/graph"
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
func renderOnce(ctx context.Context, opts Options, warnings io.Writer) ([]string, error) {
	w, h := resolveDimensionsWithOutput(opts.Width, opts.Height, warnings)
	return renderAtDimensions(ctx, opts, w, h)
}

// renderAtDimensions is shared by standalone and hosted paths. The hosted
// widget supplies its rect size directly and never probes terminal geometry.
func renderAtDimensions(ctx context.Context, opts Options, width, height int) ([]string, error) {
	root, err := readProcTree(ctx, opts.ExcludeSelf)
	if err != nil {
		return nil, err
	}
	renderOpts := buildOptions(width, height, opts.Theme, opts.ANSI, opts.ShowValues)
	renderOpts.LegendRows = opts.LegendRows
	renderOpts.LegendWidth = opts.LegendWidth
	renderOpts.LegendMinValue = opts.LegendMinValue
	if opts.LegendPosition == "right" {
		renderOpts.LegendPosition = graph.TreemapLegendRight
	}
	segments := graph.AggregateTreemap(root, opts.MaxNodes)
	return graph.RenderTreemap(segments, renderOpts), nil
}

// run prints the current process tree once and exits. It uses loom.WriteRows
// (plain, sequential, cursor-position-agnostic output), not loom.RawScreen:
// this is a one-shot "print and exit" command and must never touch cursor
// position or clear the screen -- that would clobber whatever the shell
// already has on screen above it. See loom.RawScreen's doc comment for why
// that distinction matters (it was a real regression here, caught by PTY
// testing).
func run(opts Options) error {
	rows, err := renderOnce(context.Background(), opts, os.Stderr)
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
func runWatch(ctx context.Context, opts Options) error {
	if opts.Interval <= 0 {
		return fmt.Errorf("treemap: --interval must be positive")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	cleanupQuitKey := watchForQuitKey(cancel)
	defer cleanupQuitKey()

	screen := loom.OpenRawScreen(os.Stdout)
	defer screen.Close()

	draw := func() error {
		rows, err := renderOnce(ctx, opts, os.Stderr)
		if err != nil {
			return err
		}
		return screen.Draw(rows)
	}
	if err := draw(); err != nil {
		return err
	}

	ticker := time.NewTicker(opts.Interval)
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
	opts := defaultOptions()
	cmd := &cobra.Command{
		Use:           "treemap",
		Short:         "Render the live process CPU-usage tree as a graph.RenderTreemap box layout",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateOptions(opts); err != nil {
				return err
			}
			if !opts.Watch {
				return run(opts)
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runWatch(ctx, opts)
		},
	}
	bindFlags(cmd.Flags(), &opts)
	cmd.SetArgs(args)
	return cmd.Execute()
}
