// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package splash demonstrates the Harnez startup splash screen and transition lifecycle.
package splash

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"codeberg.org/ubunatic/loom"
	"github.com/spf13/cobra"
)

var defaultTasks = []loom.ProviderTask{
	{Name: "mic", Symbol: "●", Duration: 300 * time.Millisecond},
	{Name: "claude", Symbol: "✳", Duration: 450 * time.Millisecond},
	{Name: "codex", Symbol: "֍", Duration: 350 * time.Millisecond},
	{Name: "agy", Symbol: "Λ", Duration: 400 * time.Millisecond},
}

// Options holds configuration for the splash widget and CLI execution.
type Options struct {
	Watch  bool
	Width  int
	Height int
}

type splashApp struct {
	view              *loom.SplashView
	controller        *loom.SplashController
	next              loom.Widget
	completed         bool
	completedRendered bool
	active            bool
	started           bool
	static            bool
}

func newSplashApp(opts Options) *splashApp {
	pills := []loom.ProviderPill{
		{Symbol: "●", Name: "mic", State: loom.ProviderDone},
		{Symbol: "✳", Name: "claude", State: loom.ProviderFetching},
		{Symbol: "֍", Name: "codex", State: loom.ProviderPending},
		{Symbol: "Λ", Name: "agy", State: loom.ProviderPending},
	}
	view := loom.NewSplashView("harnez usage", pills...)
	view.SpinnerFrame = 1
	view.Progress = 18.75
	view.StepText = "fetching claude..."

	if !opts.Watch {
		return &splashApp{
			view:   view,
			static: true,
		}
	}

	cfg := loom.SplashConfig{
		Title:        "harnez usage",
		Tasks:        defaultTasks,
		TickInterval: 80 * time.Millisecond,
		HoldDuration: 600 * time.Millisecond,
	}
	sc := loom.NewSplashController(cfg)
	view.Controller = sc

	return &splashApp{
		view:       view,
		controller: sc,
		next:       newInteractiveDestination(),
	}
}

func (a *splashApp) TickInterval() time.Duration {
	if a.static || a.active {
		return 0
	}
	return 80 * time.Millisecond
}

func (a *splashApp) Tick(now time.Time) {
	if a.static || a.active {
		return
	}
	if !a.started {
		if a.controller != nil {
			a.controller.Start(context.Background())
		}
		a.started = true
	}
	if a.controller != nil {
		snap := a.controller.Snapshot()
		a.view.ApplySnapshot(snap)
		if snap.Completed || snap.Dismissed {
			if !a.completed {
				a.completed = true
				return
			}
			if a.completedRendered {
				a.active = true
			}
		}
	}
}

func (a *splashApp) Draw(c *loom.Canvas, r loom.Rect) {
	if !a.active {
		a.view.Draw(c, r)
		if a.completed {
			a.completedRendered = true
		}
		return
	}
	if a.next != nil {
		a.next.Draw(c, r)
	}
}

func (a *splashApp) PaneRequest() loom.PaneRequest {
	return loom.PaneRequest{
		Resizeable: true,
		MaxCols:    0,
	}
}

func (a *splashApp) ConsumeKey(k loom.KeyEvent) (quit, consumed bool) {
	key := k.Key
	if key == "" {
		key = k.Text
	}
	if key == "ctrl-c" || key == "ctrl-q" {
		return true, true
	}
	if !a.active {
		switch key {
		case "q":
			return true, true
		case "esc", "enter":
			if a.controller != nil {
				a.controller.Dismiss()
			}
			return false, true
		}
		return false, false
	}
	if consumer, ok := a.next.(loom.KeyConsumer); ok {
		if q, c := consumer.ConsumeKey(k); c {
			return q, true
		}
	}
	if a.next != nil {
		q := a.next.HandleKey(k)
		if q {
			return true, true
		}
	}
	return false, false
}

func (a *splashApp) HandleKey(k loom.KeyEvent) bool {
	if !a.active {
		key := k.Key
		if key == "" {
			key = k.Text
		}
		switch key {
		case "q", "ctrl-c", "ctrl-q":
			return true
		case "esc", "enter":
			if a.controller != nil {
				a.controller.Dismiss()
			}
			return false
		}
		return false
	}
	if a.next != nil {
		return a.next.HandleKey(k)
	}
	return false
}

func (a *splashApp) HandleMouse(m loom.MouseEvent) bool {
	if a.active && a.next != nil {
		return a.next.HandleMouse(m)
	}
	return false
}

func (a *splashApp) Close() {
	if a.controller != nil {
		a.controller.Dismiss()
	}
}

func resolveTerminalDimensions(width, height int) (int, int) {
	if width > 0 && height > 0 {
		return width, height
	}
	cols, rows, err := loom.TerminalSize()
	if err != nil || cols <= 0 || rows <= 0 {
		cols, rows = 80, 20
	}
	if width <= 0 {
		width = cols
	}
	if height <= 0 {
		height = rows
	}
	return width, height
}

func runShowOnce(out io.Writer, opts Options) error {
	width, height := resolveTerminalDimensions(opts.Width, opts.Height)
	w, err := NewWidgetFromOptions(opts)
	if err != nil {
		return err
	}
	return loom.RenderTo(out, w, width, height)
}

func newInteractiveDestination() loom.Widget {
	items := []loom.Item{
		{Name: "mic", Desc: "Micro-agent orchestrator (300ms, active)"},
		{Name: "claude", Desc: "Claude 3.5 Sonnet provider (450ms, connected)"},
		{Name: "codex", Desc: "Codex completion engine (350ms, ready)"},
		{Name: "agy", Desc: "Antigravity runtime (400ms, initialized)"},
	}
	choice := loom.NewChoice(items)
	return &loom.Frame{
		Title:  "harnez usage",
		Status: "↑↓ select  •  Enter activate  •  q / Esc exit",
		Boxes: []loom.Box{
			{ID: "providers", Title: "Initialized Providers", Dynamic: true, FillHeight: true, Child: choice},
		},
		Actions: []loom.FrameAction{
			{ID: "quit", Action: "quit", Key: "q"},
		},
	}
}

func runWatch(ctx context.Context, opts Options) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	w, err := NewWidgetFromOptions(opts)
	if err != nil {
		return err
	}
	pane, err := loom.New(10)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	pane.MaxCols = 0

	err = pane.Run(w)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

// NewWidget builds the splash root widget from command-line arguments.
func NewWidget(args []string) (loom.Widget, error) {
	opts, err := parseOptions(args)
	if err != nil {
		return nil, err
	}
	return NewWidgetFromOptions(opts)
}

// NewWidgetFromOptions builds the splash root widget from parsed options.
func NewWidgetFromOptions(opts Options) (loom.Widget, error) {
	return newSplashApp(opts), nil
}

func parseOptions(args []string) (Options, error) {
	var opts Options
	flags := flag.NewFlagSet("splash", flag.ContinueOnError)
	flags.BoolVar(&opts.Watch, "watch", false, "run animated interactive splash lifecycle")
	flags.IntVar(&opts.Width, "width", 0, "explicit terminal width for show-once")
	flags.IntVar(&opts.Width, "w", 0, "explicit terminal width for show-once")
	flags.IntVar(&opts.Height, "height", 0, "explicit terminal height for show-once")
	flags.IntVar(&opts.Height, "H", 0, "explicit terminal height for show-once")
	if err := flags.Parse(args); err != nil {
		return opts, err
	}
	if flags.NArg() > 0 {
		return opts, fmt.Errorf("splash: unexpected arguments: %v", flags.Args())
	}
	return opts, nil
}

// Run runs the splash example with the given command-line args, writing to
// stdout, matching the Run(args []string) error signature shared by the
// other examples for loom-demo/loom-bench registration.
func Run(args []string) error {
	return Execute(context.Background(), args, os.Stdout)
}

// Execute runs the splash example's cobra command against ctx/args/out, for
// callers that need explicit context and output control (e.g. tests).
func Execute(ctx context.Context, args []string, out io.Writer) error {
	var opts Options

	cmd := &cobra.Command{
		Use:           "splash",
		Short:         "Demonstrate startup splash screen and transition lifecycle",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.Watch {
				return runWatch(cmd.Context(), opts)
			}
			return runShowOnce(out, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Watch, "watch", false, "run animated interactive splash lifecycle")
	cmd.Flags().IntVarP(&opts.Width, "width", "w", 0, "explicit terminal width for show-once")
	cmd.Flags().IntVarP(&opts.Height, "height", "H", 0, "explicit terminal height for show-once")
	cmd.SetArgs(args)
	cmd.SetOut(out)
	cmd.SetErr(out)

	return cmd.ExecuteContext(ctx)
}
