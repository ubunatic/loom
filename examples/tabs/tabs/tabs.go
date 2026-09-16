// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package tabs demonstrates the loom.Tabs widget hosting several different
// child widgets (a scrolling View, a Choice list, and a Table) under one Pane.
package tabs

import (
	"fmt"

	"codeberg.org/ubunatic/loom"
)

func viewLines() []string {
	rows := make([]string, 30)
	for i := range rows {
		rows[i] = fmt.Sprintf("  log line %02d", i+1)
	}
	return rows
}

func choiceItems() []loom.Item {
	return []loom.Item{
		{Name: "alpha", Desc: "first item"},
		{Name: "beta", Desc: "second item"},
		{Name: "gamma", Desc: "third item"},
	}
}

func tableData() ([]loom.Column, []loom.Row) {
	cols := []loom.Column{{Header: "Name"}, {Header: "Size"}}
	rows := []loom.Row{
		{Cells: []string{"main.go", "1.2K"}},
		{Cells: []string{"tabs.go", "3.4K"}},
		{Cells: []string{"README.md", "512B"}},
	}
	return cols, rows
}

// Run runs the Tabs example. It ignores args; it exists so the tabs example
// matches the Run(args []string) error signature shared by the other
// examples for loom-demo/loom-bench registration.
func Run(_ []string) error {
	cols, rows := tableData()
	root := loom.NewTabs(
		loom.Tab{Title: "Log", Widget: loom.NewView(viewLines())},
		loom.Tab{Title: "Choices", Widget: loom.NewChoice(choiceItems())},
		loom.Tab{Title: "Files", Widget: loom.NewTable(cols, rows)},
	)
	root.SwitchKey = "ctrl-t"

	pane, err := loom.New(18)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	return pane.Run(root)
}
