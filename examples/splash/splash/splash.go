// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package splash demonstrates the Harnez startup splash screen and transition lifecycle.
package splash

import (
	"context"
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

func runShowOnce(out io.Writer, width, height int) error {
	if width <= 0 || height <= 0 {
		if cols, rows, err := loom.TerminalSize(); err == nil {
			if width <= 0 {
				width = cols
			}
			if height <= 0 {
				height = rows
			}
		}
		if width <= 0 {
			width = 80
		}
		if height <= 0 {
			height = 20
		}
	}

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

	return loom.RenderTo(out, view, width, height)
}

func runWatch(ctx context.Context, out io.Writer) error {
	cfg := loom.SplashConfig{
		Title:        "harnez usage",
		Tasks:        defaultTasks,
		TickInterval: 80 * time.Millisecond,
		HoldDuration: 600 * time.Millisecond,
	}

	sc := loom.NewSplashController(cfg)
	view := loom.NewSplashView(cfg.Title)
	view.Controller = sc
	next := loom.NewView([]string{"Startup complete — press q or Esc to exit."})

	pane, err := loom.New(10)
	if err != nil {
		return err
	}
	defer pane.Close()

	cadence := loom.Cadence{
		Collect: 50 * time.Millisecond,
		Redraw:  50 * time.Millisecond,
	}
	err = pane.RunStartup(ctx, loom.StartupConfig{
		Splash:  sc,
		View:    view,
		Next:    next,
		Cadence: cadence,
	})
	if err == context.Canceled {
		return nil
	}
	return err
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
	var watch bool
	var width int
	var height int

	cmd := &cobra.Command{
		Use:           "splash",
		Short:         "Demonstrate startup splash screen and transition lifecycle",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if watch {
				return runWatch(cmd.Context(), out)
			}
			return runShowOnce(out, width, height)
		},
	}

	cmd.Flags().BoolVar(&watch, "watch", false, "run animated interactive splash lifecycle")
	cmd.Flags().IntVarP(&width, "width", "w", 0, "explicit terminal width for show-once")
	cmd.Flags().IntVarP(&height, "height", "H", 0, "explicit terminal height for show-once")
	cmd.SetArgs(args)
	cmd.SetOut(out)
	cmd.SetErr(out)

	return cmd.ExecuteContext(ctx)
}
