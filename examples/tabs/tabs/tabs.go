// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package tabs demonstrates the loom.Tabs widget hosting child widgets with
// declarative key bindings, dynamic tab manipulation, and interactive chrome.
package tabs

import (
	"fmt"

	"codeberg.org/ubunatic/loom"
	"github.com/spf13/cobra"
)

func viewLines() []string {
	rows := make([]string, 30)
	for i := range rows {
		rows[i] = fmt.Sprintf("log line %02d", i+1)
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

type tabsApp struct {
	tabs       *loom.Tabs
	tabCounter int
}

func newTabsApp() *tabsApp {
	cols, rows := tableData()
	root := loom.NewTabs(
		loom.Tab{Title: "Log", Widget: loom.NewView(viewLines())},
		loom.Tab{Title: "Choices", Widget: loom.NewChoice(choiceItems())},
		loom.Tab{Title: "Files", Widget: loom.NewTable(cols, rows)},
	)
	root.Keys = loom.TabsKeys{
		Previous: "left",
		Next:     "right",
		Cycle:    "ctrl-t",
		Select:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
	}

	return &tabsApp{tabs: root}
}

func (a *tabsApp) Draw(c *loom.Canvas, r loom.Rect) {
	if r.H <= 0 || r.W <= 0 {
		return
	}
	footerH := 1
	tabsH := max(0, r.H-footerH)
	if tabsH > 0 {
		a.tabs.Draw(c, loom.Rect{X: r.X, Y: r.Y, W: r.W, H: tabsH})
	}
	footerY := r.Y + tabsH
	c.PaintSurface(loom.Rect{X: r.X, Y: footerY, W: r.W, H: 1}, loom.Style{Dim: true})
	maxSelect := len(a.tabs.Tabs)
	if maxSelect > 9 {
		maxSelect = 9
	}
	legend := fmt.Sprintf(" 1..%d: select  •  ←/→/Ctrl-T: cycle  •  +/a: add  •  x/d: close  •  q: quit", maxSelect)
	c.Write(r.X, footerY, legend, loom.Style{Dim: true})
}

func (a *tabsApp) HandleKey(e loom.KeyEvent) bool {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	switch key {
	case "q", "ctrl-q", "ctrl-c", "esc":
		return true
	case "+", "a":
		a.tabCounter++
		title := fmt.Sprintf("Extra %d", a.tabCounter)
		lines := []string{
			fmt.Sprintf("Dynamic tab #%d", a.tabCounter),
			"",
			"Added with '+' or 'a'.",
			"Press 'x' or 'd' to remove this tab.",
			"Press 1..9 to jump to any tab directly.",
		}
		newTab := loom.Tab{
			Title:  title,
			Widget: loom.NewView(lines),
		}
		idx := a.tabs.Add(newTab)
		a.tabs.Select(idx)
		return false
	case "x", "d":
		if len(a.tabs.Tabs) > 1 {
			_ = a.tabs.Remove(a.tabs.Focus())
		}
		return false
	}
	return a.tabs.HandleKey(e)
}

func (a *tabsApp) HandleMouse(e loom.MouseEvent) bool {
	return a.tabs.HandleMouse(e)
}

// Run runs the tabs example with cobra command support.
func Run(args []string) error {
	cmd := &cobra.Command{
		Use:           "tabs",
		Short:         "Demonstrate Tabs widget with declarative keys and dynamic tab management",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run()
		},
	}
	cmd.SetArgs(args)
	return cmd.Execute()
}

func run() error {
	app := newTabsApp()
	pane, err := loom.New(18)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	pane.EnableMouseClicks()
	return pane.Run(app)
}
