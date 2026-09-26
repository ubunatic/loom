// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package treemap

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/graph"
)

// Options configures standalone rendering and the hosted treemap widget.
type Options struct {
	Width          int
	Height         int
	MaxNodes       int
	Theme          int
	ANSI           bool
	ShowValues     bool
	Watch          bool
	ExcludeSelf    bool
	Interval       time.Duration
	LegendPosition string
	LegendRows     int
	LegendWidth    int
	LegendMinValue float64
}

type treemapWidget struct {
	opts   Options
	rows   *loom.StyledRows
	output func(context.Context, Options) ([]string, error)
}

// NewWidget parses command-line arguments and builds the hosted treemap.
func NewWidget(args []string) (loom.Widget, error) {
	opts, err := parseWidgetOptions(args)
	if err != nil {
		return nil, err
	}
	return newWidgetFromOptions(opts, func(ctx context.Context, opts Options) ([]string, error) {
		return renderOnce(ctx, opts, nil)
	})
}

func newWidgetFromOptions(opts Options, output func(context.Context, Options) ([]string, error)) (*treemapWidget, error) {
	if err := validateOptions(opts); err != nil {
		return nil, err
	}
	w := &treemapWidget{opts: opts, output: output}
	if err := w.refresh(context.Background()); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *treemapWidget) refresh(ctx context.Context) error {
	lines, err := w.output(ctx, w.opts)
	if err != nil {
		return err
	}
	if w.rows == nil {
		w.rows = loom.NewStyledRows(lines...)
	} else {
		w.rows.Lines = append(w.rows.Lines[:0], lines...)
	}
	return nil
}

func (w *treemapWidget) Draw(c *loom.Canvas, r loom.Rect) {
	w.rows.Draw(c, r)
}

func (w *treemapWidget) HandleKey(e loom.KeyEvent) bool {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	switch key {
	case "q", "esc", "ctrl-c", "ctrl-q":
		return true
	default:
		return false
	}
}

func (w *treemapWidget) HandleMouse(loom.MouseEvent) bool { return false }

func (w *treemapWidget) TickInterval() time.Duration {
	if !w.opts.Watch {
		return 0
	}
	return w.opts.Interval
}

func (w *treemapWidget) Tick(_ time.Time) {
	_ = w.refresh(context.Background())
}

func validateOptions(opts Options) error {
	if err := parseTheme(opts.Theme, opts.ANSI); err != nil {
		return err
	}
	if opts.LegendPosition != "bottom" && opts.LegendPosition != "right" {
		return fmt.Errorf("treemap: --legend must be bottom or right")
	}
	if opts.LegendMinValue < 0 {
		return fmt.Errorf("treemap: --legend-min-value must be non-negative")
	}
	if opts.MaxNodes < 1 {
		return fmt.Errorf("treemap: --max-nodes must be positive")
	}
	if opts.Watch && opts.Interval <= 0 {
		return fmt.Errorf("treemap: --interval must be positive")
	}
	return nil
}

func parseWidgetOptions(args []string) (Options, error) {
	opts := Options{
		MaxNodes: graph.MaxTreemapNodes, Theme: 1, ShowValues: true,
		Interval: 2 * time.Second, LegendPosition: "bottom", LegendRows: 2,
	}
	flags := flag.NewFlagSet("treemap", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.IntVar(&opts.Width, "width", 0, "canvas width in columns")
	flags.IntVar(&opts.Width, "W", 0, "canvas width in columns")
	flags.IntVar(&opts.Height, "height", 0, "canvas height in rows")
	flags.IntVar(&opts.Height, "H", 0, "canvas height in rows")
	flags.IntVar(&opts.MaxNodes, "max-nodes", graph.MaxTreemapNodes, "node budget")
	flags.IntVar(&opts.Theme, "theme", 1, "visual style")
	flags.BoolVar(&opts.ANSI, "ansi", false, "color boxes")
	flags.BoolVar(&opts.ShowValues, "values", true, "show values")
	flags.BoolVar(&opts.Watch, "watch", false, "refresh periodically")
	flags.BoolVar(&opts.Watch, "w", false, "refresh periodically")
	flags.BoolVar(&opts.ExcludeSelf, "exclude-self", false, "exclude this process")
	flags.DurationVar(&opts.Interval, "interval", 2*time.Second, "refresh interval")
	flags.StringVar(&opts.LegendPosition, "legend", "bottom", "legend position")
	flags.IntVar(&opts.LegendRows, "legend-rows", 2, "bottom legend rows")
	flags.IntVar(&opts.LegendWidth, "legend-width", 0, "right legend width")
	flags.Float64Var(&opts.LegendMinValue, "legend-min-value", 0, "minimum legend value")
	if err := flags.Parse(args); err != nil {
		return opts, err
	}
	if flags.NArg() > 0 {
		return opts, fmt.Errorf("treemap: unexpected arguments: %v", flags.Args())
	}
	return opts, validateOptions(opts)
}
