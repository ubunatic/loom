// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package examplesreg is the static registry of examples/* programs shared
// by cmd/loom-demo (interactive launcher) and cmd/loom-bench (smoke
// harness), so both apps list, describe, and run the same set of examples
// without duplicating example logic. Adding a new example requires adding
// one entry here (see docs note in cmd/loom-demo and cmd/loom-bench).
package examplesreg

import (
	"codeberg.org/ubunatic/loom/examples/ansiviewer/ansiviewer"
	"codeberg.org/ubunatic/loom/examples/background/background"
	"codeberg.org/ubunatic/loom/examples/filebrowser/filebrowser"
	"codeberg.org/ubunatic/loom/examples/monitor/monitor"
	"codeberg.org/ubunatic/loom/examples/screens/screens"
	"codeberg.org/ubunatic/loom/examples/splash/splash"
	"codeberg.org/ubunatic/loom/examples/split/split"
	"codeberg.org/ubunatic/loom/examples/tabs/tabs"
	"codeberg.org/ubunatic/loom/examples/treemap/treemap"
	"codeberg.org/ubunatic/loom/examples/winch/winch"
)

// Example describes one examples/* program.
type Example struct {
	// Name matches the examples/<Name> directory and the loom-demo/loom-bench
	// selector argument.
	Name string
	// Description is a one-line summary shown in loom-demo's menu and
	// loom-demo --list / loom-bench output.
	Description string
	// Package is the example's importable library package path, for
	// loom-bench's build step.
	Package string
	// Run invokes the example exactly as its own main.go wrapper does.
	Run func(args []string) error
	// SupportsHelp reports whether Run(["--help"]) (or Run(["-h"])) is a fast,
	// non-blocking, non-interactive exit. Examples with a flag/cobra command
	// support this; examples without flag parsing never return from Run
	// without a live TTY, and loom-bench must not invoke them.
	SupportsHelp bool
	// DemoArgs are the arguments loom-demo passes to Run, e.g. "--watch" for
	// examples that have a live mode.
	DemoArgs []string
}

// Registry lists every examples/* program in a fixed, deterministic order.
var Registry = []Example{
	{
		Name: "ansiviewer", Description: "Browse and render text, ANSI, and file metadata",
		Package: "codeberg.org/ubunatic/loom/examples/ansiviewer",
		Run:     func(args []string) error { return ansiviewer.Run(args) }, SupportsHelp: true,
	},
	{
		Name: "background", Description: "Full-screen Astra star field behind a widget",
		Package: "codeberg.org/ubunatic/loom/examples/background",
		Run:     background.Run, SupportsHelp: false,
	},
	{
		Name:         "filebrowser",
		Description:  "File list and live metadata in split panes",
		Package:      "codeberg.org/ubunatic/loom/examples/filebrowser",
		Run:          filebrowser.Run,
		SupportsHelp: true,
	},
	{
		Name:         "monitor",
		Description:  "Embedded static dashboard shell / live collector watch",
		Package:      "codeberg.org/ubunatic/loom/examples/monitor",
		Run:          monitor.Run,
		SupportsHelp: true,
		DemoArgs:     []string{"--watch"},
	},
	{
		Name:         "splash",
		Description:  "Startup splash screen and transition lifecycle",
		Package:      "codeberg.org/ubunatic/loom/examples/splash",
		Run:          splash.Run,
		SupportsHelp: true,
		DemoArgs:     []string{"--watch"},
	},
	{
		Name:         "split",
		Description:  "Independent scrolling and keyboard focus in a Frame",
		Package:      "codeberg.org/ubunatic/loom/examples/split",
		Run:          split.Run,
		SupportsHelp: true,
	},
	{
		Name:         "tabs",
		Description:  "Tabs widget hosting a View, a Choice, and a Table",
		Package:      "codeberg.org/ubunatic/loom/examples/tabs",
		Run:          tabs.Run,
		SupportsHelp: true,
	},
	{
		Name:         "treemap",
		Description:  "Live process CPU-usage tree as a treemap layout",
		Package:      "codeberg.org/ubunatic/loom/examples/treemap",
		Run:          treemap.Run,
		SupportsHelp: true,
		DemoArgs:     []string{"--watch", "--ansi"},
	},
	{
		Name:         "screens",
		Description:  "Inline TUI that switches to full screen and the alternate screen, with auto full-screen detection",
		Package:      "codeberg.org/ubunatic/loom/examples/screens",
		Run:          screens.Run,
		SupportsHelp: true,
	},
	{
		Name:         "winch",
		Description:  "Diagnostic application exposing spec-backed resize modes",
		Package:      "codeberg.org/ubunatic/loom/examples/winch",
		Run:          winch.Run,
		SupportsHelp: true,
	},
}

// Find returns the example with the given name, or false if none matches.
func Find(name string) (Example, bool) {
	for _, e := range Registry {
		if e.Name == name {
			return e, true
		}
	}
	return Example{}, false
}
