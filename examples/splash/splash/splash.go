// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command splash demonstrates the Harnez startup splash screen and transition lifecycle.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"codeberg.org/ubunatic/loom"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var defaultTasks = []loom.ProviderTask{
	{Name: "mic", Symbol: "●", Duration: 300 * time.Millisecond},
	{Name: "claude", Symbol: "✳", Duration: 450 * time.Millisecond},
	{Name: "codex", Symbol: "֍", Duration: 350 * time.Millisecond},
	{Name: "agy", Symbol: "Λ", Duration: 400 * time.Millisecond},
}

func runShowOnce(out io.Writer, width, height int) error {
	if width <= 0 {
		width = terminalWidth(out)
		if width <= 0 {
			width = 80
		}
	}
	if height <= 0 {
		height = terminalHeight(out)
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

	cols := terminalWidth(out)
	for _, row := range loom.Render(view, width, height) {
		row = strings.TrimSuffix(row, "\x1b[0m")
		// Belt-and-suspenders against a stale/wrong terminal-width detection
		// (see loom.RawScreen's doc comment); plain sequential output, no
		// cursor control, since this is a one-shot "print once" path.
		if cols > 0 {
			row = loom.ClipRow(row, cols)
		}
		if _, err := fmt.Fprintln(out, row); err != nil {
			return err
		}
	}
	return nil
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

	pane, err := loom.New(10)
	if err != nil {
		return err
	}
	defer pane.Close()

	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	sc.Start(watchCtx)

	go func() {
		select {
		case <-sc.Done():
			// Brief delay to allow final frame render before clean exit/transition
			time.Sleep(100 * time.Millisecond)
			cancel()
		case <-watchCtx.Done():
		}
	}()

	cadence := loom.Cadence{
		Collect: 50 * time.Millisecond,
		Redraw:  50 * time.Millisecond,
	}

	collect := func(now time.Time) error {
		snap := sc.Snapshot()
		view.ApplySnapshot(snap)
		return nil
	}

	err = pane.RunWatch(watchCtx, view, cadence, collect)
	if err == context.Canceled {
		return nil
	}
	return err
}

func terminalWidth(out io.Writer) int {
	file, ok := out.(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		return 0
	}
	cols, _, err := term.GetSize(int(file.Fd()))
	if err != nil || cols < 1 {
		return 0
	}
	return cols
}

func terminalHeight(out io.Writer) int {
	file, ok := out.(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		return 0
	}
	_, rows, err := term.GetSize(int(file.Fd()))
	if err != nil || rows < 1 {
		return 0
	}
	return rows
}

func main() {
	if err := execute(context.Background(), os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func execute(ctx context.Context, args []string, out io.Writer) error {
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
