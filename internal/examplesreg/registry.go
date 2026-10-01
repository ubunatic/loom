// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package examplesreg is the static registry of examples/* programs shared
// by cmd/loom-demo (interactive launcher) and cmd/loom-bench (smoke
// harness), so both apps list, describe, and run the same set of examples
// without duplicating example logic. Adding a new example requires adding
// one entry here (see docs note in cmd/loom-demo and cmd/loom-bench).
package examplesreg

import (
	"ubunatic.com/loom/examples/ansicanvas_demo/ansicanvas_demo"
	"ubunatic.com/loom/examples/ansiedit/ansiedit"
	"ubunatic.com/loom/examples/ansiviewer/ansiviewer"
	"ubunatic.com/loom/examples/background/background"
	"ubunatic.com/loom/examples/filebrowser/filebrowser"
	"ubunatic.com/loom/examples/monitor/monitor"
	"ubunatic.com/loom/examples/screens/screens"
	"ubunatic.com/loom/examples/splash/splash"
	"ubunatic.com/loom/examples/split/split"
	"ubunatic.com/loom/examples/tabs/tabs"
	"ubunatic.com/loom/examples/textedit/textedit"
	"ubunatic.com/loom/examples/textrender/textrender"
	"ubunatic.com/loom/examples/treemap/treemap"
	"ubunatic.com/loom/examples/usage/usage"
	"ubunatic.com/loom/examples/winch/winch"
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
	// NewWidget builds the example's root widget for in-process hosting and
	// headless rendering. Nil for examples not yet converted.
	NewWidget func(args []string) (interface{}, error)
}

// Registry lists every examples/* program in a fixed, deterministic order.
var Registry = []Example{
	{
		Name:         "ansicanvas_demo",
		Description:  "Minimal demonstration of embedding loom.AnsiEditor",
		Package:      "ubunatic.com/loom/examples/ansicanvas_demo",
		Run:          ansicanvas_demo.Run,
		SupportsHelp: true,
		NewWidget:    func(args []string) (interface{}, error) { return ansicanvas_demo.NewWidget(args) },
	},
	{
		Name:         "ansiedit",
		Description:  "Full-featured ANSI art and graphic cell editor",
		Package:      "ubunatic.com/loom/examples/ansiedit",
		Run:          ansiedit.Run,
		SupportsHelp: true,
		NewWidget:    func(args []string) (interface{}, error) { return ansiedit.NewWidget(args) },
	},
	{
		Name: "ansiviewer", Description: "Browse and render text, ANSI, and file metadata",
		Package: "ubunatic.com/loom/examples/ansiviewer",
		Run:     func(args []string) error { return ansiviewer.Run(args) }, SupportsHelp: true,
	},
	{
		Name: "background", Description: "Full-screen Astra star field behind a widget",
		Package: "ubunatic.com/loom/examples/background",
		Run:     background.Run, SupportsHelp: false,
	},
	{
		Name:         "filebrowser",
		Description:  "File list and live metadata in split panes",
		Package:      "ubunatic.com/loom/examples/filebrowser",
		Run:          filebrowser.Run,
		SupportsHelp: true,
		NewWidget:    func(args []string) (interface{}, error) { return filebrowser.NewWidget(args) },
	},
	{
		Name:         "monitor",
		Description:  "Embedded static dashboard shell / live collector watch",
		Package:      "ubunatic.com/loom/examples/monitor",
		Run:          monitor.Run,
		SupportsHelp: true,
		DemoArgs:     []string{"--watch"},
		NewWidget:    func(args []string) (interface{}, error) { return monitor.NewWidget(args) },
	},
	{
		Name:         "splash",
		Description:  "Startup splash screen and transition lifecycle",
		Package:      "ubunatic.com/loom/examples/splash",
		Run:          splash.Run,
		SupportsHelp: true,
		DemoArgs:     []string{"--watch"},
		NewWidget:    func(args []string) (interface{}, error) { return splash.NewWidget(args) },
	},
	{
		Name:         "split",
		Description:  "Independent scrolling and keyboard focus in a Frame",
		Package:      "ubunatic.com/loom/examples/split",
		Run:          split.Run,
		SupportsHelp: true,
		NewWidget:    func(args []string) (interface{}, error) { return split.NewWidget(args) },
	},
	{
		Name:         "tabs",
		Description:  "Tabs widget hosting a View, a Choice, and a Table",
		Package:      "ubunatic.com/loom/examples/tabs",
		Run:          tabs.Run,
		SupportsHelp: true,
		NewWidget:    func(args []string) (interface{}, error) { return tabs.NewWidget(args) },
	},
	{
		Name:         "textedit",
		Description:  "Multi-pane text editor with keybindings, MRU, filebrowser, and embedded terminal",
		Package:      "ubunatic.com/loom/examples/textedit",
		Run:          textedit.Run,
		SupportsHelp: true,
		NewWidget:    func(args []string) (interface{}, error) { return textedit.NewWidget(args) },
	},
	{
		Name:         "textrender",
		Description:  "Non-ASCII text rendering across loom widgets",
		Package:      "ubunatic.com/loom/examples/textrender",
		Run:          textrender.Run,
		SupportsHelp: true,
		NewWidget:    func(args []string) (interface{}, error) { return textrender.NewWidget(args) },
	},
	{
		Name:         "treemap",
		Description:  "Live process CPU-usage tree as a treemap layout",
		Package:      "ubunatic.com/loom/examples/treemap",
		Run:          treemap.Run,
		SupportsHelp: true,
		DemoArgs:     []string{"--watch", "--ansi"},
		NewWidget:    func(args []string) (interface{}, error) { return treemap.NewWidget(args) },
	},
	{
		Name:         "screens",
		Description:  "Inline TUI that switches to full screen and the alternate screen, with auto full-screen detection",
		Package:      "ubunatic.com/loom/examples/screens",
		Run:          screens.Run,
		SupportsHelp: true,
	},
	{
		Name:         "usage",
		Description:  "Compact colored All Usage and local Load watch",
		Package:      "ubunatic.com/loom/examples/usage",
		Run:          usage.Run,
		SupportsHelp: true,
		DemoArgs:     []string{"--collect", "1s"},
		NewWidget:    func(args []string) (interface{}, error) { return usage.NewWidget(args) },
	},
	{
		Name:         "winch",
		Description:  "Diagnostic application exposing spec-backed resize modes",
		Package:      "ubunatic.com/loom/examples/winch",
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
