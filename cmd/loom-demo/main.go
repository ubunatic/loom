// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command loom-demo is an interactive launcher for loom's examples/*
// programs: pick one from a menu (loom.Choice, via loom.RunPane) and run it,
// or list/run by name non-interactively. It imports each example's library
// package directly (see internal/examplesreg) rather than duplicating any
// example logic or execing subprocesses.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/internal/examplesreg"
	"github.com/spf13/cobra"
)

func main() {
	if err := execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func execute(args []string) error {
	var list, hosted bool
	cmd := &cobra.Command{
		Use:           "loom-demo [example]",
		Short:         "Interactive launcher for loom's examples/* programs",
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if list {
				printList(cmd.OutOrStdout())
				return nil
			}
			if hosted {
				return runHosted()
			}
			if len(args) >= 1 {
				return runByName(args[0], args[1:]...)
			}
			return runInteractive()
		},
	}
	cmd.Flags().SetInterspersed(false)
	cmd.Flags().BoolVar(&list, "list", false, "print the example registry and exit")
	cmd.Flags().BoolVar(&hosted, "hosted", false, "run examples in hosted Tabs mode (experimental)")
	cmd.SetArgs(args)
	return cmd.Execute()
}

func printList(out io.Writer) {
	for _, e := range examplesreg.Registry {
		fmt.Fprintf(out, "%-12s %s\n", e.Name, e.Description)
	}
}

func normalizeName(name string) string {
	clean := strings.ToLower(strings.TrimSpace(name))
	clean = strings.TrimPrefix(clean, "examples/")
	clean = strings.TrimSuffix(clean, "/")
	return clean
}

func runByName(name string, extraArgs ...string) error {
	clean := normalizeName(name)
	e, ok := examplesreg.Find(clean)
	if !ok {
		e, ok = examplesreg.Find(name)
	}
	if !ok {
		return fmt.Errorf("loom-demo: unknown example %q (see loom-demo --list)", name)
	}
	runArgs := e.DemoArgs
	if len(extraArgs) > 0 {
		runArgs = extraArgs
	}
	return e.Run(runArgs)
}

// runInteractive shows a loom.Choice menu of every registered example and
// runs the one the user picks, then shows the menu again until the user aborts
// it. It uses loom.RunPane, the library's own open/run/close-pane helper,
// rather than duplicating pane lifecycle or key handling here. A failing
// example is reported and does not end the session.
func runInteractive() error {
	items := make([]loom.Item, 0, len(examplesreg.Registry))
	for _, e := range examplesreg.Registry {
		items = append(items, loom.Item{Name: e.Name, Desc: e.Description})
	}
	for {
		menu := loom.NewChoice(items)
		menu.Prompt = "> "
		item, ok, _, err := loom.RunPane(menu)
		if err != nil {
			return err
		}
		if !ok {
			return nil // user aborted the menu; nothing more to run
		}
		if err := runByName(item.Name); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

// newHostedTabs builds the hosted Tabs widget with ArrowSwitch disabled.
// The returned widget is ready to pass to a Pane.
func newHostedTabs() (*loom.Tabs, error) {
	// Collect converted examples (those with NewWidget)
	tabs := make([]loom.Tab, 0)
	for _, e := range examplesreg.Registry {
		if e.NewWidget == nil {
			continue
		}
		widget, err := e.NewWidget([]string{})
		if err != nil {
			return nil, fmt.Errorf("failed to create widget for %q: %v", e.Name, err)
		}
		w, ok := widget.(loom.Widget)
		if !ok {
			return nil, fmt.Errorf("%q NewWidget did not return a loom.Widget", e.Name)
		}
		tabs = append(tabs, loom.Tab{Title: e.Name, Widget: w})
	}

	if len(tabs) == 0 {
		return nil, fmt.Errorf("no examples with NewWidget to host")
	}

	// Create the host tabs widget with ArrowSwitch disabled
	// so left/right arrows reach the hosted app instead of switching tabs.
	host := loom.NewTabs(tabs...)
	host.ArrowSwitch = false
	host.OnChildQuit = func(i int) bool {
		// Return false to stay in hosted mode when a child quits
		return false
	}

	return host, nil
}

// runHosted runs converted examples in a Tabs host with ArrowSwitch disabled
// so that left/right arrow keys reach the hosted applications instead of
// switching tabs. This demonstrates the hosting contract from 057 and 058.
func runHosted() error {
	host, err := newHostedTabs()
	if err != nil {
		return fmt.Errorf("loom-demo: %v", err)
	}

	pane, err := loom.New(1 << 16)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	pane.MaxCols = 0
	return pane.Run(host)
}
