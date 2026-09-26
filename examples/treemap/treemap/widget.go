// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package treemap

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/graph"
	"github.com/spf13/pflag"
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

func defaultOptions() Options {
	return Options{
		MaxNodes: graph.MaxTreemapNodes, Theme: 1, ShowValues: true,
		Interval: 2 * time.Second, LegendPosition: "bottom", LegendRows: 2,
	}
}

// bindFlags defines the treemap flags for both Cobra and the hosted factory.
func bindFlags(flags *pflag.FlagSet, opts *Options) {
	flags.IntVarP(&opts.Width, "width", "W", 0, "canvas width in columns (default: terminal width)")
	flags.IntVarP(&opts.Height, "height", "H", 0, "canvas height in rows (default: terminal height - 2)")
	flags.IntVar(&opts.MaxNodes, "max-nodes", graph.MaxTreemapNodes, "node budget passed to AggregateTreemap")
	flags.BoolVar(&opts.ANSI, "ansi", false, "color each box with a cycling ANSI background")
	flags.BoolVar(&opts.ShowValues, "values", true, "append each segment's %CPU to its label")
	flags.BoolVarP(&opts.Watch, "watch", "w", false, "keep redrawing in place on an interval until interrupted (Ctrl-C)")
	flags.BoolVar(&opts.ExcludeSelf, "exclude-self", false, "exclude this process, its children, and a direct go run launcher")
	flags.DurationVar(&opts.Interval, "interval", 2*time.Second, "redraw interval in --watch mode")
	flags.IntVar(&opts.Theme, "theme", 1, "visual style: 1 (bordered boxes) or 2 (thin edges + a corner number on every box, requires --ansi)")
	flags.StringVar(&opts.LegendPosition, "legend", "bottom", "legend position: bottom or right")
	flags.IntVar(&opts.LegendRows, "legend-rows", 2, "bottom legend rows (0: library default of two, negative: unlimited)")
	flags.IntVar(&opts.LegendWidth, "legend-width", 0, "right legend width in columns (default: one third of total width)")
	flags.Float64Var(&opts.LegendMinValue, "legend-min-value", 0, "omit legend entries below this value (percent CPU; boxes remain visible)")
}

type treemapWidget struct {
	mu sync.Mutex

	opts   Options
	rows   *loom.StyledRows
	err    error
	output func(context.Context, Options, int, int) ([]string, error)

	width, height                 int
	renderedWidth, renderedHeight int
	collecting, closed            bool
	ctx                           context.Context
	cancel                        context.CancelFunc
	done                          sync.WaitGroup
}

// NewWidget parses command-line arguments and builds the hosted treemap.
func NewWidget(args []string) (loom.Widget, error) {
	opts, err := parseWidgetOptions(args)
	if err != nil {
		return nil, err
	}
	return newWidgetFromOptions(opts, renderAtDimensions)
}

func newWidgetFromOptions(opts Options, output func(context.Context, Options, int, int) ([]string, error)) (*treemapWidget, error) {
	if err := validateOptions(opts); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &treemapWidget{
		opts: opts, rows: loom.NewStyledRows("Collecting process tree…"), output: output,
		ctx: ctx, cancel: cancel,
	}, nil
}

// Draw requests data at the allocated size. Collection runs off the pane loop,
// and rows from previous dimensions remain visible until the new frame arrives.
func (w *treemapWidget) Draw(c *loom.Canvas, r loom.Rect) {
	if c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	w.mu.Lock()
	if w.width != r.W || w.height != r.H {
		w.width, w.height = r.W, r.H
	}
	if !w.collecting && (w.renderedWidth != w.width || w.renderedHeight != w.height) {
		w.startCollectionLocked()
	}
	var lines []string
	if w.err != nil {
		lines = []string{"\x1b[31m" + w.err.Error() + "\x1b[0m"}
	} else {
		lines = append(lines, w.rows.Lines...)
	}
	w.mu.Unlock()
	loom.NewStyledRows(lines...).Draw(c, r)
}

func (w *treemapWidget) startCollectionLocked() {
	if w.closed || w.collecting || w.width <= 0 || w.height <= 0 {
		return
	}
	w.collecting = true
	width, height := w.width, w.height
	opts := w.opts
	opts.Width, opts.Height = width, height
	w.done.Add(1)
	go func() {
		defer w.done.Done()
		lines, err := w.output(w.ctx, opts, width, height)
		w.mu.Lock()
		defer w.mu.Unlock()
		w.collecting = false
		if w.closed {
			return
		}
		w.renderedWidth, w.renderedHeight = width, height
		w.err = err
		if err == nil {
			w.rows.Lines = append(w.rows.Lines[:0], lines...)
		}
		if w.width != width || w.height != height {
			w.startCollectionLocked()
		}
	}()
}

// HandleKey's bool follows Widget semantics: true asks the host to quit this
// child. A tabs host may consume that child quit through its OnChildQuit hook.
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

// Tick starts collection without waiting for ps or graph rendering.
func (w *treemapWidget) Tick(_ time.Time) {
	w.mu.Lock()
	w.startCollectionLocked()
	w.mu.Unlock()
}

// Close cancels and joins an in-flight process collection.
func (w *treemapWidget) Close() {
	w.mu.Lock()
	if !w.closed {
		w.closed = true
		w.cancel()
	}
	w.mu.Unlock()
	w.done.Wait()
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
	opts := defaultOptions()
	flags := pflag.NewFlagSet("treemap", pflag.ContinueOnError)
	flags.SetOutput(io.Discard)
	bindFlags(flags, &opts)
	if err := flags.Parse(args); err != nil {
		return opts, err
	}
	if flags.NArg() > 0 {
		return opts, fmt.Errorf("treemap: unexpected arguments: %v", flags.Args())
	}
	return opts, validateOptions(opts)
}
